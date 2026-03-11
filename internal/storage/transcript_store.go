package storage

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

func (s *FileSystemStore) SaveTranscriptHeadArtifact(head TranscriptHead) (Artifact[TranscriptHead], error) {
	path := s.transcriptHeadPath(head.TranscriptID)
	return upsertArtifact(s, path, ArtifactKindTranscriptHead, artifactVersionTranscriptHead, filepath.ToSlash(filepath.Join("sandbox/transcripts", head.TranscriptID)), head.TranscriptID, head, "")
}

func (s *FileSystemStore) LoadTranscriptHeadArtifact(transcriptID string) (Artifact[TranscriptHead], error) {
	return readArtifactFile[TranscriptHead](s.transcriptHeadPath(transcriptID))
}

func (s *FileSystemStore) ListTranscriptHeadArtifacts() ([]Artifact[TranscriptHead], error) {
	root := filepath.Join(s.rootDir, "sandbox", "transcripts")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	artifacts := make([]Artifact[TranscriptHead], 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		artifact, readErr := readArtifactFile[TranscriptHead](filepath.Join(root, entry.Name(), "head.json"))
		if readErr != nil {
			return nil, readErr
		}
		artifacts = append(artifacts, artifact)
	}
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].Payload.UpdatedAt.Timestamp.After(artifacts[j].Payload.UpdatedAt.Timestamp)
	})
	return artifacts, nil
}

func (s *FileSystemStore) ListTranscriptChunkArtifacts(transcriptID string) ([]Artifact[TranscriptChunk], error) {
	root := filepath.Join(s.rootDir, "sandbox", "transcripts", transcriptID)
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	artifacts := make([]Artifact[TranscriptChunk], 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "head.json" || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		artifact, readErr := readArtifactFile[TranscriptChunk](filepath.Join(root, entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		artifacts = append(artifacts, artifact)
	}
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].Payload.ChunkIndex < artifacts[j].Payload.ChunkIndex
	})
	return artifacts, nil
}

func (s *FileSystemStore) SaveSessionSummaryArtifact(summary SessionSummary) (Artifact[SessionSummary], error) {
	path := s.sessionSummaryPath(summary.SessionID)
	return upsertArtifact(s, path, ArtifactKindSessionSummary, artifactVersionSessionSummary, "sandbox/summaries", summary.SessionID, summary, "")
}

func (s *FileSystemStore) LoadSessionSummaryArtifact(sessionID string) (Artifact[SessionSummary], error) {
	return readArtifactFile[SessionSummary](s.sessionSummaryPath(sessionID))
}

func (s *FileSystemStore) ListSessionSummaryArtifacts() ([]Artifact[SessionSummary], error) {
	root := filepath.Join(s.rootDir, "sandbox", "summaries")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	artifacts := make([]Artifact[SessionSummary], 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		artifact, readErr := readArtifactFile[SessionSummary](filepath.Join(root, entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		artifacts = append(artifacts, artifact)
	}
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].Payload.EndTime.Timestamp.After(artifacts[j].Payload.EndTime.Timestamp)
	})
	return artifacts, nil
}
