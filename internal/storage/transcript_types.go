package storage

import "github.com/fabiomigueldp/ante/internal/engine"

type TranscriptHash struct {
	Algorithm string `json:"algorithm"`
	Sum       []byte `json:"sum"`
}

type TranscriptRecord struct {
	Version      int               `json:"version"`
	SessionID    string            `json:"session_id"`
	TranscriptID string            `json:"transcript_id"`
	ChunkID      string            `json:"chunk_id"`
	Sequence     uint64            `json:"sequence"`
	HandID       int               `json:"hand_id"`
	Kind         string            `json:"kind"`
	SnapshotID   string            `json:"snapshot_id,omitempty"`
	Message      string            `json:"message,omitempty"`
	PlayerID     engine.PlayerID   `json:"player_id,omitempty"`
	PlayerName   string            `json:"player_name,omitempty"`
	Action       *engine.Action    `json:"action,omitempty"`
	Street       engine.Street     `json:"street,omitempty"`
	NewCards     []engine.Card     `json:"new_cards,omitempty"`
	ShownCards   [2]engine.Card    `json:"shown_cards,omitempty"`
	EvalName     string            `json:"eval_name,omitempty"`
	Winners      []engine.PlayerID `json:"winners,omitempty"`
	AwardAmount  int               `json:"award_amount,omitempty"`
	PotTotal     int               `json:"pot_total,omitempty"`
	Payload      []byte            `json:"payload"`
	TimeAnchor   TimeAnchor        `json:"time_anchor"`
}

type TranscriptChunk struct {
	Version          int                     `json:"version"`
	ID               string                  `json:"id"`
	SessionID        string                  `json:"session_id"`
	TranscriptID     string                  `json:"transcript_id"`
	ChunkIndex       int                     `json:"chunk_index"`
	HandID           int                     `json:"hand_id"`
	SnapshotID       string                  `json:"snapshot_id,omitempty"`
	CheckpointID     string                  `json:"checkpoint_id,omitempty"`
	PreviousHash     TranscriptHash          `json:"previous_hash"`
	CheckpointHash   TranscriptHash          `json:"checkpoint_hash"`
	Players          []engine.PlayerSnapshot `json:"players,omitempty"`
	DealerSeat       int                     `json:"dealer_seat"`
	Blinds           engine.BlindLevel       `json:"blinds"`
	FinalBoard       []engine.Card           `json:"final_board,omitempty"`
	HumanStack       int                     `json:"human_stack,omitempty"`
	ResultLabel      string                  `json:"result_label,omitempty"`
	WinningPlayerIDs []engine.PlayerID       `json:"winning_player_ids,omitempty"`
	Records          []TranscriptRecord      `json:"records"`
	CommittedAt      TimeAnchor              `json:"committed_at"`
}

type TranscriptHead struct {
	Version            int            `json:"version"`
	SessionID          string         `json:"session_id"`
	TranscriptID       string         `json:"transcript_id"`
	PlayerName         string         `json:"player_name,omitempty"`
	Mode               string         `json:"mode,omitempty"`
	LatestChunkID      string         `json:"latest_chunk_id"`
	LatestSnapshotID   string         `json:"latest_snapshot_id,omitempty"`
	LatestCheckpointID string         `json:"latest_checkpoint_id,omitempty"`
	LatestChunkHash    TranscriptHash `json:"latest_chunk_hash"`
	LatestSeq          uint64         `json:"latest_seq"`
	HandsPlayed        int            `json:"hands_played"`
	StartedAt          TimeAnchor     `json:"started_at"`
	UpdatedAt          TimeAnchor     `json:"updated_at"`
}
