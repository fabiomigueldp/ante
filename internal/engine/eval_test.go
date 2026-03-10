package engine

import "testing"

func TestEvaluateRoyalFlush(t *testing.T) {
	hole := [2]Card{NewCard(Ace, Spades), NewCard(King, Spades)}
	board := []Card{NewCard(Queen, Spades), NewCard(Jack, Spades), NewCard(Ten, Spades), NewCard(Two, Hearts), NewCard(Three, Clubs)}
	result := Evaluate(hole, board)
	if result.Rank != RoyalFlush {
		t.Fatalf("expected royal flush, got %v", result.Rank)
	}
}

func TestEvaluateWheelStraight(t *testing.T) {
	hole := [2]Card{NewCard(Ace, Spades), NewCard(Two, Hearts)}
	board := []Card{NewCard(Three, Clubs), NewCard(Four, Diamonds), NewCard(Five, Hearts), NewCard(King, Clubs), NewCard(Queen, Hearts)}
	result := Evaluate(hole, board)
	if result.Rank != Straight {
		t.Fatalf("expected straight, got %v", result.Rank)
	}
	if result.Name != "Straight, 5 high" {
		t.Fatalf("unexpected name %q", result.Name)
	}
}

func TestCompareTwoPairKicker(t *testing.T) {
	a := Evaluate([2]Card{NewCard(Ace, Spades), NewCard(King, Hearts)}, []Card{NewCard(Ace, Hearts), NewCard(King, Clubs), NewCard(Two, Spades), NewCard(Nine, Clubs), NewCard(Seven, Hearts)})
	b := Evaluate([2]Card{NewCard(Ace, Diamonds), NewCard(Queen, Hearts)}, []Card{NewCard(Ace, Clubs), NewCard(Queen, Spades), NewCard(Two, Clubs), NewCard(Nine, Hearts), NewCard(Seven, Clubs)})
	if CompareHands(a, b) <= 0 {
		t.Fatalf("expected aces and kings to beat aces and queens")
	}
}
