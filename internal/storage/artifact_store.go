package storage

import (
	"os"
	"path/filepath"
	"sync"
)

type ArtifactStore interface {
	RootDir() string
	TimeAnchorProvider() TimeAnchorProvider
	LoadConfigArtifact() (Artifact[Config], error)
	SaveConfigArtifact(cfg Config) (Artifact[Config], error)
	LoadStatsArtifact() (Artifact[StatsStore], error)
	SaveStatsArtifact(store StatsStore) (Artifact[StatsStore], error)
	LoadSaveArtifact(slot int) (Artifact[SaveSlot], error)
	SaveSaveArtifact(slot int, save SaveSlot) (Artifact[SaveSlot], error)
	DeleteSaveArtifact(slot int) error
	ListSaveArtifacts(slots int) ([]SaveInfo, error)
	LoadTranscriptChunkArtifact(transcriptID, chunkID string) (Artifact[TranscriptChunk], error)
	SaveTranscriptChunkArtifact(chunk TranscriptChunk) (Artifact[TranscriptChunk], error)
	LoadTranscriptHeadArtifact(transcriptID string) (Artifact[TranscriptHead], error)
	SaveTranscriptHeadArtifact(head TranscriptHead) (Artifact[TranscriptHead], error)
	ListTranscriptHeadArtifacts() ([]Artifact[TranscriptHead], error)
	ListTranscriptChunkArtifacts(transcriptID string) ([]Artifact[TranscriptChunk], error)
	LoadSessionSummaryArtifact(sessionID string) (Artifact[SessionSummary], error)
	SaveSessionSummaryArtifact(summary SessionSummary) (Artifact[SessionSummary], error)
	ListSessionSummaryArtifacts() ([]Artifact[SessionSummary], error)
}

type errorStore struct {
	err error
}

func (s errorStore) RootDir() string { return "" }

func (s errorStore) TimeAnchorProvider() TimeAnchorProvider { return NewLocalTimeAnchorProvider() }

func (s errorStore) LoadConfigArtifact() (Artifact[Config], error) {
	return Artifact[Config]{}, s.err
}

func (s errorStore) SaveConfigArtifact(Config) (Artifact[Config], error) {
	return Artifact[Config]{}, s.err
}

func (s errorStore) LoadStatsArtifact() (Artifact[StatsStore], error) {
	return Artifact[StatsStore]{}, s.err
}

func (s errorStore) SaveStatsArtifact(StatsStore) (Artifact[StatsStore], error) {
	return Artifact[StatsStore]{}, s.err
}

func (s errorStore) LoadSaveArtifact(int) (Artifact[SaveSlot], error) {
	return Artifact[SaveSlot]{}, s.err
}

func (s errorStore) SaveSaveArtifact(int, SaveSlot) (Artifact[SaveSlot], error) {
	return Artifact[SaveSlot]{}, s.err
}

func (s errorStore) DeleteSaveArtifact(int) error { return s.err }

func (s errorStore) ListSaveArtifacts(int) ([]SaveInfo, error) { return nil, s.err }

func (s errorStore) LoadTranscriptChunkArtifact(string, string) (Artifact[TranscriptChunk], error) {
	return Artifact[TranscriptChunk]{}, s.err
}

func (s errorStore) SaveTranscriptChunkArtifact(TranscriptChunk) (Artifact[TranscriptChunk], error) {
	return Artifact[TranscriptChunk]{}, s.err
}

func (s errorStore) LoadTranscriptHeadArtifact(string) (Artifact[TranscriptHead], error) {
	return Artifact[TranscriptHead]{}, s.err
}

func (s errorStore) SaveTranscriptHeadArtifact(TranscriptHead) (Artifact[TranscriptHead], error) {
	return Artifact[TranscriptHead]{}, s.err
}

func (s errorStore) ListTranscriptHeadArtifacts() ([]Artifact[TranscriptHead], error) {
	return nil, s.err
}

func (s errorStore) ListTranscriptChunkArtifacts(string) ([]Artifact[TranscriptChunk], error) {
	return nil, s.err
}

func (s errorStore) LoadSessionSummaryArtifact(string) (Artifact[SessionSummary], error) {
	return Artifact[SessionSummary]{}, s.err
}

func (s errorStore) SaveSessionSummaryArtifact(SessionSummary) (Artifact[SessionSummary], error) {
	return Artifact[SessionSummary]{}, s.err
}

func (s errorStore) ListSessionSummaryArtifacts() ([]Artifact[SessionSummary], error) {
	return nil, s.err
}

var (
	defaultStoreMu    sync.Mutex
	defaultStore      ArtifactStore
	defaultStoreMaker = func() (ArtifactStore, error) {
		root, err := configDir()
		if err != nil {
			return nil, err
		}
		return NewFileSystemStore(root, NewLocalTimeAnchorProvider())
	}
)

func DefaultArtifactStore() ArtifactStore {
	defaultStoreMu.Lock()
	defer defaultStoreMu.Unlock()
	if defaultStore != nil {
		return defaultStore
	}
	store, err := defaultStoreMaker()
	if err != nil {
		defaultStore = errorStore{err: err}
		return defaultStore
	}
	defaultStore = store
	return defaultStore
}

func SetDefaultArtifactStoreForTest(store ArtifactStore) func() {
	defaultStoreMu.Lock()
	oldStore := defaultStore
	oldMaker := defaultStoreMaker
	defaultStore = store
	defaultStoreMaker = func() (ArtifactStore, error) { return store, nil }
	defaultStoreMu.Unlock()
	return func() {
		defaultStoreMu.Lock()
		defaultStore = oldStore
		defaultStoreMaker = oldMaker
		defaultStoreMu.Unlock()
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".ante")
	return dir, os.MkdirAll(dir, 0o755)
}
