package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestArtifactStoreConfigRoundTripAndMetadata(t *testing.T) {
	store, _, anchor := newTestStore(t)
	cfg := DefaultConfig()
	cfg.PlayerName = "Fabio"
	cfg.SoundVolume = 42

	saved, err := store.SaveConfigArtifact(cfg)
	if err != nil {
		t.Fatalf("SaveConfigArtifact error: %v", err)
	}
	loaded, err := store.LoadConfigArtifact()
	if err != nil {
		t.Fatalf("LoadConfigArtifact error: %v", err)
	}

	if saved.Metadata.Kind != ArtifactKindConfig {
		t.Fatalf("saved kind = %q, want %q", saved.Metadata.Kind, ArtifactKindConfig)
	}
	if loaded.Metadata.Version != artifactVersionConfig {
		t.Fatalf("loaded version = %d, want %d", loaded.Metadata.Version, artifactVersionConfig)
	}
	if loaded.Metadata.Namespace != "local/config" {
		t.Fatalf("namespace = %q, want local/config", loaded.Metadata.Namespace)
	}
	if loaded.Metadata.CreatedAt != anchor {
		t.Fatalf("created_at = %+v, want %+v", loaded.Metadata.CreatedAt, anchor)
	}
	if loaded.Payload.PlayerName != "Fabio" {
		t.Fatalf("player name = %q, want Fabio", loaded.Payload.PlayerName)
	}
	if loaded.Payload.SoundVolume != 42 {
		t.Fatalf("sound volume = %d, want 42", loaded.Payload.SoundVolume)
	}
}

func TestArtifactStoreStatsRoundTrip(t *testing.T) {
	store, _, anchor := newTestStore(t)
	stats := StatsStore{
		Sessions: []SessionStats{{
			ID:            "ses_11111111111111111111111111111111",
			Mode:          "tournament",
			StartTime:     anchor.Timestamp,
			EndTime:       anchor.Timestamp,
			HandsPlayed:   12,
			FinalPosition: 1,
			TotalPlayers:  6,
			ChipsWon:      240,
			HandsWon:      5,
			BestHand:      "Straight",
		}},
	}

	if _, err := store.SaveStatsArtifact(stats); err != nil {
		t.Fatalf("SaveStatsArtifact error: %v", err)
	}
	loaded, err := store.LoadStatsArtifact()
	if err != nil {
		t.Fatalf("LoadStatsArtifact error: %v", err)
	}
	if len(loaded.Payload.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(loaded.Payload.Sessions))
	}
	if loaded.Payload.Sessions[0].BestHand != "Straight" {
		t.Fatalf("best hand = %q, want Straight", loaded.Payload.Sessions[0].BestHand)
	}
}

func TestArtifactStoreSaveSlotRoundTripListAndDelete(t *testing.T) {
	store, _, _ := newTestStore(t)
	save := SaveSlot{
		Name:        "Table One",
		Mode:        "tournament",
		HandNumber:  7,
		PlayerName:  "Hero",
		PlayerStack: 350,
		TableData: TableSaveData{
			Mode:       engine.ModeTournament,
			Seats:      6,
			DealerSeat: 2,
		},
	}

	if _, err := store.SaveSaveArtifact(1, save); err != nil {
		t.Fatalf("SaveSaveArtifact error: %v", err)
	}
	loaded, err := store.LoadSaveArtifact(1)
	if err != nil {
		t.Fatalf("LoadSaveArtifact error: %v", err)
	}
	if loaded.Payload.Name != "Table One" {
		t.Fatalf("save name = %q, want Table One", loaded.Payload.Name)
	}

	infos, err := store.ListSaveArtifacts(DefaultSaveSlotCount)
	if err != nil {
		t.Fatalf("ListSaveArtifacts error: %v", err)
	}
	if len(infos) != DefaultSaveSlotCount {
		t.Fatalf("len(infos) = %d, want %d", len(infos), DefaultSaveSlotCount)
	}
	if infos[0].Empty {
		t.Fatal("slot 1 should not be empty")
	}
	if !infos[1].Empty {
		t.Fatal("slot 2 should be empty")
	}

	if err := store.DeleteSaveArtifact(1); err != nil {
		t.Fatalf("DeleteSaveArtifact error: %v", err)
	}
	_, err = store.LoadSaveArtifact(1)
	if !errors.Is(err, ErrArtifactNotFound) {
		t.Fatalf("load after delete error = %v, want ErrArtifactNotFound", err)
	}
}

