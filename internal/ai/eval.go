package ai

import "github.com/fabiomigueldp/ante/internal/engine"

func estimateStrength(view engine.PlayerView) float64 {
	hole := view.MyCards
	r1 := int(hole[0].Rank)
	r2 := int(hole[1].Rank)
	if len(view.Board) == 0 {
		strength := float64(r1+r2) / 28.0
		if hole[0].Rank == hole[1].Rank {
			strength += 0.25
		}
		if hole[0].Suit == hole[1].Suit {
			strength += 0.05
		}
		gap := absInt(r1 - r2)
		if gap <= 1 {
			strength += 0.05
		}
		if strength > 1 {
			strength = 1
		}
		return strength
	}
	result := engine.Evaluate(view.MyCards, view.Board)
	switch result.Rank {
	case engine.RoyalFlush:
		return 1.0
	case engine.StraightFlush:
		return 0.98
	case engine.FourOfAKind:
		return 0.95
	case engine.FullHouse:
		return 0.9
	case engine.Flush:
		return 0.82
	case engine.Straight:
		return 0.75
	case engine.ThreeOfAKind:
		return 0.68
	case engine.TwoPair:
		return 0.58
	case engine.OnePair:
		return 0.42
	default:
		return 0.18
	}
}

func estimateDraws(view engine.PlayerView) float64 {
	if len(view.Board) < 3 {
		return 0
	}
	all := append([]engine.Card{}, view.Board...)
	all = append(all, view.MyCards[:]...)
	bySuit := map[engine.Suit]int{}
	byRank := map[int]bool{}
	for _, card := range all {
		bySuit[card.Suit]++
		byRank[int(card.Rank)] = true
	}
	draw := 0.0
	for _, count := range bySuit {
		if count == 4 {
			draw += 0.35
		}
	}
	if hasStraightDraw(byRank) {
		draw += 0.28
	}
	if draw > 0.6 {
		draw = 0.6
	}
	return draw
}

func tablePressure(view engine.PlayerView) float64 {
	pressure := 0.0
	if view.CurrentBet > 0 && view.MyStack > 0 {
		pressure += float64(view.CurrentBet-view.MyBet) / float64(view.MyStack+view.MyBet+1)
	}
	if view.NumActivePlayers >= 4 {
		pressure += 0.12
	}
	if pressure > 1 {
		pressure = 1
	}
	return pressure
}

func hasStraightDraw(ranks map[int]bool) bool {
	for start := 1; start <= 10; start++ {
		count := 0
		for offset := 0; offset < 5; offset++ {
			rank := start + offset
			if rank == 1 {
				rank = 14
			}
			if ranks[rank] {
				count++
			}
		}
		if count >= 4 {
			return true
		}
	}
	return false
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
