package engine

import "testing"

func TestResolveShowdownSplitPot(t *testing.T) {
	players := []*Player{
		{ID: 1, Name: "A", Stack: 0, TotalBet: 100, Status: StatusAllIn, SeatIndex: 0, HoleCards: [2]Card{NewCard(Ace, Spades), NewCard(King, Hearts)}},
		{ID: 2, Name: "B", Stack: 0, TotalBet: 100, Status: StatusAllIn, SeatIndex: 1, HoleCards: [2]Card{NewCard(Ace, Diamonds), NewCard(King, Clubs)}},
	}
	hand := &Hand{Players: players, Board: []Card{NewCard(Queen, Spades), NewCard(Jack, Hearts), NewCard(Ten, Clubs), NewCard(Two, Hearts), NewCard(Three, Diamonds)}, DealerSeat: 0}
	result := ResolveShowdown(hand)
	if len(result.Pots) != 1 {
		t.Fatalf("expected 1 pot, got %d", len(result.Pots))
	}
	if len(result.Pots[0].Winners) != 2 {
		t.Fatalf("expected split pot, got %v", result.Pots[0].Winners)
	}
}

func TestOddChipAwardedLeftOfDealer(t *testing.T) {
	players := []*Player{
		{ID: 1, Name: "A", Stack: 0, TotalBet: 100, Status: StatusAllIn, SeatIndex: 0, HoleCards: [2]Card{NewCard(Ace, Spades), NewCard(King, Hearts)}},
		{ID: 2, Name: "B", Stack: 0, TotalBet: 100, Status: StatusAllIn, SeatIndex: 1, HoleCards: [2]Card{NewCard(Ace, Diamonds), NewCard(King, Clubs)}},
		{ID: 3, Name: "C", Stack: 0, TotalBet: 1, Status: StatusFolded, SeatIndex: 2, HoleCards: [2]Card{NewCard(Two, Clubs), NewCard(Seven, Clubs)}},
	}
	hand := &Hand{Players: players, Board: []Card{NewCard(Queen, Spades), NewCard(Jack, Hearts), NewCard(Ten, Clubs), NewCard(Two, Hearts), NewCard(Three, Diamonds)}, DealerSeat: 0}
	result := ResolveShowdown(hand)
	if result.Pots[0].OddChip != 2 {
		t.Fatalf("expected odd chip for seat left of dealer, got %d", result.Pots[0].OddChip)
	}
}
