package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestResultsModelUsesAuthoritativeSessionSummary(t *testing.T) {
	store, _, anchor := newTUITestStore(t)
	restore := storage.SetDefaultArtifactStoreForTest(store)
	defer restore()
	summary := storage.SessionSummary{
		ID:            "ses_result",
		SessionID:     "ses_result",
		TranscriptID:  "trn_result",
		CheckpointID:  "ckp_result",
		PlayerName:    "Hero",
		Mode:          "tournament",
		StartTime:     anchor,
		EndTime:       storage.TimeAnchor{Timestamp: anchor.Timestamp.Add(5 * time.Minute), Source: anchor.Source},
		HandsPlayed:   12,
		BiggestPot:    88,
		BestHand:      "Straight Flush",
		LargestWin:    54,
		LongestStreak: 3,
		ResultLabel:   "Winner!",
	}
	if _, err := store.SaveSessionSummaryArtifact(summary); err != nil {
		t.Fatalf("SaveSessionSummaryArtifact error: %v", err)
	}
	sess := &session.Session{SessionID: summary.SessionID}
	m := NewResultsModel("Winner!", sess)
	view := m.View()
	if !strings.Contains(view, "Straight Flush") || !strings.Contains(view, "88") || !strings.Contains(view, "3") {
		t.Fatalf("expected authoritative summary fields in results view, got:\n%s", view)
	}
}
