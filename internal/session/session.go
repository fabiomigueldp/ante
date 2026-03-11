package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

// Config holds all parameters needed to start a game session.
type Config struct {
	Mode           engine.GameMode
	Difficulty     ai.Difficulty
	Seats          int    // total seats including human (6 or 9)
	StartingStack  int    // in chips (e.g., 100, 200, 500 BB worth)
	BlindSpeed     string // "normal", "turbo", "slow" (tournament only)
	PlayerName     string
	Seed           int64  // master seed, 0 = random
	CashGameBuyIn  int    // cash game only
	CashGameBlinds [2]int // cash game only: {SB, BB}
}

// Phase represents the high-level session lifecycle.
type Phase uint8

const (
	PhaseSetup Phase = iota
	PhasePlaying
	PhaseHandComplete
	PhaseSessionOver
)

// HandSummary is emitted after each hand completes.
type HandSummary struct {
	HandID       int
	Winners      map[int][]engine.PlayerID // pot index -> winners
	Eliminations []engine.PlayerEliminatedEvent
	BlindChange  *engine.BlindLevelChangedEvent
	PlayerStack  int // human's stack after hand
	IsFinished   bool
}

// Session orchestrates the full game: Table + Hand + Bots + Human interaction.
type Session struct {
	Config     Config
	SessionID  string
	Table      *engine.Table
	Tournament *engine.Tournament
	CashGame   *engine.CashGame
	History    *engine.SessionHistory

	Bots      map[engine.PlayerID]*ai.Bot
	HumanID   engine.PlayerID
	Phase     Phase
	HandCount int

	// Channels for TUI communication
	Updates    chan Envelope           // session -> TUI (buffered)
	ActionResp chan PlayerActionIntent // TUI -> session (unbuffered)

	// Internal
	currentHand *engine.Hand
	rng         *mathrand.Rand
	botOrder    []engine.PlayerID // for deterministic iteration
	stop        chan struct{}
	stopOnce    sync.Once
	seq         uint64
	resumed     bool
	deps        Dependencies
	transcript  *TranscriptWriter
	metrics     *MetricsAccumulator
	Summary     *storage.SessionSummary
}

