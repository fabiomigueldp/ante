package storage

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrArtifactNotFound         = errors.New("artifact not found")
	ErrLegacyFormatUnrecognized = errors.New("legacy format unrecognized")
)

type CompatibilityError struct {
	Kind       ArtifactKind
	SourcePath string
	TargetPath string
	Stage      string
	Err        error
}

func (e *CompatibilityError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s compatibility error during %s: %v", e.Kind, e.Stage, e.Err)
}

func (e *CompatibilityError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (s *FileSystemStore) ensureConfigMigrated() error {
	if exists(s.configArtifactPath()) {
		return nil
	}
	legacyPath := s.legacyConfigPath()
	if !exists(legacyPath) {
		return ErrArtifactNotFound
	}
	raw, err := os.ReadFile(legacyPath)
	if err != nil {
		return err
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return &CompatibilityError{
			Kind:       ArtifactKindConfig,
			SourcePath: legacyPath,
			TargetPath: s.configArtifactPath(),
			Stage:      "legacy_decode",
			Err:        fmt.Errorf("%w: %v", ErrLegacyFormatUnrecognized, err),
		}
	}
	if _, err := s.saveConfigArtifact(cfg, legacyPath); err != nil {
		return err
	}
	return s.recordMigration("legacy_config_json", legacyPath, ArtifactKindConfig, s.configArtifactPath(), "migrated")
}

func (s *FileSystemStore) ensureStatsMigrated() error {
	if exists(s.statsArtifactPath()) {
		return nil
	}
	legacyPath := s.legacyStatsPath()
	if !exists(legacyPath) {
		return ErrArtifactNotFound
	}
	file, err := os.Open(legacyPath)
	if err != nil {
		return err
	}
	defer file.Close()
	var store StatsStore
	if err := gob.NewDecoder(file).Decode(&store); err != nil {
		return &CompatibilityError{
			Kind:       ArtifactKindStatsStore,
			SourcePath: legacyPath,
			TargetPath: s.statsArtifactPath(),
			Stage:      "legacy_decode",
			Err:        fmt.Errorf("%w: %v", ErrLegacyFormatUnrecognized, err),
		}
	}
	if _, err := s.saveStatsArtifact(store, legacyPath); err != nil {
		return err
	}
	return s.recordMigration("legacy_stats_gob", legacyPath, ArtifactKindStatsStore, s.statsArtifactPath(), "migrated")
}

func (s *FileSystemStore) ensureSaveMigrated(slot int) error {
	if exists(s.saveArtifactPath(slot)) {
		return nil
	}
	legacyPath := s.legacySavePath(slot)
	if !exists(legacyPath) {
		return ErrArtifactNotFound
	}
	file, err := os.Open(legacyPath)
	if err != nil {
		return err
	}
	defer file.Close()
	var save SaveSlot
	if err := gob.NewDecoder(file).Decode(&save); err != nil {
		return &CompatibilityError{
			Kind:       ArtifactKindSaveSlot,
			SourcePath: legacyPath,
			TargetPath: s.saveArtifactPath(slot),
			Stage:      "legacy_decode",
			Err:        fmt.Errorf("%w: %v", ErrLegacyFormatUnrecognized, err),
		}
	}
	if _, err := s.saveSaveArtifact(slot, save, legacyPath); err != nil {
		return err
	}
	return s.recordMigration("legacy_save_gob", legacyPath, ArtifactKindSaveSlot, s.saveArtifactPath(slot), "migrated")
}

func (s *FileSystemStore) recordMigration(sourceKind, sourcePath string, targetKind ArtifactKind, targetPath, status string) error {
	anchor, err := s.timeAnchorProvider.Now()
	if err != nil {
		return err
	}
	rawID := fmt.Sprintf("%s|%s|%s|%d", sourceKind, sourcePath, targetPath, anchor.Timestamp.UnixNano())
	sum := sha256.Sum256([]byte(rawID))
	id := fmt.Sprintf("mig_%s_%x", sourceKind, sum[:6])
	manifest := Artifact[MigrationManifest]{
		Metadata: ArtifactMetadata{
			Kind:      ArtifactKindMigrationManifest,
			Version:   migrationManifestVersion,
			Namespace: "local/migrations",
			ID:        id,
			Encoding:  ArtifactEncodingJSON,
			CreatedAt: anchor,
			UpdatedAt: anchor,
		},
		Payload: MigrationManifest{
			ID:         id,
			SourceKind: sourceKind,
			SourcePath: sourcePath,
			TargetKind: string(targetKind),
			TargetPath: targetPath,
			Status:     status,
			ExecutedAt: anchor,
		},
	}
	return writeArtifactFile(s.migrationManifestPath(id), manifest)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureParentDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}
