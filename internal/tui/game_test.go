package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
)

func updateGameFromEvent(t *testing.T, m GameModel, ev session.SessionEvent) GameModel {
	t.Helper()
	model, _ := m.handleSessionEvent(ev)
	next, ok := model.(GameModel)
	if !ok {
		t.Fatalf("expected GameModel, got %T", model)
	}
	return next
}

func TestGameHandleActionReqDoesNotReturnPointerModel(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	m := NewGameModel(sess, true)
	req := session.ActionRequest{
		View: engine.PlayerView{
			MyCards:    [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts)},
			MyStack:    200,
			MyBet:      0,
			Pot:        3,
			CurrentBet: 2,
			Street:     engine.StreetPreflop,
		},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionCall, MinAmount: 2, MaxAmount: 2}},
		Snapshot: session.TableState{
			HandNum: 1,
			Blinds:  engine.BlindLevel{SB: 1, BB: 2},
			Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 200, IsHuman: true}},
		},
	}

	model, _ := m.handleActionReq(req)
	if _, ok := model.(GameModel); !ok {
		t.Fatalf("expected GameModel, got %T", model)
	}
}

func TestGameActionTakenSoundsRespectMapping(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := make([]audio.SoundType, 0, 3)
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: sess.HumanID, Action: engine.Action{Type: engine.ActionCheck}, PotTotal: 3}})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionRaise}, PotTotal: 6}})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionCall}, PotTotal: 6}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionAllIn}, PotTotal: 20}})

	if len(plays) != 3 {
		t.Fatalf("len(plays) = %d, want 3", len(plays))
	}
	if plays[0] != audio.SoundCheck {
		t.Fatalf("first sound = %v, want SoundCheck", plays[0])
	}
	if plays[1] != audio.SoundOpponentPressure {
		t.Fatalf("second sound = %v, want SoundOpponentPressure", plays[1])
	}
	if plays[2] != audio.SoundAllIn {
		t.Fatalf("third sound = %v, want SoundAllIn", plays[2])
	}
}

func TestGameWaitingForHumanPlaysOnce(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	count := 0
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) {
		if sound == audio.SoundYourTurn {
			count++
		}
	}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "waiting_for_human", Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}}}})
	_, _ = m.handleActionReq(session.ActionRequest{Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}}}})

	if count != 1 {
		t.Fatalf("your turn sound count = %d, want 1", count)
	}
}

func TestGameHoleCardsDealPlaysOnceForHuman(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	count := 0
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) {
		if sound == audio.SoundHoleCards {
			count++
		}
	}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "hand_started", Event: engine.HandStartedEvent{}, Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: sess.HumanID, Name: "Hero", Stack: 100, IsHuman: true}}}})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "hole_cards_dealt", Event: engine.HoleCardsDealtEvent{PlayerID: sess.HumanID}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "hole_cards_dealt", Event: engine.HoleCardsDealtEvent{PlayerID: sess.HumanID}})

	if count != 1 {
		t.Fatalf("hole cards sound count = %d, want 1", count)
	}
}

func TestGameStreetAdvanceDifferentiatesFlopAndTurnRiver(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := []audio.SoundType{}
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "street_advanced", Event: engine.StreetAdvancedEvent{Street: engine.StreetFlop, NewCards: []engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Diamonds)}}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "street_advanced", Event: engine.StreetAdvancedEvent{Street: engine.StreetTurn, NewCards: []engine.Card{engine.NewCard(engine.Jack, engine.Clubs)}}})

	if len(plays) != 2 {
		t.Fatalf("len(plays) = %d, want 2", len(plays))
	}
	if plays[0] != audio.SoundFlop {
		t.Fatalf("first street sound = %v, want SoundFlop", plays[0])
	}
	if plays[1] != audio.SoundTurnRiver {
		t.Fatalf("second street sound = %v, want SoundTurnRiver", plays[1])
	}
}

