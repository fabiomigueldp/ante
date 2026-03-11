package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	envelopeMsg      session.Envelope
	sessionDoneMsg   struct{}
	tickAnimationMsg struct{}
)

var playGameSound = audio.Play

var listSaves = storage.ListSavesResult
var saveSessionToSlot = func(sess *session.Session, slot int) error { return sess.SaveToSlot(slot) }

type GameModel struct {
	width  int
	height int

	sess *session.Session
	vm   session.GameVM

	betInput string
	betMode  bool

	showPotOdds      bool
	paused           bool
	localMessage     string
	localMessageKind session.MessageKind

	handSoundState handSoundState
}

type handSoundState struct {
	holeCuePlayed bool
	potCuePlayed  bool
	bustCuePlayed bool
}

func NewGameModel(sess *session.Session, showPotOdds bool) GameModel {
	return GameModel{
		sess:        sess,
		vm:          session.BootstrapGameVM(sess.SessionID, sess.TableState()),
		showPotOdds: showPotOdds,
	}
}

func (m GameModel) Init() tea.Cmd {
	go func(sess *session.Session) {
		defer func() {
			if r := recover(); r != nil {
				sess.Stop()
			}
		}()
		sess.Run()
	}(m.sess)
	return m.waitForEnvelope()
}

func (m GameModel) waitForEnvelope() tea.Cmd {
	return func() tea.Msg {
		env, ok := <-m.sess.Updates
		if !ok {
			return sessionDoneMsg{}
		}
		return envelopeMsg(env)
	}
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case envelopeMsg:
		return m.handleEnvelope(session.Envelope(msg))
	case sessionDoneMsg:
		if !m.vm.Finished && m.vm.Result == "" {
			m.vm.Finished = true
			m.vm.Result = "Session ended."
		}
		return m, nil
	case tickAnimationMsg:
		return m, nil
	case tea.KeyMsg:
		if m.paused {
			return m.handlePauseKey(msg)
		}
		if m.vm.Finished {
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenResults} }
		}
		return m.handleGameKey(msg)
	}
	return m, nil
}

func (m GameModel) handleEnvelope(env session.Envelope) (tea.Model, tea.Cmd) {
	prev := m.vm
	m.vm = session.ReduceGameVM(m.vm, env)
	if m.vm.Prompt == nil {
		m.betMode = false
		m.betInput = ""
	}
	m.applyEnvelopeAudio(env, prev)
	if env.IsTerminal() {
		return m, nil
	}
	m.clearLocalMessage()
	return m, m.waitForEnvelope()
}

func (m *GameModel) applyEnvelopeAudio(env session.Envelope, prev session.GameVM) {
	if env.Notice != nil {
		switch env.Notice.Type {
		case "hand_started":
			m.handSoundState = handSoundState{}
		case "hole_cards_dealt":
			if dealt, ok := env.Notice.Event.(engine.HoleCardsDealtEvent); ok && dealt.PlayerID == m.sess.HumanID && !m.handSoundState.holeCuePlayed {
				playGameSound(audio.SoundHoleCards)
				m.handSoundState.holeCuePlayed = true
			}
		case "action_taken":
			if actionTaken, ok := env.Notice.Event.(engine.ActionTakenEvent); ok {
				if sound, ok := m.soundForAction(actionTaken); ok {
					playGameSound(sound)
				}
			}
		case "street_advanced":
			if advanced, ok := env.Notice.Event.(engine.StreetAdvancedEvent); ok {
				playGameSound(streetAdvanceSound(advanced))
			}
		case "showdown_started":
			playGameSound(audio.SoundShowdown)
		case "pot_awarded":
			if awarded, ok := env.Notice.Event.(engine.PotAwardedEvent); ok && m.humanWonPot(awarded.Winners) && !m.handSoundState.potCuePlayed {
				playGameSound(audio.SoundPotWon)
				m.handSoundState.potCuePlayed = true
			}
		case "player_eliminated":
			if eliminated, ok := env.Notice.Event.(engine.PlayerEliminatedEvent); ok && m.shouldPlayBustout(eliminated) {
				playGameSound(audio.SoundBustout)
				m.handSoundState.bustCuePlayed = true
			}
		case "blind_level_changed":
			playGameSound(audio.SoundBlindIncrease)
		case "waiting_for_human":
			if prev.Prompt == nil && m.vm.Prompt != nil {
				playGameSound(audio.SoundYourTurn)
			}
		case "tournament_finished", "session_ended":
			playGameSound(m.sessionEndSound(env))
		}
	}
	if env.Error != nil {
		switch env.Error.Code {
		case "invalid_action", "stale_action":
			playGameSound(audio.SoundInvalidAction)
		}
	}
}

