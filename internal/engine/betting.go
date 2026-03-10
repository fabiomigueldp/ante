package engine

import "fmt"

type BettingRound struct {
	Street             Street
	BigBlind           int
	CurrentBet         int
	MinRaise           int
	LastAggressor      PlayerID
	LastFullRaiseSize  int
	LastFullRaiseTotal int
	ActedPlayers       map[PlayerID]bool
	ReopenedPlayers    map[PlayerID]bool
}

func NewBettingRound(street Street, bigBlind, currentBet int) *BettingRound {
	if bigBlind <= 0 {
		bigBlind = 1
	}
	if currentBet < 0 {
		currentBet = 0
	}
	return &BettingRound{
		Street:             street,
		BigBlind:           bigBlind,
		CurrentBet:         currentBet,
		MinRaise:           bigBlind,
		LastFullRaiseSize:  bigBlind,
		LastFullRaiseTotal: currentBet,
		ActedPlayers:       make(map[PlayerID]bool),
		ReopenedPlayers:    make(map[PlayerID]bool),
	}
}

func (br *BettingRound) ResetForNextStreet(street Street) {
	br.Street = street
	br.CurrentBet = 0
	br.MinRaise = br.BigBlind
	br.LastAggressor = 0
	br.LastFullRaiseSize = br.BigBlind
	br.LastFullRaiseTotal = 0
	br.ActedPlayers = make(map[PlayerID]bool)
	br.ReopenedPlayers = make(map[PlayerID]bool)
}

func (br *BettingRound) LegalActions(player *Player) []LegalAction {
	if player == nil || !player.CanAct() {
		return nil
	}

	toCall := br.CurrentBet - player.Bet
	if toCall < 0 {
		toCall = 0
	}

	legal := make([]LegalAction, 0, 4)
	if toCall == 0 {
		legal = append(legal, LegalAction{Type: ActionCheck})
		minTarget := player.Bet + br.MinRaise
		raiseType := ActionBet
		if br.CurrentBet > 0 {
			raiseType = ActionRaise
			minTarget = br.CurrentBet + br.MinRaise
		}
		if player.Bet+player.Stack >= minTarget {
			legal = append(legal, LegalAction{Type: raiseType, MinAmount: minTarget, MaxAmount: player.Bet + player.Stack})
		}
		if player.Stack > 0 {
			legal = append(legal, LegalAction{Type: ActionAllIn, MinAmount: player.Bet + player.Stack, MaxAmount: player.Bet + player.Stack})
		}
		return legal
	}

	legal = append(legal, LegalAction{Type: ActionFold})
	if player.Stack <= toCall {
		legal = append(legal, LegalAction{Type: ActionAllIn, MinAmount: player.Bet + player.Stack, MaxAmount: player.Bet + player.Stack})
		return legal
	}

	legal = append(legal, LegalAction{Type: ActionCall, MinAmount: toCall, MaxAmount: toCall})
	minTarget := br.CurrentBet + br.MinRaise
	if player.Bet+player.Stack >= minTarget {
		legal = append(legal, LegalAction{Type: ActionRaise, MinAmount: minTarget, MaxAmount: player.Bet + player.Stack})
	}
	legal = append(legal, LegalAction{Type: ActionAllIn, MinAmount: player.Bet + player.Stack, MaxAmount: player.Bet + player.Stack})
	return legal
}

