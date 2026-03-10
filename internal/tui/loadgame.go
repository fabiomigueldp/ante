package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/storage"
)

// LoadGameModel shows save slots for loading.
type LoadGameModel struct {
	width  int
	height int
	saves  []storage.SaveInfo
	cursor int
}

func NewLoadGameModel() LoadGameModel {
	return LoadGameModel{
		saves: storage.ListSaves(),
	}
}

func (m LoadGameModel) Init() tea.Cmd { return nil }

func (m LoadGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if len(m.saves) == 0 {
			switch msg.String() {
			case "escape", "esc", "q":
				return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "escape", "esc", "q":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.saves) - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.saves) {
				m.cursor = 0
			}
		case "enter":
			if m.cursor < len(m.saves) && !m.saves[m.cursor].Empty {
				// TODO: actually load and resume the saved game
				return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
			}
		case "d", "delete":
			if m.cursor < len(m.saves) && !m.saves[m.cursor].Empty {
				_ = storage.DeleteSave(m.saves[m.cursor].Slot)
				m.saves = storage.ListSaves()
			}
		}
	}
	return m, nil
}

func (m LoadGameModel) View() string {
	title := StyleTitle.Render("LOAD GAME")

	var sections []string
	sections = append(sections, title, "")

	if len(m.saves) == 0 {
		sections = append(sections, StyleDim.Render("No save slots available."))
		sections = append(sections, "", StyleDim.Render("[Esc] Back"))
		content := lipgloss.JoinVertical(lipgloss.Center, sections...)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	for i, save := range m.saves {
		cursor := "  "
		style := StyleMenuItem
		if i == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}

		if save.Empty {
			line := fmt.Sprintf("Slot %d: %s", save.Slot, StyleDim.Render("(empty)"))
			sections = append(sections, cursor+style.Render(line))
		} else {
			line := fmt.Sprintf("Slot %d: %s  %s  Hand #%d  Stack: %s  %s",
				save.Slot,
				save.Name,
				save.Mode,
				save.HandNum,
				ChipStr(save.Stack),
				StyleDim.Render(save.Timestamp.Format("Jan 2 15:04")),
			)
			sections = append(sections, cursor+style.Render(line))
		}
	}

	sections = append(sections, "")
	sections = append(sections, StyleDim.Render("[Enter] Load  |  [D] Delete  |  [Esc] Back"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
