package tui

import (
	"testing"
	"time"

	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestReplayRevealsBoardStreetByStreetFromTranscript(t *testing.T) {
	anchor := storage.TimeAnchor{Timestamp: time.Date(2026, time.March, 11, 10, 30, 0, 0, time.UTC), Source: "test_clock"}
	chunk := sampleReplayChunk(anchor)
	m := NewReplayModel(&chunk)

	if len(m.visibleBoard()) != 0 {
		t.Fatal("expected no board before stepping transcript")
	}
	m.step = 2
	flop := m.visibleBoard()
	if len(flop) != 3 {
		t.Fatalf("len(flop) = %d, want 3", len(flop))
	}
	m.step = 3
	turn := m.visibleBoard()
	if len(turn) != 4 || turn[3] != engine.NewCard(engine.Jack, engine.Diamonds) {
		t.Fatalf("turn board = %+v", turn)
	}
	m.step = 4
	river := m.visibleBoard()
	if len(river) != 5 || river[4] != engine.NewCard(engine.Ten, engine.Spades) {
		t.Fatalf("river board = %+v", river)
	}
}
