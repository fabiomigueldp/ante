package engine

import "sort"

type ShowdownResult struct {
	Pots     []PotResult
	Revealed []PlayerHand
	RawPots  []Pot
}

type PotResult struct {
	PotIndex int
	Pot      Pot
	Winners  []PlayerID
	Amount   int
	OddChip  PlayerID
	Hands    []PlayerHand
}

type PlayerHand struct {
	PlayerID PlayerID
	Cards    [2]Card
	Eval     EvalResult
	Mucked   bool
}

func ResolveShowdown(hand *Hand) ShowdownResult {
	rawPots := CalculatePots(hand.Players)
	result := ShowdownResult{RawPots: rawPots}
	if len(rawPots) == 0 {
		return result
	}

	revealOrder := showdownOrder(hand)
	revealedByID := make(map[PlayerID]PlayerHand)

	for potIndex, pot := range rawPots {
		eligible := eligiblePlayersForPot(hand, pot)
		potResult := PotResult{PotIndex: potIndex, Pot: pot, Amount: pot.Amount}
		if len(eligible) == 0 {
			result.Pots = append(result.Pots, potResult)
			continue
		}
		if len(eligible) == 1 {
			winner := eligible[0]
			winner.Stack += pot.Amount
			potResult.Winners = []PlayerID{winner.ID}
			result.Pots = append(result.Pots, potResult)
			continue
		}

		bestScore := uint64(0)
		for _, player := range eligible {
			eval := Evaluate(player.HoleCards, hand.Board)
			entry := PlayerHand{PlayerID: player.ID, Cards: player.HoleCards, Eval: eval}
			potResult.Hands = append(potResult.Hands, entry)
			revealedByID[player.ID] = entry
			if eval.Score > bestScore {
				bestScore = eval.Score
			}
		}
		for _, entry := range potResult.Hands {
			if entry.Eval.Score == bestScore {
				potResult.Winners = append(potResult.Winners, entry.PlayerID)
			}
		}
		share := pot.Amount / len(potResult.Winners)
		for _, winnerID := range potResult.Winners {
			if winner := hand.playerByID(winnerID); winner != nil {
				winner.Stack += share
			}
		}
		remainder := pot.Amount - share*len(potResult.Winners)
		if remainder > 0 {
			odd := oddChipWinner(hand, potResult.Winners)
			potResult.OddChip = odd
			if winner := hand.playerByID(odd); winner != nil {
				winner.Stack += remainder
			}
		}
		result.Pots = append(result.Pots, potResult)
	}

	for _, playerID := range revealOrder {
		if revealed, ok := revealedByID[playerID]; ok {
			result.Revealed = append(result.Revealed, revealed)
		}
	}
	return result
}

func eligiblePlayersForPot(hand *Hand, pot Pot) []*Player {
	eligible := make([]*Player, 0, len(pot.Eligible))
	for _, id := range pot.Eligible {
		player := hand.playerByID(id)
		if player == nil || player.Status == StatusFolded || player.Status == StatusOut || player.Status == StatusSittingOut {
			continue
		}
		eligible = append(eligible, player)
	}
	return eligible
}

func showdownOrder(hand *Hand) []PlayerID {
	contenders := make([]*Player, 0, len(hand.Players))
	for _, player := range hand.Players {
		if player == nil || player.Status == StatusFolded || player.Status == StatusOut || player.Status == StatusSittingOut {
			continue
		}
		contenders = append(contenders, player)
	}
	if len(contenders) == 0 {
		return nil
	}
	startSeat := hand.ShowdownStartSeat
	if startSeat < 0 {
		if hand.LastAggressor != 0 {
			if aggressor := hand.playerByID(hand.LastAggressor); aggressor != nil {
				startSeat = aggressor.SeatIndex
			}
		}
		if startSeat < 0 {
			startSeat = nextOccupiedSeat(hand.Players, hand.DealerSeat)
		}
	}
	order := make([]PlayerID, 0, len(contenders))
	visited := make(map[PlayerID]bool, len(contenders))
	max := maxSeat(hand.Players)
	for offset := 0; offset <= max; offset++ {
		seat := (startSeat + offset) % (max + 1)
		player := hand.playerAtSeat(seat)
		if player == nil || player.Status == StatusFolded || player.Status == StatusOut || player.Status == StatusSittingOut {
			continue
		}
		if !visited[player.ID] {
			visited[player.ID] = true
			order = append(order, player.ID)
		}
	}
	return order
}

func oddChipWinner(hand *Hand, winners []PlayerID) PlayerID {
	if len(winners) == 0 {
		return 0
	}
	sort.Slice(winners, func(i, j int) bool {
		left := hand.playerByID(winners[i])
		right := hand.playerByID(winners[j])
		if left == nil || right == nil {
			return winners[i] < winners[j]
		}
		leftDistance := oddChipDistance(hand.DealerSeat, left.SeatIndex, maxSeat(hand.Players)+1)
		rightDistance := oddChipDistance(hand.DealerSeat, right.SeatIndex, maxSeat(hand.Players)+1)
		if leftDistance == rightDistance {
			return winners[i] < winners[j]
		}
		return leftDistance < rightDistance
	})
	return winners[0]
}

func oddChipDistance(from, to, tableSize int) int {
	if tableSize <= 0 {
		return 0
	}
	distance := seatDistance(from, to, tableSize)
	if distance == 0 {
		return tableSize
	}
	return distance
}

func seatDistance(from, to, tableSize int) int {
	if tableSize <= 0 {
		return 0
	}
	if to >= from {
		return to - from
	}
	return tableSize - from + to
}
