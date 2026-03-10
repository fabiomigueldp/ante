package engine

import "fmt"

type GameMode uint8

const (
	ModeTournament GameMode = iota
	ModeCashGame
	ModeHeadsUpDuel
)

type Table struct {
	Mode         GameMode
	Players      []*Player
	Seats        int
	DealerSeat   int
	HandNumber   int
	BlindsConfig BlindStructure
	CurrentLevel int
	MasterSeed   int64
}

func (t *Table) CurrentBlinds() BlindLevel {
	if len(t.BlindsConfig.Levels) == 0 {
		return BlindLevel{Level: 1, SB: 1, BB: 2}
	}
	idx := t.CurrentLevel
	if idx < 0 {
		idx = 0
	}
	if idx >= len(t.BlindsConfig.Levels) {
		idx = len(t.BlindsConfig.Levels) - 1
	}
	return t.BlindsConfig.Levels[idx]
}

func (t *Table) NextHand() *Hand {
	t.prepareForNextHand()
	if t.IsFinished() {
		return nil
	}
	if t.HandNumber == 0 && t.DealerSeat < 0 {
		t.DealerSeat = nextOccupiedSeat(t.Players, -1)
	} else if t.HandNumber > 0 {
		t.AdvanceDealer()
	}
	hand := NewHand(t.HandNumber+1, t.Players, t.DealerSeat, t.CurrentBlinds(), t.handSeed(t.HandNumber+1))
	t.HandNumber++
	return hand
}

func (t *Table) ApplyHandResults(hand *Hand) {
	if t == nil || hand == nil {
		return
	}
	for _, tablePlayer := range t.Players {
		if tablePlayer == nil {
			continue
		}
		handPlayer := playerByID(hand.Players, tablePlayer.ID)
		if handPlayer == nil {
			continue
		}
		tablePlayer.Stack = handPlayer.Stack
		tablePlayer.Bet = 0
		tablePlayer.TotalBet = 0
		tablePlayer.HoleCards = [2]Card{}

		switch {
		case tablePlayer.Status == StatusOut:
			continue
		case tablePlayer.Status == StatusSittingOut || handPlayer.Status == StatusSittingOut:
			tablePlayer.Status = StatusSittingOut
		case handPlayer.Stack == 0 && t.Mode == ModeCashGame:
			tablePlayer.Status = StatusSittingOut
		default:
			tablePlayer.Status = StatusActive
		}
	}
	if t.DealerSeat != hand.DealerSeat {
		t.DealerSeat = hand.DealerSeat
	}
}

func (t *Table) prepareForNextHand() {
	for _, player := range t.Players {
		if player == nil {
			continue
		}
		if player.Stack > 0 {
			if player.Status != StatusOut && player.Status != StatusSittingOut {
				player.Status = StatusActive
			}
			continue
		}
		switch t.Mode {
		case ModeCashGame:
			if player.Status != StatusOut {
				player.Status = StatusSittingOut
			}
		default:
			player.Status = StatusOut
		}
	}
}

func (t *Table) handSeed(handNumber int) int64 {
	return t.MasterSeed + int64(handNumber*7919)
}

func (t *Table) AdvanceDealer() {
	next := nextOccupiedSeat(t.Players, t.DealerSeat)
	if next != -1 {
		t.DealerSeat = next
	}
}

func (t *Table) EliminatePlayer(id PlayerID) {
	for _, player := range t.Players {
		if player != nil && player.ID == id {
			player.Status = StatusOut
			player.Stack = 0
			return
		}
	}
}

func (t *Table) ActivePlayers() []*Player {
	out := make([]*Player, 0, len(t.Players))
	for _, player := range t.Players {
		if player != nil && player.Status != StatusOut && player.Status != StatusSittingOut {
			out = append(out, player)
		}
	}
	return out
}

func (t *Table) IsFinished() bool {
	return len(t.ActivePlayers()) <= 1
}

func NewTable(mode GameMode, seats int, structure BlindStructure, seed int64, players []*Player) (*Table, error) {
	if seats <= 0 {
		return nil, fmt.Errorf("seats must be positive")
	}
	return &Table{
		Mode:         mode,
		Players:      players,
		Seats:        seats,
		DealerSeat:   -1,
		BlindsConfig: structure,
		CurrentLevel: 0,
		MasterSeed:   seed,
	}, nil
}
