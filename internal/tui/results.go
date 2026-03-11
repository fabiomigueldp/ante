package tui

import (
	"fmt"

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
		case "enter", " ", "escape", "esc", "q":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ResultsModel) View() string {
	sections := []string{StyleTitle.Render("GAME RESULTS"), "", StyleBold.Render(m.result), ""}
	if m.summary != nil {
		sections = append(sections,
			StyleSubtitle.Render("AUTHORITATIVE SUMMARY"),
			"",
			fmt.Sprintf("  Mode:            %s", m.summary.Mode),
			fmt.Sprintf("  Hands Played:    %d", m.summary.HandsPlayed),
			fmt.Sprintf("  Result:          %s", m.summary.ResultLabel),
			fmt.Sprintf("  Biggest Pot:     %s", ChipStr(m.summary.BiggestPot)),
			fmt.Sprintf("  Best Hand:       %s", m.summary.BestHand),
			fmt.Sprintf("  Largest Win:     %s", ChipStr(m.summary.LargestWin)),
			fmt.Sprintf("  Longest Streak:  %d", m.summary.LongestStreak),
			fmt.Sprintf("  Checkpoint ID:   %s", truncateID(m.summary.CheckpointID)),
		)
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
