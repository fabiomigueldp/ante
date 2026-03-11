package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func newGameTestModel(t *testing.T) GameModel {
	t.Helper()
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
	m.width = 120
	m.height = 40
	return m
}

func baseSnapshot(sess *session.Session) session.TableState {
	return session.TableState{
		HandNum:    1,
		HandID:     1,
		Blinds:     engine.BlindLevel{SB: 1, BB: 2},
		DealerSeat: 1,
		Street:     engine.StreetPreflop,
		Pot:        3,
		HumanCards: [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts)},
		Players: []session.PlayerInfo{
			{ID: sess.HumanID, Name: "Hero", Stack: 200, Bet: 0, IsHuman: true, Seat: 0},
			{ID: 2, Name: "Bot", Stack: 200, Bet: 2, Seat: 1},
		},
	}
}

func applyEnvelopeToGame(t *testing.T, m GameModel, env session.Envelope) GameModel {
	t.Helper()
	model, _ := m.handleEnvelope(env)
	next, ok := model.(GameModel)
	if !ok {
		t.Fatalf("expected GameModel, got %T", model)
	}
	return next
}

func TestGameActionBarFollowsPromptAtomically(t *testing.T) {
	m := newGameTestModel(t)
	snapshot := baseSnapshot(m.sess)

	m = applyEnvelopeToGame(t, m, session.Envelope{
		Seq:       1,
		SessionID: m.sess.SessionID,
		HandID:    1,
		Snapshot:  snapshot,
		Prompt: &session.Prompt{
			Seq:      1,
			HandID:   1,
			PlayerID: m.sess.HumanID,
			View: engine.PlayerView{
				MyID:       m.sess.HumanID,
				MyCards:    snapshot.HumanCards,
				MyStack:    200,
				MyBet:      0,
				Pot:        3,
				CurrentBet: 2,
				Street:     engine.StreetPreflop,
			},
			LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionCall, MinAmount: 2, MaxAmount: 2}},
		},
		Notice: &session.Notice{Type: "waiting_for_human", Message: "Your turn"},
	})

	barWithPrompt := m.renderActionBar()
	if !strings.Contains(barWithPrompt, "Fold") || !strings.Contains(barWithPrompt, "Call") {
		t.Fatalf("expected action bar to render prompt actions, got:\n%s", barWithPrompt)
	}

	m = applyEnvelopeToGame(t, m, session.Envelope{
		Seq:       2,
		SessionID: m.sess.SessionID,
		HandID:    1,
		Snapshot:  snapshot,
		Notice:    &session.Notice{Type: "action_taken", Message: "Bot calls 2"},
	})

	barWithoutPrompt := m.renderActionBar()
	if strings.Contains(barWithoutPrompt, "Fold") || strings.Contains(barWithoutPrompt, "Call") {
		t.Fatalf("expected action bar to hide prompt actions once prompt clears, got:\n%s", barWithoutPrompt)
	}
	if !strings.Contains(barWithoutPrompt, "Bot calls 2") {
		t.Fatalf("expected action bar to show latest status line, got:\n%s", barWithoutPrompt)
	}
}

func TestGameErrorRendersAtomicallyWithCurrentPrompt(t *testing.T) {
	m := newGameTestModel(t)
	snapshot := baseSnapshot(m.sess)
	prompt := &session.Prompt{
		Seq:      3,
		HandID:   1,
		PlayerID: m.sess.HumanID,
		View: engine.PlayerView{
			MyID:       m.sess.HumanID,
			MyCards:    snapshot.HumanCards,
			MyStack:    200,
			MyBet:      0,
			Pot:        3,
			CurrentBet: 2,
			Street:     engine.StreetPreflop,
		},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionRaise, MinAmount: 4, MaxAmount: 200}},
	}

	m = applyEnvelopeToGame(t, m, session.Envelope{
		Seq:       3,
		SessionID: m.sess.SessionID,
		HandID:    1,
		Snapshot:  snapshot,
		Prompt:    prompt,
		Error:     &session.SessionError{Code: "invalid_action", Message: "Minimum raise is 4."},
	})

	bar := m.renderActionBar()
	if !strings.Contains(bar, "Minimum raise is 4.") {
		t.Fatalf("expected error text in action bar, got:\n%s", bar)
	}
	if !strings.Contains(bar, "Raise") {
		t.Fatalf("expected prompt actions to remain visible with the error, got:\n%s", bar)
	}
}

