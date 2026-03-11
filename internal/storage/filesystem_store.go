package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileSystemStore struct {
	rootDir            string
	timeAnchorProvider TimeAnchorProvider
}

func NewFileSystemStore(rootDir string, provider TimeAnchorProvider) (*FileSystemStore, error) {
	if rootDir == "" {
		return nil, fmt.Errorf("root directory is required")
	}
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return nil, err
	}
	if provider == nil {
		local := NewLocalTimeAnchorProvider()
		provider = local
	}
	return &FileSystemStore{rootDir: rootDir, timeAnchorProvider: provider}, nil
}

func (s *FileSystemStore) RootDir() string {
	return s.rootDir
}

func (s *FileSystemStore) TimeAnchorProvider() TimeAnchorProvider {
	return s.timeAnchorProvider
}

func (s *FileSystemStore) LoadConfigArtifact() (Artifact[Config], error) {
	err := s.ensureConfigMigrated()
	if err != nil && !errors.Is(err, ErrArtifactNotFound) {
		return Artifact[Config]{}, err
	}
	artifact, err := readArtifactFile[Config](s.configArtifactPath())
	if err != nil {
		return Artifact[Config]{}, err
	}
	artifact.Payload.SoundVolume = clampInt(artifact.Payload.SoundVolume, 0, 100)
	return artifact, nil
}

func (s *FileSystemStore) SaveConfigArtifact(cfg Config) (Artifact[Config], error) {
	return s.saveConfigArtifact(cfg, "")
}

func (s *FileSystemStore) saveConfigArtifact(cfg Config, legacySource string) (Artifact[Config], error) {
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	return upsertArtifact(s, s.configArtifactPath(), ArtifactKindConfig, artifactVersionConfig, "local/config", "current", cfg, legacySource)
}

func (s *FileSystemStore) LoadStatsArtifact() (Artifact[StatsStore], error) {
	err := s.ensureStatsMigrated()
	if err != nil && !errors.Is(err, ErrArtifactNotFound) {
		return Artifact[StatsStore]{}, err
	}
	return readArtifactFile[StatsStore](s.statsArtifactPath())
}

func (s *FileSystemStore) SaveStatsArtifact(store StatsStore) (Artifact[StatsStore], error) {
	return s.saveStatsArtifact(store, "")
}

func (s *FileSystemStore) saveStatsArtifact(store StatsStore, legacySource string) (Artifact[StatsStore], error) {
	return upsertArtifact(s, s.statsArtifactPath(), ArtifactKindStatsStore, artifactVersionStatsStore, "sandbox/stats/aggregates", "current", store, legacySource)
}

func (s *FileSystemStore) LoadSaveArtifact(slot int) (Artifact[SaveSlot], error) {
	if err := validateSlot(slot); err != nil {
		return Artifact[SaveSlot]{}, err
	}
	err := s.ensureSaveMigrated(slot)
	if err != nil && !errors.Is(err, ErrArtifactNotFound) {
		return Artifact[SaveSlot]{}, err
	}
	return readArtifactFile[SaveSlot](s.saveArtifactPath(slot))
}

func (s *FileSystemStore) SaveSaveArtifact(slot int, save SaveSlot) (Artifact[SaveSlot], error) {
	return s.saveSaveArtifact(slot, save, "")
}

func (s *FileSystemStore) saveSaveArtifact(slot int, save SaveSlot, legacySource string) (Artifact[SaveSlot], error) {
	if err := validateSlot(slot); err != nil {
		return Artifact[SaveSlot]{}, err
	}
	id := fmt.Sprintf("slot-%d", slot)
	return upsertArtifact(s, s.saveArtifactPath(slot), ArtifactKindSaveSlot, artifactVersionSaveSlot, "sandbox/saves", id, save, legacySource)
}

