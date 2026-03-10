package session

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestNewSession(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "TestPlayer",
		Seed:          42,
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if sess.HumanID != 1 {
		t.Errorf("HumanID = %d, want 1", sess.HumanID)
	}
	if len(sess.Bots) != 5 {
		t.Errorf("len(Bots) = %d, want 5", len(sess.Bots))
	}
	if len(sess.Table.Players) != 6 {
		t.Errorf("len(Players) = %d, want 6", len(sess.Table.Players))
	}
	if sess.Phase != PhaseSetup {
		t.Errorf("Phase = %d, want PhaseSetup", sess.Phase)
	}

	// Human is seat 0
	human := playerByID(sess.Table.Players, 1)
	if human == nil {
		t.Fatal("human player not found")
	}
	if human.SeatIndex != 0 {
		t.Errorf("human.SeatIndex = %d, want 0", human.SeatIndex)
	}
	if human.Name != "TestPlayer" {
		t.Errorf("human.Name = %q, want %q", human.Name, "TestPlayer")
	}

	// All players have same starting stack
	startChips := sess.startingChips()
	for _, p := range sess.Table.Players {
		if p.Stack != startChips {
			t.Errorf("player %d stack = %d, want %d", p.ID, p.Stack, startChips)
		}
	}
}

func TestNewSession_HeadsUp(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeHeadsUpDuel,
		Difficulty:    ai.DifficultyMedium,
		Seats:         2,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          123,
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if len(sess.Bots) != 1 {
		t.Errorf("len(Bots) = %d, want 1", len(sess.Bots))
	}
	if len(sess.Table.Players) != 2 {
		t.Errorf("len(Players) = %d, want 2", len(sess.Table.Players))
	}
	if sess.Tournament == nil {
		t.Error("Tournament should be set for HeadsUp mode")
	}
}

func TestNewSession_CashGame(t *testing.T) {
	cfg := Config{
		Mode:           engine.ModeCashGame,
		Difficulty:     ai.DifficultyEasy,
		Seats:          6,
		StartingStack:  100,
		PlayerName:     "Cash",
		Seed:           456,
		CashGameBuyIn:  1000,
		CashGameBlinds: [2]int{5, 10},
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if sess.CashGame == nil {
		t.Fatal("CashGame should be set")
	}
	if sess.Tournament != nil {
		t.Error("Tournament should not be set for CashGame mode")
	}

	// Check all players have cash game buy-in
	for _, p := range sess.Table.Players {
		if p.Stack != 1000 {
			t.Errorf("player %d stack = %d, want 1000", p.ID, p.Stack)
		}
	}
}

func TestNewSession_InvalidSeats(t *testing.T) {
	cfg := Config{Mode: engine.ModeTournament, Seats: 1}
	_, err := New(cfg)
	if err == nil {
		t.Error("expected error for seats < 2")
	}

	cfg.Seats = 10
	_, err = New(cfg)
	if err == nil {
		t.Error("expected error for seats > 9")
	}
}

func TestSessionRun_SmallTournament(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyEasy,
		Seats:         3,
		StartingStack: 50,
		PlayerName:    "AutoPlayer",
		Seed:          999,
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Run session in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		sess.Run()
	}()

	// Automated play: always pick first legal action (fold/check)
	eventsDrained := make(chan struct{})
	go func() {
		defer close(eventsDrained)
		for range sess.Events {
			// drain events
		}
	}()

	// Handle action requests: auto-play
	go func() {
		for req := range sess.ActionReq {
			// Choose the safest action: check > call > fold
			var chosen engine.Action
			chosen.PlayerID = sess.HumanID
			for _, la := range req.LegalActions {
				if la.Type == engine.ActionCheck {
					chosen.Type = engine.ActionCheck
					break
				}
				if la.Type == engine.ActionCall {
					chosen.Type = engine.ActionCall
					break
				}
				if la.Type == engine.ActionFold {
					chosen.Type = engine.ActionFold
				}
			}
			if chosen.Type == 0 && len(req.LegalActions) > 0 {
				chosen.Type = req.LegalActions[0].Type
				chosen.Amount = req.LegalActions[0].MinAmount
			}
			sess.ActionResp <- chosen
		}
	}()

	<-done
	<-eventsDrained

	if sess.Phase != PhaseSessionOver {
		t.Errorf("Phase = %d, want PhaseSessionOver", sess.Phase)
	}
	if sess.HandCount == 0 {
		t.Error("expected at least one hand to be played")
	}
	if len(sess.History.Records) == 0 {
		t.Error("expected hand history to be recorded")
	}

	t.Logf("Tournament ended after %d hands", sess.HandCount)
}