func TestGameWaitingForHumanPlaysSoundWhenPromptAppears(t *testing.T) {
	m := newGameTestModel(t)
	snapshot := baseSnapshot(m.sess)
	count := 0
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) {
		if sound == audio.SoundYourTurn {
			count++
		}
	}
	defer func() { playGameSound = prev }()

	_ = applyEnvelopeToGame(t, m, session.Envelope{
		Seq:       1,
		SessionID: m.sess.SessionID,
		HandID:    1,
		Snapshot:  snapshot,
		Prompt: &session.Prompt{
			Seq:          1,
			HandID:       1,
			PlayerID:     m.sess.HumanID,
			LegalActions: []engine.LegalAction{{Type: engine.ActionCheck}},
		},
		Notice: &session.Notice{Type: "waiting_for_human", Message: "Your turn"},
	})

	if count != 1 {
		t.Fatalf("your turn sound count = %d, want 1", count)
	}
}

func TestPauseSaveShowsMidHandError(t *testing.T) {
	m := newGameTestModel(t)
	m.paused = true
	prevList := listSaves
	prevSave := saveSessionToSlot
	listSaves = func() ([]storage.SaveInfo, error) { return []storage.SaveInfo{{Slot: 1, Empty: true}}, nil }
	saveSessionToSlot = func(_ *session.Session, _ int) error { return session.ErrSaveMidHandNotSupported }
	defer func() {
		listSaves = prevList
		saveSessionToSlot = prevSave
	}()

	model, _ := m.handlePauseKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	next := model.(GameModel)
	if next.localMessage != session.ErrSaveMidHandNotSupported.Error() {
		t.Fatalf("local message = %q, want %q", next.localMessage, session.ErrSaveMidHandNotSupported.Error())
	}
	if next.localMessageKind != session.MessageKindError {
		t.Fatalf("local message kind = %q, want error", next.localMessageKind)
	}
}

func TestPauseSaveShowsSuccessMessage(t *testing.T) {
	m := newGameTestModel(t)
	m.paused = true
	prevList := listSaves
	prevSave := saveSessionToSlot
	listSaves = func() ([]storage.SaveInfo, error) { return []storage.SaveInfo{{Slot: 2, Empty: true}}, nil }
	saveSessionToSlot = func(_ *session.Session, slot int) error {
		if slot != 2 {
			return errors.New("unexpected slot")
		}
		return nil
	}
	defer func() {
		listSaves = prevList
		saveSessionToSlot = prevSave
	}()

	model, _ := m.handlePauseKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	next := model.(GameModel)
	if next.localMessage != "Game saved to slot 2." {
		t.Fatalf("local message = %q", next.localMessage)
	}
	if next.localMessageKind != session.MessageKindInfo {
		t.Fatalf("local message kind = %q, want info", next.localMessageKind)
	}
}

func TestGamePromptFlowResetsShowdownOnNextHand(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyEasy,
		Seats:         3,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	m := NewGameModel(sess, true)
	m.width = 120
	m.height = 40
	cmd := m.Init()

	sawShowdown := false
	showdownHandID := 0
	for step := 0; step < 2000; step++ {
		if cmd == nil {
			t.Fatal("unexpected nil command while session still in progress")
		}

		msg := cmd()
		env, ok := msg.(envelopeMsg)
		if !ok {
			if _, done := msg.(sessionDoneMsg); done {
				t.Fatal("session ended before observing showdown rollover")
			}
			t.Fatalf("unexpected message type %T", msg)
		}

		model, nextCmd := m.handleEnvelope(session.Envelope(env))
		m = model.(GameModel)

		if env.Notice != nil {
			switch env.Notice.Type {
			case "hand_revealed":
				sawShowdown = true
				showdownHandID = env.HandID
			case "hand_started":
				if sawShowdown && env.HandID > showdownHandID {
					if m.vm.Showdown {
						t.Fatal("expected showdown flag to be cleared on next hand")
					}
					if len(m.vm.Revealed) != 0 {
						t.Fatalf("expected revealed hands to reset, got %d", len(m.vm.Revealed))
					}
					if len(m.vm.PotAwards) != 0 {
						t.Fatalf("expected pot awards to reset, got %d", len(m.vm.PotAwards))
					}
					return
				}
			}
		}

		if env.Prompt != nil {
			if nextCmd != nil {
				t.Fatal("expected no waiter while a human prompt is active")
			}
			if env.Prompt.Kind == session.PromptKindBetweenHands {
				model, submitCmd := m.submitControl(session.ControlIntent{Kind: session.ControlIntentReadyNextHand})
				m = model.(GameModel)
				if submitCmd == nil {
					t.Fatal("expected submitControl to arm the next waiter")
				}
				cmd = submitCmd
				continue
			}
			model, submitCmd := m.submitAction(defaultGameAction(env.Prompt))
			m = model.(GameModel)
			if submitCmd == nil {
				t.Fatal("expected submitAction to arm the next waiter")
			}
			cmd = submitCmd
			continue
		}

		if session.Envelope(env).IsTerminal() {
			t.Fatal("session terminated before next hand started after a showdown")
		}
		cmd = nextCmd
	}

	t.Fatal("did not observe showdown state resetting on the next hand")
}

