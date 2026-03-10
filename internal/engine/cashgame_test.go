package engine

import "testing"

func TestCashGameRebuy(t *testing.T) {
	players := []*Player{{ID: 1, Stack: 0, SeatIndex: 0}}
	table, _ := NewTable(ModeCashGame, 1, CashGameBlinds(1, 2), 1, players)
	cash := NewCashGame(table, 200)
	if !cash.OfferRebuy(1) {
		t.Fatal("expected rebuy to succeed")
	}
	if players[0].Stack != 200 {
		t.Fatalf("expected stack 200, got %d", players[0].Stack)
	}
}