func TestLegacyFacadesUseDefaultArtifactStore(t *testing.T) {
	store, _, anchor := newTestStore(t)
	useDefaultStoreForTest(t, store)

	cfg := DefaultConfig()
	cfg.PlayerName = "Facade"
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}
	loadedCfg, err := LoadConfigResult()
	if err != nil {
		t.Fatalf("LoadConfigResult error: %v", err)
	}
	if loadedCfg.PlayerName != "Facade" {
		t.Fatalf("config player name = %q, want Facade", loadedCfg.PlayerName)
	}

	stats := &StatsStore{Sessions: []SessionStats{{ID: "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", StartTime: anchor.Timestamp, EndTime: anchor.Timestamp}}}
	if err := SaveStats(stats); err != nil {
		t.Fatalf("SaveStats error: %v", err)
	}
	loadedStats, err := LoadStatsResult()
	if err != nil {
		t.Fatalf("LoadStatsResult error: %v", err)
	}
	if len(loadedStats.Sessions) != 1 {
		t.Fatalf("len(stats sessions) = %d, want 1", len(loadedStats.Sessions))
	}

	save := &SaveSlot{Name: "Facade Save", Mode: "cash", Timestamp: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)}
	if err := SaveGame(1, save); err != nil {
		t.Fatalf("SaveGame error: %v", err)
	}
	loadedSave, err := LoadGame(1)
	if err != nil {
		t.Fatalf("LoadGame error: %v", err)
	}
	if loadedSave.Name != "Facade Save" {
		t.Fatalf("save name = %q, want Facade Save", loadedSave.Name)
	}
	infos, err := ListSavesResult()
	if err != nil {
		t.Fatalf("ListSavesResult error: %v", err)
	}
	if infos[0].Name != "Facade Save" {
		t.Fatalf("slot 1 name = %q, want Facade Save", infos[0].Name)
	}
}

func TestTranscriptChunkArtifactRoundTrip(t *testing.T) {
	store, _, anchor := newTestStore(t)
	chunk := TranscriptChunk{
		Version:      artifactVersionTranscriptChunk,
		ID:           "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001",
		SessionID:    "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		TranscriptID: "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ChunkIndex:   1,
		HandID:       3,
		CommittedAt:  anchor,
		Records: []TranscriptRecord{{
			Version:      1,
			SessionID:    "ses_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			TranscriptID: "trn_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ChunkID:      "tch_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_000001",
			Sequence:     12,
			HandID:       3,
			Kind:         "action_taken",
			Payload:      []byte{1, 2, 3},
			TimeAnchor:   anchor,
		}},
	}

	if _, err := store.SaveTranscriptChunkArtifact(chunk); err != nil {
		t.Fatalf("SaveTranscriptChunkArtifact error: %v", err)
	}
	loaded, err := store.LoadTranscriptChunkArtifact(chunk.TranscriptID, chunk.ID)
	if err != nil {
		t.Fatalf("LoadTranscriptChunkArtifact error: %v", err)
	}
	if loaded.Payload.ID != chunk.ID {
		t.Fatalf("chunk id = %q, want %q", loaded.Payload.ID, chunk.ID)
	}
	if len(loaded.Payload.Records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(loaded.Payload.Records))
	}
}

func TestStableArtifactIDMapping(t *testing.T) {
	sessionID := "ses_0123456789abcdef0123456789abcdef"
	transcriptID, err := TranscriptIDFromSessionID(sessionID)
	if err != nil {
		t.Fatalf("TranscriptIDFromSessionID error: %v", err)
	}
	if transcriptID != "trn_0123456789abcdef0123456789abcdef" {
		t.Fatalf("transcriptID = %q", transcriptID)
	}
	chunkID, _ := ChunkIDFromSessionID(sessionID, 3)
	if chunkID != "tch_0123456789abcdef0123456789abcdef_000003" {
		t.Fatalf("chunkID = %q", chunkID)
	}
	checkpointID, _ := CheckpointIDFromSessionID(sessionID, 9)
	if checkpointID != "ckp_0123456789abcdef0123456789abcdef_000009" {
		t.Fatalf("checkpointID = %q", checkpointID)
	}
	snapshotID, _ := SnapshotIDFromSessionID(sessionID, 9, 17)
	if snapshotID != "snp_0123456789abcdef0123456789abcdef_000009_000000017" {
		t.Fatalf("snapshotID = %q", snapshotID)
	}
}
