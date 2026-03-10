package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const splashArt = `
     █████╗ ███╗   ██╗████████╗███████╗
    ██╔══██╗████╗  ██║╚══██╔══╝██╔════╝
    ███████║██╔██╗ ██║   ██║   █████╗
    ██╔══██║██║╚██╗██║   ██║   ██╔══╝
    ██║  ██║██║ ╚████║   ██║   ███████╗
    ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝

      No-Limit Texas Hold'em Poker
`

const splashDuration = 2 * time.Second

type splashDoneMsg struct{}

type SplashModel struct {
	width  int
	height int
	done   bool
}

func NewSplashModel() SplashModel {
	return SplashModel{}
}

func (m SplashModel) Init() tea.Cmd {
	return tea.Tick(splashDuration, func(t time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

func (m SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m, func() tea.Msg { return splashDoneMsg{} }
	case splashDoneMsg:
		m.done = true
		return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
	}
	return m, nil
}

func (m SplashModel) View() string {
	art := StyleSplash.Render(splashArt)

	subtitle := lipgloss.NewStyle().
		Foreground(ColorGray).
		Italic(true).
		Render("Press any key to continue...")

	content := lipgloss.JoinVertical(lipgloss.Center, art, "", subtitle)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
