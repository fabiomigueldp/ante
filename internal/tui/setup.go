package tui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"
)

type setupField int

const (
	fieldMode setupField = iota
	fieldSeats
	fieldDifficulty
	fieldStack
	fieldName
	fieldStart
	fieldCount
)

type SetupModel struct {
	width      int
	height     int
	cursor     setupField
	mode       int    // 0=tournament, 1=cash, 2=headsup
	seats      int    // 2,6,9
	difficulty int    // 0=easy, 1=medium, 2=hard
	stack      int    // starting stack in BB
	name       string // player name
	nameEdit   bool
	config     storage.Config
}

var modeNames = []string{"Tournament (Sit & Go)", "Cash Game", "Heads-Up Duel"}
var modeValues = []engine.GameMode{engine.ModeTournament, engine.ModeCashGame, engine.ModeHeadsUpDuel}
var diffNames = []string{"Easy", "Medium", "Hard"}
var diffValues = []ai.Difficulty{ai.DifficultyEasy, ai.DifficultyMedium, ai.DifficultyHard}
var seatOptions = []int{6, 9, 2}
var stackOptions = []int{50, 100, 200}

func NewSetupModel(cfg storage.Config) SetupModel {
	m := SetupModel{
		seats:  cfg.DefaultSeats,
		stack:  cfg.StartingStack,
		name:   cfg.PlayerName,
		config: cfg,
	}
	switch cfg.DefaultDiff {
	case "easy":
		m.difficulty = 0
	case "hard":
		m.difficulty = 2
	default:
		m.difficulty = 1
	}
	switch cfg.DefaultMode {
	case "cash":
		m.mode = 1
	case "headsup":
		m.mode = 2
	default:
		m.mode = 0
	}
	return m
}

func (m SetupModel) Init() tea.Cmd { return nil }

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.nameEdit {
			return m.handleNameEdit(msg)
		}
		switch msg.String() {
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = fieldCount - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= fieldCount {
				m.cursor = 0
			}
		case "left", "h":
			m.adjustField(-1)
		case "right", "l":
			m.adjustField(1)
		case "enter", " ":
			if m.cursor == fieldName {
				m.nameEdit = true
				return m, nil
			}
			if m.cursor == fieldStart {
				return m, m.startGame()
			}
			m.adjustField(1)
		case "escape", "esc":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *SetupModel) handleNameEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "escape", "esc":
		m.nameEdit = false
		if m.name == "" {
			m.name = "Player"
		}
	case "backspace":
		runes := []rune(m.name)
		if len(runes) > 0 {
			m.name = string(runes[:len(runes)-1])
		}
	default:
		ch := msg.String()
		if len([]rune(ch)) == 1 && len([]rune(m.name)) < 16 {
			m.name += ch
		}
	}
	return *m, nil
}

func (m *SetupModel) adjustField(dir int) {
	switch m.cursor {
	case fieldMode:
		m.mode = (m.mode + dir + len(modeNames)) % len(modeNames)
		if m.mode == 2 { // headsup forces 2 seats
			m.seats = 2
		} else if m.seats == 2 {
			m.seats = 6
		}
	case fieldSeats:
		if m.mode == 2 {
			return // locked to 2 for headsup
		}
		idx := 0
		for i, s := range seatOptions {
			if s == m.seats {
				idx = i
			}
		}
		idx = (idx + dir + len(seatOptions)) % len(seatOptions)
		m.seats = seatOptions[idx]
		if m.seats == 2 && m.mode != 2 {
			// Skip 2-seat for non-headsup
			idx = (idx + dir + len(seatOptions)) % len(seatOptions)
			m.seats = seatOptions[idx]
		}
	case fieldDifficulty:
		m.difficulty = (m.difficulty + dir + len(diffNames)) % len(diffNames)
	case fieldStack:
		idx := 0
		for i, s := range stackOptions {
			if s == m.stack {
				idx = i
			}
		}
		idx = (idx + dir + len(stackOptions)) % len(stackOptions)
		m.stack = stackOptions[idx]
	}
}

func (m SetupModel) startGame() tea.Cmd {
	return func() tea.Msg {
		seats := m.seats
		if m.mode == 2 {
			seats = 2
		}
		cfg := session.Config{
			Mode:          modeValues[m.mode],
			Difficulty:    diffValues[m.difficulty],
			Seats:         seats,
			StartingStack: m.stack,
			PlayerName:    m.name,
		}
		return startGameMsg{config: cfg}
	}
}

func (m SetupModel) View() string {
	title := StyleTitle.Render("NEW GAME SETUP")

	fields := []struct {
		label string
		value string
	}{
		{"Game Mode", modeNames[m.mode]},
		{"Seats", m.seatsDisplay()},
		{"Difficulty", diffNames[m.difficulty]},
		{"Starting Stack", strconv.Itoa(m.stack) + " BB"},
		{"Player Name", m.nameDisplay()},
		{"", "[ START GAME ]"},
	}

	var rows string
	for i, f := range fields {
		cursor := "  "
		style := StyleMenuItem
		if setupField(i) == m.cursor {
			cursor = "> "
			style = StyleMenuItemActive
		}

		label := ""
		if f.label != "" {
			label = StyleDim.Render(fmt.Sprintf("%-16s", f.label))
		}

		value := style.Render(f.value)
		if setupField(i) == m.cursor && setupField(i) != fieldName && setupField(i) != fieldStart {
			value = StyleKey.Render("< ") + value + StyleKey.Render(" >")
		}

		rows += fmt.Sprintf("%s%s %s\n", cursor, label, value)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(rows)

	help := StyleDim.Render("Arrow keys to navigate  |  Left/Right to change  |  Enter to select  |  Esc to go back")

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", box, "", help)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m SetupModel) seatsDisplay() string {
	if m.mode == 2 {
		return "2 (Heads-Up)"
	}
	return strconv.Itoa(m.seats) + "-Max"
}

func (m SetupModel) nameDisplay() string {
	if m.nameEdit {
		return m.name + "_"
	}
	return m.name
}
