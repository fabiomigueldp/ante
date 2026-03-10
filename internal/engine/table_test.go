package engine

import "testing"

func TestTableNextHandDeterministicSeed(t *testing.T) {
	players := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 100, SeatIndex: 1}}
	table, err := NewTable(ModeTournament, 2, TournamentBlinds("normal"), 1000, players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h1 := table.NextHand()
	if h1 == nil {
		t.Fatal("expected hand")
	}
	h2 := table.NextHand()
	if h2 == nil {
		t.Fatal("expected second hand")
	}
	if h1.Seed() == h2.Seed() {
		t.Fatal("expected different hand seeds")
	}
}

func TestTournamentZeroStackMarkedOutBeforeNextHand(t *testing.T) {
	players := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 0, SeatIndex: 1}}
	table, err := NewTable(ModeTournament, 2, TournamentBlinds("normal"), 1000, players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if table.NextHand() != nil {
		t.Fatal("expected no hand when one tournament player remains")
	}
	if players[1].Status != StatusOut {
		t.Fatalf("expected busted player to be marked out, got %v", players[1].Status)
	}
}

func TestApplyHandResultsPersistsStacksWithoutLeakingHandState(t *testing.T) {
	players := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 100, SeatIndex: 1}}
	table, err := NewTable(ModeTournament, 2, TournamentBlinds("normal"), 1000, players)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hand := table.NextHand()
	if hand == nil {
		t.Fatal("expected hand")
	}
	handPlayer1 := playerByID(hand.Players, 1)
	handPlayer2 := playerByID(hand.Players, 2)
	if handPlayer1 == nil || handPlayer2 == nil {
		t.Fatal("expected cloned hand players")
	}

	handPlayer1.Stack = 135
	handPlayer1.Bet = 12
	handPlayer1.TotalBet = 12
	handPlayer1.HoleCards = [2]Card{NewCard(Ace, Spades), NewCard(King, Hearts)}
	handPlayer1.Status = StatusActive
	handPlayer2.Stack = 65
	handPlayer2.Status = StatusActive

	table.ApplyHandResults(hand)

	if players[0].Stack != 135 {
		t.Fatalf("player 1 stack = %d, want 135", players[0].Stack)
	}
	if players[1].Stack != 65 {
		t.Fatalf("player 2 stack = %d, want 65", players[1].Stack)
	}
	if players[0].Bet != 0 || players[0].TotalBet != 0 {
		t.Fatalf("expected hand-local bets cleared on table player, got bet=%d total=%d", players[0].Bet, players[0].TotalBet)
	}
	if players[0].HoleCards != [2]Card{} {
		t.Fatal("expected table player hole cards cleared after reconciliation")
	}
}
