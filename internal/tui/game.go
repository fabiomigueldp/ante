package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Messages for game screen communication.
type (
	sessionEventMsg  session.SessionEvent
	actionReqMsg     session.ActionRequest
	sessionDoneMsg   struct{}
	tickAnimationMsg struct{}
)

var playGameSound = audio.Play

// GameModel is the main game table screen.
type GameModel struct {
	width  int
	height int

	// Session
	sess *session.Session

	// Display state
	players    []session.PlayerInfo
	board      []engine.Card
	pot        int
	street     engine.Street
	handNum    int
	blinds     engine.BlindLevel
	dealerSeat int
	lastAction string
	message    string
	msgExpiry  time.Time

	// Human cards
	myCards [2]engine.Card
	myStack int
	myBet   int

	// Action state
	needsAction  bool
	actionReq    *session.ActionRequest
	legalActions []engine.LegalAction
	betInput     string
	betMode      bool
	potOddsStr   string

	// Animation
	showdown       bool
	revealed       []revealedHand
	potAwards      []string
	animPhase      int
	eventQueue     []session.SessionEvent
	processing     bool
	handSoundState handSoundState

	// Session ended
	finished bool
	result   string

	// Settings
	showPotOdds bool

	// Pause
	paused bool
}

type revealedHand struct {
	playerID engine.PlayerID
	name     string
	cards    [2]engine.Card
	eval     string
}

type handSoundState struct {
	holeCuePlayed bool
	potCuePlayed  bool
	bustCuePlayed bool
}

func NewGameModel(sess *session.Session, showPotOdds bool) GameModel {
	ts := sess.TableState()
	return GameModel{
		sess:        sess,
		players:     ts.Players,
		handNum:     ts.HandNum,
		blinds:      ts.Blinds,
		showPotOdds: showPotOdds,
	}
}

func (m GameModel) Init() tea.Cmd {
	// Start session in background goroutine
	go func(sess *session.Session) {
		defer func() {
			if r := recover(); r != nil {
				sess.Stop()
			}
		}()
		sess.Run()
	}(m.sess)
	return tea.Batch(
		m.waitForSession(),
	)
}

func (m GameModel) waitForSession() tea.Cmd {
	return func() tea.Msg {
		select {
		case ev, ok := <-m.sess.Events:
			if !ok {
				return sessionDoneMsg{}
			}
			return sessionEventMsg(ev)
		case ar, ok := <-m.sess.ActionReq:
			if !ok {
				return sessionDoneMsg{}
			}
			return actionReqMsg(ar)
		}
	}
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case sessionEventMsg:
		return m.handleSessionEvent(session.SessionEvent(msg))

	case actionReqMsg:
		return m.handleActionReq(session.ActionRequest(msg))

	case sessionDoneMsg:
		m.finished = true
		if m.result == "" {
			m.result = "Session ended."
		}
		return m, nil

	case tickAnimationMsg:
		return m.advanceAnimation()

	case tea.KeyMsg:
		if m.paused {
			return m.handlePauseKey(msg)
		}
		if m.finished {
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenResults} }
		}
		return m.handleGameKey(msg)
	}
	return m, nil
}