func TestGameIgnoresAdditionalInputWhileAwaitingActionResult(t *testing.T) {
	m := newGameTestModel(t)
	m.vm.Prompt = &session.Prompt{
		Seq:      10,
		HandID:   4,
		PlayerID: m.sess.HumanID,
		LegalActions: []engine.LegalAction{
			{Type: engine.ActionFold},
			{Type: engine.ActionCall, MinAmount: 10, MaxAmount: 10},
		},
	}

	model, cmd := m.submitAction(engine.Action{Type: engine.ActionCall})
	next := model.(GameModel)
	if !next.awaitingAction {
		t.Fatal("expected model to enter awaitingAction state after submitAction")
	}
	if cmd == nil {
		t.Fatal("expected submitAction to return a waiter command")
	}

	blockedModel, blockedCmd := next.handleGameKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	blocked := blockedModel.(GameModel)
	if !blocked.awaitingAction {
		t.Fatal("expected awaitingAction state to remain active while waiting for envelope")
	}
	if blockedCmd != nil {
		t.Fatal("expected no second command while awaiting action result")
	}
}

func TestGamePotOddsUsesEffectiveCallForShortStack(t *testing.T) {
	m := newGameTestModel(t)
	m.vm.Prompt = &session.Prompt{
		View: engine.PlayerView{
			MyID:       m.sess.HumanID,
			MyStack:    15,
			MyBet:      0,
			Pot:        500,
			CurrentBet: 300,
			Players: []engine.OpponentView{
				{ID: 2, Name: "Bot", Bet: 300, Status: engine.StatusActive},
			},
		},
		LegalActions: []engine.LegalAction{
			{Type: engine.ActionFold},
			{Type: engine.ActionAllIn, MinAmount: 15, MaxAmount: 15},
		},
	}

	line := m.potOddsLine()
	if !strings.Contains(line, "Pot: 215") {
		t.Fatalf("expected effective contestable pot, got %q", line)
	}
	if !strings.Contains(line, "Call: 15") {
		t.Fatalf("expected effective call amount, got %q", line)
	}
	if !strings.Contains(line, "Odds: 14.3:1") {
		t.Fatalf("expected effective odds, got %q", line)
	}
}

func TestGamePotOddsUsesAllInAmountWhenCallEqualsStack(t *testing.T) {
	m := newGameTestModel(t)
	m.vm.Prompt = &session.Prompt{
		View: engine.PlayerView{
			MyID:       m.sess.HumanID,
			MyStack:    50,
			MyBet:      0,
			Pot:        120,
			CurrentBet: 50,
			Players: []engine.OpponentView{
				{ID: 2, Name: "Bot", Bet: 50, Status: engine.StatusActive},
			},
		},
		LegalActions: []engine.LegalAction{
			{Type: engine.ActionFold},
			{Type: engine.ActionAllIn, MinAmount: 50, MaxAmount: 50},
		},
	}

	line := m.potOddsLine()
	if !strings.Contains(line, "Pot: 120") {
		t.Fatalf("expected full contestable pot before the all-in call, got %q", line)
	}
	if !strings.Contains(line, "Call: 50") {
		t.Fatalf("expected exact all-in call amount, got %q", line)
	}
}

func TestGamePotOddsDiscountsUncoverableOpponentExcess(t *testing.T) {
	m := newGameTestModel(t)
	m.vm.Prompt = &session.Prompt{
		View: engine.PlayerView{
			MyID:       m.sess.HumanID,
			MyStack:    40,
			MyBet:      10,
			Pot:        210,
			CurrentBet: 110,
			Players: []engine.OpponentView{
				{ID: 2, Name: "Bot A", Bet: 110, Status: engine.StatusActive},
				{ID: 3, Name: "Bot B", Bet: 160, Status: engine.StatusAllIn},
			},
		},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionAllIn, MinAmount: 50, MaxAmount: 50}},
	}

	line := m.potOddsLine()
	if !strings.Contains(line, "Call: 40") {
		t.Fatalf("expected effective call amount to be capped by stack, got %q", line)
	}
	if !strings.Contains(line, "Pot: 40") {
		t.Fatalf("expected pot display to discount uncoverable excess, got %q", line)
	}
}

