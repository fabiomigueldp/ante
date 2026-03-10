package engine

type CashGame struct {
	Table  *Table
	BuyIn  int
	Profit map[PlayerID]int
}

func NewCashGame(table *Table, buyIn int) *CashGame {
	return &CashGame{Table: table, BuyIn: buyIn, Profit: make(map[PlayerID]int)}
}

func (cg *CashGame) OfferRebuy(id PlayerID) bool {
	player := playerByID(cg.Table.Players, id)
	if player == nil || player.Stack > 0 || player.Status == StatusOut {
		return false
	}
	player.Stack = cg.BuyIn
	player.Status = StatusActive
	return true
}

func (cg *CashGame) ReplacePlayer(seatIndex int, newPlayer *Player) {
	if cg.Table == nil || newPlayer == nil {
		return
	}
	newPlayer.SeatIndex = seatIndex
	for i, player := range cg.Table.Players {
		if player != nil && player.SeatIndex == seatIndex {
			cg.Table.Players[i] = newPlayer
			return
		}
	}
	cg.Table.Players = append(cg.Table.Players, newPlayer)
}

func (cg *CashGame) CashOut(id PlayerID) int {
	player := playerByID(cg.Table.Players, id)
	if player == nil {
		return 0
	}
	profit := player.Stack - cg.BuyIn
	cg.Profit[id] += profit
	return profit
}
