package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/session"
)

// ResultsModel shows end-of-game results.
type ResultsModel struct {
	width  int
	height int
	result string
	sess   *session.Session
}

func NewResultsModel(result string, sess *session.Session) ResultsModel {
	return ResultsModel{
		result: result,
		sess:   sess,
	}
}

func (m ResultsModel) Init() tea.Cmd { return nil }

func (m ResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ", "escape", "q":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ResultsModel) View() string {
	title := StyleTitle.Render("GAME RESULTS")

	var sections []string
	sections = append(sections, title, "")

	// Main result message
	resultLine := StyleBold.Render(m.result)
	sections = append(sections, resultLine, "")

	// Session stats if available
	if m.sess != nil {
		sections = append(sections, m.renderStats()...)
	}

	sections = append(sections, "", StyleDim.Render("Press any key to return to menu..."))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGold).
		Padding(2, 4).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m ResultsModel) renderStats() []string {
	if m.sess == nil {
		return nil
	}

	var lines []string

	lines = append(lines, StyleSubtitle.Render("SESSION SUMMARY"))
	lines = append(lines, "")

	handsPlayed := m.sess.HandCount
	lines = append(lines, fmt.Sprintf("  Hands Played:  %s", StyleChips.Render(ChipStr(handsPlayed))))

	// Final standings from table
	ts := m.sess.TableState()
	lines = append(lines, "")
	lines = append(lines, StyleSubtitle.Render("FINAL STANDINGS"))
	lines = append(lines, "")

	// Sort by stack descending
	players := make([]session.PlayerInfo, len(ts.Players))
	copy(players, ts.Players)
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[j].Stack > players[i].Stack {
				players[i], players[j] = players[j], players[i]
			}
		}
	}

	for rank, p := range players {
		name := p.Name
		if p.IsHuman {
			name = StyleHumanLabel.Render(name)
		}
		status := ""
		if p.Stack == 0 {
			status = StyleDim.Render(" (eliminated)")
		}

		lines = append(lines, fmt.Sprintf("  #%d  %-16s  %s%s",
			rank+1, name, StyleChips.Render(ChipStr(p.Stack)), status))
	}

	// History summary
	if m.sess.History != nil && len(m.sess.History.Records) > 0 {
		lines = append(lines, "")
		biggest := 0
		for _, r := range m.sess.History.Records {
			for _, e := range r.Events {
				if strings.Contains(e.EventType(), "pot_awarded") {
					// We can't easily extract amount from interface, skip
				}
			}
			_ = biggest
		}
	}

	return lines
}
