package engine

import "testing"

func TestNewStandardDeckCardsUnique(t *testing.T) {
	deck := NewStandardDeckCards()
	seen := make(map[Card]bool, len(deck))
	for _, card := range deck {
		if seen[card] {
			t.Fatalf("duplicate card found: %v", card)
		}
		seen[card] = true
	}
	if len(seen) != 52 {
		t.Fatalf("expected 52 unique cards, got %d", len(seen))
	}
}
