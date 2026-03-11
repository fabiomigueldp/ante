package storage

type TranscriptHash struct {
	Algorithm string `json:"algorithm"`
	Sum       []byte `json:"sum"`
}

type TranscriptRecord struct {
	Version      int        `json:"version"`
	SessionID    string     `json:"session_id"`
	TranscriptID string     `json:"transcript_id"`
	ChunkID      string     `json:"chunk_id"`
	Sequence     uint64     `json:"sequence"`
	HandID       int        `json:"hand_id"`
	Kind         string     `json:"kind"`
	SnapshotID   string     `json:"snapshot_id,omitempty"`
	Payload      []byte     `json:"payload"`
	TimeAnchor   TimeAnchor `json:"time_anchor"`
}

type TranscriptChunk struct {
	Version        int                `json:"version"`
	ID             string             `json:"id"`
	SessionID      string             `json:"session_id"`
	TranscriptID   string             `json:"transcript_id"`
	ChunkIndex     int                `json:"chunk_index"`
	HandID         int                `json:"hand_id"`
	PreviousHash   TranscriptHash     `json:"previous_hash"`
	CheckpointHash TranscriptHash     `json:"checkpoint_hash"`
	Records        []TranscriptRecord `json:"records"`
	CommittedAt    TimeAnchor         `json:"committed_at"`
}

type TranscriptHead struct {
	Version         int            `json:"version"`
	SessionID       string         `json:"session_id"`
	TranscriptID    string         `json:"transcript_id"`
	LatestChunkID   string         `json:"latest_chunk_id"`
	LatestChunkHash TranscriptHash `json:"latest_chunk_hash"`
	LatestSeq       uint64         `json:"latest_seq"`
	UpdatedAt       TimeAnchor     `json:"updated_at"`
}
