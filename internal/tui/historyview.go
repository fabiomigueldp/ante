package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/storage"
)

var (
	listHistorySessions = storage.ListHistorySessionsResult
	listSessionHands    = storage.ListSessionHandsResult
	loadReplayChunk     = storage.LoadReplayChunkResult
)

type historyMode uint8

const (
	historyModeSessions historyMode = iota
	historyModeHands
)

// HistoryViewModel browses transcript-backed session history.
type HistoryViewModel struct {
	width    int
	height   int
	mode     historyMode
	sessions []storage.HistorySessionEntry
	hands    []storage.HistoryHandEntry
	selected storage.HistorySessionEntry
	cursor   int
	scroll   int
	errorMsg string
}

func NewHistoryViewModel() HistoryViewModel {
	sessions, err := listHistorySessions()
	model := HistoryViewModel{mode: historyModeSessions, sessions: sessions}
	if err != nil {
		model.errorMsg = err.Error()
	}
	return model
}

func (m HistoryViewModel) Init() tea.Cmd { return nil }

func (m HistoryViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "escape", "esc", "q":
			if m.mode == historyModeHands {
				m.mode = historyModeSessions
				m.hands = nil
				m.cursor = 0
				m.scroll = 0
				m.errorMsg = ""
				return m, nil
			}
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "enter":
			if m.mode == historyModeSessions {
				if len(m.sessions) == 0 {
					return m, nil
				}
				selected := m.sessions[m.cursor]
				hands, err := listSessionHands(selected.TranscriptID)
				if err != nil {
					m.errorMsg = err.Error()
					return m, nil
				}
				m.selected = selected
				m.hands = hands
				m.mode = historyModeHands
				m.cursor = 0
				m.scroll = 0
				m.errorMsg = ""
				return m, nil
			}
			if len(m.hands) == 0 {
				return m, nil
			}
			selected := m.hands[m.cursor]
			chunk, err := loadReplayChunk(selected.TranscriptID, selected.ChunkID)
			if err != nil {
				m.errorMsg = err.Error()
				return m, nil
			}
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenReplay, data: chunk} }
		}
	}
	return m, nil
}

func (m *HistoryViewModel) moveCursor(delta int) {
	items := m.activeLength()
	if items == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = items - 1
	}
	if m.cursor >= items {
		m.cursor = 0
	}
	maxVisible := m.maxVisible()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+maxVisible {
		m.scroll = m.cursor - maxVisible + 1
	}
}

func (m HistoryViewModel) activeLength() int {
	if m.mode == historyModeHands {
		return len(m.hands)
	}
	return len(m.sessions)
}

func (m HistoryViewModel) maxVisible() int {
	v := m.height - 12
	if v < 5 {
		v = 5
	}
	return v
}

func (m HistoryViewModel) View() string {
	title := StyleTitle.Render("HAND HISTORY")
	sections := []string{title, ""}
	if m.errorMsg != "" {
		sections = append(sections, StyleError.Render(m.errorMsg), "")
	}
	if m.mode == historyModeSessions {
		return m.renderSessions(sections)
	}
	return m.renderHands(sections)
}

func (m HistoryViewModel) renderSessions(sections []string) string {
	if len(m.sessions) == 0 {
		sections = append(sections,
			StyleDim.Render("No transcript sessions recorded yet."),
			StyleDim.Render("Play some hands to build your history!"),
			"",
			StyleDim.Render("[Esc] Back"),
		)
		content := lipgloss.JoinVertical(lipgloss.Center, sections...)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	sections = append(sections,
		StyleSubtitle.Render("SESSIONS"),
		StyleBold.Render("  Player          Mode         Hands  Result                Session ID"),
		StyleDim.Render("  "+repeatStr("-", 78)),
	)
	maxVis := m.maxVisible()
	end := min(len(m.sessions), m.scroll+maxVis)
	for i := m.scroll; i < end; i++ {
		entry := m.sessions[i]
		cursor := "  "
		style := StyleMenuItem
		if i == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}
		line := fmt.Sprintf("%-14s %-12s %-5d %-20s %s",
			entry.PlayerName,
			entry.Mode,
			entry.HandsPlayed,
			entry.ResultLabel,
			truncateID(entry.SessionID),
		)
		sections = append(sections, cursor+style.Render(line))
	}
	sections = append(sections, "", StyleDim.Render("[Enter] Open Session  |  [Esc] Back"))
	return renderHistoryBox(m.width, m.height, sections)
}

func (m HistoryViewModel) renderHands(sections []string) string {
	sections = append(sections,
		StyleSubtitle.Render(fmt.Sprintf("SESSION %s", truncateID(m.selected.SessionID))),
		StyleDim.Render(fmt.Sprintf("Player: %s  |  Mode: %s  |  Transcript: %s", m.selected.PlayerName, m.selected.Mode, truncateID(m.selected.TranscriptID))),
		"",
	)
	if len(m.hands) == 0 {
		sections = append(sections,
			StyleDim.Render("No transcript chunks recorded for this session."),
			"",
			StyleDim.Render("[Esc] Back"),
		)
		return renderHistoryBox(m.width, m.height, sections)
	}
	sections = append(sections,
		StyleBold.Render("  Hand  Result               Blinds    Chunk ID"),
		StyleDim.Render("  "+repeatStr("-", 70)),
	)
	maxVis := m.maxVisible()
	end := min(len(m.hands), m.scroll+maxVis)
	for i := m.scroll; i < end; i++ {
		entry := m.hands[i]
		cursor := "  "
		style := StyleMenuItem
		if i == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}
		line := fmt.Sprintf("#%-4d %-20s %2d/%-2d    %s",
			entry.HandID,
			entry.ResultLabel,
			entry.Blinds.SB,
			entry.Blinds.BB,
			truncateID(entry.ChunkID),
		)
		sections = append(sections, cursor+style.Render(line))
	}
	sections = append(sections, "", StyleDim.Render("[Enter] Replay Hand  |  [Esc] Back"))
	return renderHistoryBox(m.width, m.height, sections)
}

func renderHistoryBox(width, height int, sections []string) string {
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func truncateID(id string) string {
	if len(id) <= 18 {
		return id
	}
	return id[:18]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func repeatStr(s string, n int) string {
	result := ""
	for range n {
		result += s
	}
	return result
}
