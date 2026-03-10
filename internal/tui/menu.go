package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	label string
	desc  string
	key   string
}

var mainMenuItems = []menuItem{
	{label: "New Game", desc: "Start a new poker session", key: "n"},
	{label: "Continue", desc: "Load a saved game", key: "c"},
	{label: "Statistics", desc: "View your poker stats", key: "s"},
	{label: "Hand History", desc: "Browse past hands", key: "h"},
	{label: "Settings", desc: "Configure game options", key: "o"},
	{label: "Help", desc: "How to play & controls", key: "?"},
	{label: "Quit", desc: "Exit the game", key: "q"},
}

type MenuModel struct {
	cursor int
	width  int
	height int
}

func NewMenuModel() MenuModel {
	return MenuModel{}
}

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(mainMenuItems) - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(mainMenuItems) {
				m.cursor = 0
			}
		case "enter", " ":
			return m, m.selectItem(m.cursor)
		case "n":
			return m, m.selectItem(0)
		case "c":
			return m, m.selectItem(1)
		case "s":
			return m, m.selectItem(2)
		case "h":
			return m, m.selectItem(3)
		case "o":
			return m, m.selectItem(4)
		case "?":
			return m, m.selectItem(5)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) selectItem(index int) tea.Cmd {
	switch index {
	case 6:
		return tea.Quit
	default:
		return func() tea.Msg {
			switch index {
			case 0:
				return switchScreenMsg{screen: ScreenSetup}
			case 1:
				return switchScreenMsg{screen: ScreenLoadGame}
			case 2:
				return switchScreenMsg{screen: ScreenStats}
			case 3:
				return switchScreenMsg{screen: ScreenHistory}
			case 4:
				return switchScreenMsg{screen: ScreenSettings}
			case 5:
				return switchScreenMsg{screen: ScreenHelp}
			}
			return nil
		}
	}
}

func (m MenuModel) View() string {
	title := StyleSplash.Render(`
   █████╗ ███╗   ██╗████████╗███████╗
  ██╔══██╗████╗  ██║╚══██╔══╝██╔════╝
  ███████║██╔██╗ ██║   ██║   █████╗
  ██╔══██║██║╚██╗██║   ██║   ██╔══╝
  ██║  ██║██║ ╚████║   ██║   ███████╗
  ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝`)

	var items string
	for i, item := range mainMenuItems {
		cursor := "  "
		style := StyleMenuItem
		if i == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}
		key := StyleKey.Render(fmt.Sprintf("[%s]", item.key))
		label := style.Render(item.label)
		desc := StyleDim.Render(item.desc)
		items += fmt.Sprintf("%s%s %s  %s\n", cursor, key, label, desc)
	}

	menuBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(items)

	footer := StyleDim.Render("Use arrow keys or highlighted keys to navigate")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title, "", menuBox, "", footer,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