func TestGameRenderSeatShowsStreetBetWhileActionBarShowsCallDelta(t *testing.T) {
	m := newGameTestModel(t)
	snapshot := session.TableState{
		HandNum:    25,
		HandID:     25,
		Blinds:     engine.BlindLevel{SB: 3, BB: 6, Ante: 1},
		DealerSeat: 0,
		Street:     engine.StreetPreflop,
		Pot:        24,
		HumanCards: [2]engine.Card{engine.NewCard(engine.Jack, engine.Clubs), engine.NewCard(engine.King, engine.Spades)},
		Players: []session.PlayerInfo{
			{ID: m.sess.HumanID, Name: "Hero", Stack: 502, Bet: 7, IsHuman: true, Seat: 0},
			{ID: 2, Name: "Carlos Rivera", Stack: 674, Bet: 17, Seat: 1},
		},
	}

	m = applyEnvelopeToGame(t, m, session.Envelope{
		Seq:       1,
		SessionID: m.sess.SessionID,
		HandID:    25,
		Snapshot:  snapshot,
		Prompt: &session.Prompt{
			Seq:      1,
			HandID:   25,
			PlayerID: m.sess.HumanID,
			View: engine.PlayerView{
				MyID:       m.sess.HumanID,
				MyCards:    snapshot.HumanCards,
				MyStack:    502,
				MyBet:      7,
				Pot:        24,
				CurrentBet: 17,
				Street:     engine.StreetPreflop,
			},
			LegalActions: []engine.LegalAction{
				{Type: engine.ActionFold},
				{Type: engine.ActionCall, MinAmount: 10, MaxAmount: 10},
				{Type: engine.ActionRaise, MinAmount: 28, MaxAmount: 509},
				{Type: engine.ActionAllIn, MinAmount: 509, MaxAmount: 509},
			},
		},
		Notice: &session.Notice{Type: "waiting_for_human", Message: "Your turn"},
	})

	opponentSeat := m.renderSeat(snapshot.Players[1])
	if !strings.Contains(opponentSeat, "Bet: 17") {
		t.Fatalf("expected opponent seat to show total street bet, got:\n%s", opponentSeat)
	}
	if strings.Contains(opponentSeat, "Bet: 10") {
		t.Fatalf("opponent seat should not show call delta, got:\n%s", opponentSeat)
	}

	bar := m.renderActionBar()
	if !strings.Contains(bar, "Call 10") {
		t.Fatalf("expected action bar to show incremental call amount, got:\n%s", bar)
	}
}

func TestBetweenHandsActionBarShowsReadyAndLeave(t *testing.T) {
	m := newGameTestModel(t)
	m.vm = session.GameVM{
		Prompt:       &session.Prompt{Kind: session.PromptKindBetweenHands, HandID: 3, PlayerID: m.sess.HumanID},
		PromptKind:   session.PromptKindBetweenHands,
		BetweenHands: true,
		CanSave:      true,
		StatusLine:   "Hand #3 complete. Press Enter for the next hand or L to leave the table.",
	}
	bar := m.renderActionBar()
	if !strings.Contains(bar, "Next Hand") || !strings.Contains(bar, "Leave Table") {
		t.Fatalf("expected between-hands controls, got:\n%s", bar)
	}
	if !strings.Contains(bar, "Pause/Save") {
		t.Fatalf("expected between-hands prompt to advertise pause/save, got:\n%s", bar)
	}
}

func TestBetweenHandsPreservesShowdownView(t *testing.T) {
	m := newGameTestModel(t)
	m.vm.Board = []engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Clubs), engine.NewCard(engine.Jack, engine.Diamonds), engine.NewCard(engine.Ten, engine.Spades)}
	m.vm.HumanCards = [2]engine.Card{engine.NewCard(engine.Nine, engine.Hearts), engine.NewCard(engine.Nine, engine.Diamonds)}
	m.vm.Pot = 22
	m.vm.Street = engine.StreetRiver
	m.vm.Showdown = true
	m.vm.Revealed = []session.RevealedHand{{PlayerID: 2, Name: "Bot", Cards: [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.Ace, engine.Hearts)}, Eval: "One Pair"}}
	m.vm.ShowdownPayouts = []session.ShowdownPayout{{PotIndex: 0, Winners: []engine.PlayerID{2}, Amount: 22}}
	m.vm.PotAwards = []string{"Bot wins 22"}
	m.vm.Prompt = &session.Prompt{Kind: session.PromptKindBetweenHands, HandID: 4, PlayerID: m.sess.HumanID}
	m.vm.PromptKind = session.PromptKindBetweenHands
	m.vm.BetweenHands = true
	view := m.renderTable()
	if !strings.Contains(view, "One Pair") || !strings.Contains(view, "+ 22") {
		t.Fatalf("expected showdown information to remain visible between hands, got:\n%s", view)
	}
	if !strings.Contains(view, "╭") {
		t.Fatalf("expected showdown summary to render inside a centered panel, got:\n%s", view)
	}
	if strings.Contains(view, "Hand Result") {
		t.Fatalf("expected showdown panel to avoid redundant heading text, got:\n%s", view)
	}
}