func (m GameModel) soundForAction(e engine.ActionTakenEvent) (audio.SoundType, bool) {
	if e.Action.Type == engine.ActionAllIn {
		return audio.SoundAllIn, true
	}
	if e.PlayerID != m.sess.HumanID {
		switch e.Action.Type {
		case engine.ActionBet, engine.ActionRaise:
			return audio.SoundOpponentPressure, true
		default:
			return 0, false
		}
	}
	switch e.Action.Type {
	case engine.ActionCheck:
		return audio.SoundCheck, true
	case engine.ActionFold:
		return audio.SoundFold, true
	case engine.ActionCall, engine.ActionBet, engine.ActionRaise:
		return audio.SoundChip, true
	default:
		return 0, false
	}
}

func (m GameModel) humanWonPot(winners []engine.PlayerID) bool {
	for _, pid := range winners {
		if pid == m.sess.HumanID {
			return true
		}
	}
	return false
}

func (m GameModel) shouldPlayBustout(e engine.PlayerEliminatedEvent) bool {
	if m.handSoundState.bustCuePlayed {
		return false
	}
	if e.PlayerID == m.sess.HumanID {
		return true
	}
	return m.sess.Config.Mode != engine.ModeCashGame
}

func (m GameModel) sessionEndSound(env session.Envelope) audio.SoundType {
	if m.didHumanWinSession(env) {
		return audio.SoundVictory
	}
	return audio.SoundDefeat
}

func (m GameModel) didHumanWinSession(env session.Envelope) bool {
	if env.Notice == nil {
		return false
	}
	switch env.Notice.Type {
	case "tournament_finished":
		if finished, ok := env.Notice.Event.(engine.TournamentFinishedEvent); ok {
			for _, result := range finished.Results {
				if result.PlayerID == m.sess.HumanID {
					return result.Position == 1
				}
			}
		}
	case "session_ended":
		for _, player := range m.vm.Players {
			if player.ID == m.sess.HumanID {
				return player.Stack >= m.sess.Config.CashGameBuyIn
			}
		}
	}
	return false
}

func streetAdvanceSound(e engine.StreetAdvancedEvent) audio.SoundType {
	if len(e.NewCards) >= 3 {
		return audio.SoundFlop
	}
	return audio.SoundTurnRiver
}

func (m GameModel) handleGameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "escape" || msg.String() == "esc" {
		m.paused = true
		return m, nil
	}

	if !m.vm.HasPrompt() {
		return m, nil
	}

	if m.betMode {
		return m.handleBetInput(msg)
	}

	switch msg.String() {
	case "q":
		if m.hasLegal(engine.ActionFold) {
			return m.submitAction(engine.Action{Type: engine.ActionFold})
		}
		return m, nil
	case "w":
		if m.hasLegal(engine.ActionCheck) {
			return m.submitAction(engine.Action{Type: engine.ActionCheck})
		}
	case "e":
		if m.hasLegal(engine.ActionCall) {
			return m.submitAction(engine.Action{Type: engine.ActionCall})
		}
		if m.hasLegal(engine.ActionCheck) {
			return m.submitAction(engine.Action{Type: engine.ActionCheck})
		}
	case "t":
		if m.hasLegal(engine.ActionRaise) || m.hasLegal(engine.ActionBet) {
			m.betMode = true
			m.betInput = ""
			for _, legal := range m.vm.Prompt.LegalActions {
				if legal.Type == engine.ActionRaise || legal.Type == engine.ActionBet {
					m.betInput = strconv.Itoa(legal.MinAmount)
					break
				}
			}
		}
		return m, nil
	case "a":
		if m.hasLegal(engine.ActionAllIn) {
			return m.submitAction(engine.Action{Type: engine.ActionAllIn})
		}
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if m.hasLegal(engine.ActionRaise) || m.hasLegal(engine.ActionBet) {
			m.betMode = true
			m.betInput = msg.String()
		}
		return m, nil
	}

	return m, nil
}

