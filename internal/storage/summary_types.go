package storage

import (
	"time"

	"github.com/fabiomigueldp/ante/internal/engine"
)

type SessionMetricsSnapshot struct {
	StartTime          TimeAnchor     `json:"start_time"`
	HandsPlayed        int            `json:"hands_played"`
	HandsWon           int            `json:"hands_won"`
	FlopsSeen          int            `json:"flops_seen"`
	ShowdownsWon       int            `json:"showdowns_won"`
	ShowdownsSeen      int            `json:"showdowns_seen"`
	AllInsWon          int            `json:"allins_won"`
	AllInsSeen         int            `json:"allins_seen"`
	BiggestPot         int            `json:"biggest_pot"`
	LargestWin         int            `json:"largest_win"`
	BestHand           string         `json:"best_hand"`
	CurrentStreak      int            `json:"current_streak"`
	LongestStreak      int            `json:"longest_streak"`
	LastChunkID        string         `json:"last_chunk_id,omitempty"`
	LastCheckpointID   string         `json:"last_checkpoint_id,omitempty"`
	LastSnapshotID     string         `json:"last_snapshot_id,omitempty"`
	LastCheckpointHash TranscriptHash `json:"last_checkpoint_hash"`
}

type SessionSummary struct {
	ID               string         `json:"id"`
	SessionID        string         `json:"session_id"`
	TranscriptID     string         `json:"transcript_id"`
	LatestChunkID    string         `json:"latest_chunk_id,omitempty"`
	LatestSnapshotID string         `json:"latest_snapshot_id,omitempty"`
	CheckpointID     string         `json:"checkpoint_id,omitempty"`
	CheckpointHash   TranscriptHash `json:"checkpoint_hash"`
	PlayerName       string         `json:"player_name"`
	Mode             string         `json:"mode"`
	StartTime        TimeAnchor     `json:"start_time"`
	EndTime          TimeAnchor     `json:"end_time"`
	HandsPlayed      int            `json:"hands_played"`
	FinalPosition    int            `json:"final_position"`
	TotalPlayers     int            `json:"total_players"`
	FinalStack       int            `json:"final_stack"`
	StartingChips    int            `json:"starting_chips"`
	ChipsWon         int            `json:"chips_won"`
	BiggestPot       int            `json:"biggest_pot"`
	HandsWon         int            `json:"hands_won"`
	FlopsSeen        int            `json:"flops_seen"`
	ShowdownsWon     int            `json:"showdowns_won"`
	ShowdownsSeen    int            `json:"showdowns_seen"`
	AllInsWon        int            `json:"allins_won"`
	AllInsSeen       int            `json:"allins_seen"`
	BestHand         string         `json:"best_hand"`
	LargestWin       int            `json:"largest_win"`
	LongestStreak    int            `json:"longest_streak"`
	ResultLabel      string         `json:"result_label"`
}

type HistorySessionEntry struct {
	SessionID    string    `json:"session_id"`
	TranscriptID string    `json:"transcript_id"`
	PlayerName   string    `json:"player_name"`
	Mode         string    `json:"mode"`
	HandsPlayed  int       `json:"hands_played"`
	ResultLabel  string    `json:"result_label"`
	UpdatedAt    time.Time `json:"updated_at"`
	Completed    bool      `json:"completed"`
}

type HistoryHandEntry struct {
	SessionID    string                  `json:"session_id"`
	TranscriptID string                  `json:"transcript_id"`
	ChunkID      string                  `json:"chunk_id"`
	SnapshotID   string                  `json:"snapshot_id"`
	CheckpointID string                  `json:"checkpoint_id"`
	HandID       int                     `json:"hand_id"`
	ChunkIndex   int                     `json:"chunk_index"`
	DealerSeat   int                     `json:"dealer_seat"`
	Blinds       engine.BlindLevel       `json:"blinds"`
	Players      []engine.PlayerSnapshot `json:"players"`
	ResultLabel  string                  `json:"result_label"`
	CommittedAt  time.Time               `json:"committed_at"`
	FinalBoard   []engine.Card           `json:"final_board,omitempty"`
}