func TestRenderShowdownAggregatesMultiplePayoutsPerWinner(t *testing.T) {
	m := newGameTestModel(t)
	m.vm.Showdown = true
	m.vm.Revealed = []session.RevealedHand{
		{PlayerID: 2, Name: "Riley Banks", Cards: [2]engine.Card{engine.NewCard(engine.Five, engine.Spades), engine.NewCard(engine.Nine, engine.Clubs)}, Eval: "One Pair, 9s"},
		{PlayerID: m.sess.HumanID, Name: "Player", Cards: [2]engine.Card{engine.NewCard(engine.Four, engine.Diamonds), engine.NewCard(engine.Four, engine.Hearts)}, Eval: "One Pair, 4s"},
	}
	m.vm.ShowdownPayouts = []session.ShowdownPayout{
		{PotIndex: 0, Winners: []engine.PlayerID{2}, Amount: 12},
		{PotIndex: 1, Winners: []engine.PlayerID{2}, Amount: 27},
	}
	m.vm.PotAwards = []string{"Riley Banks wins 12", "Riley Banks wins 27"}
	panel := m.renderShowdown(120)
	if !strings.Contains(panel, "+ 39") {
		t.Fatalf("expected aggregated payout total in showdown panel, got:\n%s", panel)
	}
	if strings.Contains(panel, "wins 12") || strings.Contains(panel, "wins 27") {
		t.Fatalf("expected structured table, not raw payout receipt lines, got:\n%s", panel)
	}
}

func TestBetweenHandsEnvelopePreservesBoardPotAndHumanCards(t *testing.T) {
	m := newGameTestModel(t)
	env := session.Envelope{
		Seq:       9,
		SessionID: m.sess.SessionID,
		HandID:    5,
		Snapshot: session.TableState{
			HandNum:    5,
			HandID:     5,
			Blinds:     engine.BlindLevel{SB: 2, BB: 4},
			Board:      []engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Clubs), engine.NewCard(engine.Jack, engine.Diamonds), engine.NewCard(engine.Ten, engine.Spades)},
			Pot:        88,
			Street:     engine.StreetRiver,
			DealerSeat: 1,
			HumanCards: [2]engine.Card{engine.NewCard(engine.Nine, engine.Hearts), engine.NewCard(engine.Nine, engine.Diamonds)},
			Players: []session.PlayerInfo{
				{ID: m.sess.HumanID, Name: "Hero", Stack: 140, IsHuman: true, Seat: 0},
				{ID: 2, Name: "Bot", Stack: 0, Seat: 1, Status: engine.StatusOut},
			},
		},
		Prompt: &session.Prompt{Kind: session.PromptKindBetweenHands, HandID: 5, PlayerID: m.sess.HumanID},
		Notice: &session.Notice{Type: "waiting_for_ready", Message: "Hand #5 complete. Press Enter for the next hand or L to leave the table."},
	}
	m = applyEnvelopeToGame(t, m, env)
	table := m.renderTable()
	if strings.Contains(table, "0♠") {
		t.Fatalf("expected no zero-value card render in between-hands table, got:\n%s", table)
	}
	if !strings.Contains(table, "88") {
		t.Fatalf("expected preserved pot in between-hands table, got:\n%s", table)
	}
	if !strings.Contains(table, "River") {
		t.Fatalf("expected preserved street in between-hands table, got:\n%s", table)
	}
	if strings.Count(table, "┌──┐") < 7 || !strings.Contains(table, "│A ") || !strings.Contains(table, "│9 ") {
		t.Fatalf("expected preserved board and human cards in between-hands table, got:\n%s", table)
	}
}

func defaultGameAction(prompt *session.Prompt) engine.Action {
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
		if action.Type == engine.ActionAllIn {
			action.Amount = 0
		}
	}
	return action
}