func TestGameShowdownAndBustoutAndEndUsePremiumCues(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := []audio.SoundType{}
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "showdown_started"})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "player_eliminated", Event: engine.PlayerEliminatedEvent{PlayerID: sess.HumanID, Position: 2}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "tournament_finished", Event: engine.TournamentFinishedEvent{Results: []engine.TournamentResult{{PlayerID: sess.HumanID, Position: 1, Name: "Hero"}}}})

	if len(plays) != 3 {
		t.Fatalf("len(plays) = %d, want 3", len(plays))
	}
	if plays[0] != audio.SoundShowdown {
		t.Fatalf("first sound = %v, want SoundShowdown", plays[0])
	}
	if plays[1] != audio.SoundBustout {
		t.Fatalf("second sound = %v, want SoundBustout", plays[1])
	}
	if plays[2] != audio.SoundVictory {
		t.Fatalf("third sound = %v, want SoundVictory", plays[2])
	}
}

func TestGameCashSessionEndCanUseDefeatCue(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:           engine.ModeCashGame,
		Difficulty:     ai.DifficultyMedium,
		Seats:          6,
		StartingStack:  100,
		CashGameBuyIn:  1000,
		CashGameBlinds: [2]int{5, 10},
		PlayerName:     "Hero",
		Seed:           42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := []audio.SoundType{}
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m.players = []session.PlayerInfo{{ID: sess.HumanID, Name: "Hero", Stack: 800, IsHuman: true}}
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "session_ended", Message: "Session over."})

	if len(plays) != 1 {
		t.Fatalf("len(plays) = %d, want 1", len(plays))
	}
	if plays[0] != audio.SoundDefeat {
		t.Fatalf("sound = %v, want SoundDefeat", plays[0])
	}
}

func TestHandleActionReqClearsThinkingMessage(t *testing.T) {
	// Fix #3: When entering action mode, stale "thinking..." messages
	// from bot_thinking events should be cleared.
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	// Suppress sounds
	prev := playGameSound
	playGameSound = func(audio.SoundType) {}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)

	// Simulate a bot_thinking event setting lastAction
	m = updateGameFromEvent(t, m, session.SessionEvent{
		Type:      "bot_thinking",
		BotName:   "Shark",
		ThinkTime: 500,
	})
	if m.lastAction != "Shark is thinking..." {
		t.Fatalf("lastAction = %q, want %q", m.lastAction, "Shark is thinking...")
	}

	// Now simulate an action request arriving
	req := session.ActionRequest{
		View: engine.PlayerView{
			MyCards: [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts)},
			MyStack: 200,
			MyBet:   0,
			Pot:     3,
			Street:  engine.StreetPreflop,
		},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionCall, MinAmount: 2, MaxAmount: 2}},
		Snapshot: session.TableState{
			HandNum: 1,
			Blinds:  engine.BlindLevel{SB: 1, BB: 2},
			Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 200, IsHuman: true}},
		},
	}

	model, _ := m.handleActionReq(req)
	updated, ok := model.(GameModel)
	if !ok {
		t.Fatalf("expected GameModel, got %T", model)
	}

	if updated.lastAction != "" {
		t.Errorf("lastAction should be cleared after handleActionReq, got %q", updated.lastAction)
	}
}

func TestHandleActionReqKeepsNonThinkingMessage(t *testing.T) {
	// Verify that non-thinking lastAction messages are preserved.
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	prev := playGameSound
	playGameSound = func(audio.SoundType) {}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m.lastAction = "Shark raises 20"

	req := session.ActionRequest{
		View:         engine.PlayerView{MyStack: 200},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}},
		Snapshot: session.TableState{
			Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 200, IsHuman: true}},
		},
	}

	model, _ := m.handleActionReq(req)
	updated := model.(GameModel)

	if updated.lastAction != "Shark raises 20" {
		t.Errorf("lastAction = %q, want %q (non-thinking messages should be preserved)", updated.lastAction, "Shark raises 20")
	}
}

