package session

import "github.com/fabiomigueldp/ante/internal/storage"

type Dependencies struct {
	ArtifactStore      storage.ArtifactStore
	TimeAnchorProvider storage.TimeAnchorProvider
}

var sessionDependenciesProvider = DefaultDependencies

func DefaultDependencies() Dependencies {
	store := storage.DefaultArtifactStore()
	return Dependencies{
		ArtifactStore:      store,
		TimeAnchorProvider: store.TimeAnchorProvider(),
	}
}
