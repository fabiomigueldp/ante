package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

// HistoryViewModel browses saved hand history.
type HistoryViewModel struct {
	width   int
	height  int
	store   *storage.StatsStore
	records []handHistoryEntry
	cursor  int
	scroll  int
}

type handHistoryEntry struct {
	index     int
	sessionID string
	handNum   int
	mode      string
	result    string
}

func NewHistoryViewModel() HistoryViewModel {
	store := storage.LoadStats()
	var entries []handHistoryEntry
	for i, sess := range store.Sessions {
		entries = append(entries, handHistoryEntry{
			index:     i,
			sessionID: sess.ID,
			handNum:   sess.HandsPlayed,
			mode:      sess.Mode,
			result:    resultStr(sess),
		})
	}
	return HistoryViewModel{
		store:   store,
		records: entries,
	}
}

func resultStr(sess storage.SessionStats) string {
	if sess.Mode == "tournament" || sess.Mode == "headsup" {
		if sess.FinalPosition == 1 {
			return "Winner!"
		}
		return fmt.Sprintf("#%d of %d", sess.FinalPosition, sess.TotalPlayers)
	}
	if sess.ChipsWon >= 0 {
		return fmt.Sprintf("+%s chips", ChipStr(sess.ChipsWon))
	}
	return fmt.Sprintf("%s chips", ChipStr(sess.ChipsWon))
}

func (m HistoryViewModel) Init() tea.Cmd { return nil }

func (m HistoryViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "escape", "q":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scroll {
					m.scroll = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.records)-1 {
				m.cursor++
				maxVisible := m.maxVisible()
				if m.cursor >= m.scroll+maxVisible {
					m.scroll = m.cursor - maxVisible + 1
				}
			}
		case "enter":
			// TODO: open replay for selected session
		}
	}
	return m, nil
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

	if len(m.records) == 0 {
		empty := lipgloss.JoinVertical(lipgloss.Center,
			title, "",
			StyleDim.Render("No sessions recorded yet."),
			StyleDim.Render("Play some hands to build your history!"),
			"",
			StyleDim.Render("[Esc] Back"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, empty)
	}

	var sections []string
	sections = append(sections, title, "")

	header := fmt.Sprintf("  %-4s %-14s %-8s %-20s", "#", "Mode", "Hands", "Result")
	sections = append(sections, StyleBold.Render(header))
	sections = append(sections, StyleDim.Render("  "+repeatStr("-", 50)))

	maxVis := m.maxVisible()
	end := m.scroll + maxVis
	if end > len(m.records) {
		end = len(m.records)
	}

	for i := m.scroll; i < end; i++ {
		entry := m.records[i]
		cursor := "  "
		style := StyleMenuItem
		if i == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}
		line := fmt.Sprintf("%-4d %-14s %-8d %-20s",
			entry.index+1, entry.mode, entry.handNum, entry.result)
		sections = append(sections, cursor+style.Render(line))
	}

	if m.scroll > 0 {
		sections = append(sections, StyleDim.Render("  ... more above"))
	}
	if end < len(m.records) {
		sections = append(sections, StyleDim.Render("  ... more below"))
	}

	sections = append(sections, "")
	sections = append(sections, StyleDim.Render("[Esc] Back  |  [Enter] View Details"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func repeatStr(s string, n int) string {
	result := ""
	for range n {
		result += s
	}
	return result
}

// Ensure we use engine to avoid import error
var _ = engine.Card{}
