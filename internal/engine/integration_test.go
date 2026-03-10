package engine

import "testing"

func TestDeterministicHoleCardsForSameSeed(t *testing.T) {
	playersA := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 100, SeatIndex: 1}, {ID: 3, Stack: 100, SeatIndex: 2}}
	playersB := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 100, SeatIndex: 1}, {ID: 3, Stack: 100, SeatIndex: 2}}
	h1 := NewHand(1, playersA, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 77)
	h2 := NewHand(1, playersB, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 77)
	h1.PostBlinds()
	h2.PostBlinds()
	h1.DealHoleCards()
	h2.DealHoleCards()
	for i := range playersA {
		if playersA[i].HoleCards != playersB[i].HoleCards {
			t.Fatalf("expected same hole cards for player %d", playersA[i].ID)
		}
	}
}

func TestPlayerViewDoesNotExposeOpponentCards(t *testing.T) {
	players := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 100, SeatIndex: 1}}
	hand := NewHand(1, players, 0, BlindLevel{Level: 1, SB: 1, BB: 2}, 12)
	hand.PostBlinds()
	hand.DealHoleCards()
	view := hand.PlayerView(1)
	if len(view.Players) != 1 {
		t.Fatalf("expected 1 opponent, got %d", len(view.Players))
	}
	if view.Players[0].ID != 2 {
		t.Fatalf("expected opponent id 2, got %d", view.Players[0].ID)
	}
}