func TestSessionRun_DeterministicSeeds(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeHeadsUpDuel,
		Difficulty:    ai.DifficultyEasy,
		Seats:         2,
		StartingStack: 20,
		PlayerName:    "Seed",
		Seed:          777,
	}

	// Run twice, compare hand count (deterministic with same decisions)
	counts := [2]int{}
	for run := range 2 {
		sess, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}

		done := make(chan struct{})
		drained := make(chan struct{})
		go func() {
			defer close(done)
			sess.Run()
		}()

		go func() {
			defer close(drained)
			for range sess.Events {
			}
		}()

		actionsDone := make(chan struct{})
		go func() {
			defer close(actionsDone)
			for req := range sess.ActionReq {
				// Pick best available: check > call > fold
				action := engine.Action{PlayerID: sess.HumanID}
				picked := false
				for _, la := range req.LegalActions {
					if la.Type == engine.ActionCheck {
						action.Type = engine.ActionCheck
						picked = true
						break
					}
				}
				if !picked {
					for _, la := range req.LegalActions {
						if la.Type == engine.ActionFold {
							action.Type = engine.ActionFold
							picked = true
							break
						}
					}
				}
				if !picked && len(req.LegalActions) > 0 {
					action.Type = req.LegalActions[0].Type
					action.Amount = req.LegalActions[0].MinAmount
				}
				sess.ActionResp <- action
			}
		}()

		<-done
		<-drained
		<-actionsDone
		counts[run] = sess.HandCount
		t.Logf("Run %d: %d hands", run, counts[run])
	}

	if counts[0] != counts[1] {
		t.Errorf("non-deterministic: run1=%d hands, run2=%d hands", counts[0], counts[1])
	}
	t.Logf("Both runs: %d hands (deterministic)", counts[0])
}

func TestPlayerName(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         3,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	}
	sess, _ := New(cfg)

	if sess.PlayerName(sess.HumanID) != "Hero" {
		t.Errorf("PlayerName(human) = %q, want %q", sess.PlayerName(sess.HumanID), "Hero")
	}

	// Bot names should be non-empty
	for pid := range sess.Bots {
		name := sess.PlayerName(pid)
		if name == "" {
			t.Errorf("bot %d has empty name", pid)
		}
	}
}

func TestHandStartedEventEmitted(t *testing.T) {
	// Verify that "hand_started" is emitted to the Events channel and
	// arrives before blind_posted for the first hand.
	cfg := Config{
		Mode:          engine.ModeHeadsUpDuel,
		Difficulty:    ai.DifficultyEasy,
		Seats:         2,
		StartingStack: 20,
		PlayerName:    "TestHSE",
		Seed:          12345,
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		sess.Run()
	}()

	// Auto-respond to actions in background
	go func() {
		for req := range sess.ActionReq {
			action := engine.Action{PlayerID: sess.HumanID, Type: engine.ActionFold}
			for _, la := range req.LegalActions {
				if la.Type == engine.ActionCheck {
					action.Type = engine.ActionCheck
					break
				}
			}
			sess.ActionResp <- action
		}
	}()

	// Track first hand's event ordering
	foundHandStarted := false
	blindBeforeHandStarted := false
	eventsDrained := make(chan struct{})
	go func() {
		defer close(eventsDrained)
		firstHandChecked := false
		for ev := range sess.Events {
			if firstHandChecked {
				continue // drain rest
			}
			switch ev.Type {
			case "session_started":
				continue
			case "hand_started":
				foundHandStarted = true
				firstHandChecked = true // only check first hand
			case "blind_posted":
				if !foundHandStarted {
					blindBeforeHandStarted = true
				}
				firstHandChecked = true
			}
		}
	}()

	<-done
	<-eventsDrained

	if !foundHandStarted {
		t.Fatal("hand_started event was never emitted")
	}
	if blindBeforeHandStarted {
		t.Error("blind_posted arrived before hand_started for the first hand")
	}
}

