package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"
)

var loadResultSummary = storage.LoadSessionSummaryResult

// ResultsModel shows end-of-game results.
type ResultsModel struct {
	width   int
	height  int
	result  string
	sess    *session.Session
	summary *storage.SessionSummary
}

type resultsRow struct {
	label      string
	value      string
	alignRight bool
}

func NewResultsModel(result string, sess *session.Session) ResultsModel {
	model := ResultsModel{result: result, sess: sess}
	if sess != nil && sess.Summary != nil {
		model.summary = sess.Summary
	} else if sess != nil && sess.SessionID != "" {
		if summary, err := loadResultSummary(sess.SessionID); err == nil {
			model.summary = summary
		}
	}
	return model
}

func (m ResultsModel) Init() tea.Cmd { return nil }

func (m ResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ResultsModel) View() string {
	panelWidth := m.resultsPanelWidth()
	innerWidth := panelWidth - 8
	sections := []string{
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(StyleTitle.Render("GAME RESULTS")),
		"",
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(StyleBold.Render(m.result)),
	}
	if m.summary != nil {
		sections = append(sections,
			"",
			m.renderResultsSection("Session", []resultsRow{
				{label: "Mode", value: resultsModeLabel(m.summary.Mode)},
				{label: "Hands Played", value: fmt.Sprintf("%d", m.summary.HandsPlayed), alignRight: true},
				{label: "Result", value: m.summary.ResultLabel},
			}),
			"",
			m.renderResultsSection("Highlights", []resultsRow{
				{label: "Biggest Pot", value: ChipStr(m.summary.BiggestPot), alignRight: true},
				{label: "Best Hand", value: m.summary.BestHand},
				{label: "Largest Win", value: ChipStr(m.summary.LargestWin), alignRight: true},
				{label: "Longest Streak", value: fmt.Sprintf("%d", m.summary.LongestStreak), alignRight: true},
			}),
			"",
			m.renderResultsSection("Integrity", []resultsRow{{label: "Checkpoint ID", value: truncateID(m.summary.CheckpointID)}}),
		)
	} else {
		sections = append(sections, "", lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(StyleDim.Render("No persisted summary available.")))
	}
	sections = append(sections, "", lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(StyleDim.Render("[Enter] Return to Menu")))
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	box := lipgloss.NewStyle().
		Width(panelWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGold).
		Padding(2, 4).
		Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m ResultsModel) resultsPanelWidth() int {
	if m.width <= 0 {
		return 76
	}
	width := m.width - 16
	if width > 76 {
		width = 76
	}
	if width < 44 {
		width = 44
	}
	return width
}

func (m ResultsModel) renderResultsSection(title string, rows []resultsRow) string {
	innerWidth := m.resultsPanelWidth() - 8
	labelWidth := 0
	for _, row := range rows {
		labelWidth = max(labelWidth, lipgloss.Width(row.label))
	}
	if labelWidth < 12 {
		labelWidth = 12
	}
	if labelWidth > 18 {
		labelWidth = 18
	}
	valueWidth := innerWidth - labelWidth - 3
	if valueWidth < 10 {
		valueWidth = 10
	}
	lines := []string{StyleSubtitle.Render(strings.ToUpper(title))}
	for _, row := range rows {
		label := lipgloss.NewStyle().Width(labelWidth).Foreground(ColorGray).Render(truncateText(row.label, labelWidth))
		valueStyle := lipgloss.NewStyle().Width(valueWidth).Foreground(ColorWhite)
		if row.alignRight {
			valueStyle = valueStyle.Align(lipgloss.Right)
		}
		value := valueStyle.Render(truncateText(row.value, valueWidth))
		lines = append(lines, label+" : "+value)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func resultsModeLabel(mode string) string {
	switch mode {
	case "cash":
		return "Cash Game"
	case "headsup":
		return "Heads-Up"
	default:
		return "Tournament"
	}
}