func (m GameModel) handleBetInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc":
		m.betMode = false
		m.betInput = ""
		return m, nil
	case "enter":
		amount, err := strconv.Atoi(m.betInput)
		if err != nil || amount <= 0 {
			m.betMode = false
			m.betInput = ""
			return m, nil
		}
		actType := engine.ActionRaise
		for _, legal := range m.vm.Prompt.LegalActions {
			if legal.Type == engine.ActionBet {
				actType = engine.ActionBet
				break
			}
		}
		return m.submitAction(engine.Action{Type: actType, Amount: amount})
	case "backspace":
		if len(m.betInput) > 0 {
			m.betInput = m.betInput[:len(m.betInput)-1]
		}
		return m, nil
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if len(m.betInput) < 10 {
			m.betInput += msg.String()
		}
		return m, nil
	}
	return m, nil
}

func (m GameModel) submitAction(action engine.Action) (tea.Model, tea.Cmd) {
	if m.vm.Prompt == nil {
		return m, nil
	}
	m.betMode = false
	m.betInput = ""
	prompt := m.vm.Prompt
	return m, func() tea.Msg {
		m.sess.ActionResp <- session.PlayerActionIntent{
			PromptSeq: prompt.Seq,
			HandID:    prompt.HandID,
			Action:    action,
		}
		return m.waitForEnvelope()()
	}
}

func (m GameModel) hasLegal(actionType engine.ActionType) bool {
	if m.vm.Prompt == nil {
		return false
	}
	for _, legal := range m.vm.Prompt.LegalActions {
		if legal.Type == actionType {
			return true
		}
	}
	return false
}

func (m GameModel) handlePauseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc", "p":
		m.paused = false
	case "s":
		slot, err := nextSaveSlot()
		if err != nil {
			m.setLocalMessage(err.Error(), session.MessageKindError)
			return m, nil
		}
		if err := saveSessionToSlot(m.sess, slot); err != nil {
			m.setLocalMessage(err.Error(), session.MessageKindError)
			return m, nil
		}
		m.setLocalMessage(fmt.Sprintf("Game saved to slot %d.", slot), session.MessageKindInfo)
		return m, nil
	case "q":
		m.sess.Stop()
		return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
	case "h":
		return m, func() tea.Msg { return switchScreenMsg{screen: ScreenHelp} }
	}
	return m, nil
}

func (m GameModel) View() string {
	if m.width < 80 || m.height < 24 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			StyleError.Render("Terminal too small. Need at least 80x24."))
	}

	if m.paused {
		return m.renderPauseOverlay()
	}

	if m.vm.Finished {
		return m.renderFinished()
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.renderTable(),
		m.renderActionBar(),
	)
	return content
}

func (m GameModel) renderHeader() string {
	modeStr := "Tournament"
	switch m.sess.Config.Mode {
	case engine.ModeCashGame:
		modeStr = "Cash Game"
	case engine.ModeHeadsUpDuel:
		modeStr = "Heads-Up"
	}

	left := StyleBold.Render(fmt.Sprintf("ANTE  Hand #%d", m.vm.HandNum))
	mid := StyleDim.Render(fmt.Sprintf("Blinds %d/%d", m.vm.Blinds.SB, m.vm.Blinds.BB))
	if m.vm.Blinds.Ante > 0 {
		mid += StyleDim.Render(fmt.Sprintf(" (ante %d)", m.vm.Blinds.Ante))
	}
	right := StyleInfo.Render(modeStr) + "  " + StyleChips.Render(fmt.Sprintf("Stack: %s", ChipStr(m.vm.MyStack)))

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(mid) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}
	header := left + strings.Repeat(" ", gap/2) + mid + strings.Repeat(" ", gap-gap/2) + right
	return StyleHeader.Width(m.width).Render(header)
}

