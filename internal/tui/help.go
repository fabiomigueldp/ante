package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel shows game controls and rules.
type HelpModel struct {
	width  int
	height int
	scroll int
}

func NewHelpModel() HelpModel {
	return HelpModel{}
}

func (m HelpModel) Init() tea.Cmd { return nil }

func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "escape", "q":
			return m, func() tea.Msg { return switchScreenMsg{screen: ScreenMenu} }
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			m.scroll++
		}
	}
	return m, nil
}

func (m HelpModel) View() string {
	title := StyleTitle.Render("HELP & CONTROLS")

	sections := []string{
		title,
		"",
		StyleSubtitle.Render("GAME CONTROLS"),
		"",
		"  " + StyleKey.Render("[Q]") + "     Fold",
		"  " + StyleKey.Render("[W]") + "     Check",
		"  " + StyleKey.Render("[E]") + "     Call",
		"  " + StyleKey.Render("[T]") + "     Raise / Bet",
		"  " + StyleKey.Render("[A]") + "     All-In",
		"  " + StyleKey.Render("[0-9]") + "   Enter bet amount",
		"  " + StyleKey.Render("[Enter]") + " Confirm bet",
		"  " + StyleKey.Render("[Esc]") + "   Pause / Back",
		"",
		StyleSubtitle.Render("NAVIGATION"),
		"",
		"  " + StyleKey.Render("[Up/Down]") + "    Navigate menus",
		"  " + StyleKey.Render("[Left/Right]") + " Change values",
		"  " + StyleKey.Render("[Enter]") + "      Select / Confirm",
		"  " + StyleKey.Render("[Esc]") + "        Go back",
		"",
		StyleSubtitle.Render("PAUSE MENU"),
		"",
		"  " + StyleKey.Render("[Esc]") + " Resume game",
		"  " + StyleKey.Render("[S]") + "   Save game",
		"  " + StyleKey.Render("[H]") + "   Show help",
		"  " + StyleKey.Render("[Q]") + "   Quit to menu",
		"",
		StyleSubtitle.Render("POT ODDS"),
		"",
		"  When facing a bet, pot odds are shown as:",
		"  " + StyleInfo.Render("Pot: $X | Call: $Y | Odds: Z:1 (N%)"),
		"",
		StyleSubtitle.Render("GAME MODES"),
		"",
		"  " + StyleBold.Render("Tournament (Sit & Go)"),
		"  Play until one player has all the chips.",
		"  Blinds increase over time. No rebuys.",
		"",
		"  " + StyleBold.Render("Cash Game"),
		"  Play as long as you want. Fixed blinds.",
		"  Leave anytime with your current stack.",
		"",
		"  " + StyleBold.Render("Heads-Up Duel"),
		"  One-on-one tournament format. Fast blinds.",
		"",
		StyleSubtitle.Render("HAND RANKINGS (Best to Worst)"),
		"",
		"  Royal Flush     A K Q J T of same suit",
		"  Straight Flush  Five sequential same suit",
		"  Four of a Kind  Four cards same rank",
		"  Full House      Three + Two of a kind",
		"  Flush           Five cards same suit",
		"  Straight        Five sequential cards",
		"  Three of a Kind Three cards same rank",
		"  Two Pair        Two different pairs",
		"  One Pair        Two cards same rank",
		"  High Card       Highest card plays",
		"",
		StyleDim.Render("[Esc] Back  |  [Up/Down] Scroll"),
	}

	// Apply scroll
	maxVis := m.height - 4
	if maxVis < 10 {
		maxVis = 10
	}
	if m.scroll > len(sections)-maxVis {
		m.scroll = len(sections) - maxVis
	}
	if m.scroll < 0 {
		m.scroll = 0
	}

	end := m.scroll + maxVis
	if end > len(sections) {
		end = len(sections)
	}
	visible := sections[m.scroll:end]

	content := lipgloss.JoinVertical(lipgloss.Left, visible...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDarkGray).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
