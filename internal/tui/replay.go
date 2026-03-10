package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/engine"
)

// ReplayModel replays a recorded hand step by step.
type ReplayModel struct {
	width  int
	height int
	record *engine.HandRecord
	step   int // current action index
	total  int
}

func NewReplayModel(record *engine.HandRecord) ReplayModel {
	total := 0
	if record != nil {
		total = len(record.Actions)
	}
	return ReplayModel{
		record: record,
		total:  total,
	}
}

func (m ReplayModel) Init() tea.Cmd { return nil }

func (m ReplayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "escape", "esc", "q":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenHistory} }
		case "ctrl+c":
			return m, tea.Quit
		case "right", "l", "enter", " ":
			if m.step < m.total {
				m.step++
			}
		case "left", "h":
			if m.step > 0 {
				m.step--
			}
		case "home", "0":
			m.step = 0
		case "end", "$":
			m.step = m.total
		}
	}
	return m, nil
}

func (m ReplayModel) View() string {
	if m.record == nil {
		content := lipgloss.JoinVertical(lipgloss.Center,
			StyleTitle.Render("HAND REPLAY"),
			"",
			StyleDim.Render("No hand data to replay."),
			"",
			StyleDim.Render("[Esc] Back"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	title := StyleTitle.Render(fmt.Sprintf("HAND #%d REPLAY", m.record.HandID))

	var sections []string
	sections = append(sections, title, "")

	// Hand info
	sections = append(sections, fmt.Sprintf("  Blinds: %d/%d", m.record.Blinds.SB, m.record.Blinds.BB))
	sections = append(sections, fmt.Sprintf("  Dealer: Seat %d", m.record.DealerSeat))

	// Players
	sections = append(sections, "")
	sections = append(sections, StyleSubtitle.Render("  Players:"))
	for _, p := range m.record.Players {
		sections = append(sections, fmt.Sprintf("    Seat %d: %s (%s)", p.Seat, p.Name, ChipStr(p.Stack)))
	}

	// Board (show cards dealt up to current step)
	sections = append(sections, "")
	board := m.boardAtStep()
	if len(board) > 0 {
		sections = append(sections, "  Board: "+RenderBoard(board))
	} else {
		sections = append(sections, "  Board: "+StyleDim.Render("(none)"))
	}

	// Actions up to current step
	sections = append(sections, "")
	sections = append(sections, StyleSubtitle.Render("  Actions:"))

	if m.total == 0 {
		sections = append(sections, StyleDim.Render("    (no actions recorded)"))
	} else {
		for i := 0; i < m.step && i < m.total; i++ {
			action := m.record.Actions[i]
			name := m.playerName(action.PlayerID)
			actStr := ActionStr(action.Type)
			line := fmt.Sprintf("    %s %s", name, actStr)
			if action.Amount > 0 {
				line += " " + ChipStr(action.Amount)
			}
			sections = append(sections, line)
		}
		if m.step < m.total {
			sections = append(sections, StyleDim.Render("    ..."))
		}
	}

	// Progress
	sections = append(sections, "")
	progress := fmt.Sprintf("  Step %d / %d", m.step, m.total)
	bar := m.renderProgressBar()
	sections = append(sections, progress+"  "+bar)

	sections = append(sections, "")
	sections = append(sections, StyleDim.Render("[Left/Right] Step  |  [Home/End] Jump  |  [Esc] Back"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m ReplayModel) boardAtStep() []engine.Card {
	if m.record == nil || len(m.record.Board) == 0 {
		return nil
	}
	// Show board cards based on the current action step
	// Simple heuristic: show board as-is up to what's been dealt
	if m.step >= m.total {
		return m.record.Board
	}
	// Count street transitions to determine visible board cards
	// For simplicity, show all board cards once we're past preflop
	streets := 0
	for i := 0; i < m.step && i < len(m.record.Actions); i++ {
		if m.record.Actions[i].Type == engine.ActionFold ||
			m.record.Actions[i].Type == engine.ActionCheck ||
			m.record.Actions[i].Type == engine.ActionCall ||
			m.record.Actions[i].Type == engine.ActionBet ||
			m.record.Actions[i].Type == engine.ActionRaise ||
			m.record.Actions[i].Type == engine.ActionAllIn {
			// Count approximate street boundaries
		}
		_ = streets
	}
	// For now, just show all board cards when past first action
	if m.step > 0 {
		return m.record.Board
	}
	return nil
}

func (m ReplayModel) playerName(pid engine.PlayerID) string {
	if m.record == nil {
		return fmt.Sprintf("P%d", pid)
	}
	for _, p := range m.record.Players {
		if p.ID == pid {
			return p.Name
		}
	}
	return fmt.Sprintf("P%d", pid)
}

func (m ReplayModel) renderProgressBar() string {
	w := 20
	if m.total == 0 {
		return StyleDim.Render("[" + strings.Repeat("-", w) + "]")
	}
	filled := w * m.step / m.total
	if filled > w {
		filled = w
	}
	bar := strings.Repeat("=", filled) + strings.Repeat("-", w-filled)
	return StyleDim.Render("[") + StyleChips.Render(bar[:filled]) + StyleDim.Render(bar[filled:]+"]")
}
