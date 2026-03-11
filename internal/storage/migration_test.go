package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestConfigMigrationFromLegacyJSON(t *testing.T) {
	store, root, _ := newTestStore(t)
	legacy := DefaultConfig()
	legacy.PlayerName = "Migrated"
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.json"), raw, 0o644); err != nil {
		t.Fatalf("write legacy config error: %v", err)
	}

	loaded, err := store.LoadConfigArtifact()
	if err != nil {
		t.Fatalf("LoadConfigArtifact error: %v", err)
	}
	if loaded.Payload.PlayerName != "Migrated" {
		t.Fatalf("player name = %q, want Migrated", loaded.Payload.PlayerName)
	}
	if !exists(store.configArtifactPath()) {
		t.Fatal("expected migrated config artifact")
	}
	migrations, err := filepath.Glob(filepath.Join(root, "local", "migrations", "*.json"))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least one migration manifest")
	}
}

func TestStatsMigrationFromLegacyGob(t *testing.T) {
	store, root, anchor := newTestStore(t)
	legacy := StatsStore{Sessions: []SessionStats{{
		ID:        "ses_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Mode:      "cash",
		StartTime: anchor.Timestamp,
		EndTime:   anchor.Timestamp,
		ChipsWon:  100,
	}}}
	writeLegacyGob(t, filepath.Join(root, "stats.gob"), legacy)

	loaded, err := store.LoadStatsArtifact()
	if err != nil {
		t.Fatalf("LoadStatsArtifact error: %v", err)
	}
	if len(loaded.Payload.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(loaded.Payload.Sessions))
	}
	if loaded.Payload.Sessions[0].ChipsWon != 100 {
		t.Fatalf("chips won = %d, want 100", loaded.Payload.Sessions[0].ChipsWon)
	}
}

func TestSaveMigrationFromLegacyGob(t *testing.T) {
	store, root, _ := newTestStore(t)
	legacy := SaveSlot{
		Name:        "Legacy Save",
		Mode:        "tournament",
		HandNumber:  4,
		PlayerName:  "Hero",
		PlayerStack: 220,
		Timestamp:   time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
		TableData: TableSaveData{
			Mode:       engine.ModeTournament,
			Seats:      6,
			DealerSeat: 3,
		},
	}
	writeLegacyGob(t, filepath.Join(root, "saves", "slot_1.gob"), legacy)

	loaded, err := store.LoadSaveArtifact(1)
	if err != nil {
		t.Fatalf("LoadSaveArtifact error: %v", err)
	}
	if loaded.Payload.Name != "Legacy Save" {
		t.Fatalf("save name = %q, want Legacy Save", loaded.Payload.Name)
	}
	if !exists(store.saveArtifactPath(1)) {
		t.Fatal("expected migrated save artifact")
	}
}

func TestIrrecongnizableLegacyConfigReturnsTypedError(t *testing.T) {
	store, root, _ := newTestStore(t)
	if err := os.WriteFile(filepath.Join(root, "config.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write broken config error: %v", err)
	}

	_, err := store.LoadConfigArtifact()
	if err == nil {
		t.Fatal("expected error")
	}
	var compatErr *CompatibilityError
	if !errors.As(err, &compatErr) {
		t.Fatalf("expected CompatibilityError, got %T", err)
	}
	if !errors.Is(err, ErrLegacyFormatUnrecognized) {
		t.Fatalf("expected ErrLegacyFormatUnrecognized, got %v", err)
	}
}