func TestHandStartedEventContainsCorrectData(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeHeadsUpDuel,
		Difficulty:    ai.DifficultyEasy,
		Seats:         2,
		StartingStack: 20,
		PlayerName:    "HSData",
		Seed:          54321,
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		sess.Run()
	}()

	go func() {
		for req := range sess.ActionReq {
			action := engine.Action{PlayerID: sess.HumanID, Type: engine.ActionFold}
			for _, la := range req.LegalActions {
				if la.Type == engine.ActionCheck {
					action.Type = engine.ActionCheck
					break
				}
			}
			sess.ActionResp <- action
		}
	}()

	var handStartedEvent engine.HandStartedEvent
	found := false
	eventsDrained := make(chan struct{})
	go func() {
		defer close(eventsDrained)
		for ev := range sess.Events {
			if ev.Type == "hand_started" && !found {
				if e, ok := ev.Event.(engine.HandStartedEvent); ok {
					handStartedEvent = e
					found = true
				}
			}
		}
	}()

	<-done
	<-eventsDrained

	if !found {
		t.Fatal("hand_started event not found")
	}
	if handStartedEvent.Blinds.SB != 1 || handStartedEvent.Blinds.BB != 2 {
		t.Errorf("blinds = %d/%d, want 1/2", handStartedEvent.Blinds.SB, handStartedEvent.Blinds.BB)
	}
	if handStartedEvent.HandID != 1 {
		t.Errorf("HandID = %d, want 1", handStartedEvent.HandID)
	}
}

func TestTableState(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         4,
		StartingStack: 100,
		PlayerName:    "Test",
		Seed:          42,
	}
	sess, _ := New(cfg)

	ts := sess.TableState()
	if len(ts.Players) != 4 {
		t.Errorf("len(Players) = %d, want 4", len(ts.Players))
	}

	humanFound := false
	for _, pi := range ts.Players {
		if pi.IsHuman {
			humanFound = true
			if pi.Name != "Test" {
				t.Errorf("human name = %q, want %q", pi.Name, "Test")
			}
		}
	}
	if !humanFound {
		t.Error("human player not found in TableState")
	}
}

func TestSnapshotUsesCurrentHandStateWhileHandActive(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         3,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	}
	sess, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	hand := sess.Table.NextHand()
	if hand == nil {
		t.Fatal("expected hand")
	}
	sess.currentHand = hand
	handPlayer := playerByID(hand.Players, sess.HumanID)
	tablePlayer := playerByID(sess.Table.Players, sess.HumanID)
	if handPlayer == nil || tablePlayer == nil {
		t.Fatal("expected both hand and table players")
	}

	handPlayer.Stack = 123
	handPlayer.Bet = 7

	ts := sess.TableState()
	human := playerInfoByID(ts.Players, sess.HumanID)
	if human == nil {
		t.Fatal("expected human in table state")
	}
	if human.Stack != 123 {
		t.Fatalf("snapshot stack = %d, want 123 from current hand", human.Stack)
	}
	if human.Bet != 7 {
		t.Fatalf("snapshot bet = %d, want 7 from current hand", human.Bet)
	}
	if tablePlayer.Stack == 123 {
		t.Fatal("table player should not have been mutated by hand-only change")
	}
}

func playerInfoByID(players []PlayerInfo, id engine.PlayerID) *PlayerInfo {
	for i := range players {
		if players[i].ID == id {
			return &players[i]
		}
	}
	return nil
}
