package engine

import (
	"fmt"
	"sort"
	"strings"
)

type HandRank uint8

const (
	HighCard HandRank = iota + 1
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

type EvalResult struct {
	Rank  HandRank
	Score uint64
	Cards [5]Card
	Name  string
}

func (r HandRank) String() string {
	switch r {
	case HighCard:
		return "High Card"
	case OnePair:
		return "One Pair"
	case TwoPair:
		return "Two Pair"
	case ThreeOfAKind:
		return "Three of a Kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full House"
	case FourOfAKind:
		return "Four of a Kind"
	case StraightFlush:
		return "Straight Flush"
	case RoyalFlush:
		return "Royal Flush"
	default:
		return "Unknown"
	}
}

func Evaluate(holeCards [2]Card, board []Card) EvalResult {
	all := make([]Card, 0, 2+len(board))
	all = append(all, holeCards[:]...)
	all = append(all, board...)
	if len(all) < 5 {
		panic("evaluate requires at least 5 cards")
	}
	if len(all) == 5 {
		var combo [5]Card
		copy(combo[:], all)
		return eval5(combo)
	}

	var best EvalResult
	first := true
	for a := 0; a < len(all)-4; a++ {
		for b := a + 1; b < len(all)-3; b++ {
			for c := b + 1; c < len(all)-2; c++ {
				for d := c + 1; d < len(all)-1; d++ {
					for e := d + 1; e < len(all); e++ {
						combo := [5]Card{all[a], all[b], all[c], all[d], all[e]}
						current := eval5(combo)
						if first || current.Score > best.Score {
							best = current
							first = false
						}
					}
				}
			}
		}
	}
	return best
}

func CompareHands(a, b EvalResult) int {
	switch {
	case a.Score > b.Score:
		return 1
	case a.Score < b.Score:
		return -1
	default:
		return 0
	}
}

func eval5(cards [5]Card) EvalResult {
	ranks := make([]Rank, 5)
	counts := make(map[Rank]int, 5)
	suit := cards[0].Suit
	flush := true
	for i, card := range cards {
		ranks[i] = card.Rank
		counts[card.Rank]++
		if card.Suit != suit {
			flush = false
		}
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })

	straight, straightHigh := isStraight(ranks)
	if flush && straight {
		if straightHigh == Ace {
			return buildEval(RoyalFlush, cards, []Rank{Ace}, "Royal Flush")
		}
		return buildEval(StraightFlush, cards, []Rank{straightHigh}, fmt.Sprintf("Straight Flush, %s high", rankWord(straightHigh)))
	}

	type rankCount struct {
		rank  Rank
		count int
	}
	groups := make([]rankCount, 0, len(counts))
	for rank, count := range counts {
		groups = append(groups, rankCount{rank: rank, count: count})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].count == groups[j].count {
			return groups[i].rank > groups[j].rank
		}
		return groups[i].count > groups[j].count
	})

	switch {
	case groups[0].count == 4:
		kicker := groups[1].rank
		return buildEval(FourOfAKind, cards, []Rank{groups[0].rank, kicker}, fmt.Sprintf("Four of a Kind, %ss", rankWord(groups[0].rank)))
	case groups[0].count == 3 && groups[1].count == 2:
		return buildEval(FullHouse, cards, []Rank{groups[0].rank, groups[1].rank}, fmt.Sprintf("Full House, %ss over %ss", rankWord(groups[0].rank), rankWord(groups[1].rank)))
	case flush:
		return buildEval(Flush, cards, ranks, fmt.Sprintf("Flush, %s high", rankWord(ranks[0])))
	case straight:
		return buildEval(Straight, cards, []Rank{straightHigh}, fmt.Sprintf("Straight, %s high", rankWord(straightHigh)))
	case groups[0].count == 3:
		kickers := make([]Rank, 0, 2)
		for _, group := range groups[1:] {
			kickers = append(kickers, group.rank)
		}
		sort.Slice(kickers, func(i, j int) bool { return kickers[i] > kickers[j] })
		return buildEval(ThreeOfAKind, cards, append([]Rank{groups[0].rank}, kickers...), fmt.Sprintf("Three of a Kind, %ss", rankWord(groups[0].rank)))
	case groups[0].count == 2 && groups[1].count == 2:
		highPair := groups[0].rank
		lowPair := groups[1].rank
		kicker := groups[2].rank
		return buildEval(TwoPair, cards, []Rank{highPair, lowPair, kicker}, fmt.Sprintf("Two Pair, %ss and %ss", rankWord(highPair), rankWord(lowPair)))
	case groups[0].count == 2:
		kickers := make([]Rank, 0, 3)
		for _, group := range groups[1:] {
			kickers = append(kickers, group.rank)
		}
		sort.Slice(kickers, func(i, j int) bool { return kickers[i] > kickers[j] })
		return buildEval(OnePair, cards, append([]Rank{groups[0].rank}, kickers...), fmt.Sprintf("One Pair, %ss", rankWord(groups[0].rank)))
	default:
		return buildEval(HighCard, cards, ranks, fmt.Sprintf("High Card, %s", rankWord(ranks[0])))
	}
}

func buildEval(rank HandRank, cards [5]Card, tie []Rank, name string) EvalResult {
	ordered := orderCardsForDisplay(cards)
	return EvalResult{
		Rank:  rank,
		Score: encodeScore(rank, tie),
		Cards: ordered,
		Name:  name,
	}
}

func encodeScore(rank HandRank, tie []Rank) uint64 {
	parts := make([]Rank, 5)
	copy(parts, tie)
	score := uint64(rank)
	for _, part := range parts {
		score = score*15 + uint64(part)
	}
	return score
}

func isStraight(ranks []Rank) (bool, Rank) {
	unique := make([]int, 0, len(ranks))
	seen := make(map[Rank]bool, len(ranks))
	for _, rank := range ranks {
		if !seen[rank] {
			seen[rank] = true
			unique = append(unique, int(rank))
		}
	}
	if len(unique) != 5 {
		return false, 0
	}
	sort.Sort(sort.Reverse(sort.IntSlice(unique)))
	if unique[0] == int(Ace) && unique[1] == 5 && unique[2] == 4 && unique[3] == 3 && unique[4] == 2 {
		return true, Five
	}
	for i := 0; i < 4; i++ {
		if unique[i]-1 != unique[i+1] {
			return false, 0
		}
	}
	return true, Rank(unique[0])
}

func orderCardsForDisplay(cards [5]Card) [5]Card {
	ordered := cards
	sort.Slice(ordered[:], func(i, j int) bool {
		if ordered[i].Rank == ordered[j].Rank {
			return ordered[i].Suit < ordered[j].Suit
		}
		return ordered[i].Rank > ordered[j].Rank
	})
	return ordered
}

func rankWord(rank Rank) string {
	switch rank {
	case Ace:
		return "Ace"
	case King:
		return "King"
	case Queen:
		return "Queen"
	case Jack:
		return "Jack"
	case Ten:
		return "Ten"
	default:
		return fmt.Sprintf("%d", rank)
	}
}

func CardsString(cards []Card) string {
	parts := make([]string, 0, len(cards))
	for _, card := range cards {
		parts = append(parts, card.String())
	}
	return strings.Join(parts, " ")
}
