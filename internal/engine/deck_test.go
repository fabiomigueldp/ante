package engine

import "testing"

func TestDeckShuffleDeterministic(t *testing.T) {
	a := NewDeck(42)
	b := NewDeck(42)
	a.Shuffle()
	b.Shuffle()
	for i := 0; i < 52; i++ {
		if a.cards[i] != b.cards[i] {
			t.Fatalf("expected same order at position %d", i)
		}
	}
}

func TestDeckDealRemaining(t *testing.T) {
	deck := NewDeck(7)
	deck.Shuffle()
	for i := 0; i < 5; i++ {
		_ = deck.Deal()
	}
	if got := deck.Remaining(); got != 47 {
		t.Fatalf("expected 47 remaining, got %d", got)
	}
}