// New creates a session from config. Does not start the game loop.
func New(cfg Config) (*Session, error) {
	if cfg.Seats < 2 || cfg.Seats > 9 {
		return nil, fmt.Errorf("seats must be between 2 and 9")
	}
	if cfg.PlayerName == "" {
		cfg.PlayerName = "You"
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	if cfg.StartingStack <= 0 {
		cfg.StartingStack = 200 // default 200 BB
	}

	rng := mathrand.New(mathrand.NewSource(cfg.Seed))

	// Build blind structure
	var structure engine.BlindStructure
	switch cfg.Mode {
	case engine.ModeHeadsUpDuel:
		structure = engine.HeadsUpBlinds()
		cfg.Seats = 2
	case engine.ModeTournament:
		speed := cfg.BlindSpeed
		if speed == "" {
			speed = "normal"
		}
		structure = engine.TournamentBlinds(speed)
	case engine.ModeCashGame:
		sb, bb := cfg.CashGameBlinds[0], cfg.CashGameBlinds[1]
		if sb <= 0 {
			sb = 1
		}
		if bb <= 0 {
			bb = 2
		}
		structure = engine.CashGameBlinds(sb, bb)
		cfg.CashGameBlinds = [2]int{sb, bb}
		if cfg.CashGameBuyIn <= 0 {
			cfg.CashGameBuyIn = cfg.StartingStack * bb
		}
	}

	// Compute starting stack in chips from BB multiplier
	startChips := cfg.StartingStack
	if cfg.Mode == engine.ModeCashGame {
		startChips = cfg.CashGameBuyIn
	} else {
		bb := 2
		if len(structure.Levels) > 0 {
			bb = structure.Levels[0].BB
		}
		startChips = cfg.StartingStack * bb
	}

	// Select bot characters
	numBots := cfg.Seats - 1
	characters := ai.SelectCharacters(cfg.Difficulty, numBots, rng.Int63())

	// Create players: human is always PlayerID 1, seat 0
	players := make([]*engine.Player, 0, cfg.Seats)
	humanPlayer := &engine.Player{
		ID:        1,
		Name:      cfg.PlayerName,
		Stack:     startChips,
		Status:    engine.StatusActive,
		SeatIndex: 0,
	}
	players = append(players, humanPlayer)

	// Create bots
	bots := make(map[engine.PlayerID]*ai.Bot, numBots)
	botOrder := make([]engine.PlayerID, 0, numBots)
	for i, char := range characters {
		pid := engine.PlayerID(i + 2) // IDs 2, 3, 4, ...
		seatIdx := i + 1
		players = append(players, &engine.Player{
			ID:        pid,
			Name:      char.Profile.Name,
			Stack:     startChips,
			Status:    engine.StatusActive,
			SeatIndex: seatIdx,
		})
		bots[pid] = ai.NewBot(char, rng.Int63())
		botOrder = append(botOrder, pid)
	}

	// Create table
	table, err := engine.NewTable(cfg.Mode, cfg.Seats, structure, cfg.Seed, players)
	if err != nil {
		return nil, fmt.Errorf("creating table: %w", err)
	}

	sess := &Session{
		Config:     cfg,
		SessionID:  mustNewSessionID(),
		Table:      table,
		History:    &engine.SessionHistory{},
		Bots:       bots,
		HumanID:    1,
		Phase:      PhaseSetup,
		Updates:    make(chan Envelope, 1024),
		ActionResp: make(chan PlayerActionIntent),
		rng:        rng,
		botOrder:   botOrder,
		stop:       make(chan struct{}),
	}
	deps := sessionDependenciesProvider()
	startAnchor, err := deps.TimeAnchorProvider.Now()
	if err != nil {
		return nil, fmt.Errorf("creating session start time anchor: %w", err)
	}
	transcript, err := newTranscriptWriter(deps.ArtifactStore, deps.TimeAnchorProvider, sess.SessionID, cfg.PlayerName, modeString(cfg.Mode), startAnchor)
	if err != nil {
		return nil, fmt.Errorf("creating transcript writer: %w", err)
	}
	sess.deps = deps
	sess.transcript = transcript
	sess.metrics = newMetricsAccumulator(startAnchor)

	// Wire tournament/cash game manager
	switch cfg.Mode {
	case engine.ModeTournament, engine.ModeHeadsUpDuel:
		sess.Tournament = engine.NewTournament(table, startChips)
	case engine.ModeCashGame:
		sess.CashGame = engine.NewCashGame(table, startChips)
	}

	return sess, nil
}

// Run starts the game loop. Blocks until the session ends.
// The TUI should call this in a goroutine and listen on Updates.
func (s *Session) Run() {
	defer func() {
		if r := recover(); r != nil {
			s.emitEnvelope(Envelope{Error: &SessionError{Code: "session_error", Message: fmt.Sprintf("Session crashed: %v", r)}})
			_ = string(debug.Stack())
		}
		close(s.Updates)
	}()

	s.Phase = PhasePlaying
	if !s.resumed {
		if _, ok := s.emitEnvelope(Envelope{Notice: &Notice{Type: "session_started", Message: fmt.Sprintf("Welcome to the table, %s!", s.Config.PlayerName)}}); !ok {
			return
		}
	} else {
		s.resumed = false
	}
	if s.isStopped() {
		return
	}

	for {
		if s.isStopped() {
			return
		}
		hand := s.Table.NextHand()
		if hand == nil {
			break
		}
		s.currentHand = hand
		s.HandCount++

		// Emit hand_started so TUI resets state between hands.
		// The HandStartedEvent is recorded inside NewHand() but never
		// returned from any method, so we emit it explicitly here.
		if _, ok := s.emitEnvelope(Envelope{
			HandID: hand.ID,
			Notice: &Notice{
				Type:    "hand_started",
				Event:   engine.HandStartedEvent{HandID: hand.ID, DealerSeat: hand.DealerSeat, SBSeat: hand.SBSeat, BBSeat: hand.BBSeat, Blinds: hand.Blinds},
				Message: fmt.Sprintf("Hand #%d begins", hand.ID),
			},
		}); !ok {
			return
		}

		summary, ok := s.playHand(hand)
		if !ok || s.isStopped() {
			return
		}

		if s.Tournament == nil {
			s.Table.ApplyHandResults(hand)
		}

		// Record authoritative hand artifacts and metrics.
		if err := s.recordHand(hand); err != nil {
			s.emitEnvelope(Envelope{Error: &SessionError{Code: "session_error", Message: fmt.Sprintf("Unable to persist hand transcript: %v", err)}})
			return
		}

		// Emit hand summary
		s.emitEnvelope(Envelope{HandID: hand.ID, Notice: &Notice{Type: "hand_complete", Message: fmt.Sprintf("Hand #%d complete", hand.ID)}})
		s.emitHandSummary(summary)

		// Check session end conditions
		if summary.IsFinished {
			break
		}

		// Cool down all bots between hands
		for _, bot := range s.Bots {
			bot.CoolDown()
		}
	}

	if s.isStopped() {
		return
	}
	s.Phase = PhaseSessionOver
	s.emitSessionEnd()
}

// playHand runs a single hand from start to finish.
func (s *Session) playHand(hand *engine.Hand) (HandSummary, bool) {
	// Phase 1: Post blinds (auto-advance through init -> post blinds -> deal)
	for {
		if s.isStopped() {
			return HandSummary{}, false
		}
		step := hand.NextStep()
		switch step.Type {
		case engine.StepComplete:
			return s.buildHandSummary(hand), true
		case engine.StepAutoAdvance:
			events := hand.AdvanceStreet()
			s.emitEngineEvents(hand.ID, events)
		case engine.StepNeedAction:
			s.handleAction(hand, step.PlayerID)
		}

		if hand.Phase == engine.PhaseComplete {
			break
		}
	}

	return s.buildHandSummary(hand), true
}

// handleAction routes a needed action to either the human or a bot.
func (s *Session) handleAction(hand *engine.Hand, playerID engine.PlayerID) {
	if playerID == s.HumanID {
		s.handleHumanAction(hand, playerID)
	} else {
		s.handleBotAction(hand, playerID)
	}
}

// handleHumanAction sends the action request to the TUI and waits for a response.
func (s *Session) handleHumanAction(hand *engine.Hand, playerID engine.PlayerID) {
	// Skip if player has no legal actions (e.g., already all-in)
	preCheck := hand.LegalActions(playerID)
	if len(preCheck) == 0 {
		return
	}
	const maxRetries = 10
	for range maxRetries {
		view := hand.PlayerView(playerID)
		legal := hand.LegalActions(playerID)
		promptSeq, ok := s.emitEnvelope(Envelope{
			HandID: hand.ID,
			Prompt: &Prompt{
				PlayerID:     playerID,
				View:         view,
				LegalActions: legal,
				HandID:       hand.ID,
			},
			Notice: &Notice{
				Type:     "waiting_for_human",
				PlayerID: playerID,
				Message:  "Your turn",
			},
		})
		if !ok {
			return
		}

		// Wait for response from TUI
		var intent PlayerActionIntent
		select {
		case <-s.stop:
			return
		case intent = <-s.ActionResp:
		}
		if intent.PromptSeq != promptSeq || intent.HandID != hand.ID {
			s.emitEnvelope(Envelope{
				HandID: hand.ID,
				Prompt: &Prompt{PlayerID: playerID, View: view, LegalActions: legal, HandID: hand.ID},
				Error:  &SessionError{Code: "stale_action", Message: "That action expired. Use the current prompt."},
			})
			continue
		}
		action := intent.Action
		action.PlayerID = playerID

		events, err := hand.ApplyAction(playerID, action)
		if err != nil {
			s.emitEnvelope(Envelope{
				HandID: hand.ID,
				Prompt: &Prompt{PlayerID: playerID, View: view, LegalActions: legal, HandID: hand.ID},
				Error:  &SessionError{Code: "invalid_action", Message: humanizeActionError(action, legal, err)},
			})
			continue
		}

		s.emitEngineEvents(hand.ID, events)
		return
	}

	// Exhausted retries — force fold/check
	legal := hand.LegalActions(playerID)
	fallback := engine.Action{PlayerID: playerID, Type: engine.ActionFold}
	for _, la := range legal {
		if la.Type == engine.ActionCheck {
			fallback.Type = engine.ActionCheck
			break
		}
	}
	events, _ := hand.ApplyAction(playerID, fallback)
	s.emitEngineEvents(hand.ID, events)
}

// handleBotAction gets the bot's decision and applies it.
func (s *Session) handleBotAction(hand *engine.Hand, playerID engine.PlayerID) {
	bot, ok := s.Bots[playerID]
	if !ok {
		// Fallback: fold or check
		legal := hand.LegalActions(playerID)
		fallbackAction := engine.Action{PlayerID: playerID, Type: engine.ActionFold}
		for _, la := range legal {
			if la.Type == engine.ActionCheck {
				fallbackAction.Type = engine.ActionCheck
				break
			}
		}
		hand.ApplyAction(playerID, fallbackAction)
		return
	}

	view := hand.PlayerView(playerID)
	decision := bot.Decide(view)

	// Emit thinking event (TUI will animate this)
	s.emitEnvelope(Envelope{HandID: hand.ID, Notice: &Notice{Type: "bot_thinking", PlayerID: playerID, BotName: bot.Character.Profile.Name, ThinkTime: decision.Think, Reason: decision.Reason, Message: fmt.Sprintf("%s is thinking...", bot.Character.Profile.Name)}})

	events, err := hand.ApplyAction(playerID, decision.Action)
	if err != nil {
		// Bot made invalid action — fallback to fold/check
		legal := hand.LegalActions(playerID)
		fallback := engine.Action{PlayerID: playerID, Type: engine.ActionFold}
		for _, la := range legal {
			if la.Type == engine.ActionCheck {
				fallback.Type = engine.ActionCheck
				break
			}
		}
		events, _ = hand.ApplyAction(playerID, fallback)
	}

	s.emitEngineEvents(hand.ID, events)

	// Check if bot lost big (for tilt)
	player := playerByID(hand.Players, playerID)
	if player != nil {
		stackFraction := 1.0 - float64(player.Stack)/float64(s.startingChips())
		if stackFraction > 0.4 {
			bot.ObserveBigLoss(stackFraction)
		}
	}
}

// buildHandSummary creates the summary after a hand.
func (s *Session) buildHandSummary(hand *engine.Hand) HandSummary {
	summary := HandSummary{
		HandID:  hand.ID,
		Winners: hand.Winners,
	}

	// Handle tournament-specific post-hand logic
	if s.Tournament != nil {
		elims := s.Tournament.HandleEliminations(hand)
		summary.Eliminations = elims
		for _, e := range elims {
			s.emitEngineEvents(hand.ID, []engine.Event{e})
		}

		blindChange := s.Tournament.CheckBlindIncrease()
		if blindChange != nil {
			summary.BlindChange = blindChange
			s.emitEnvelope(Envelope{HandID: hand.ID, Notice: &Notice{Type: "blind_level_changed", Event: *blindChange, Message: fmt.Sprintf("Blinds increase to %d/%d (ante: %d)", blindChange.SB, blindChange.BB, blindChange.Ante)}})
		}
	}

	// Get human's stack
	humanPlayer := playerByID(s.Table.Players, s.HumanID)
	if humanPlayer != nil {
		summary.PlayerStack = humanPlayer.Stack
	}

	summary.IsFinished = s.Table.IsFinished()

	// Check if human was eliminated
	if humanPlayer != nil && humanPlayer.Status == engine.StatusOut {
		summary.IsFinished = true // End session if human is out
	}

	return summary
}

// recordHand saves the hand to history and transcript.
func (s *Session) recordHand(hand *engine.Hand) error {
	anchor, err := s.deps.TimeAnchorProvider.Now()
	if err != nil {
		return err
	}
	snapshots := make([]engine.PlayerSnapshot, 0, len(hand.Players))
	for _, p := range hand.Players {
		if p != nil {
			snapshots = append(snapshots, engine.PlayerSnapshot{
				ID:    p.ID,
				Name:  p.Name,
				Seat:  p.SeatIndex,
				Stack: p.Stack,
			})
		}
	}
	record := engine.HandRecord{
		HandID:     hand.ID,
		Seed:       hand.Seed(),
		Players:    snapshots,
		DealerSeat: hand.DealerSeat,
		Blinds:     hand.Blinds,
		Board:      append([]engine.Card(nil), hand.Board...),
		Actions:    append([]engine.Action(nil), hand.Actions...),
		Events:     append([]engine.Event(nil), hand.Events...),
		Timestamp:  anchor.Timestamp,
	}
	s.History.Add(record)
	refs, err := s.transcript.CommitHand(s, hand, record)
	if err != nil {
		return err
	}
	s.metrics.ObserveHand(hand, s.HumanID, refs)
	return nil
}

// emitEngineEvents wraps engine events as envelopes and sends them.
func (s *Session) emitEngineEvents(handID int, events []engine.Event) {
	for _, e := range events {
		notice := s.noticeForEngineEvent(e)
		s.emitEnvelope(Envelope{HandID: handID, Notice: &notice})
	}
}

// emitHandSummary sends the hand summary as an event.
func (s *Session) emitHandSummary(summary HandSummary) {
	s.emitEnvelope(Envelope{HandID: summary.HandID, Notice: &Notice{Type: "hand_summary", Message: fmt.Sprintf("Hand #%d - Your stack: %d chips", summary.HandID, summary.PlayerStack)}})
}

// emitSessionEnd sends the final session event.
func (s *Session) emitSessionEnd() {
	if err := s.persistSessionSummary(); err != nil {
		s.emitEnvelope(Envelope{Error: &SessionError{Code: "session_error", Message: fmt.Sprintf("Unable to persist session summary: %v", err)}})
		return
	}
	msg := "Session over."
	if s.Tournament != nil {
		results := s.Tournament.Results()
		for _, r := range results {
			if r.PlayerID == s.HumanID {
				msg = fmt.Sprintf("You finished in position #%d!", r.Position)
				break
			}
		}
		s.emitEnvelope(Envelope{Notice: &Notice{Type: "tournament_finished", Event: engine.TournamentFinishedEvent{Results: results}, Message: msg}})
	} else {
		humanPlayer := playerByID(s.Table.Players, s.HumanID)
		profit := 0
		if humanPlayer != nil {
			profit = humanPlayer.Stack - s.startingChips()
		}
		if profit >= 0 {
			msg = fmt.Sprintf("Session over. You walk away with %d chips profit!", profit)
		} else {
			msg = fmt.Sprintf("Session over. You lost %d chips.", -profit)
		}
		s.emitEnvelope(Envelope{Notice: &Notice{Type: "session_ended", Message: msg}})
	}
}

func (s *Session) startingChips() int {
	if s.Config.Mode == engine.ModeCashGame {
		return s.Config.CashGameBuyIn
	}
	bb := 2
	if len(s.Table.BlindsConfig.Levels) > 0 {
		bb = s.Table.BlindsConfig.Levels[0].BB
	}
	return s.Config.StartingStack * bb
}

// CurrentView returns the human's current view of the hand, if a hand is active.
func (s *Session) CurrentView() *engine.PlayerView {
	if s.currentHand == nil {
		return nil
	}
	view := s.currentHand.PlayerView(s.HumanID)
	return &view
}

// BotInfo returns the character info for a given player ID.
func (s *Session) BotInfo(pid engine.PlayerID) *ai.Character {
	bot, ok := s.Bots[pid]
	if !ok {
		return nil
	}
	return &bot.Character
}

// PlayerName returns the display name for a player ID.
func (s *Session) PlayerName(pid engine.PlayerID) string {
	if pid == s.HumanID {
		return s.Config.PlayerName
	}
	if bot, ok := s.Bots[pid]; ok {
		return bot.Character.Profile.Name
	}
	return fmt.Sprintf("Player %d", pid)
}

// PlayerNickname returns the bot nickname or empty for human.
func (s *Session) PlayerNickname(pid engine.PlayerID) string {
	if bot, ok := s.Bots[pid]; ok {
		return bot.Character.Profile.Nickname
	}
	return ""
}

// IsHuman returns whether the given player ID is the human.
func (s *Session) IsHuman(pid engine.PlayerID) bool {
	return pid == s.HumanID
}

// TableState returns a snapshot of all player stacks and statuses.
type TableState struct {
	Players    []PlayerInfo
	HandNum    int
	HandID     int
	Blinds     engine.BlindLevel
	Board      []engine.Card
	Pot        int
	Street     engine.Street
	DealerSeat int
	HumanCards [2]engine.Card
}

type PlayerInfo struct {
	ID       engine.PlayerID
	Name     string
	Nickname string
	Stack    int
	Bet      int
	Status   engine.PlayerStatus
	Seat     int
	IsHuman  bool
}

func (s *Session) TableState() TableState {
	return s.snapshot()
}

func (s *Session) snapshot() TableState {
	ts := TableState{
		HandNum: s.HandCount,
		Blinds:  s.Table.CurrentBlinds(),
	}
	players := s.Table.Players
	if s.currentHand != nil {
		ts.HandID = s.currentHand.ID
		ts.Board = append([]engine.Card(nil), s.currentHand.Board...)
		ts.Street = s.currentHand.Street
		ts.DealerSeat = s.currentHand.DealerSeat
		view := s.currentHand.PlayerView(s.HumanID)
		ts.Pot = view.Pot
		ts.HumanCards = view.MyCards
		if s.currentHand.Phase != engine.PhaseComplete {
			players = s.currentHand.Players
		}
	}
	for _, p := range players {
		if p == nil {
			continue
		}
		pi := PlayerInfo{
			ID:      p.ID,
			Name:    p.Name,
			Stack:   p.Stack,
			Bet:     p.Bet,
			Status:  p.Status,
			Seat:    p.SeatIndex,
			IsHuman: p.ID == s.HumanID,
		}
		if bot, ok := s.Bots[p.ID]; ok {
			pi.Nickname = bot.Character.Profile.Nickname
		}
		ts.Players = append(ts.Players, pi)
	}
	return ts
}

func (s *Session) emitEnvelope(env Envelope) (uint64, bool) {
	if env.Snapshot.Players == nil {
		env.Snapshot = s.snapshot()
	}
	s.seq++
	env.Seq = s.seq
	env.SessionID = s.SessionID
	if env.HandID == 0 {
		env.HandID = env.Snapshot.HandID
	}
	if env.Prompt != nil {
		env.Prompt.Seq = env.Seq
		if env.Prompt.HandID == 0 {
			env.Prompt.HandID = env.HandID
		}
	}
	select {
	case <-s.stop:
		return 0, false
	case s.Updates <- env:
		return env.Seq, true
	}
}

func (s *Session) noticeForEngineEvent(event engine.Event) Notice {
	notice := Notice{Type: event.EventType(), Event: event, Message: event.EventType()}
	switch e := event.(type) {
	case engine.BlindsPostedEvent:
		blind := "SB"
		if e.Type == engine.BlindBig {
			blind = "BB"
		} else if e.Type == engine.BlindAnte {
			blind = "ante"
		}
		notice.PlayerID = e.PlayerID
		notice.Message = fmt.Sprintf("%s posts %s %d", s.PlayerName(e.PlayerID), blind, e.Amount)
	case engine.HoleCardsDealtEvent:
		notice.PlayerID = e.PlayerID
		notice.Message = fmt.Sprintf("%s is dealt hole cards", s.PlayerName(e.PlayerID))
	case engine.ActionTakenEvent:
		notice.PlayerID = e.PlayerID
		notice.Message = formatActionNotice(s.PlayerName(e.PlayerID), e.Action)
		if bot, ok := s.Bots[e.PlayerID]; ok {
			notice.BotName = bot.Character.Profile.Name
		}
	case engine.StreetAdvancedEvent:
		notice.Message = streetAdvanceMessage(e)
	case engine.ShowdownStartedEvent:
		notice.Message = "Showdown"
	case engine.HandRevealedEvent:
		notice.PlayerID = e.PlayerID
		notice.Message = fmt.Sprintf("%s shows %s", s.PlayerName(e.PlayerID), e.Eval.Name)
	case engine.PotAwardedEvent:
		notice.Message = formatPotAwardMessage(s, e)
	case engine.PlayerEliminatedEvent:
		notice.PlayerID = e.PlayerID
		notice.Message = fmt.Sprintf("%s eliminated (#%d)", s.PlayerName(e.PlayerID), e.Position)
	case engine.BlindLevelChangedEvent:
		notice.Message = fmt.Sprintf("Blinds increase to %d/%d (ante: %d)", e.SB, e.BB, e.Ante)
	case engine.TournamentFinishedEvent:
		notice.Message = "Tournament finished"
	}
	return notice
}

func formatActionNotice(name string, action engine.Action) string {
	label := actionLabel(action.Type)
	if action.Type == engine.ActionCheck {
		label = "checks"
	}
	if action.Type == engine.ActionFold {
		label = "folds"
	}
	if action.Type == engine.ActionCall {
		label = "calls"
	}
	if action.Type == engine.ActionBet {
		label = "bets"
	}
	if action.Type == engine.ActionRaise {
		label = "raises"
	}
	if action.Type == engine.ActionAllIn {
		label = "moves all-in"
	}
	if action.Amount > 0 {
		return fmt.Sprintf("%s %s %d", name, label, action.Amount)
	}
	return fmt.Sprintf("%s %s", name, label)
}

func streetAdvanceMessage(event engine.StreetAdvancedEvent) string {
	switch event.Street {
	case engine.StreetFlop:
		return "Flop"
	case engine.StreetTurn:
		return "Turn"
	case engine.StreetRiver:
		return "River"
	default:
		return "Next street"
	}
}

func formatPotAwardMessage(s *Session, event engine.PotAwardedEvent) string {
	names := make([]string, 0, len(event.Winners))
	for _, playerID := range event.Winners {
		names = append(names, s.PlayerName(playerID))
	}
	return fmt.Sprintf("%s wins %d", joinNames(names), event.Amount)
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return "No one"
	}
	if len(names) == 1 {
		return names[0]
	}
	return fmt.Sprintf("%s and %s", strings.Join(names[:len(names)-1], ", "), names[len(names)-1])
}

