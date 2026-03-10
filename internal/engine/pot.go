package engine

import "sort"

type Pot struct {
	Amount   int
	Eligible []PlayerID
}

func CalculatePots(players []*Player) []Pot {
	type contribution struct {
		player   *Player
		amount   int
		eligible bool
	}

	contribs := make([]contribution, 0, len(players))
	for _, p := range players {
		if p.TotalBet <= 0 {
			continue
		}
		contribs = append(contribs, contribution{
			player:   p,
			amount:   p.TotalBet,
			eligible: p.Status != StatusFolded && p.Status != StatusOut && p.Status != StatusSittingOut,
		})
	}
	if len(contribs) == 0 {
		return nil
	}

	sort.Slice(contribs, func(i, j int) bool {
		if contribs[i].amount == contribs[j].amount {
			return contribs[i].player.ID < contribs[j].player.ID
		}
		return contribs[i].amount < contribs[j].amount
	})

	var pots []Pot
	previous := 0
	for i, current := range contribs {
		if current.amount == previous {
			continue
		}
		level := current.amount
		participants := len(contribs) - i
		amount := (level - previous) * participants
		eligible := make([]PlayerID, 0, participants)
		for j := i; j < len(contribs); j++ {
			if contribs[j].eligible {
				eligible = append(eligible, contribs[j].player.ID)
			}
		}
		if amount > 0 {
			pots = append(pots, Pot{Amount: amount, Eligible: eligible})
		}
		previous = level
	}
	return pots
}
