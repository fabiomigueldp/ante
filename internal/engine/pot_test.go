package engine

import "testing"

func TestCalculatePotsSidePots(t *testing.T) {
	players := []*Player{
		{ID: 1, TotalBet: 50, Status: StatusAllIn},
		{ID: 2, TotalBet: 100, Status: StatusActive},
		{ID: 3, TotalBet: 100, Status: StatusActive},
	}
	pots := CalculatePots(players)
	if len(pots) != 2 {
		t.Fatalf("expected 2 pots, got %d", len(pots))
	}
	if pots[0].Amount != 150 {
		t.Fatalf("expected main pot 150, got %d", pots[0].Amount)
	}
	if pots[1].Amount != 100 {
		t.Fatalf("expected side pot 100, got %d", pots[1].Amount)
	}
}

func TestCalculatePotsFoldedPlayerNotEligible(t *testing.T) {
	players := []*Player{
		{ID: 1, TotalBet: 100, Status: StatusFolded},
		{ID: 2, TotalBet: 100, Status: StatusActive},
	}
	pots := CalculatePots(players)
	if len(pots) != 1 {
		t.Fatalf("expected 1 pot, got %d", len(pots))
	}
	if len(pots[0].Eligible) != 1 || pots[0].Eligible[0] != 2 {
		t.Fatalf("expected only player 2 eligible, got %v", pots[0].Eligible)
	}
}