func (m GameModel) renderTable() string {
	tableH := m.height - 8
	if tableH < 16 {
		tableH = 16
	}

	human, opponents := m.splitPlayers()
	parts := []string{m.renderOpponentArea(opponents, m.width)}
	if m.vm.Showdown && len(m.vm.Revealed) > 0 {
		parts = append(parts, m.renderShowdown())
	}
	parts = append(parts, m.renderBoardArea(m.width), m.renderHumanArea(human, m.width))
	if m.messageText() != "" {
		parts = append(parts, CenterH(m.width).Render(m.renderMessage()))
	}
	table := lipgloss.JoinVertical(lipgloss.Left, parts...)
	tableLines := strings.Count(table, "\n") + 1
	if tableLines < tableH {
		table += strings.Repeat("\n", tableH-tableLines)
	}
	return table
}

func (m GameModel) renderMessage() string {
	message := m.messageText()
	if message == "" {
		return ""
	}
	if m.messageKind() == session.MessageKindError {
		return StyleError.Render(message)
	}
	return StyleInfo.Render(message)
}

func (m GameModel) splitPlayers() (*session.PlayerInfo, []session.PlayerInfo) {
	var human *session.PlayerInfo
	opponents := make([]session.PlayerInfo, 0, len(m.vm.Players))
	for i := range m.vm.Players {
		player := m.vm.Players[i]
		if player.IsHuman {
			h := player
			human = &h
			continue
		}
		if player.Status == engine.StatusOut || player.Status == engine.StatusSittingOut {
			continue
		}
		opponents = append(opponents, player)
	}
	return human, opponents
}

func (m GameModel) renderOpponentArea(opponents []session.PlayerInfo, width int) string {
	if len(opponents) == 0 {
		return ""
	}
	seats := make([]string, 0, len(opponents))
	for _, opponent := range opponents {
		seats = append(seats, m.renderSeat(opponent))
	}
	maxPerRow := (width + SeatGap) / (SeatTotalWidth + SeatGap)
	if maxPerRow < 1 {
		maxPerRow = 1
	}
	rows := make([]string, 0, (len(seats)+maxPerRow-1)/maxPerRow)
	for i := 0; i < len(seats); i += maxPerRow {
		end := i + maxPerRow
		if end > len(seats) {
			end = len(seats)
		}
		rows = append(rows, joinSeats(seats[i:end], width))
	}
	return strings.Join(rows, "\n")
}

func joinSeats(seats []string, width int) string {
	if len(seats) == 0 {
		return ""
	}
	parts := make([]string, 0, len(seats))
	for i, seat := range seats {
		if i < len(seats)-1 {
			seat = lipgloss.NewStyle().MarginRight(SeatGap).Render(seat)
		}
		parts = append(parts, seat)
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, parts...))
}

func (m GameModel) renderSeat(player session.PlayerInfo) string {
	isFolded := player.Status == engine.StatusFolded
	isAllIn := player.Status == engine.StatusAllIn
	isOut := player.Status == engine.StatusOut || player.Status == engine.StatusSittingOut
	isDealer := player.Seat == m.vm.DealerSeat

	nameLine := player.Name
	if isDealer {
		badge := " " + StyleDealer.Render("D")
		maxName := SeatContentWidth - 2
		runes := []rune(player.Name)
		if len(runes) > maxName {
			nameLine = string(runes[:maxName-1]) + "..."
		}
		nameLine += badge
	} else if lipgloss.Width(nameLine) > SeatContentWidth {
		runes := []rune(player.Name)
		if len(runes) > SeatContentWidth-1 {
			nameLine = string(runes[:SeatContentWidth-1]) + "..."
		}
	}

	if isOut {
		content := strings.Join([]string{nameLine, ChipStr(player.Stack), "OUT", ""}, "\n")
		return lipgloss.NewStyle().Width(SeatTotalWidth).Height(SeatHeight).Padding(0, 1).Foreground(ColorDim).Render(content)
	}

	stackLine := StyleChips.Render(ChipStr(player.Stack))
	statusLine := ""
	if isFolded {
		statusLine = StyleFolded.Render("FOLDED")
	} else if isAllIn {
		statusLine = StyleAllIn.Render("ALL-IN")
	} else if player.Bet > 0 {
		statusLine = StyleBet.Render("Bet: " + ChipStr(player.Bet))
	}

	cardLine := ""
	if !isFolded {
		cardLine = "[" + CardBack() + "][" + CardBack() + "]"
		for _, revealed := range m.vm.Revealed {
			if revealed.PlayerID == player.ID {
				cardLine = RenderHoleCards(revealed.Cards, true)
				break
			}
		}
	}

	style := SeatStyle(true, false, isFolded, isAllIn, isDealer)
	content := strings.Join([]string{nameLine, stackLine, statusLine, cardLine}, "\n")
	return style.Render(content)
}