func (m GameModel) handleSessionEvent(ev session.SessionEvent) (tea.Model, tea.Cmd) {
	switch ev.Type {
	case "hand_started":
		if e, ok := ev.Event.(engine.HandStartedEvent); ok {
			m.dealerSeat = e.DealerSeat
			m.blinds = e.Blinds
			m.board = nil
			m.pot = 0
			m.showdown = false
			m.revealed = nil
			m.potAwards = nil
			m.myBet = 0
			m.lastAction = ""
			m.handSoundState = handSoundState{}
		}
		m.applySnapshot(ev.Snapshot)

	case "blind_posted":
		if e, ok := ev.Event.(engine.BlindsPostedEvent); ok {
			name := m.sess.PlayerName(e.PlayerID)
			blind := "SB"
			if e.Type == engine.BlindBig {
				blind = "BB"
			} else if e.Type == engine.BlindAnte {
				blind = "ante"
			}
			m.lastAction = fmt.Sprintf("%s posts %s %s", name, blind, ChipStr(e.Amount))
		}

	case "hole_cards_dealt":
		if e, ok := ev.Event.(engine.HoleCardsDealtEvent); ok {
			if e.PlayerID == m.sess.HumanID {
				m.myCards = e.Cards
				if !m.handSoundState.holeCuePlayed {
					playGameSound(audio.SoundHoleCards)
					m.handSoundState.holeCuePlayed = true
				}
			}
		}

	case "action_taken":
		if e, ok := ev.Event.(engine.ActionTakenEvent); ok {
			name := m.sess.PlayerName(e.PlayerID)
			actStr := ActionStr(e.Action.Type)
			if e.Action.Amount > 0 {
				m.lastAction = fmt.Sprintf("%s %s %s", name, actStr, ChipStr(e.Action.Amount))
			} else {
				m.lastAction = fmt.Sprintf("%s %s", name, actStr)
			}
			m.pot = e.PotTotal
			if sound, ok := m.soundForAction(e); ok {
				playGameSound(sound)
			}
		}
		m.applySnapshot(ev.Snapshot)

	case "street_advanced":
		if e, ok := ev.Event.(engine.StreetAdvancedEvent); ok {
			m.board = append(m.board, e.NewCards...)
			m.street = e.Street
			m.myBet = 0
			playGameSound(streetAdvanceSound(e))
		}
		m.applySnapshot(ev.Snapshot)

	case "showdown_started":
		m.showdown = true
		playGameSound(audio.SoundShowdown)

	case "hand_revealed":
		if e, ok := ev.Event.(engine.HandRevealedEvent); ok {
			m.revealed = append(m.revealed, revealedHand{
				playerID: e.PlayerID,
				name:     m.sess.PlayerName(e.PlayerID),
				cards:    e.Cards,
				eval:     e.Eval.Name,
			})
		}

	case "pot_awarded":
		if e, ok := ev.Event.(engine.PotAwardedEvent); ok {
			names := make([]string, len(e.Winners))
			for i, pid := range e.Winners {
				names[i] = m.sess.PlayerName(pid)
			}
			m.potAwards = append(m.potAwards, fmt.Sprintf("%s wins %s", strings.Join(names, ", "), ChipStr(e.Amount)))
			if m.humanWonPot(e.Winners) && !m.handSoundState.potCuePlayed {
				playGameSound(audio.SoundPotWon)
				m.handSoundState.potCuePlayed = true
			}
		}

	case "player_eliminated":
		if e, ok := ev.Event.(engine.PlayerEliminatedEvent); ok {
			name := m.sess.PlayerName(e.PlayerID)
			m.setMessage(fmt.Sprintf("%s eliminated (#%d)", name, e.Position))
			if m.shouldPlayBustout(e) {
				playGameSound(audio.SoundBustout)
				m.handSoundState.bustCuePlayed = true
			}
		}

	case "blind_level_changed":
		m.setMessage(ev.Message)
		playGameSound(audio.SoundBlindIncrease)
		m.applySnapshot(ev.Snapshot)

	case "hand_complete":
		m.handNum++
		m.applySnapshot(ev.Snapshot)

	case "hand_summary":
		m.applySnapshot(ev.Snapshot)

	case "action_error":
		m.setMessage(ev.Message)
		playGameSound(audio.SoundInvalidAction)
		return m, m.waitForSession()

	case "session_error":
		m.finished = true
		m.result = ev.Message
		return m, nil

	case "bot_thinking":
		if ev.ThinkTime > 0 {
			m.lastAction = fmt.Sprintf("%s is thinking...", ev.BotName)
		}

	case "tournament_finished", "session_ended":
		m.finished = true
		m.result = ev.Message
		playGameSound(m.sessionEndSound(ev))
		return m, nil

	case "waiting_for_human":
		m.applySnapshot(ev.Snapshot)
		playGameSound(audio.SoundYourTurn)
	}

	return m, m.waitForSession()
}