func (s *FileSystemStore) DeleteSaveArtifact(slot int) error {
	if err := validateSlot(slot); err != nil {
		return err
	}
	newPath := s.saveArtifactPath(slot)
	legacyPath := s.legacySavePath(slot)
	if exists(newPath) {
		if err := os.Remove(newPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if exists(legacyPath) {
		if err := os.Remove(legacyPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func (s *FileSystemStore) ListSaveArtifacts(slots int) ([]SaveInfo, error) {
	if slots <= 0 {
		slots = DefaultSaveSlotCount
	}
	infos := make([]SaveInfo, 0, slots)
	var firstErr error
	for slot := 1; slot <= slots; slot++ {
		err := s.ensureSaveMigrated(slot)
		switch {
		case err == nil:
			artifact, readErr := readArtifactFile[SaveSlot](s.saveArtifactPath(slot))
			if readErr != nil {
				if firstErr == nil {
					firstErr = readErr
				}
				infos = append(infos, SaveInfo{Slot: slot, Incompatible: true, Error: readErr.Error()})
				continue
			}
			infos = append(infos, saveInfoFromArtifact(slot, artifact))
		case errors.Is(err, ErrArtifactNotFound):
			infos = append(infos, SaveInfo{Slot: slot, Empty: true})
		default:
			if firstErr == nil {
				firstErr = err
			}
			infos = append(infos, SaveInfo{Slot: slot, Incompatible: true, Error: err.Error()})
		}
	}
	return infos, firstErr
}

func (s *FileSystemStore) SaveTranscriptChunkArtifact(chunk TranscriptChunk) (Artifact[TranscriptChunk], error) {
	path := s.transcriptChunkPath(chunk.TranscriptID, chunk.ID)
	return upsertArtifact(s, path, ArtifactKindTranscriptChunk, artifactVersionTranscriptChunk, filepath.ToSlash(filepath.Join("sandbox/transcripts", chunk.TranscriptID)), chunk.ID, chunk, "")
}

func (s *FileSystemStore) LoadTranscriptChunkArtifact(transcriptID, chunkID string) (Artifact[TranscriptChunk], error) {
	return readArtifactFile[TranscriptChunk](s.transcriptChunkPath(transcriptID, chunkID))
}

func upsertArtifact[T any](s *FileSystemStore, path string, kind ArtifactKind, version int, namespace, id string, payload T, legacySource string) (Artifact[T], error) {
	var createdAt TimeAnchor
	if existing, err := readArtifactFile[T](path); err == nil {
		createdAt = existing.Metadata.CreatedAt
	} else if !errors.Is(err, ErrArtifactNotFound) {
		return Artifact[T]{}, err
	}
	anchor, err := s.timeAnchorProvider.Now()
	if err != nil {
		return Artifact[T]{}, err
	}
	if createdAt.Timestamp.IsZero() {
		createdAt = anchor
	}
	artifact := Artifact[T]{
		Metadata: ArtifactMetadata{
			Kind:         kind,
			Version:      version,
			Namespace:    namespace,
			ID:           id,
			Encoding:     ArtifactEncodingJSON,
			CreatedAt:    createdAt.normalized(),
			UpdatedAt:    anchor.normalized(),
			LegacySource: legacySource,
		},
		Payload: payload,
	}
	if err := writeArtifactFile(path, artifact); err != nil {
		return Artifact[T]{}, err
	}
	return artifact, nil
}

func readArtifactFile[T any](path string) (Artifact[T], error) {
	var artifact Artifact[T]
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Artifact[T]{}, ErrArtifactNotFound
		}
		return Artifact[T]{}, err
	}
	if err := json.Unmarshal(data, &artifact); err != nil {
		return Artifact[T]{}, &CompatibilityError{
			Stage:      "artifact_decode",
			SourcePath: path,
			Err:        err,
		}
	}
	return artifact, nil
}

func writeArtifactFile[T any](path string, artifact Artifact[T]) error {
	if err := ensureParentDir(path); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func (s *FileSystemStore) configArtifactPath() string {
	return filepath.Join(s.rootDir, "local", "config", "current.json")
}

func (s *FileSystemStore) statsArtifactPath() string {
	return filepath.Join(s.rootDir, "sandbox", "stats", "aggregates", "current.json")
}

func (s *FileSystemStore) saveArtifactPath(slot int) string {
	return filepath.Join(s.rootDir, "sandbox", "saves", fmt.Sprintf("slot-%d.json", slot))
}

func (s *FileSystemStore) migrationManifestPath(id string) string {
	return filepath.Join(s.rootDir, "local", "migrations", id+".json")
}

func (s *FileSystemStore) transcriptChunkPath(transcriptID, chunkID string) string {
	return filepath.Join(s.rootDir, "sandbox", "transcripts", transcriptID, chunkID+".json")
}

func (s *FileSystemStore) transcriptHeadPath(transcriptID string) string {
	return filepath.Join(s.rootDir, "sandbox", "transcripts", transcriptID, "head.json")
}

func (s *FileSystemStore) sessionSummaryPath(sessionID string) string {
	return filepath.Join(s.rootDir, "sandbox", "summaries", sessionID+".json")
}

func (s *FileSystemStore) legacyConfigPath() string {
	return filepath.Join(s.rootDir, "config.json")
}

func (s *FileSystemStore) legacyStatsPath() string {
	return filepath.Join(s.rootDir, "stats.gob")
}

func (s *FileSystemStore) legacySavePath(slot int) string {
	return filepath.Join(s.rootDir, "saves", fmt.Sprintf("slot_%d.gob", slot))
}

func validateSlot(slot int) error {
	if slot < 1 || slot > DefaultSaveSlotCount {
		return fmt.Errorf("slot must be between 1 and %d", DefaultSaveSlotCount)
	}
	return nil
}

func saveInfoFromArtifact(slot int, artifact Artifact[SaveSlot]) SaveInfo {
	save := artifact.Payload
	return SaveInfo{
		Slot:      slot,
		Empty:     false,
		Name:      save.Name,
		Mode:      save.Mode,
		HandNum:   save.HandNumber,
		Stack:     save.PlayerStack,
		Timestamp: artifact.Metadata.UpdatedAt.Timestamp,
	}
}
