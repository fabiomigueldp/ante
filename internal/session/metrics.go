package session

import (
	"fmt"
	"strings"

	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

type MetricsAccumulator struct {
	startTime          storage.TimeAnchor
	handsPlayed        int
	handsWon           int
	flopsSeen          int
	showdownsWon       int
	showdownsSeen      int
	allInsWon          int
	allInsSeen         int
	biggestPot         int
	largestWin         int
	bestHand           string
	currentStreak      int
	longestStreak      int
	lastChunkID        string
	lastCheckpointID   string
	lastSnapshotID     string
	lastCheckpointHash storage.TranscriptHash
}

func newMetricsAccumulator(start storage.TimeAnchor) *MetricsAccumulator {
	return &MetricsAccumulator{startTime: start}
}

func metricsFromSnapshot(snapshot storage.SessionMetricsSnapshot) *MetricsAccumulator {
	acc := &MetricsAccumulator{}
	acc.applySnapshot(snapshot)
	return acc
}

func (m *MetricsAccumulator) applySnapshot(snapshot storage.SessionMetricsSnapshot) {
	m.startTime = snapshot.StartTime
	m.handsPlayed = snapshot.HandsPlayed
	m.handsWon = snapshot.HandsWon
	m.flopsSeen = snapshot.FlopsSeen
	m.showdownsWon = snapshot.ShowdownsWon
	m.showdownsSeen = snapshot.ShowdownsSeen
	m.allInsWon = snapshot.AllInsWon
	m.allInsSeen = snapshot.AllInsSeen
	m.biggestPot = snapshot.BiggestPot
	m.largestWin = snapshot.LargestWin
	m.bestHand = snapshot.BestHand
	m.currentStreak = snapshot.CurrentStreak
	m.longestStreak = snapshot.LongestStreak
	m.lastChunkID = snapshot.LastChunkID
	m.lastCheckpointID = snapshot.LastCheckpointID
	m.lastSnapshotID = snapshot.LastSnapshotID
	m.lastCheckpointHash = snapshot.LastCheckpointHash
}

func (m *MetricsAccumulator) Snapshot() storage.SessionMetricsSnapshot {
	if m == nil {
		return storage.SessionMetricsSnapshot{}
	}
	return storage.SessionMetricsSnapshot{
		StartTime:          m.startTime,
		HandsPlayed:        m.handsPlayed,
		HandsWon:           m.handsWon,
		FlopsSeen:          m.flopsSeen,
		ShowdownsWon:       m.showdownsWon,
		ShowdownsSeen:      m.showdownsSeen,
		AllInsWon:          m.allInsWon,
		AllInsSeen:         m.allInsSeen,
		BiggestPot:         m.biggestPot,
		LargestWin:         m.largestWin,
		BestHand:           m.bestHand,
		CurrentStreak:      m.currentStreak,
		LongestStreak:      m.longestStreak,
		LastChunkID:        m.lastChunkID,
		LastCheckpointID:   m.lastCheckpointID,
		LastSnapshotID:     m.lastSnapshotID,
		LastCheckpointHash: m.lastCheckpointHash,
	}
}

func (m *MetricsAccumulator) ObserveHand(hand *engine.Hand, humanID engine.PlayerID, refs transcriptRefs) {
	if m == nil || hand == nil {
		return
	}
	m.handsPlayed++
	human := playerByID(hand.Players, humanID)
	if human == nil {
		return
	}
	humanInHand := human.Status != engine.StatusOut && human.Status != engine.StatusSittingOut
	humanSawFlop := false
	humanReachedShowdown := false
	humanWentAllIn := false
	humanWonHand := false
	humanWinnings := 0
	totalPot := 0
	for _, event := range hand.Events {
		switch e := event.(type) {
		case engine.ActionTakenEvent:
			if e.PlayerID == humanID {
				if e.Action.Type == engine.ActionFold {
					humanInHand = false
				}
				if e.Action.Type == engine.ActionAllIn {
					humanWentAllIn = true
				}
			}
		case engine.StreetAdvancedEvent:
			if e.Street == engine.StreetFlop && humanInHand {
				humanSawFlop = true
			}
		case engine.ShowdownStartedEvent:
			if humanInHand {
				humanReachedShowdown = true
			}
		case engine.PotAwardedEvent:
			totalPot += e.Amount
			for _, winner := range e.Winners {
				if winner == humanID {
					share := e.Amount
					if len(e.Winners) > 0 {
						share = e.Amount / len(e.Winners)
						if e.OddChip == humanID {
							share += e.Amount - share*len(e.Winners)
						}
					}
					humanWonHand = true
					humanWinnings += share
				}
			}
		}
	}
	if humanSawFlop {
		m.flopsSeen++
	}
	if humanReachedShowdown {
		m.showdownsSeen++
		if humanWonHand {
			m.showdownsWon++
		}
	}
	if humanWentAllIn {
		m.allInsSeen++
		if humanWonHand {
			m.allInsWon++
		}
	}
	if humanWonHand {
		m.handsWon++
		m.currentStreak++
		if m.currentStreak > m.longestStreak {
			m.longestStreak = m.currentStreak
		}
	} else {
		m.currentStreak = 0
	}
	if totalPot > m.biggestPot {
		m.biggestPot = totalPot
	}
	if humanWinnings > m.largestWin {
		m.largestWin = humanWinnings
	}
	if human.Status != engine.StatusFolded && human.Status != engine.StatusOut && len(hand.Board) == 5 {
		eval := engine.Evaluate(human.HoleCards, hand.Board)
		if eval.Name != "" && betterHandName(eval.Name, m.bestHand) {
			m.bestHand = eval.Name
		}
	}
	m.lastChunkID = refs.chunkID
	m.lastCheckpointID = refs.checkpointID
	m.lastSnapshotID = refs.snapshotID
	m.lastCheckpointHash = refs.checkpointHash
}

func betterHandName(candidate, current string) bool {
	if candidate == "" {
		return false
	}
	if current == "" {
		return true
	}
	return handRankWeight(candidate) > handRankWeight(current)
}

func handRankWeight(name string) int {
	switch {
	case name == "Royal Flush":
		return 10
	case strings.HasPrefix(name, "Straight Flush"):
		return 9
	case strings.HasPrefix(name, "Four of a Kind"):
		return 8
	case strings.HasPrefix(name, "Full House"):
		return 7
	case strings.HasPrefix(name, "Flush"):
		return 6
	case strings.HasPrefix(name, "Straight"):
		return 5
	case strings.HasPrefix(name, "Three of a Kind"):
		return 4
	case strings.HasPrefix(name, "Two Pair"):
		return 3
	case strings.HasPrefix(name, "One Pair"):
		return 2
	case strings.HasPrefix(name, "High Card"):
		return 1
	default:
		return 0
	}
}

func (m *MetricsAccumulator) BuildSummary(s *Session, end storage.TimeAnchor, transcriptHead storage.TranscriptHead) storage.SessionSummary {
	human := playerByID(s.Table.Players, s.HumanID)
	finalStack := 0
	if human != nil {
		finalStack = human.Stack
	}
	finalPosition := 0
	resultLabel := "Session complete"
	if s.Tournament != nil {
		for _, result := range s.Tournament.Results() {
			if result.PlayerID == s.HumanID {
				finalPosition = result.Position
				if result.Position == 1 {
					resultLabel = "Winner!"
				} else {
					resultLabel = fmt.Sprintf("#%d of %d", result.Position, len(s.Table.Players))
				}
				break
			}
		}
	} else {
		chipsWon := finalStack - s.startingChips()
		if chipsWon >= 0 {
			resultLabel = fmt.Sprintf("+%d chips", chipsWon)
		} else {
			resultLabel = fmt.Sprintf("%d chips", chipsWon)
		}
	}
	return storage.SessionSummary{
		ID:               s.SessionID,
		SessionID:        s.SessionID,
		TranscriptID:     transcriptHead.TranscriptID,
		LatestChunkID:    transcriptHead.LatestChunkID,
		LatestSnapshotID: transcriptHead.LatestSnapshotID,
		CheckpointID:     transcriptHead.LatestCheckpointID,
		CheckpointHash:   transcriptHead.LatestChunkHash,
		PlayerName:       s.Config.PlayerName,
		Mode:             modeString(s.Config.Mode),
		StartTime:        m.startTime,
		EndTime:          end,
		HandsPlayed:      m.handsPlayed,
		FinalPosition:    finalPosition,
		TotalPlayers:     len(s.Table.Players),
		FinalStack:       finalStack,
		StartingChips:    s.startingChips(),
		ChipsWon:         finalStack - s.startingChips(),
		BiggestPot:       m.biggestPot,
		HandsWon:         m.handsWon,
		FlopsSeen:        m.flopsSeen,
		ShowdownsWon:     m.showdownsWon,
		ShowdownsSeen:    m.showdownsSeen,
		AllInsWon:        m.allInsWon,
		AllInsSeen:       m.allInsSeen,
		BestHand:         defaultIfEmpty(m.bestHand, "N/A"),
		LargestWin:       m.largestWin,
		LongestStreak:    m.longestStreak,
		ResultLabel:      resultLabel,
	}
}

func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
