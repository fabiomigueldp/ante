package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestStatsViewUsesSummaryBackedReadModel(t *testing.T) {
	store, _, anchor := newTUITestStore(t)
	restore := storage.SetDefaultArtifactStoreForTest(store)
	defer restore()
	summary := storage.SessionSummary{
		ID:            "ses_stats",
		SessionID:     "ses_stats",
		TranscriptID:  "trn_stats",
		PlayerName:    "Hero",
		Mode:          "tournament",
		StartTime:     anchor,
		EndTime:       storage.TimeAnchor{Timestamp: anchor.Timestamp.Add(10 * time.Minute), Source: anchor.Source},
		HandsPlayed:   6,
		FinalPosition: 1,
		TotalPlayers:  4,
		ChipsWon:      32,
		BiggestPot:    22,
		HandsWon:      3,
		BestHand:      "Full House",
	}
	if _, err := store.SaveSessionSummaryArtifact(summary); err != nil {
		t.Fatalf("SaveSessionSummaryArtifact error: %v", err)
	}
	m := NewStatsViewModel()
	if m.store.TotalSessions() != 1 {
		t.Fatalf("TotalSessions = %d, want 1", m.store.TotalSessions())
	}
	view := m.View()
	if !strings.Contains(view, "Full House") || !strings.Contains(view, "32") {
		t.Fatalf("expected summary-backed stats in view, got:\n%s", view)
	}
}
