package tui

import (
	"strings"
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
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