func (m *GameModel) handleActionReq(req session.ActionRequest) (tea.Model, tea.Cmd) {
	m.needsAction = true
	ar := req
	m.actionReq = &ar
	m.legalActions = req.LegalActions
	m.betInput = ""
	m.betMode = false

	// Update display from view
	m.myCards = req.View.MyCards
	m.myStack = req.View.MyStack
	m.myBet = req.View.MyBet
	m.pot = req.View.Pot
	m.street = req.View.Street
	m.board = req.View.Board
	m.applySnapshot(req.Snapshot)

	// Calculate pot odds
	toCall := req.View.CurrentBet - req.View.MyBet
	if m.showPotOdds && toCall > 0 {
		odds := float64(req.View.Pot+toCall) / float64(toCall)
		pct := 100.0 / odds
		m.potOddsStr = fmt.Sprintf("Pot: %s | Call: %s | Odds: %.1f:1 (%.1f%%)",
			ChipStr(req.View.Pot), ChipStr(toCall), odds-1, pct)
	} else {
		m.potOddsStr = ""
	}

	return *m, nil
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

func (m GameModel) sessionEndSound(ev session.SessionEvent) audio.SoundType {
	if m.didHumanWinSession(ev) {
		return audio.SoundVictory
	}
	return audio.SoundDefeat
}

func (m GameModel) didHumanWinSession(ev session.SessionEvent) bool {
	switch ev.Type {
	case "tournament_finished":
		if e, ok := ev.Event.(engine.TournamentFinishedEvent); ok {
			for _, result := range e.Results {
				if result.PlayerID == m.sess.HumanID {
					return result.Position == 1
				}
			}
		}
	case "session_ended":
		human := m.findPlayer(m.sess.HumanID)
		if human == nil {
			return false
		}
		return human.Stack >= m.sess.Config.CashGameBuyIn
	}
	return false
}

func (m GameModel) findPlayer(pid engine.PlayerID) *session.PlayerInfo {
	for i := range m.players {
		if m.players[i].ID == pid {
			return &m.players[i]
		}
	}
	return nil
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

	if !m.needsAction {
		return m, nil
	}

	if m.betMode {
		return m.handleBetInput(msg)
	}

	switch msg.String() {
	case "q": // Fold
		return m.submitAction(engine.Action{Type: engine.ActionFold})
	case "w": // Check
		if m.hasLegal(engine.ActionCheck) {
			return m.submitAction(engine.Action{Type: engine.ActionCheck})
		}
	case "e": // Call
		if m.hasLegal(engine.ActionCall) {
			return m.submitAction(engine.Action{Type: engine.ActionCall})
		}
		if m.hasLegal(engine.ActionCheck) {
			return m.submitAction(engine.Action{Type: engine.ActionCheck})
		}
	case "t": // Raise/Bet - enter bet mode
		if m.hasLegal(engine.ActionRaise) || m.hasLegal(engine.ActionBet) {
			m.betMode = true
			m.betInput = ""
			// Pre-fill with minimum raise
			for _, la := range m.legalActions {
				if la.Type == engine.ActionRaise || la.Type == engine.ActionBet {
					m.betInput = strconv.Itoa(la.MinAmount)
					break
				}
			}
			return m, nil
		}
	case "a": // All-in
		if m.hasLegal(engine.ActionAllIn) {
			return m.submitAction(engine.Action{Type: engine.ActionAllIn})
		}
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Start bet mode with this digit
		if m.hasLegal(engine.ActionRaise) || m.hasLegal(engine.ActionBet) {
			m.betMode = true
			m.betInput = msg.String()
			return m, nil
		}
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
			return m, nil
		}
		// Determine if this is a bet or raise
		actType := engine.ActionRaise
		for _, la := range m.legalActions {
			if la.Type == engine.ActionBet {
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
	if !m.needsAction {
		return m, nil
	}
	m.needsAction = false
	m.betMode = false
	m.potOddsStr = ""

	return m, func() tea.Msg {
		m.sess.ActionResp <- action
		return m.waitForSession()()
	}
}

func (m *GameModel) hasLegal(t engine.ActionType) bool {
	for _, la := range m.legalActions {
		if la.Type == t {
			return true
		}
	}
	return false
}

func (m *GameModel) applySnapshot(ts session.TableState) {
	if ts.Players == nil {
		return
	}
	m.players = ts.Players
	m.handNum = ts.HandNum
	m.blinds = ts.Blinds
	for _, p := range ts.Players {
		if p.IsHuman {
			m.myStack = p.Stack
			break
		}
	}
}

func (m *GameModel) setMessage(msg string) {
	m.message = msg
	m.msgExpiry = time.Now().Add(4 * time.Second)
}

func (m GameModel) advanceAnimation() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m GameModel) handlePauseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "esc", "p":
		m.paused = false
	case "s":
		// TODO: Save game
		m.setMessage("Game saved!")
		m.paused = false
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

	if m.finished {
		return m.renderFinished()
	}

	header := m.renderHeader()
	table := m.renderTable()
	actionBar := m.renderActionBar()

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		table,
		actionBar,
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

	left := StyleBold.Render(fmt.Sprintf("ANTE  Hand #%d", m.handNum))
	mid := StyleDim.Render(fmt.Sprintf("Blinds %d/%d", m.blinds.SB, m.blinds.BB))
	if m.blinds.Ante > 0 {
		mid += StyleDim.Render(fmt.Sprintf(" (ante %d)", m.blinds.Ante))
	}
	right := StyleInfo.Render(modeStr) + "  " + StyleChips.Render(fmt.Sprintf("Stack: %s", ChipStr(m.myStack)))

	w := m.width
	gap := w - lipgloss.Width(left) - lipgloss.Width(mid) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}
	header := left + strings.Repeat(" ", gap/2) + mid + strings.Repeat(" ", gap-gap/2) + right
	return StyleHeader.Width(w).Render(header)
}

