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
