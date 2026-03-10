package engine

import "testing"

func TestTournamentBlindIncrease(t *testing.T) {
	players := []*Player{{ID: 1, Stack: 100, SeatIndex: 0}, {ID: 2, Stack: 100, SeatIndex: 1}}
	table, _ := NewTable(ModeTournament, 2, TournamentBlinds("turbo"), 1, players)
	tournament := NewTournament(table, 100)
	var event *BlindLevelChangedEvent
	for i := 0; i < 6; i++ {
		event = tournament.CheckBlindIncrease()
	}
	if event == nil {
		t.Fatal("expected blind increase event after 6 hands in turbo")
	}
}