func TestHandStartedResetsGameState(t *testing.T) {
	// Fix #1 (TUI side): hand_started event should reset board, pot,
	// showdown state, etc.
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	prev := playGameSound
	playGameSound = func(audio.SoundType) {}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)

	// Simulate stale state from previous hand
	m.pot = 150
	m.board = []engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Diamonds)}
	m.showdown = true
	m.revealed = []revealedHand{{playerID: 2, name: "Bot"}}
	m.potAwards = []string{"Bot wins 150"}
	m.myBet = 50
	m.lastAction = "Bot raises 100"

	// Process hand_started event
	m = updateGameFromEvent(t, m, session.SessionEvent{
		Type: "hand_started",
		Event: engine.HandStartedEvent{
			HandID:     2,
			DealerSeat: 1,
			Blinds:     engine.BlindLevel{SB: 1, BB: 2},
		},
		Snapshot: session.TableState{
			HandNum: 2,
			Blinds:  engine.BlindLevel{SB: 1, BB: 2},
			Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}},
		},
	})

	if m.pot != 0 {
		t.Errorf("pot = %d, want 0 after hand_started", m.pot)
	}
	if m.board != nil {
		t.Errorf("board = %v, want nil after hand_started", m.board)
	}
	if m.showdown {
		t.Error("showdown should be false after hand_started")
	}
	if m.revealed != nil {
		t.Error("revealed should be nil after hand_started")
	}
	if m.potAwards != nil {
		t.Error("potAwards should be nil after hand_started")
	}
	if m.myBet != 0 {
		t.Errorf("myBet = %d, want 0 after hand_started", m.myBet)
	}
	if m.lastAction != "" {
		t.Errorf("lastAction = %q, want empty after hand_started", m.lastAction)
	}
	if m.dealerSeat != 1 {
		t.Errorf("dealerSeat = %d, want 1 after hand_started", m.dealerSeat)
	}
}

func TestRenderSeatFitsWithinWidth(t *testing.T) {
	// Fix #5: Each line of a rendered seat must fit within the
	// 22-char total width (20 content + 2 padding).
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	prev := playGameSound
	playGameSound = func(audio.SoundType) {}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m.dealerSeat = 1

	cases := []struct {
		name string
		info session.PlayerInfo
	}{
		{
			name: "active with bet and dealer",
			info: session.PlayerInfo{ID: 2, Name: "LongPlayerName", Stack: 99999, Bet: 5000, Status: engine.StatusActive, Seat: 1},
		},
		{
			name: "all-in",
			info: session.PlayerInfo{ID: 3, Name: "AllInPlayer", Stack: 0, Bet: 15000, Status: engine.StatusAllIn, Seat: 2},
		},
		{
			name: "folded",
			info: session.PlayerInfo{ID: 4, Name: "FoldedGuy", Stack: 500, Status: engine.StatusFolded, Seat: 3},
		},
		{
			name: "very long name with dealer",
			info: session.PlayerInfo{ID: 5, Name: "AVeryVeryLongNameThatOverflows", Stack: 12345, Bet: 678, Status: engine.StatusActive, Seat: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rendered := m.renderSeat(tc.info)
			width := lipgloss.Width(rendered)
			if width > 22 {
				t.Errorf("renderSeat width = %d, want <= 22\nrendered:\n%s", width, rendered)
			}
		})
	}
}

func TestWaitForSessionPrioritizesBufferedEventsOverActionRequests(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         3,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	m := NewGameModel(sess, true)
	sess.Events <- session.SessionEvent{Type: "action_taken", Message: "buffered"}

	actionSent := make(chan struct{})
	go func() {
		sess.ActionReq <- session.ActionRequest{}
		close(actionSent)
	}()

	msg := m.waitForSession()()
	if ev, ok := msg.(sessionEventMsg); !ok || session.SessionEvent(ev).Type != "action_taken" {
		t.Fatalf("expected buffered session event first, got %T (%v)", msg, msg)
	}

	msg = m.waitForSession()()
	if _, ok := msg.(actionReqMsg); !ok {
		t.Fatalf("expected action request after draining buffered event, got %T", msg)
	}

	select {
	case <-actionSent:
	case <-time.After(2 * time.Second):
		t.Fatal("expected action request sender to complete after the second receive")
	}

	sess.Stop()
}
