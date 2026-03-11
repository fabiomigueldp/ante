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
	if sess.SessionID == "" {
		t.Fatal("expected non-empty SessionID")
	}

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
	for _, p := range sess.Table.Players {
		if p.Stack != 1000 {
			t.Errorf("player %d stack = %d, want 1000", p.ID, p.Stack)
		}
	}
}

func TestNewSession_InvalidSeats(t *testing.T) {
	cfg := Config{Mode: engine.ModeTournament, Seats: 1}
	if _, err := New(cfg); err == nil {
		t.Error("expected error for seats < 2")
	}

	cfg.Seats = 10
	if _, err := New(cfg); err == nil {
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

	runSessionWithAutoPlay(t, sess, defaultActionForPrompt, nil)

	if sess.Phase != PhaseSessionOver {
		t.Errorf("Phase = %d, want PhaseSessionOver", sess.Phase)
	}
	if sess.HandCount == 0 {
		t.Error("expected at least one hand to be played")
	}
	if len(sess.History.Records) == 0 {
		t.Error("expected hand history to be recorded")
	}
}

func TestSessionStopsAtBetweenHandsUntilHumanReady(t *testing.T) {
	store, _, _ := newSessionTestStore(t)
	useSessionDependenciesForTest(t, Dependencies{ArtifactStore: store, TimeAnchorProvider: store.TimeAnchorProvider()})
	sess, err := New(Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyEasy,
		Seats:         3,
		StartingStack: 50,
		PlayerName:    "Hero",
		Seed:          77,
	})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		sess.Run()
	}()
	seenWaitingReady := false
	seenNextHand := false
	for env := range sess.Updates {
		if env.Prompt != nil && env.Prompt.Kind == PromptKindAction {
			select {
			case sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Action: defaultActionForPrompt(env.Prompt)}:
			case <-sess.stop:
			}
			continue
		}
		if env.Prompt != nil && env.Prompt.Kind == PromptKindBetweenHands {
			seenWaitingReady = true
			if sess.Phase != PhaseWaitingReady {
				t.Fatalf("phase = %d, want PhaseWaitingReady", sess.Phase)
			}
			if !sess.CanSave() {
				t.Fatal("expected CanSave to be true while waiting for ready")
			}
			if sess.readyState == nil {
				t.Fatal("expected readyState to be populated")
			}
			if env.Snapshot.HandID != sess.readyState.Snapshot.HandID {
				t.Fatalf("boundary snapshot hand id = %d, want %d", env.Snapshot.HandID, sess.readyState.Snapshot.HandID)
			}
			if env.Snapshot.Pot != sess.readyState.Snapshot.Pot {
				t.Fatalf("boundary snapshot pot = %d, want %d", env.Snapshot.Pot, sess.readyState.Snapshot.Pot)
			}
			if env.Snapshot.Street != sess.readyState.Snapshot.Street {
				t.Fatalf("boundary snapshot street = %d, want %d", env.Snapshot.Street, sess.readyState.Snapshot.Street)
			}
			if !equalCards(env.Snapshot.Board, sess.readyState.Snapshot.Board) {
				t.Fatalf("boundary snapshot board = %+v, want %+v", env.Snapshot.Board, sess.readyState.Snapshot.Board)
			}
			if env.Snapshot.HumanCards != sess.readyState.Snapshot.HumanCards {
				t.Fatalf("boundary human cards = %+v, want %+v", env.Snapshot.HumanCards, sess.readyState.Snapshot.HumanCards)
			}
			sess.Stop()
			continue
		}
		if env.Notice != nil && env.Notice.Type == "hand_started" && env.HandID > 1 {
			seenNextHand = true
			sess.Stop()
		}
	}
	<-done
	if !seenWaitingReady {
		t.Fatal("expected session to stop at between-hands boundary")
	}
	if seenNextHand {
		t.Fatal("next hand should not start before the explicit ready signal is processed")
	}
}

