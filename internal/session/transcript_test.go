package session

import (
	"bytes"
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestTranscriptCheckpointHashChainsAcrossHands(t *testing.T) {
	store, _, _ := newSessionTestStore(t)
	useSessionDependenciesForTest(t, Dependencies{ArtifactStore: store, TimeAnchorProvider: store.TimeAnchorProvider()})
	sess, err := New(Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyEasy,
		Seats:         3,
		StartingStack: 40,
		PlayerName:    "Hero",
		Seed:          77,
	})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	playCompletedHands(t, sess, 2)
	transcriptID, err := storage.TranscriptIDFromSessionID(sess.SessionID)
	if err != nil {
		t.Fatalf("TranscriptIDFromSessionID error: %v", err)
	}
	chunks, err := store.ListTranscriptChunkArtifacts(transcriptID)
	if err != nil {
		t.Fatalf("ListTranscriptChunkArtifacts error: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	first := chunks[0].Payload
	second := chunks[1].Payload
	if !bytes.Equal(second.PreviousHash.Sum, first.CheckpointHash.Sum) {
		t.Fatalf("second previous hash = %x, want %x", second.PreviousHash.Sum, first.CheckpointHash.Sum)
	}
	if second.PreviousHash.Algorithm != first.CheckpointHash.Algorithm {
		t.Fatalf("algorithm = %q, want %q", second.PreviousHash.Algorithm, first.CheckpointHash.Algorithm)
	}
	if first.CheckpointID == "" || second.CheckpointID == "" {
		t.Fatal("expected checkpoint ids to be populated")
	}
}

func playCompletedHands(t *testing.T, sess *Session, handCount int) {
	t.Helper()
	for i := 0; i < handCount; i++ {
		playExactlyOneHand(t, sess)
		sess.currentHand = nil
	}
}
