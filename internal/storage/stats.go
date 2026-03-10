package storage

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"time"
)

// SessionStats records statistics for a completed session.
type SessionStats struct {
	ID            string    `json:"id"`
	Mode          string    `json:"mode"` // "tournament", "cash", "headsup"
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	HandsPlayed   int       `json:"hands_played"`
	FinalPosition int       `json:"final_position"` // tournament: 1=winner
	TotalPlayers  int       `json:"total_players"`
	ChipsWon      int       `json:"chips_won"` // net profit/loss
	BiggestPot    int       `json:"biggest_pot"`
	HandsWon      int       `json:"hands_won"`
	FlopsSeen     int       `json:"flops_seen"`
	ShowdownsWon  int       `json:"showdowns_won"`
	ShowdownsSeen int       `json:"showdowns_seen"`
	AllInsWon     int       `json:"allins_won"`
	AllInsSeen    int       `json:"allins_seen"`
	BestHand      string    `json:"best_hand"`
	LargestWin    int       `json:"largest_win"`
	LongestStreak int       `json:"longest_streak"` // consecutive wins
}

// StatsStore holds all session statistics.
type StatsStore struct {
	Sessions []SessionStats
}

func statsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stats.gob"), nil
}

func LoadStats() *StatsStore {
	path, err := statsPath()
	if err != nil {
		return &StatsStore{}
	}
	f, err := os.Open(path)
	if err != nil {
		return &StatsStore{}
	}
	defer f.Close()
	var store StatsStore
	if err := gob.NewDecoder(f).Decode(&store); err != nil {
		return &StatsStore{}
	}
	return &store
}

func SaveStats(store *StatsStore) error {
	path, err := statsPath()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(store)
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
