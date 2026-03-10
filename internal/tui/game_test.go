package tui

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
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

	m := NewGameModel(sess, true, true)
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
