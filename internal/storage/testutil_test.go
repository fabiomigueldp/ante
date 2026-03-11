package storage

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type staticTimeAnchorProvider struct {
	anchor TimeAnchor
	err    error
}

func (p staticTimeAnchorProvider) Now() (TimeAnchor, error) {
	if p.err != nil {
		return TimeAnchor{}, p.err
	}
	return p.anchor.normalized(), nil
}

func newTestStore(t *testing.T) (*FileSystemStore, string, TimeAnchor) {
	t.Helper()
	root := t.TempDir()
	anchor := TimeAnchor{Timestamp: time.Date(2026, time.March, 11, 10, 30, 0, 123, time.UTC), Source: "test_clock"}
	store, err := NewFileSystemStore(root, staticTimeAnchorProvider{anchor: anchor})
	if err != nil {
		t.Fatalf("NewFileSystemStore error: %v", err)
	}
	return store, root, anchor
}

func useDefaultStoreForTest(t *testing.T, store ArtifactStore) {
	t.Helper()
	defaultStoreMu.Lock()
	oldStore := defaultStore
	oldMaker := defaultStoreMaker
	defaultStore = store
	defaultStoreMaker = func() (ArtifactStore, error) { return store, nil }
	defaultStoreMu.Unlock()
	t.Cleanup(func() {
		defaultStoreMu.Lock()
		defaultStore = oldStore
		defaultStoreMaker = oldMaker
		defaultStoreMu.Unlock()
	})
}

func writeLegacyGob[T any](t *testing.T, path string, value T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file error: %v", err)
	}
	defer file.Close()
	if err := gob.NewEncoder(file).Encode(value); err != nil {
		t.Fatalf("gob encode error: %v", err)
	}
}
