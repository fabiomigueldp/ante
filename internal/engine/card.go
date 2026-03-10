package engine

import "fmt"

type Suit uint8

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

type Rank uint8

const (
	Two Rank = 2 + iota
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

type Card struct {
	Rank Rank
	Suit Suit
}

func NewCard(rank Rank, suit Suit) Card {
	return Card{Rank: rank, Suit: suit}
}

func (s Suit) String() string {
	switch s {
	case Spades:
		return "s"
	case Hearts:
		return "h"
	case Diamonds:
		return "d"
	case Clubs:
		return "c"
	default:
		return "?"
	}
}

func (r Rank) String() string {
	switch r {
	case Ten:
		return "T"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	default:
		return fmt.Sprintf("%d", r)
	}
}

func (c Card) String() string {
	return c.Rank.String() + c.Suit.String()
}

func NewStandardDeckCards() [52]Card {
	var cards [52]Card
	idx := 0
	for suit := Spades; suit <= Clubs; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			cards[idx] = NewCard(rank, suit)
			idx++
		}
	}
	return cards
}
