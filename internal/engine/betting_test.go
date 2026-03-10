package engine

import "testing"

func TestLegalActionsFacingBet(t *testing.T) {
	br := NewBettingRound(StreetPreflop, 2, 10)
	player := &Player{ID: 1, Stack: 100, Bet: 2, Status: StatusActive}
	legal := br.LegalActions(player)
	if len(legal) != 4 {
		t.Fatalf("expected 4 legal actions, got %d", len(legal))
	}
}

func TestApplyRaiseUpdatesCurrentBet(t *testing.T) {
	br := NewBettingRound(StreetPreflop, 2, 10)
	player := &Player{ID: 1, Stack: 100, Bet: 10, Status: StatusActive}
	resolved, err := br.Apply(player, Action{Type: ActionRaise, Amount: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Amount != 20 || br.CurrentBet != 20 || br.MinRaise != 10 {
		t.Fatalf("unexpected raise state: resolved=%d current=%d minraise=%d", resolved.Amount, br.CurrentBet, br.MinRaise)
	}
}

func TestAllInBelowMinRaiseDoesNotReopen(t *testing.T) {
	br := NewBettingRound(StreetTurn, 10, 50)
	br.MinRaise = 40
	br.ActedPlayers[1] = true
	player := &Player{ID: 2, Stack: 20, Bet: 40, Status: StatusActive}
	_, err := br.Apply(player, Action{Type: ActionAllIn})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if br.LastAggressor != 0 {
		t.Fatalf("expected no reopened action, got aggressor %d", br.LastAggressor)
	}
}

func TestRaiseReopensAction(t *testing.T) {
	br := NewBettingRound(StreetFlop, 10, 20)
	br.ActedPlayers[1] = true
	player := &Player{ID: 2, Stack: 100, Bet: 20, Status: StatusActive}
	_, err := br.Apply(player, Action{Type: ActionRaise, Amount: 40})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if br.LastAggressor != 2 {
		t.Fatalf("expected player 2 as aggressor, got %d", br.LastAggressor)
	}
	if br.ActedPlayers[1] {
		t.Fatal("expected prior acted players to be reopened")
	}
}

func TestAllInAmountIsTargetBet(t *testing.T) {
	// Hypothesis A: player with Stack=439, Bet=10 should show AllIn MinAmount=449 (total target)
	br := NewBettingRound(StreetPreflop, 2, 10)
	player := &Player{ID: 1, Stack: 439, Bet: 10, Status: StatusActive}
	legal := br.LegalActions(player)
	for _, la := range legal {
		if la.Type == ActionAllIn {
			expected := player.Bet + player.Stack // 449
			if la.MinAmount != expected {
				t.Fatalf("AllIn MinAmount = %d, want %d (Bet %d + Stack %d)", la.MinAmount, expected, player.Bet, player.Stack)
			}
			// The TUI should display MinAmount - Bet = 439 (remaining stack)
			display := la.MinAmount - player.Bet
			if display != player.Stack {
				t.Fatalf("AllIn display = %d, want %d (remaining stack)", display, player.Stack)
			}
			return
		}
	}
	t.Fatal("AllIn action not found in legal actions")
}
