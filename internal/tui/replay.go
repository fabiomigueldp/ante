package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

// ReplayModel replays a transcript-backed hand step by step.
type ReplayModel struct {
	width  int
	height int
	chunk  *storage.TranscriptChunk
	step   int
	total  int
}

func NewReplayModel(chunk *storage.TranscriptChunk) ReplayModel {
	total := 0
	if chunk != nil {
		total = len(chunk.Records)
	}
	return ReplayModel{chunk: chunk, total: total}
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
	if m.chunk == nil {
		content := lipgloss.JoinVertical(lipgloss.Center,
			StyleTitle.Render("HAND REPLAY"),
			"",
			StyleDim.Render("No transcript data to replay."),
			"",
			StyleDim.Render("[Esc] Back"),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	sections := []string{
		StyleTitle.Render(fmt.Sprintf("HAND #%d REPLAY", m.chunk.HandID)),
		"",
		fmt.Sprintf("  Blinds: %d/%d", m.chunk.Blinds.SB, m.chunk.Blinds.BB),
		fmt.Sprintf("  Dealer: Seat %d", m.chunk.DealerSeat),
		fmt.Sprintf("  Chunk: %s", truncateID(m.chunk.ID)),
		"",
		StyleSubtitle.Render("  Players:"),
	}
	for _, player := range m.chunk.Players {
		sections = append(sections, fmt.Sprintf("    Seat %d: %s (%s)", player.Seat, player.Name, ChipStr(player.Stack)))
	}
	sections = append(sections, "", "  Board: "+m.renderBoard(), "", StyleSubtitle.Render("  Timeline:"))
	if m.total == 0 {
		sections = append(sections, StyleDim.Render("    (no transcript records)"))
	} else {
		for i := 0; i < m.step && i < m.total; i++ {
			sections = append(sections, "    "+m.describeRecord(m.chunk.Records[i]))
		}
		if m.step < m.total {
			sections = append(sections, StyleDim.Render("    ..."))
		}
	}
	sections = append(sections,
		"",
		fmt.Sprintf("  Step %d / %d  %s", m.step, m.total, m.renderProgressBar()),
		"",
		StyleDim.Render("[Left/Right] Step  |  [Home/End] Jump  |  [Esc] Back"),
	)
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m ReplayModel) renderBoard() string {
	board := m.visibleBoard()
	if len(board) == 0 {
		return StyleDim.Render("(none)")
	}
	return RenderBoard(board)
}

func (m ReplayModel) visibleBoard() []engine.Card {
	if m.chunk == nil {
		return nil
	}
	board := make([]engine.Card, 0, 5)
	for i := 0; i < m.step && i < len(m.chunk.Records); i++ {
		record := m.chunk.Records[i]
		if record.Kind == "street_advanced" {
			board = append(board, record.NewCards...)
		}
	}
	return board
}

func (m ReplayModel) describeRecord(record storage.TranscriptRecord) string {
	switch record.Kind {
	case "street_advanced":
		return fmt.Sprintf("%s -> %s", StreetStr(record.Street), RenderBoard(record.NewCards))
	case "action_taken":
		name := m.playerName(record.PlayerID)
		if record.Action == nil {
			return name + " acts"
		}
		line := fmt.Sprintf("%s %s", name, ActionStr(record.Action.Type))
		if record.Action.Amount > 0 {
			line += " " + ChipStr(record.Action.Amount)
		}
		return line
	case "hand_revealed":
		return fmt.Sprintf("%s shows %s  %s", m.playerName(record.PlayerID), RenderHoleCards(record.ShownCards, true), StyleHandRank.Render(record.EvalName))
	case "pot_awarded":
		winners := make([]string, 0, len(record.Winners))
		for _, winner := range record.Winners {
			winners = append(winners, m.playerName(winner))
		}
		return fmt.Sprintf("%s wins %s", strings.Join(winners, ", "), ChipStr(record.AwardAmount))
	case "showdown_started":
		return "Showdown"
	case "blind_posted":
		return fmt.Sprintf("%s posts %s", m.playerName(record.PlayerID), ChipStr(record.AwardAmount))
	default:
		if record.Message != "" {
			return record.Message
		}
		return record.Kind
	}
}

func (m ReplayModel) playerName(pid engine.PlayerID) string {
	if m.chunk == nil {
		return fmt.Sprintf("P%d", pid)
	}
	for _, player := range m.chunk.Players {
		if player.ID == pid {
			return player.Name
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
