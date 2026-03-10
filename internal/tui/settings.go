package tui

import (
	"fmt"

	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingsField int

const (
	sfPlayerName settingsField = iota
	sfSound
	sfVolume
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

var (
	saveSettingsConfig = storage.SaveConfig
	setAudioEnabled    = audio.SetEnabled
	setAudioVolume     = func(v int) { audio.SetVolume(float64(v) / 100.0) }
	playSettingsSound  = func() { audio.Play(audio.SoundChip) }
)

// SettingsModel provides an interactive settings editor.
type SettingsModel struct {
	width         int
	height        int
	cursor        settingsField
	config        storage.Config
	initialConfig storage.Config
	nameEdit      bool
	saved         bool
}

func NewSettingsModel(cfg storage.Config) SettingsModel {
	cfg.SoundVolume = clampSettingVolume(cfg.SoundVolume)
	return SettingsModel{
		config:        cfg,
		initialConfig: cfg,
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
		case "escape", "esc":
			m.restoreAudioPreview()
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			m.restoreAudioPreview()
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
			m.saved = false
		case "down", "j":
			m.moveCursor(1)
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
	case "enter", "escape", "esc":
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

func (m *SettingsModel) moveCursor(dir int) {
	fields := m.visibleFields()
	idx := 0
	for i, field := range fields {
		if field == m.cursor {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(fields)) % len(fields)
	m.cursor = fields[idx]
}

func (m *SettingsModel) adjustField(dir int) {
	switch m.cursor {
	case sfSound:
		m.config.SoundEnabled = !m.config.SoundEnabled
		m.applyAudioPreview(true)
		if !m.config.SoundEnabled && m.cursor == sfVolume {
			m.cursor = sfPotOdds
		}
	case sfVolume:
		if !m.config.SoundEnabled {
			return
		}
		m.config.SoundVolume = clampSettingVolume(m.config.SoundVolume + dir*10)
		m.applyAudioPreview(true)
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

	if !m.config.SoundEnabled && m.cursor == sfVolume {
		m.cursor = sfPotOdds
	}
}

func (m SettingsModel) saveConfig() (tea.Model, tea.Cmd) {
	m.config.SoundVolume = clampSettingVolume(m.config.SoundVolume)
	if err := saveSettingsConfig(m.config); err != nil {
		return m, func() tea.Msg {
			return appErrorMsg{message: fmt.Sprintf("Unable to save settings: %v", err)}
		}
	}
	m.initialConfig = m.config
	m.saved = true
	m.applyAudioPreview(false)
	return m, func() tea.Msg { return settingsSavedMsg{config: m.config} }
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
		field settingsField
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

	fields := []fieldDisplay{{field: sfPlayerName, label: "Player Name", value: nameDisplay}, {field: sfSound, label: "Sound", value: soundStr}}
	if m.config.SoundEnabled {
		fields = append(fields, fieldDisplay{field: sfVolume, label: "Volume", value: fmt.Sprintf("%d%%", m.config.SoundVolume)})
	}
	fields = append(fields,
		fieldDisplay{field: sfPotOdds, label: "Pot Odds Helper", value: potOddsStr},
		fieldDisplay{field: sfAnimSpeed, label: "Animation Speed", value: m.config.AnimationSpeed},
		fieldDisplay{field: sfDefaultMode, label: "Default Mode", value: m.config.DefaultMode},
		fieldDisplay{field: sfDefaultDiff, label: "Default Difficulty", value: m.config.DefaultDiff},
		fieldDisplay{field: sfDefaultSeats, label: "Default Seats", value: fmt.Sprintf("%d", m.config.DefaultSeats)},
		fieldDisplay{field: sfStartStack, label: "Starting Stack", value: fmt.Sprintf("%d BB", m.config.StartingStack)},
		fieldDisplay{field: sfTheme, label: "Theme", value: m.config.Theme},
		fieldDisplay{field: sfSave, label: "", value: "[ SAVE SETTINGS ]"},
	)

	var rows string
	for _, f := range fields {
		cursor := "  "
		style := StyleMenuItem
		if f.field == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}

		label := ""
		if f.label != "" {
			label = StyleDim.Render(fmt.Sprintf("%-20s", f.label))
		}

		value := style.Render(f.value)
		if f.field == m.cursor && f.field != sfPlayerName && f.field != sfSave {
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

func (m *SettingsModel) visibleFields() []settingsField {
	fields := []settingsField{sfPlayerName, sfSound}
	if m.config.SoundEnabled {
		fields = append(fields, sfVolume)
	}
	fields = append(fields, sfPotOdds, sfAnimSpeed, sfDefaultMode, sfDefaultDiff, sfDefaultSeats, sfStartStack, sfTheme, sfSave)
	return fields
}

func (m SettingsModel) applyAudioPreview(playSample bool) {
	setAudioEnabled(m.config.SoundEnabled)
	setAudioVolume(clampSettingVolume(m.config.SoundVolume))
	if playSample && m.config.SoundEnabled {
		playSettingsSound()
	}
}

func (m SettingsModel) restoreAudioPreview() {
	setAudioEnabled(m.initialConfig.SoundEnabled)
	setAudioVolume(clampSettingVolume(m.initialConfig.SoundVolume))
}

func clampSettingVolume(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