func TestSessionLeaveTableEndsCleanlyFromBetweenHands(t *testing.T) {
	store, _, _ := newSessionTestStore(t)
	useSessionDependenciesForTest(t, Dependencies{ArtifactStore: store, TimeAnchorProvider: store.TimeAnchorProvider()})
	sess, err := New(Config{
		Mode:          engine.ModeCashGame,
		Difficulty:    ai.DifficultyEasy,
		Seats:         3,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          88,
		CashGameBuyIn: 200,
	})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		sess.Run()
	}()
	seenBoundary := false
	seenSessionEnd := false
	for env := range sess.Updates {
		if env.Prompt != nil && env.Prompt.Kind == PromptKindAction {
			select {
			case sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Action: defaultActionForPrompt(env.Prompt)}:
			case <-sess.stop:
			}
			continue
		}
		if env.Prompt != nil && env.Prompt.Kind == PromptKindBetweenHands {
			seenBoundary = true
			select {
			case sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Control: ControlIntent{Kind: ControlIntentLeaveTable}}:
			case <-sess.stop:
			}
			continue
		}
		if env.Notice != nil && env.Notice.Type == "session_ended" {
			seenSessionEnd = true
		}
	}
	<-done
	if !seenBoundary {
		t.Fatal("expected between-hands prompt before leaving the table")
	}
	if !seenSessionEnd {
		t.Fatal("expected clean session_ended notice after leave-table intent")
	}
	if sess.Summary == nil {
		t.Fatal("expected session summary to be persisted on leave-table flow")
	}
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

	counts := [2]int{}
	for run := range 2 {
		sess, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		runSessionWithAutoPlay(t, sess, defaultActionForPrompt, nil)
		counts[run] = sess.HandCount
	}

	if counts[0] != counts[1] {
		t.Errorf("non-deterministic: run1=%d hands, run2=%d hands", counts[0], counts[1])
	}
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
	for pid := range sess.Bots {
		if name := sess.PlayerName(pid); name == "" {
			t.Errorf("bot %d has empty name", pid)
		}
	}
}

func TestHandStartedEnvelopeOrdering(t *testing.T) {
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

	foundHandStarted := false
	blindBeforeHandStarted := false
	runSessionWithAutoPlay(t, sess, defaultActionForPrompt, func(env Envelope) {
		if env.Notice == nil {
			return
		}
		switch env.Notice.Type {
		case "session_started":
			return
		case "hand_started":
			if !foundHandStarted {
				foundHandStarted = true
			}
		case "blind_posted":
			if !foundHandStarted {
				blindBeforeHandStarted = true
			}
		}
	})

	if !foundHandStarted {
		t.Fatal("hand_started notice was never emitted")
	}
	if blindBeforeHandStarted {
		t.Error("blind_posted arrived before hand_started for the first hand")
	}
}

func TestHandStartedEnvelopeContainsCorrectData(t *testing.T) {
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

	var started engine.HandStartedEvent
	found := false
	runSessionWithAutoPlay(t, sess, defaultActionForPrompt, func(env Envelope) {
		if found || env.Notice == nil || env.Notice.Type != "hand_started" {
			return
		}
		event, ok := env.Notice.Event.(engine.HandStartedEvent)
		if !ok {
			t.Fatalf("expected HandStartedEvent, got %T", env.Notice.Event)
		}
		started = event
		found = true
	})

	if !found {
		t.Fatal("hand_started notice not found")
	}
	if started.Blinds.SB != 1 || started.Blinds.BB != 2 {
		t.Errorf("blinds = %d/%d, want 1/2", started.Blinds.SB, started.Blinds.BB)
	}
	if started.HandID != 1 {
		t.Errorf("HandID = %d, want 1", started.HandID)
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

func runSessionWithAutoPlay(t *testing.T, sess *Session, chooser func(*Prompt) engine.Action, observe func(Envelope)) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		sess.Run()
	}()
	for env := range sess.Updates {
		if observe != nil {
			observe(env)
		}
		if env.Prompt != nil {
			if env.Prompt.Kind == PromptKindBetweenHands {
				select {
				case sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Control: ControlIntent{Kind: ControlIntentReadyNextHand}}:
				case <-sess.stop:
				}
				continue
			}
			action := chooser(env.Prompt)
			select {
			case sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Action: action}:
			case <-sess.stop:
			}
		}
	}
	<-done
}

func defaultActionForPrompt(prompt *Prompt) engine.Action {
	action := engine.Action{PlayerID: prompt.PlayerID, Type: engine.ActionFold}
	for _, legal := range prompt.LegalActions {
		if legal.Type == engine.ActionCheck {
			action.Type = engine.ActionCheck
			return action
		}
	}
	for _, legal := range prompt.LegalActions {
		if legal.Type == engine.ActionCall {
			action.Type = engine.ActionCall
			action.Amount = legal.MinAmount
			return action
		}
	}
	if len(prompt.LegalActions) > 0 {
		action.Type = prompt.LegalActions[0].Type
		action.Amount = prompt.LegalActions[0].MinAmount
	}
	return action
}

func playerInfoByID(players []PlayerInfo, id engine.PlayerID) *PlayerInfo {
	for i := range players {
		if players[i].ID == id {
			return &players[i]
		}
	}
	return nil
}

func equalCards(a, b []engine.Card) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
