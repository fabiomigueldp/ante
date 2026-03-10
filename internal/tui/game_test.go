package tui

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
)

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

	plays := make([]audio.SoundType, 0, 2)
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	_, _ = m.handleSessionEvent(session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: sess.HumanID, Action: engine.Action{Type: engine.ActionCheck}, PotTotal: 3}})
	_, _ = m.handleSessionEvent(session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionCall}, PotTotal: 6}})
	_, _ = m.handleSessionEvent(session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionAllIn}, PotTotal: 20}})

	if len(plays) != 2 {
		t.Fatalf("len(plays) = %d, want 2", len(plays))
	}
	if plays[0] != audio.SoundCheck {
		t.Fatalf("first sound = %v, want SoundCheck", plays[0])
	}
	if plays[1] != audio.SoundAllIn {
		t.Fatalf("second sound = %v, want SoundAllIn", plays[1])
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
	_, _ = m.handleSessionEvent(session.SessionEvent{Type: "waiting_for_human", Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}}}})
	_, _ = m.handleActionReq(session.ActionRequest{Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}}}})

	if count != 1 {
		t.Fatalf("your turn sound count = %d, want 1", count)
	}
}