func mustNewSessionID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("ses_%032x", time.Now().UnixNano())
	}
	return "ses_" + hex.EncodeToString(raw[:])
}

func (s *Session) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

func (s *Session) isStopped() bool {
	select {
	case <-s.stop:
		return true
	default:
		return false
	}
}

// humanizeActionError converts a raw engine error into a user-friendly message.
func humanizeActionError(action engine.Action, legal []engine.LegalAction, _ error) string {
	// Check if fold was attempted when check is free
	if action.Type == engine.ActionFold {
		for _, la := range legal {
			if la.Type == engine.ActionCheck {
				return "You can't fold — checking is free."
			}
		}
	}

	// Check for below-minimum raise/bet
	if action.Type == engine.ActionRaise || action.Type == engine.ActionBet {
		for _, la := range legal {
			if la.Type == action.Type && action.Amount < la.MinAmount {
				return fmt.Sprintf("Minimum %s is %d.", actionLabel(action.Type), la.MinAmount)
			}
		}
	}

	return "That action is not available right now."
}

func actionLabel(t engine.ActionType) string {
	switch t {
	case engine.ActionBet:
		return "bet"
	case engine.ActionRaise:
		return "raise"
	default:
		return "amount"
	}
}

// helper
func playerByID(players []*engine.Player, id engine.PlayerID) *engine.Player {
	for _, p := range players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}