func (m GameModel) renderTable() string {
	w := m.width
	// Calculate available height: total - header(2) - action bar(3) - footer(2)
	tableH := m.height - 8
	if tableH < 16 {
		tableH = 16
	}

	// Separate players into rows
	humanPlayer, opponents := m.splitPlayers()

	// Top row of opponents
	topRow := m.renderOpponentRow(opponents, 0, w)
	// Middle: board + pot
	boardRow := m.renderBoardArea(w)
	// Bottom: human player
	humanRow := m.renderHumanArea(humanPlayer, w)

	// Add showdown info if applicable
	showdownInfo := ""
	if m.showdown && len(m.revealed) > 0 {
		showdownInfo = m.renderShowdown()
	}

	// Compose table
	var parts []string
	parts = append(parts, topRow)
	if showdownInfo != "" {
		parts = append(parts, showdownInfo)
	}
	parts = append(parts, boardRow)
	parts = append(parts, humanRow)

	// Add message if present
	if m.message != "" && time.Now().Before(m.msgExpiry) {
		parts = append(parts, CenterH(w).Render(StyleInfo.Render(m.message)))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Pad to fill available height
	tableLines := strings.Count(table, "\n") + 1
	if tableLines < tableH {
		table += strings.Repeat("\n", tableH-tableLines)
	}

	return table
}

func (m GameModel) splitPlayers() (*session.PlayerInfo, []session.PlayerInfo) {
	var human *session.PlayerInfo
	var opponents []session.PlayerInfo
	for i := range m.players {
		if m.players[i].IsHuman {
			h := m.players[i]
			human = &h
		} else {
			opponents = append(opponents, m.players[i])
		}
	}
	return human, opponents
}

func (m GameModel) renderOpponentRow(opponents []session.PlayerInfo, startIdx int, width int) string {
	if len(opponents) == 0 {
		return ""
	}

	var seats []string
	for _, opp := range opponents {
		seats = append(seats, m.renderSeat(opp))
	}

	// Distribute seats evenly across width
	totalSeatWidth := 0
	for _, s := range seats {
		totalSeatWidth += lipgloss.Width(s)
	}

	gap := (width - totalSeatWidth) / (len(seats) + 1)
	if gap < 1 {
		gap = 1
	}

	var row string
	for _, s := range seats {
		row += strings.Repeat(" ", gap) + s
	}

	return row + "\n"
}

func (m GameModel) renderSeat(p session.PlayerInfo) string {
	isFolded := p.Status == engine.StatusFolded
	isAllIn := p.Status == engine.StatusAllIn
	isOut := p.Status == engine.StatusOut || p.Status == engine.StatusSittingOut
	isDealer := p.Seat == m.dealerSeat

	if isOut {
		return lipgloss.NewStyle().Foreground(ColorDim).Width(22).Render(
			fmt.Sprintf("%s\n  (out)", p.Name))
	}

	// Name line with dealer button
	nameLine := p.Name
	if isDealer {
		nameLine += " " + StyleDealer.Render("(D)")
	}

	// Truncate name if too long
	if lipgloss.Width(nameLine) > 22 {
		runes := []rune(nameLine)
		if len(runes) > 20 {
			nameLine = string(runes[:19]) + "…"
		}
	}

	// Stack + bet on same line
	stackLine := StyleChips.Render(ChipStr(p.Stack))
	if p.Bet > 0 {
		stackLine += "  " + StyleBet.Render("Bet: "+ChipStr(p.Bet))
	}

	// Status
	statusLine := ""
	if isFolded {
		statusLine = StyleFolded.Render("FOLDED")
	} else if isAllIn {
		statusLine = StyleAllIn.Render("ALL-IN")
	}

	// Cards (face down for opponents in active hand)
	cardLine := "[##][##]"
	// Check if this player has been revealed in showdown
	for _, rev := range m.revealed {
		if rev.playerID == p.ID {
			cardLine = RenderHoleCards(rev.cards, true)
			break
		}
	}
	if isFolded {
		cardLine = ""
	}

	style := SeatStyle(true, false, isFolded, isAllIn, isDealer)

	content := nameLine + "\n" + stackLine
	if statusLine != "" {
		content += "  " + statusLine
	}
	if cardLine != "" {
		content += "\n" + cardLine
	}

	return style.Render(content)
}

func (m GameModel) renderBoardArea(width int) string {
	// Community cards
	boardStr := RenderBoardLarge(m.board)

	// Pot
	potLine := StylePot.Render(fmt.Sprintf("Pot: %s", ChipStr(m.pot)))

	// Street
	streetLine := StyleDim.Render(StreetStr(m.street))

	content := lipgloss.JoinVertical(lipgloss.Center,
		"", boardStr, potLine, streetLine, "",
	)

	return CenterH(width).Render(content)
}

func (m GameModel) renderHumanArea(human *session.PlayerInfo, width int) string {
	if human == nil {
		return ""
	}

	label := StyleHumanLabel.Render("* " + human.Name + " *")
	cards := RenderBigCards(m.myCards)
	stack := StyleChips.Render(fmt.Sprintf("Stack: %s", ChipStr(m.myStack)))

	betStr := ""
	if m.myBet > 0 {
		betStr = StyleDim.Render(fmt.Sprintf("Bet: %s", ChipStr(m.myBet)))
	}

	isDealer := human.Seat == m.dealerSeat
	dealerStr := ""
	if isDealer {
		dealerStr = StyleDealer.Render(" (D)")
	}

	info := stack
	if betStr != "" {
		info += "  " + betStr
	}
	if dealerStr != "" {
		info += dealerStr
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		label,
		cards,
		info,
	)

	return CenterH(width).Render(content)
}

func (m GameModel) renderShowdown() string {
	var lines []string
	for _, rev := range m.revealed {
		cards := RenderHoleCards(rev.cards, true)
		line := fmt.Sprintf("  %s: %s  %s", rev.name, cards, StyleHandRank.Render(rev.eval))
		lines = append(lines, line)
	}
	for _, award := range m.potAwards {
		lines = append(lines, "  "+StyleWinner.Render(award))
	}
	return strings.Join(lines, "\n")
}

func (m GameModel) renderActionBar() string {
	w := m.width

	if !m.needsAction {
		// Show last action or waiting message
		info := m.lastAction
		if info == "" {
			info = "Waiting..."
		}
		bar := StyleDim.Render(info)
		return StyleFooter.Width(w).Render(bar)
	}

	// Pot odds line
	oddsLine := ""
	if m.potOddsStr != "" {
		oddsLine = StyleInfo.Render(m.potOddsStr)
	}

	if m.betMode {
		// Bet input mode
		var minMax string
		for _, la := range m.legalActions {
			if la.Type == engine.ActionRaise || la.Type == engine.ActionBet {
				minMax = fmt.Sprintf("(min: %s  max: %s)", ChipStr(la.MinAmount), ChipStr(la.MaxAmount))
				break
			}
		}

		betLine := fmt.Sprintf("%s Amount: %s %s  %s",
			StyleKey.Render("[Enter]"), StyleBold.Render(m.betInput+"_"), minMax,
			StyleDim.Render("[Esc] Cancel"))

		content := betLine
		if oddsLine != "" {
			content = oddsLine + "\n" + betLine
		}
		return StyleFooter.Width(w).Render(content)
	}

	// Normal action bar
	var actions []string

	if m.hasLegal(engine.ActionFold) {
		actions = append(actions, StyleKey.Render("[Q]")+" Fold")
	}
	if m.hasLegal(engine.ActionCheck) {
		actions = append(actions, StyleKey.Render("[W]")+" Check")
	}
	if m.hasLegal(engine.ActionCall) {
		for _, la := range m.legalActions {
			if la.Type == engine.ActionCall {
				actions = append(actions, StyleKey.Render("[E]")+" Call "+ChipStr(la.MinAmount))
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
		for _, la := range m.legalActions {
			if la.Type == engine.ActionAllIn {
				allInDisplay := la.MinAmount - m.myBet
				if allInDisplay < 0 {
					allInDisplay = 0
				}
				actions = append(actions, StyleKey.Render("[A]")+" All-In "+ChipStr(allInDisplay))
				break
			}
		}
	}

	actionLine := strings.Join(actions, "   ")
	if m.lastAction != "" {
		actionLine += "   " + StyleDim.Render("| "+m.lastAction)
	}

	content := actionLine
	if oddsLine != "" {
		content = oddsLine + "\n" + actionLine
	}
	return StyleFooter.Width(w).Render(content)
}

func (m GameModel) renderPauseOverlay() string {
	menu := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorGold).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			StyleTitle.Render("GAME PAUSED"),
			"",
			StyleKey.Render("[Esc]")+" Resume",
			StyleKey.Render("[S]")+"   Save Game",
			StyleKey.Render("[H]")+"   Help",
			StyleKey.Render("[Q]")+"   Quit to Menu",
		))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, menu)
}

func (m GameModel) renderFinished() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		StyleTitle.Render("GAME OVER"),
		"",
		StyleBold.Render(m.result),
		"",
		StyleDim.Render("Press any key to continue..."),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
