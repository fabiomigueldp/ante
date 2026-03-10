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

func TestAnteCallAmountIsFullBB(t *testing.T) {
	// With antes, a player who has posted ante should still need to call
	// the full BB to stay in, not BB-ante.
	//
	// Level 3: SB=3, BB=6, Ante=1.  Three players with 200 chips.
	// After PostBlinds + DealHoleCards:
	//   P0 (dealer): Bet=1 (ante only)
	//   P1 (SB):     Bet=1+3=4
	//   P2 (BB):     Bet=1+6=7
	// CurrentBet must equal 7, so P0 toCall = 7-1 = 6 = BB.
	players := []*Player{
		{ID: 1, Name: "A", Stack: 200, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 200, SeatIndex: 1},
		{ID: 3, Name: "C", Stack: 200, SeatIndex: 2},
	}
	blinds := BlindLevel{Level: 3, SB: 3, BB: 6, Ante: 1}
	hand := NewHand(1, players, 0, blinds, 42)

	hand.PostBlinds()
	hand.DealHoleCards()

	// dealer=0, SB=1, BB=2. First preflop actor is after BB → seat 0.
	step := hand.NextStep()
	if step.Type != StepNeedAction {
		t.Fatalf("expected NeedAction, got %v", step.Type)
	}

	p0 := hand.playerByID(step.PlayerID)
	if p0 == nil {
		t.Fatal("first actor is nil")
	}

	// Check that CurrentBet includes ante
	if hand.Betting.CurrentBet != 7 {
		t.Fatalf("CurrentBet = %d, want 7 (ante %d + BB %d)", hand.Betting.CurrentBet, blinds.Ante, blinds.BB)
	}

	// The call amount for P0 (who posted ante=1) should be 6 (= BB)
	legal := hand.LegalActions(p0.ID)
	for _, la := range legal {
		if la.Type == ActionCall {
			if la.MinAmount != 6 {
				t.Fatalf("call amount = %d, want 6 (full BB)", la.MinAmount)
			}
			return
		}
	}
	t.Fatal("ActionCall not found in legal actions")
}

func TestAnteCallAmountSBPaysCorrectDifference(t *testing.T) {
	// SB has posted ante+SB. To call BB, the SB should pay BB-SB.
	players := []*Player{
		{ID: 1, Name: "A", Stack: 200, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 200, SeatIndex: 1},
		{ID: 3, Name: "C", Stack: 200, SeatIndex: 2},
	}
	blinds := BlindLevel{Level: 3, SB: 3, BB: 6, Ante: 1}
	hand := NewHand(1, players, 0, blinds, 42)
	hand.PostBlinds()
	hand.DealHoleCards()

	// SB is player at seat 1
	sbPlayer := hand.playerAtSeat(hand.SBSeat)
	if sbPlayer == nil {
		t.Fatal("SB player not found")
	}
	// SB should have Bet = ante + SB = 4
	if sbPlayer.Bet != 4 {
		t.Fatalf("SB Bet = %d, want 4 (ante %d + SB %d)", sbPlayer.Bet, blinds.Ante, blinds.SB)
	}

	legal := hand.LegalActions(sbPlayer.ID)
	for _, la := range legal {
		if la.Type == ActionCall {
			// toCall = CurrentBet(7) - sbBet(4) = 3 = BB - SB
			if la.MinAmount != 3 {
				t.Fatalf("SB call amount = %d, want 3 (BB-SB)", la.MinAmount)
			}
			return
		}
	}
	t.Fatal("ActionCall not found for SB")
}

func TestNoAnteCurrentBetEqualsBB(t *testing.T) {
	// Without antes, CurrentBet should equal BB.
	players := []*Player{
		{ID: 1, Name: "A", Stack: 200, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 200, SeatIndex: 1},
		{ID: 3, Name: "C", Stack: 200, SeatIndex: 2},
	}
	blinds := BlindLevel{Level: 1, SB: 1, BB: 2, Ante: 0}
	hand := NewHand(1, players, 0, blinds, 42)
	hand.PostBlinds()
	hand.DealHoleCards()

	if hand.Betting.CurrentBet != 2 {
		t.Fatalf("CurrentBet = %d, want 2 (BB with no ante)", hand.Betting.CurrentBet)
	}
}

func TestNewHandClonesPlayersInsteadOfReusingPointers(t *testing.T) {
	players := []*Player{
		{ID: 1, Name: "A", Stack: 100, SeatIndex: 0},
		{ID: 2, Name: "B", Stack: 100, SeatIndex: 1},
	}
	hand := NewHand(1, players, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 42)

	if hand.Players[0] == players[0] {
		t.Fatal("expected hand player to be a clone, not the same pointer as table player")
	}

	hand.Players[0].Stack = 77
	if players[0].Stack != 100 {
		t.Fatalf("table player stack mutated via hand clone: got %d want 100", players[0].Stack)
	}

	if players[0].Bet != 0 || players[0].TotalBet != 0 {
		t.Fatalf("table player hand-local fields should remain untouched, got bet=%d total=%d", players[0].Bet, players[0].TotalBet)
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
