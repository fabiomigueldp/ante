package storage

import (
	"path/filepath"
	"time"
)

// SessionStats records statistics for a completed session.
type SessionStats struct {
	ID            string    `json:"id"`
	Mode          string    `json:"mode"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	HandsPlayed   int       `json:"hands_played"`
	FinalPosition int       `json:"final_position"`
	TotalPlayers  int       `json:"total_players"`
	ChipsWon      int       `json:"chips_won"`
	BiggestPot    int       `json:"biggest_pot"`
	HandsWon      int       `json:"hands_won"`
	FlopsSeen     int       `json:"flops_seen"`
	ShowdownsWon  int       `json:"showdowns_won"`
	ShowdownsSeen int       `json:"showdowns_seen"`
	AllInsWon     int       `json:"allins_won"`
	AllInsSeen    int       `json:"allins_seen"`
	BestHand      string    `json:"best_hand"`
	LargestWin    int       `json:"largest_win"`
	LongestStreak int       `json:"longest_streak"`
}

// StatsStore holds all session statistics.
type StatsStore struct {
	Sessions []SessionStats `json:"sessions"`
}

func statsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stats.gob"), nil
}

func LoadStatsResult() (*StatsStore, error) {
	store := DefaultArtifactStore()
	summaries, err := store.ListSessionSummaryArtifacts()
	if err == nil && len(summaries) > 0 {
		aggregated := aggregateStatsFromSummaries(summaries)
		if _, saveErr := store.SaveStatsArtifact(*aggregated); saveErr == nil {
			return aggregated, nil
		}
		return aggregated, nil
	}
	artifact, err := store.LoadStatsArtifact()
	if err == nil {
		return &artifact.Payload, nil
	}
	if err == ErrArtifactNotFound {
		return &StatsStore{}, nil
	}
	return &StatsStore{}, err
}

func LoadStats() *StatsStore {
	store, err := LoadStatsResult()
	if err != nil {
		return &StatsStore{}
	}
	return store
}

func SaveStats(store *StatsStore) error {
	if store == nil {
		store = &StatsStore{}
	}
	_, err := DefaultArtifactStore().SaveStatsArtifact(*store)
	return err
}

func (s *StatsStore) Add(stats SessionStats) {
	s.Sessions = append(s.Sessions, stats)
}

func (s *StatsStore) TotalSessions() int { return len(s.Sessions) }

func (s *StatsStore) TotalHandsPlayed() int {
	total := 0
	for _, sess := range s.Sessions {
		total += sess.HandsPlayed
	}
	return total
}

func (s *StatsStore) TournamentWins() int {
	wins := 0
	for _, sess := range s.Sessions {
		if (sess.Mode == "tournament" || sess.Mode == "headsup") && sess.FinalPosition == 1 {
			wins++
		}
	}
	return wins
}

func (s *StatsStore) TotalProfit() int {
	total := 0
	for _, sess := range s.Sessions {
		total += sess.ChipsWon
	}
	return total
}

func (s *StatsStore) AvgFinish() float64 {
	count := 0
	total := 0
	for _, sess := range s.Sessions {
		if sess.Mode == "tournament" || sess.Mode == "headsup" {
			total += sess.FinalPosition
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

func (s *StatsStore) WinRate() float64 {
	if len(s.Sessions) == 0 {
		return 0
	}
	hands := 0
	won := 0
	for _, sess := range s.Sessions {
		hands += sess.HandsPlayed
		won += sess.HandsWon
	}
	if hands == 0 {
		return 0
	}
	return float64(won) / float64(hands) * 100
}

func (s *StatsStore) BestHandEver() string {
	best := ""
	for _, sess := range s.Sessions {
		if sess.BestHand != "" && (best == "" || sess.BestHand > best) {
			best = sess.BestHand
		}
	}
	if best == "" {
		return "N/A"
	}
	return best
}

func (s *StatsStore) RecentSessions(n int) []SessionStats {
	if n >= len(s.Sessions) {
		out := make([]SessionStats, len(s.Sessions))
		copy(out, s.Sessions)
		return out
	}
	out := make([]SessionStats, n)
	copy(out, s.Sessions[len(s.Sessions)-n:])
	return out
}

func aggregateStatsFromSummaries(summaries []Artifact[SessionSummary]) *StatsStore {
	store := &StatsStore{Sessions: make([]SessionStats, 0, len(summaries))}
	for _, artifact := range summaries {
		summary := artifact.Payload
		store.Sessions = append(store.Sessions, SessionStats{
			ID:            summary.SessionID,
			Mode:          summary.Mode,
			StartTime:     summary.StartTime.Timestamp,
			EndTime:       summary.EndTime.Timestamp,
			HandsPlayed:   summary.HandsPlayed,
			FinalPosition: summary.FinalPosition,
			TotalPlayers:  summary.TotalPlayers,
			ChipsWon:      summary.ChipsWon,
			BiggestPot:    summary.BiggestPot,
			HandsWon:      summary.HandsWon,
			FlopsSeen:     summary.FlopsSeen,
			ShowdownsWon:  summary.ShowdownsWon,
			ShowdownsSeen: summary.ShowdownsSeen,
			AllInsWon:     summary.AllInsWon,
			AllInsSeen:    summary.AllInsSeen,
			BestHand:      summary.BestHand,
			LargestWin:    summary.LargestWin,
			LongestStreak: summary.LongestStreak,
		})
	}
	return store
}