func (m GameModel) renderBoardArea(width int) string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		RenderBoardLarge(m.vm.Board),
		StylePot.Render(fmt.Sprintf("Pot: %s", ChipStr(m.vm.Pot))),
		StyleDim.Render(StreetStr(m.vm.Street)),
		"",
	)
	return CenterH(width).Render(content)
}

func (m GameModel) renderHumanArea(human *session.PlayerInfo, width int) string {
	if human == nil {
		return ""
	}
	label := StyleHumanLabel.Render("* " + human.Name + " *")
	stack := StyleChips.Render(fmt.Sprintf("Stack: %s", ChipStr(m.vm.MyStack)))
	betStr := ""
	if m.vm.MyBet > 0 {
		betStr = StyleDim.Render(fmt.Sprintf("Bet: %s", ChipStr(m.vm.MyBet)))
	}
	dealerStr := ""
	if human.Seat == m.vm.DealerSeat {
		dealerStr = " " + StyleDealer.Render("D")
	}
	info := stack
	if betStr != "" {
		info += "  " + betStr
	}
	if dealerStr != "" {
		info += dealerStr
	}
	content := lipgloss.JoinVertical(lipgloss.Center, label, RenderBigCards(m.vm.HumanCards), info)
	return CenterH(width).Render(content)
}

func (m GameModel) renderShowdown() string {
	lines := make([]string, 0, len(m.vm.Revealed)+len(m.vm.PotAwards))
	for _, revealed := range m.vm.Revealed {
		lines = append(lines, fmt.Sprintf("  %s: %s  %s", revealed.Name, RenderHoleCards(revealed.Cards, true), StyleHandRank.Render(revealed.Eval)))
	}
	for _, award := range m.vm.PotAwards {
		lines = append(lines, "  "+StyleWinner.Render(award))
	}
	return strings.Join(lines, "\n")
}

func (m GameModel) renderActionBar() string {
	if m.vm.Prompt == nil {
		info := m.vm.StatusLine
		if info == "" {
			info = "Waiting..."
		}
		return StyleFooter.Width(m.width).Render(StyleDim.Render(info))
	}

	extraLine := ""
	if m.messageText() != "" {
		extraLine = m.renderMessage()
	} else if odds := m.potOddsLine(); odds != "" {
		extraLine = StyleInfo.Render(odds)
	}

	if m.betMode {
		minMax := ""
		for _, legal := range m.vm.Prompt.LegalActions {
			if legal.Type == engine.ActionRaise || legal.Type == engine.ActionBet {
				minMax = fmt.Sprintf("(min: %s  max: %s)", ChipStr(legal.MinAmount), ChipStr(legal.MaxAmount))
				break
			}
		}
		betLine := fmt.Sprintf("%s Amount: %s %s  %s", StyleKey.Render("[Enter]"), StyleBold.Render(m.betInput+"_"), minMax, StyleDim.Render("[Esc] Cancel"))
		if extraLine != "" {
			return StyleFooter.Width(m.width).Render(extraLine + "\n" + betLine)
		}
		return StyleFooter.Width(m.width).Render(betLine)
	}

	actions := make([]string, 0, len(m.vm.Prompt.LegalActions))
	if m.hasLegal(engine.ActionFold) {
		actions = append(actions, StyleKey.Render("[Q]")+" Fold")
	}
	if m.hasLegal(engine.ActionCheck) {
		actions = append(actions, StyleKey.Render("[W]")+" Check")
	}
	if m.hasLegal(engine.ActionCall) {
		for _, legal := range m.vm.Prompt.LegalActions {
			if legal.Type == engine.ActionCall {
				actions = append(actions, StyleKey.Render("[E]")+" Call "+ChipStr(legal.MinAmount))
				break
			}
		}
	}
	if m.hasLegal(engine.ActionRaise) {
		actions = append(actions, StyleKey.Render("[T]")+" Raise")
	} else if m.hasLegal(engine.ActionBet) {
		actions = append(actions, StyleKey.Render("[T]")+" Bet")
	}
	if m.hasLegal(engine.ActionAllIn) {
		for _, legal := range m.vm.Prompt.LegalActions {
			if legal.Type == engine.ActionAllIn {
				actions = append(actions, StyleKey.Render("[A]")+" All-In "+ChipStr(max(0, legal.MinAmount-m.vm.MyBet)))
				break
			}
		}
	}
	actionLine := strings.Join(actions, "   ")
	if m.vm.StatusLine != "" {
		actionLine += "   " + StyleDim.Render("| "+m.vm.StatusLine)
	}
	if extraLine != "" {
		return StyleFooter.Width(m.width).Render(extraLine + "\n" + actionLine)
	}
	return StyleFooter.Width(m.width).Render(actionLine)
}

