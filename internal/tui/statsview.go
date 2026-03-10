package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/storage"
)

// StatsViewModel shows aggregate player statistics.
type StatsViewModel struct {
	width  int
	height int
	store  *storage.StatsStore
	scroll int
}

func NewStatsViewModel() StatsViewModel {
	return StatsViewModel{
		store: storage.LoadStats(),
	}
}

func (m StatsViewModel) Init() tea.Cmd { return nil }

func (m StatsViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			m.scroll++
		}
	}
	return m, nil
}

func (m StatsViewModel) View() string {
	title := StyleTitle.Render("STATISTICS")

	if m.store.TotalSessions() == 0 {
		empty := lipgloss.JoinVertical(lipgloss.Center,
			title, "",
			StyleDim.Render("No sessions played yet."),
			StyleDim.Render("Start a new game to begin tracking your stats!"),
			"",
			StyleDim.Render("[Esc] Back"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, empty)
	}

	var sections []string
	sections = append(sections, title, "")

	// Aggregate stats
	sections = append(sections, StyleSubtitle.Render("OVERVIEW"))
	sections = append(sections, "")
	sections = append(sections, fmt.Sprintf("  Sessions Played:    %d", m.store.TotalSessions()))
	sections = append(sections, fmt.Sprintf("  Hands Played:       %d", m.store.TotalHandsPlayed()))
	sections = append(sections, fmt.Sprintf("  Tournament Wins:    %d", m.store.TournamentWins()))
	sections = append(sections, fmt.Sprintf("  Total Profit:       %s", ChipStr(m.store.TotalProfit())))
	sections = append(sections, fmt.Sprintf("  Hand Win Rate:      %.1f%%", m.store.WinRate()))
	sections = append(sections, fmt.Sprintf("  Avg Tournament:     #%.1f", m.store.AvgFinish()))
	sections = append(sections, fmt.Sprintf("  Best Hand:          %s", m.store.BestHandEver()))
	sections = append(sections, "")

	// Recent sessions
	recent := m.store.RecentSessions(10)
	if len(recent) > 0 {
		sections = append(sections, StyleSubtitle.Render("RECENT SESSIONS"))
		sections = append(sections, "")
		sections = append(sections, fmt.Sprintf("  %-6s %-12s %-8s %-10s %-8s",
			"#", "Mode", "Hands", "Result", "Profit"))
		sections = append(sections, StyleDim.Render("  "+fmt.Sprintf("%s", "----------------------------------------------")))

		for i, sess := range recent {
			result := ""
			if sess.Mode == "tournament" || sess.Mode == "headsup" {
				result = fmt.Sprintf("#%d/%d", sess.FinalPosition, sess.TotalPlayers)
			} else {
				result = "Cash"
			}
			profit := ChipStr(sess.ChipsWon)
			if sess.ChipsWon >= 0 {
				profit = "+" + profit
			}
			sections = append(sections, fmt.Sprintf("  %-6d %-12s %-8d %-10s %-8s",
				i+1, sess.Mode, sess.HandsPlayed, result, profit))
		}
	}

	sections = append(sections, "")
	sections = append(sections, StyleDim.Render("[Esc] Back"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
