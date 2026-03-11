package storage

import (
	"fmt"
	"sort"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func ListHistorySessionsResult() ([]HistorySessionEntry, error) {
	store := DefaultArtifactStore()
	summaries, err := store.ListSessionSummaryArtifacts()
	if err != nil {
		return nil, err
	}
	summaryByTranscript := make(map[string]SessionSummary, len(summaries))
	for _, artifact := range summaries {
		summaryByTranscript[artifact.Payload.TranscriptID] = artifact.Payload
	}
	heads, err := store.ListTranscriptHeadArtifacts()
	if err != nil {
		return nil, err
	}
	entries := make([]HistorySessionEntry, 0, len(heads))
	for _, artifact := range heads {
		head := artifact.Payload
		summary, completed := summaryByTranscript[head.TranscriptID]
		entry := HistorySessionEntry{
			SessionID:    head.SessionID,
			TranscriptID: head.TranscriptID,
			PlayerName:   head.PlayerName,
			Mode:         head.Mode,
			HandsPlayed:  head.HandsPlayed,
			UpdatedAt:    head.UpdatedAt.Timestamp,
			Completed:    completed,
			ResultLabel:  inProgressResultLabel(head.HandsPlayed),
		}
		if completed {
			entry.ResultLabel = summary.ResultLabel
			entry.HandsPlayed = summary.HandsPlayed
			entry.UpdatedAt = summary.EndTime.Timestamp
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})
	return entries, nil
}

func ListSessionHandsResult(transcriptID string) ([]HistoryHandEntry, error) {
	artifacts, err := DefaultArtifactStore().ListTranscriptChunkArtifacts(transcriptID)
	if err != nil {
		return nil, err
	}
	entries := make([]HistoryHandEntry, 0, len(artifacts))
	for _, artifact := range artifacts {
		chunk := artifact.Payload
		entries = append(entries, HistoryHandEntry{
			SessionID:    chunk.SessionID,
			TranscriptID: chunk.TranscriptID,
			ChunkID:      chunk.ID,
			SnapshotID:   chunk.SnapshotID,
			CheckpointID: chunk.CheckpointID,
			HandID:       chunk.HandID,
			ChunkIndex:   chunk.ChunkIndex,
			DealerSeat:   chunk.DealerSeat,
			Blinds:       chunk.Blinds,
			Players:      append([]engine.PlayerSnapshot(nil), chunk.Players...),
			ResultLabel:  chunk.ResultLabel,
			CommittedAt:  chunk.CommittedAt.Timestamp,
			FinalBoard:   append([]engine.Card(nil), chunk.FinalBoard...),
		})
	}
	return entries, nil
}

func LoadReplayChunkResult(transcriptID, chunkID string) (*TranscriptChunk, error) {
	artifact, err := DefaultArtifactStore().LoadTranscriptChunkArtifact(transcriptID, chunkID)
	if err != nil {
		return nil, err
	}
	chunk := artifact.Payload
	return &chunk, nil
}

func LoadSessionSummaryResult(sessionID string) (*SessionSummary, error) {
	artifact, err := DefaultArtifactStore().LoadSessionSummaryArtifact(sessionID)
	if err != nil {
		return nil, err
	}
	summary := artifact.Payload
	return &summary, nil
}

func inProgressResultLabel(handsPlayed int) string {
	if handsPlayed <= 0 {
		return "In progress"
	}
	return fmt.Sprintf("In progress (%d hands)", handsPlayed)
}
