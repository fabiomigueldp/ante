package storage

import "testing"

func TestAggregateStatsFromSummariesDoesNotCountForfeitsAsWins(t *testing.T) {
	store := aggregateStatsFromSummaries([]Artifact[SessionSummary]{
		{Payload: SessionSummary{SessionID: "ses_forfeit", Mode: "tournament", FinalPosition: 4, ResultLabel: "Forfeited (#4 of 4)", TerminationReason: "forfeited"}},
		{Payload: SessionSummary{SessionID: "ses_win", Mode: "tournament", FinalPosition: 1, ResultLabel: "Winner!", TerminationReason: "completed"}},
	})
	if store.TournamentWins() != 1 {
		t.Fatalf("TournamentWins = %d, want 1", store.TournamentWins())
	}
}
