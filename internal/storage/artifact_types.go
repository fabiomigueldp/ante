package storage

const DefaultSaveSlotCount = 5

type ArtifactKind string

const (
	ArtifactKindConfig            ArtifactKind = "config"
	ArtifactKindSaveSlot          ArtifactKind = "sandbox_save_slot"
	ArtifactKindStatsStore        ArtifactKind = "sandbox_stats_store"
	ArtifactKindTranscriptChunk   ArtifactKind = "sandbox_transcript_chunk"
	ArtifactKindTranscriptHead    ArtifactKind = "sandbox_transcript_head"
	ArtifactKindMigrationManifest ArtifactKind = "migration_manifest"
)

type ArtifactEncoding string

const (
	ArtifactEncodingJSON ArtifactEncoding = "json/v1"
)

const (
	artifactVersionConfig          = 1
	artifactVersionSaveSlot        = 1
	artifactVersionStatsStore      = 1
	artifactVersionTranscriptChunk = 1
	migrationManifestVersion       = 1
)

type ArtifactMetadata struct {
	Kind         ArtifactKind     `json:"kind"`
	Version      int              `json:"version"`
	Namespace    string           `json:"namespace"`
	ID           string           `json:"id,omitempty"`
	Encoding     ArtifactEncoding `json:"encoding"`
	CreatedAt    TimeAnchor       `json:"created_at"`
	UpdatedAt    TimeAnchor       `json:"updated_at"`
	LegacySource string           `json:"legacy_source,omitempty"`
}

type Artifact[T any] struct {
	Metadata ArtifactMetadata `json:"metadata"`
	Payload  T                `json:"payload"`
}

type MigrationManifest struct {
	ID         string     `json:"id"`
	SourceKind string     `json:"source_kind"`
	SourcePath string     `json:"source_path"`
	TargetKind string     `json:"target_kind"`
	TargetPath string     `json:"target_path"`
	Status     string     `json:"status"`
	ExecutedAt TimeAnchor `json:"executed_at"`
	Details    string     `json:"details,omitempty"`
}