func (m GameModel) potOddsLine() string {
	if !m.showPotOdds || m.vm.Prompt == nil {
		return ""
	}
	toCall := m.vm.Prompt.View.CurrentBet - m.vm.Prompt.View.MyBet
	if toCall <= 0 {
		return ""
	}
	odds := float64(m.vm.Prompt.View.Pot+toCall) / float64(toCall)
	pct := 100.0 / odds
	return fmt.Sprintf("Pot: %s | Call: %s | Odds: %.1f:1 (%.1f%%)", ChipStr(m.vm.Prompt.View.Pot), ChipStr(toCall), odds-1, pct)
}

func (m GameModel) renderPauseOverlay() string {
	parts := []string{StyleTitle.Render("GAME PAUSED"), ""}
	if m.localMessage != "" {
		parts = append(parts, m.renderPauseMessage(), "")
	}
	parts = append(parts,
		StyleKey.Render("[Esc]")+" Resume",
		StyleKey.Render("[S]")+"   Save Game",
		StyleKey.Render("[H]")+"   Help",
		StyleKey.Render("[Q]")+"   Quit to Menu",
	)
	menu := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorGold).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, menu)
}

func (m GameModel) renderFinished() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		StyleTitle.Render("GAME OVER"),
		"",
		StyleBold.Render(m.vm.Result),
		"",
		StyleDim.Render("Press any key to continue..."),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *GameModel) setLocalMessage(message string, kind session.MessageKind) {
	m.localMessage = message
	m.localMessageKind = kind
}

func (m *GameModel) clearLocalMessage() {
	m.localMessage = ""
	m.localMessageKind = session.MessageKindNone
}

func (m GameModel) messageText() string {
	if m.localMessage != "" {
		return m.localMessage
	}
	return m.vm.Message
}

func (m GameModel) messageKind() session.MessageKind {
	if m.localMessage != "" {
		return m.localMessageKind
	}
	return m.vm.MessageKind
}

func (m GameModel) renderPauseMessage() string {
	if m.localMessage == "" {
		return ""
	}
	if m.localMessageKind == session.MessageKindError {
		return StyleError.Render(m.localMessage)
	}
	return StyleSuccess.Render(m.localMessage)
}

func nextSaveSlot() (int, error) {
	saves, err := listSaves()
	if err != nil && len(saves) == 0 {
		return 0, err
	}
	for _, save := range saves {
		if save.Empty {
			return save.Slot, nil
		}
	}
	if len(saves) == 0 {
		return 0, session.ErrSaveSlotUnavailable
	}
	selected := saves[0]
	for _, save := range saves[1:] {
		if save.Timestamp.Before(selected.Timestamp) {
			selected = save
		}
	}
	return selected.Slot, nil
}
