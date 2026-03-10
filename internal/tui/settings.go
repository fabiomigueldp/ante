package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/storage"
)

type settingsField int

const (
	sfPlayerName settingsField = iota
	sfSound
	sfPotOdds
	sfAnimSpeed
	sfDefaultMode
	sfDefaultDiff
	sfDefaultSeats
	sfStartStack
	sfTheme
	sfSave
	sfFieldCount
)

// SettingsModel provides an interactive settings editor.
type SettingsModel struct {
	width    int
	height   int
	cursor   settingsField
	config   storage.Config
	nameEdit bool
	saved    bool
}

func NewSettingsModel(cfg storage.Config) SettingsModel {
	return SettingsModel{
		config: cfg,
	}
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.nameEdit {
			return m.handleNameEdit(msg)
		}
		switch msg.String() {
		case "escape":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = sfFieldCount - 1
			}
			m.saved = false
		case "down", "j":
			m.cursor++
			if m.cursor >= sfFieldCount {
				m.cursor = 0
			}
			m.saved = false
		case "left", "h":
			m.adjustField(-1)
			m.saved = false
		case "right", "l":
			m.adjustField(1)
			m.saved = false
		case "enter", " ":
			if m.cursor == sfPlayerName {
				m.nameEdit = true
				return m, nil
			}
			if m.cursor == sfSave {
				return m.saveConfig()
			}
			m.adjustField(1)
			m.saved = false
		}
	}
	return m, nil
}

func (m *SettingsModel) handleNameEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "escape":
		m.nameEdit = false
		if m.config.PlayerName == "" {
			m.config.PlayerName = "Player"
		}
	case "backspace":
		runes := []rune(m.config.PlayerName)
		if len(runes) > 0 {
			m.config.PlayerName = string(runes[:len(runes)-1])
		}
	default:
		ch := msg.String()
		if len([]rune(ch)) == 1 && len([]rune(m.config.PlayerName)) < 16 {
			m.config.PlayerName += ch
		}
	}
	return *m, nil
}

func (m *SettingsModel) adjustField(dir int) {
	switch m.cursor {
	case sfSound:
		m.config.SoundEnabled = !m.config.SoundEnabled
	case sfPotOdds:
		m.config.ShowPotOdds = !m.config.ShowPotOdds
	case sfAnimSpeed:
		speeds := []string{"slow", "normal", "fast", "off"}
		idx := indexOf(speeds, m.config.AnimationSpeed)
		idx = (idx + dir + len(speeds)) % len(speeds)
		m.config.AnimationSpeed = speeds[idx]
	case sfDefaultMode:
		modes := []string{"tournament", "cash", "headsup"}
		idx := indexOf(modes, m.config.DefaultMode)
		idx = (idx + dir + len(modes)) % len(modes)
		m.config.DefaultMode = modes[idx]
	case sfDefaultDiff:
		diffs := []string{"easy", "medium", "hard"}
		idx := indexOf(diffs, m.config.DefaultDiff)
		idx = (idx + dir + len(diffs)) % len(diffs)
		m.config.DefaultDiff = diffs[idx]
	case sfDefaultSeats:
		seats := []int{2, 6, 9}
		idx := 0
		for i, s := range seats {
			if s == m.config.DefaultSeats {
				idx = i
			}
		}
		idx = (idx + dir + len(seats)) % len(seats)
		m.config.DefaultSeats = seats[idx]
	case sfStartStack:
		stacks := []int{50, 100, 200}
		idx := 0
		for i, s := range stacks {
			if s == m.config.StartingStack {
				idx = i
			}
		}
		idx = (idx + dir + len(stacks)) % len(stacks)
		m.config.StartingStack = stacks[idx]
	case sfTheme:
		themes := []string{"classic", "dark", "green"}
		idx := indexOf(themes, m.config.Theme)
		idx = (idx + dir + len(themes)) % len(themes)
		m.config.Theme = themes[idx]
	}
}

func (m SettingsModel) saveConfig() (tea.Model, tea.Cmd) {
	_ = storage.SaveConfig(m.config)
	m.saved = true
	return m, nil
}

func indexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return 0
}

func (m SettingsModel) View() string {
	title := StyleTitle.Render("SETTINGS")

	type fieldDisplay struct {
		label string
		value string
	}

	nameDisplay := m.config.PlayerName
	if m.nameEdit {
		nameDisplay += "_"
	}

	soundStr := "Off"
	if m.config.SoundEnabled {
		soundStr = "On"
	}
	potOddsStr := "Off"
	if m.config.ShowPotOdds {
		potOddsStr = "On"
	}

	fields := []fieldDisplay{
		{"Player Name", nameDisplay},
		{"Sound (Bell)", soundStr},
		{"Pot Odds Helper", potOddsStr},
		{"Animation Speed", m.config.AnimationSpeed},
		{"Default Mode", m.config.DefaultMode},
		{"Default Difficulty", m.config.DefaultDiff},
		{"Default Seats", fmt.Sprintf("%d", m.config.DefaultSeats)},
		{"Starting Stack", fmt.Sprintf("%d BB", m.config.StartingStack)},
		{"Theme", m.config.Theme},
		{"", "[ SAVE SETTINGS ]"},
	}

	var rows string
	for i, f := range fields {
		cursor := "  "
		style := StyleMenuItem
		if settingsField(i) == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}

		label := ""
		if f.label != "" {
			label = StyleDim.Render(fmt.Sprintf("%-20s", f.label))
		}

		value := style.Render(f.value)
		if settingsField(i) == m.cursor && settingsField(i) != sfPlayerName && settingsField(i) != sfSave {
			value = StyleKey.Render("< ") + value + StyleKey.Render(" >")
		}

		rows += fmt.Sprintf("%s%s %s\n", cursor, label, value)
	}

	savedMsg := ""
	if m.saved {
		savedMsg = StyleSuccess.Render("Settings saved!")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(rows)

	help := StyleDim.Render("Arrow keys to navigate  |  Left/Right to change  |  Enter to select  |  Esc to go back")

	var sections []string
	sections = append(sections, title, "", box)
	if savedMsg != "" {
		sections = append(sections, savedMsg)
	}
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
