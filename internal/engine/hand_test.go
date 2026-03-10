package engine

import "testing"

func TestHandFlowCompletes(t *testing.T) {
	players := []*Player{
		{ID: 1, Name: "A", Stack: 100, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 100, SeatIndex: 1},
		{ID: 3, Name: "C", Stack: 100, SeatIndex: 2},
	}
	hand := NewHand(1, players, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 42)
	hand.PostBlinds()
	hand.DealHoleCards()
	steps := 0
	for hand.Phase != PhaseComplete && steps < 64 {
		steps++
		step := hand.NextStep()
		switch step.Type {
		case StepAutoAdvance:
			if hand.Phase == PhaseShowdown {
				hand.ResolveShowdown()
			} else {
				hand.AdvanceStreet()
			}
		case StepNeedAction:
			legal := hand.LegalActions(step.PlayerID)
			if len(legal) == 0 {
				t.Fatalf("player %d had no legal actions", step.PlayerID)
			}
			action := Action{PlayerID: step.PlayerID, Type: legal[0].Type, Amount: legal[0].MinAmount}
			if action.Type == ActionAllIn {
				action.Amount = 0
			}
			if _, err := hand.ApplyAction(step.PlayerID, action); err != nil {
				t.Fatalf("unexpected action error: %v", err)
			}
		case StepComplete:
			break
		}
	}
	if hand.Phase != PhaseComplete {
		t.Fatalf("hand did not complete; phase=%v steps=%d", hand.Phase, steps)
	}
}

func TestHeadsUpBlindAssignment(t *testing.T) {
	players := []*Player{
		{ID: 1, Name: "A", Stack: 100, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 100, SeatIndex: 1},
	}
	hand := NewHand(1, players, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 99)
	if hand.SBSeat != hand.DealerSeat {
		t.Fatalf("expected dealer to post small blind heads-up; dealer=%d sb=%d", hand.DealerSeat, hand.SBSeat)
	}
	if hand.preflopFirstToActSeat() != hand.SBSeat {
		t.Fatalf("expected dealer/sb to act first preflop")
	}
}

func TestAllInSkipsToShowdown(t *testing.T) {
	// Hypothesis B: both players go all-in preflop → should skip to showdown
	// without producing any StepNeedAction after the preflop actions complete.
	players := []*Player{
		{ID: 1, Name: "A", Stack: 100, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 100, SeatIndex: 1},
	}
	hand := NewHand(1, players, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 42)
	hand.PostBlinds()
	hand.DealHoleCards()

	// SB (player 1) goes all-in preflop
	step := hand.NextStep()
	if step.Type != StepNeedAction {
		t.Fatalf("expected NeedAction for SB, got %v", step.Type)
	}
	if _, err := hand.ApplyAction(step.PlayerID, Action{Type: ActionAllIn}); err != nil {
		t.Fatalf("SB all-in error: %v", err)
	}

	// BB (player 2) calls all-in
	step = hand.NextStep()
	if step.Type != StepNeedAction {
		t.Fatalf("expected NeedAction for BB, got %v", step.Type)
	}
	if _, err := hand.ApplyAction(step.PlayerID, Action{Type: ActionAllIn}); err != nil {
		t.Fatalf("BB all-in error: %v", err)
	}

	// After both all-in, should auto-advance through streets to showdown.
	// No StepNeedAction should appear.
	steps := 0
	for hand.Phase != PhaseComplete && steps < 32 {
		steps++
		step = hand.NextStep()
		if step.Type == StepNeedAction {
			t.Fatalf("unexpected NeedAction for player %d after all-in at phase %v street %v", step.PlayerID, hand.Phase, hand.Street)
		}
		if step.Type == StepAutoAdvance {
			if hand.Phase == PhaseShowdown {
				hand.ResolveShowdown()
			} else {
				hand.AdvanceStreet()
			}
		}
		if step.Type == StepComplete {
			break
		}
	}
	if hand.Phase != PhaseComplete {
		t.Fatalf("hand did not complete after all-in; phase=%v steps=%d", hand.Phase, steps)
	}
}