func (br *BettingRound) Apply(player *Player, action Action) (Action, error) {
	if player == nil || !player.CanAct() {
		return Action{}, fmt.Errorf("player cannot act")
	}
	if !actionAllowed(action, br.LegalActions(player)) {
		return Action{}, fmt.Errorf("illegal action %v amount %d", action.Type, action.Amount)
	}

	resolved := Action{PlayerID: player.ID, Type: action.Type}
	toCall := br.CurrentBet - player.Bet
	if toCall < 0 {
		toCall = 0
	}
	wasCurrent := player.Bet == br.CurrentBet

	switch action.Type {
	case ActionFold:
		player.Status = StatusFolded
		br.markActed(player.ID)
		return resolved, nil
	case ActionCheck:
		br.markActed(player.ID)
		return resolved, nil
	case ActionCall:
		resolved.Amount = player.Contribute(toCall)
		br.markActed(player.ID)
		return resolved, nil
	case ActionBet:
		target := clampTarget(player, action.Amount)
		if target < player.Bet+br.MinRaise {
			return Action{}, fmt.Errorf("bet below minimum")
		}
		contrib := player.Contribute(target - player.Bet)
		resolved.Amount = player.Bet
		br.CurrentBet = player.Bet
		br.MinRaise = maxInt(br.BigBlind, contrib)
		br.LastFullRaiseSize = br.MinRaise
		br.LastFullRaiseTotal = br.CurrentBet
		br.LastAggressor = player.ID
		br.reopenAction(player.ID)
		return resolved, nil
	case ActionRaise:
		target := clampTarget(player, action.Amount)
		if target < br.CurrentBet+br.MinRaise {
			return Action{}, fmt.Errorf("raise below minimum")
		}
		previousBet := br.CurrentBet
		player.Contribute(target - player.Bet)
		resolved.Amount = player.Bet
		raiseSize := player.Bet - previousBet
		br.CurrentBet = player.Bet
		br.MinRaise = raiseSize
		br.LastFullRaiseSize = raiseSize
		br.LastFullRaiseTotal = br.CurrentBet
		br.LastAggressor = player.ID
		br.reopenAction(player.ID)
		return resolved, nil
	case ActionAllIn:
		target := player.Bet + player.Stack
		if action.Amount > 0 {
			target = clampTarget(player, action.Amount)
		}
		previousBet := br.CurrentBet
		player.Contribute(target - player.Bet)
		resolved.Amount = player.Bet
		if player.Bet <= previousBet {
			br.markActed(player.ID)
			return resolved, nil
		}
		raiseSize := player.Bet - previousBet
		br.CurrentBet = player.Bet
		if previousBet == 0 || raiseSize >= br.MinRaise {
			br.MinRaise = maxInt(br.BigBlind, raiseSize)
			br.LastFullRaiseSize = br.MinRaise
			br.LastFullRaiseTotal = br.CurrentBet
			br.LastAggressor = player.ID
			br.reopenAction(player.ID)
			return resolved, nil
		}
		if wasCurrent {
			br.markActed(player.ID)
		} else {
			br.markActed(player.ID)
		}
		return resolved, nil
	default:
		return Action{}, fmt.Errorf("unsupported action type %v", action.Type)
	}
}

func (br *BettingRound) IsComplete(players []*Player) bool {
	contenders := 0
	for _, player := range players {
		if player == nil || player.Status == StatusOut || player.Status == StatusSittingOut {
			continue
		}
		if player.Status != StatusFolded {
			contenders++
		}
		if player.CanAct() {
			if player.Bet != br.CurrentBet {
				return false
			}
			if !br.ActedPlayers[player.ID] {
				return false
			}
		}
	}
	if contenders <= 1 {
		return true
	}
	return true
}

func (br *BettingRound) markActed(playerID PlayerID) {
	br.ActedPlayers[playerID] = true
	br.ReopenedPlayers[playerID] = true
}

func (br *BettingRound) reopenAction(aggressor PlayerID) {
	br.ActedPlayers = make(map[PlayerID]bool)
	br.ReopenedPlayers = make(map[PlayerID]bool)
	br.ActedPlayers[aggressor] = true
	br.ReopenedPlayers[aggressor] = true
}

func actionAllowed(action Action, legal []LegalAction) bool {
	for _, candidate := range legal {
		if candidate.Type != action.Type {
			continue
		}
		switch action.Type {
		case ActionFold, ActionCheck, ActionCall:
			return true
		case ActionAllIn:
			return action.Amount == 0 || (action.Amount >= candidate.MinAmount && action.Amount <= candidate.MaxAmount)
		default:
			return action.Amount >= candidate.MinAmount && action.Amount <= candidate.MaxAmount
		}
	}
	return false
}

func clampTarget(player *Player, target int) int {
	if target < player.Bet {
		return player.Bet
	}
	maxTarget := player.Bet + player.Stack
	if target > maxTarget {
		return maxTarget
	}
	return target
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
