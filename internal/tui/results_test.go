package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	if !strings.Contains(view, "SESSION") || !strings.Contains(view, "HIGHLIGHTS") || !strings.Contains(view, "INTEGRITY") {
		t.Fatalf("expected structured results sections, got:\n%s", view)
	}
	if !strings.Contains(view, "Straight Flush") || !strings.Contains(view, "88") || !strings.Contains(view, "3") {
		t.Fatalf("expected authoritative summary fields in results view, got:\n%s", view)
	}
	if !strings.Contains(view, "[Enter] Return to Menu") {
		t.Fatalf("expected strict enter instruction in results view, got:\n%s", view)
	}
}

func TestResultsModelOnlyExitsOnEnter(t *testing.T) {
	m := NewResultsModel("Done", nil)
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Fatal("expected escape to no longer leave the results screen")
	}
	if _, ok := model.(ResultsModel); !ok {
		t.Fatalf("expected ResultsModel, got %T", model)
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return to the menu")
	}
}

func TestResultsModelDisplaysForfeitedSummary(t *testing.T) {
	summary := &storage.SessionSummary{
		Mode:              "tournament",
		HandsPlayed:       6,
		ResultLabel:       "Forfeited (#4 of 4)",
		TerminationReason: "forfeited",
		BiggestPot:        40,
		BestHand:          "Two Pair",
		LargestWin:        18,
		LongestStreak:     2,
		CheckpointID:      "ckp_forfeit",
	}
	m := ResultsModel{result: "You forfeited the session in position #4.", summary: summary}
	view := m.View()
	if !strings.Contains(view, "Forfeited (#4 of 4)") {
		t.Fatalf("expected forfeited label in results view, got:\n%s", view)
	}
}
