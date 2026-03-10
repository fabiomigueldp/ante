package engine

import (
	"fmt"
	"math/rand"
)

type Deck struct {
	cards [52]Card
	pos   int
	rng   *rand.Rand
	seed  int64
}

func NewDeck(seed int64) *Deck {
	return &Deck{
		cards: NewStandardDeckCards(),
		rng:   rand.New(rand.NewSource(seed)),
		seed:  seed,
	}
}

func (d *Deck) Seed() int64 {
	return d.seed
}

func (d *Deck) Shuffle() {
	for i := len(d.cards) - 1; i > 0; i-- {
		j := d.rng.Intn(i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
	d.pos = 0
}

func (d *Deck) Deal() Card {
	if d.pos >= len(d.cards) {
		panic(fmt.Sprintf("deck exhausted at position %d", d.pos))
	}
	card := d.cards[d.pos]
	d.pos++
	return card
}

func (d *Deck) DealN(n int) []Card {
	cards := make([]Card, n)
	for i := 0; i < n; i++ {
		cards[i] = d.Deal()
	}
	return cards
}

func (d *Deck) Burn() {
	_ = d.Deal()
}

func (d *Deck) Remaining() int {
	return len(d.cards) - d.pos
}
