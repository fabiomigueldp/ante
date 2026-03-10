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
