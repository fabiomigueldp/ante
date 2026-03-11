package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestHistoryAndReplayWorkWithEmptyLegacyStatsIfTranscriptExists(t *testing.T) {
	store, _, anchor := newTUITestStore(t)
	restore := storage.SetDefaultArtifactStoreForTest(store)
	defer restore()
	transcriptID := "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	head := storage.TranscriptHead{
		Version:            1,
		SessionID:          "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TranscriptID:       transcriptID,
		PlayerName:         "Hero",
		Mode:               "tournament",
		LatestChunkID:      "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001",
		LatestSnapshotID:   "snp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001_000000010",
		LatestCheckpointID: "ckp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001",
		LatestChunkHash:    storage.TranscriptHash{Algorithm: "sha256", Sum: []byte{0x1}},
		LatestSeq:          10,
		HandsPlayed:        1,
		StartedAt:          anchor,
		UpdatedAt:          anchor,
	}
	if _, err := store.SaveTranscriptHeadArtifact(head); err != nil {
		t.Fatalf("SaveTranscriptHeadArtifact error: %v", err)
	}
	chunk := sampleReplayChunk(anchor)
	if _, err := store.SaveTranscriptChunkArtifact(chunk); err != nil {
		t.Fatalf("SaveTranscriptChunkArtifact error: %v", err)
	}
	if _, err := store.SaveStatsArtifact(storage.StatsStore{}); err != nil {
		t.Fatalf("SaveStatsArtifact error: %v", err)
	}

	m := NewHistoryViewModel()
	if len(m.sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(m.sessions))
	}
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(HistoryViewModel)
	if m.mode != historyModeHands {
		t.Fatalf("mode = %d, want historyModeHands", m.mode)
	}
	if len(m.hands) != 1 {
		t.Fatalf("len(hands) = %d, want 1", len(m.hands))
	}
	if cmd != nil {
		t.Fatal("expected no command when opening session hand list")
	}
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected replay navigation command")
	}
	msg := cmd()
	switchMsg, ok := msg.(switchScreenMsg)
	if !ok {
		t.Fatalf("expected switchScreenMsg, got %T", msg)
	}
	if switchMsg.screen != ScreenReplay {
		t.Fatalf("screen = %v, want ScreenReplay", switchMsg.screen)
	}
	replayChunk, ok := switchMsg.data.(*storage.TranscriptChunk)
	if !ok {
		t.Fatalf("expected *storage.TranscriptChunk, got %T", switchMsg.data)
	}
	if replayChunk.ID != chunk.ID {
		t.Fatalf("chunk id = %q, want %q", replayChunk.ID, chunk.ID)
	}
}

func newTUITestStore(t *testing.T) (*storage.FileSystemStore, string, storage.TimeAnchor) {
	t.Helper()
	root := t.TempDir()
	anchor := storage.TimeAnchor{Timestamp: time.Date(2026, time.March, 11, 10, 30, 0, 0, time.UTC), Source: "test_clock"}
	store, err := storage.NewFileSystemStore(root, tuiStaticTimeProvider{anchor: anchor})
	if err != nil {
		t.Fatalf("NewFileSystemStore error: %v", err)
	}
	return store, root, anchor
}

type tuiStaticTimeProvider struct{ anchor storage.TimeAnchor }

func (p tuiStaticTimeProvider) Now() (storage.TimeAnchor, error) { return p.anchor, nil }

func sampleReplayChunk(anchor storage.TimeAnchor) storage.TranscriptChunk {
	return storage.TranscriptChunk{
		Version:          1,
		ID:               "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001",
		SessionID:        "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TranscriptID:     "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ChunkIndex:       1,
		HandID:           1,
		SnapshotID:       "snp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001_000000010",
		CheckpointID:     "ckp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001",
		CheckpointHash:   storage.TranscriptHash{Algorithm: "sha256", Sum: []byte{0x1}},
		Players:          []engine.PlayerSnapshot{{ID: 1, Name: "Hero", Seat: 0, Stack: 100}, {ID: 2, Name: "Bot", Seat: 1, Stack: 100}},
		DealerSeat:       1,
		Blinds:           engine.BlindLevel{SB: 1, BB: 2},
		ResultLabel:      "Won 12",
		WinningPlayerIDs: []engine.PlayerID{1},
		CommittedAt:      anchor,
		Records: []storage.TranscriptRecord{
			{Version: 1, SessionID: "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TranscriptID: "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ChunkID: "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001", Sequence: 1, HandID: 1, Kind: "hand_started", Message: "Hand #1 begins", TimeAnchor: anchor},
			{Version: 1, SessionID: "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TranscriptID: "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ChunkID: "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001", Sequence: 2, HandID: 1, Kind: "street_advanced", Street: engine.StreetFlop, NewCards: []engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Clubs)}, TimeAnchor: anchor},
			{Version: 1, SessionID: "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TranscriptID: "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ChunkID: "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001", Sequence: 3, HandID: 1, Kind: "street_advanced", Street: engine.StreetTurn, NewCards: []engine.Card{engine.NewCard(engine.Jack, engine.Diamonds)}, TimeAnchor: anchor},
			{Version: 1, SessionID: "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TranscriptID: "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ChunkID: "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001", Sequence: 4, HandID: 1, Kind: "street_advanced", Street: engine.StreetRiver, NewCards: []engine.Card{engine.NewCard(engine.Ten, engine.Spades)}, TimeAnchor: anchor},
		},
	}
}
