package ai

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestBotDecisionUsesOnlyLegalActions(t *testing.T) {
	character := Characters()[0]
	bot := NewBot(character, 42)
	view := engine.PlayerView{
		MyID:       1,
		MyCards:    [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Spades)},
		MyStack:    100,
		MyBet:      0,
		Street:     engine.StreetPreflop,
		Pot:        3,
		CurrentBet: 2,
		LegalActions: []engine.LegalAction{
			{Type: engine.ActionFold},
			{Type: engine.ActionCall, MinAmount: 2, MaxAmount: 2},
			{Type: engine.ActionRaise, MinAmount: 6, MaxAmount: 100},
			{Type: engine.ActionAllIn, MinAmount: 100, MaxAmount: 100},
		},
	}
	decision := bot.Decide(view)
	legal := false
	for _, action := range view.LegalActions {
		if action.Type == decision.Action.Type {
			legal = true
		}
	}
	if !legal {
		t.Fatalf("bot selected illegal action: %+v", decision.Action)
	}
}

func TestSelectCharactersCount(t *testing.T) {
	selected := SelectCharacters(DifficultyEasy, 8, 7)
	if len(selected) != 8 {
		t.Fatalf("expected 8 characters, got %d", len(selected))
	}
}
