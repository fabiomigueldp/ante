package session

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestReduceGameVMIgnoresStaleEnvelope(t *testing.T) {
	base := TableState{
		HandNum: 1,
		Blinds:  engine.BlindLevel{SB: 1, BB: 2},
		Players: []PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}},
	}
	vm := ReduceGameVM(GameVM{}, Envelope{Seq: 2, SessionID: "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Snapshot: base, Notice: &Notice{Type: "action_taken", Message: "Villain calls 2"}})
	stale := Envelope{Seq: 1, SessionID: vm.SessionID, Snapshot: base, Notice: &Notice{Type: "bot_thinking", Message: "Old notice"}}
	updated := ReduceGameVM(vm, stale)

	if updated.Seq != 2 {
		t.Fatalf("seq = %d, want 2", updated.Seq)
	}
	if updated.StatusLine != "Villain calls 2" {
		t.Fatalf("status line = %q, want original value", updated.StatusLine)
	}
}

func TestReduceGameVMKeepsPromptAndErrorAtomic(t *testing.T) {
	snapshot := TableState{
		HandNum: 2,
		Blinds:  engine.BlindLevel{SB: 1, BB: 2},
		Players: []PlayerInfo{{ID: 1, Name: "Hero", Stack: 200, IsHuman: true}},
		Pot:     3,
	}
	prompt := &Prompt{
		PlayerID: 1,
		HandID:   2,
		View: engine.PlayerView{
			MyID:       1,
			MyStack:    200,
			Pot:        3,
			CurrentBet: 2,
		},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionCall, MinAmount: 2, MaxAmount: 2}},
	}

	vm := ReduceGameVM(GameVM{}, Envelope{
		Seq:       3,
		SessionID: "ses_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		HandID:    2,
		Snapshot:  snapshot,
		Prompt:    prompt,
		Error:     &SessionError{Code: "invalid_action", Message: "Minimum raise is 4."},
	})

	if vm.Prompt == nil {
		t.Fatal("expected prompt to remain present")
	}
	if vm.Message != "Minimum raise is 4." {
		t.Fatalf("message = %q, want error text", vm.Message)
	}
	if vm.MessageKind != MessageKindError {
		t.Fatalf("message kind = %q, want error", vm.MessageKind)
	}
}

func TestReduceGameVMClearsPromptWhenEnvelopeHasNoPrompt(t *testing.T) {
	snapshot := TableState{
		HandNum: 2,
		Blinds:  engine.BlindLevel{SB: 1, BB: 2},
		Players: []PlayerInfo{{ID: 1, Name: "Hero", Stack: 200, IsHuman: true}},
	}
	vm := ReduceGameVM(GameVM{}, Envelope{
		Seq:       1,
		SessionID: "ses_cccccccccccccccccccccccccccccccc",
		HandID:    2,
		Snapshot:  snapshot,
		Prompt: &Prompt{
			Seq:          1,
			HandID:       2,
			PlayerID:     1,
			LegalActions: []engine.LegalAction{{Type: engine.ActionFold}},
		},
	})

	updated := ReduceGameVM(vm, Envelope{
		Seq:       2,
		SessionID: vm.SessionID,
		HandID:    2,
		Snapshot:  snapshot,
		Notice:    &Notice{Type: "action_taken", Message: "Hero folds"},
	})

	if updated.Prompt != nil {
		t.Fatal("expected prompt to be cleared")
	}
	if updated.StatusLine != "Hero folds" {
		t.Fatalf("status line = %q, want Hero folds", updated.StatusLine)
	}
}

func TestReduceGameVMBetweenHandsKeepsShowdownVisible(t *testing.T) {
	vm := GameVM{
		Seq:       4,
		SessionID: "ses_waitingready",
		HandID:    7,
		Showdown:  true,
		Revealed:  []RevealedHand{{PlayerID: 2, Name: "Bot", Eval: "Straight"}},
		PotAwards: []string{"Bot wins 40"},
	}
	updated := ReduceGameVM(vm, Envelope{
		Seq:       5,
		SessionID: vm.SessionID,
		HandID:    7,
		Snapshot: TableState{
			HandNum:         7,
			HandID:          7,
			Blinds:          engine.BlindLevel{SB: 2, BB: 4},
			Pot:             3,
			Street:          engine.StreetFlop,
			Board:           []engine.Card{engine.NewCard(engine.Ace, engine.Spades)},
			HumanCards:      [2]engine.Card{engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Diamonds)},
			Boundary:        true,
			Showdown:        true,
			Players:         []PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}},
			Revealed:        []RevealedHand{{PlayerID: 2, Name: "Bot", Eval: "Straight"}},
			ShowdownPayouts: []ShowdownPayout{{PotIndex: 0, Winners: []engine.PlayerID{2}, Amount: 40}},
			PotAwards:       []string{"Bot wins 40"},
		},
		Prompt: &Prompt{Kind: PromptKindBetweenHands, HandID: 7, PlayerID: 1},
		Notice: &Notice{Type: "waiting_for_ready", Message: "Press Enter for the next hand or L to leave the table."},
	})
	if !updated.BetweenHands {
		t.Fatal("expected between-hands state")
	}
	if !updated.CanSave {
		t.Fatal("expected CanSave to be enabled while waiting between hands")
	}
	if !updated.Showdown {
		t.Fatal("expected showdown to remain visible while waiting for ready")
	}
	if updated.Pot != 3 {
		t.Fatalf("pot = %d, want 3", updated.Pot)
	}
	if updated.Street != engine.StreetFlop {
		t.Fatalf("street = %d, want flop", updated.Street)
	}
	if len(updated.Board) != 1 || updated.Board[0] != engine.NewCard(engine.Ace, engine.Spades) {
		t.Fatalf("board = %+v, want preserved showdown board", updated.Board)
	}
	if updated.HumanCards != [2]engine.Card{engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Diamonds)} {
		t.Fatalf("human cards = %+v, want preserved showdown hand", updated.HumanCards)
	}
	if len(updated.Revealed) != 1 || len(updated.PotAwards) != 1 {
		t.Fatal("expected showdown details to remain intact")
	}
	if len(updated.ShowdownPayouts) != 1 || updated.ShowdownPayouts[0].Amount != 40 {
		t.Fatalf("expected showdown payouts to remain intact, got %+v", updated.ShowdownPayouts)
	}
}
