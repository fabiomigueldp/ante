package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorGreen      = lipgloss.Color("#2d5016")
	ColorDarkGreen  = lipgloss.Color("#1a3009")
	ColorFelt       = lipgloss.Color("#1e4d1e")
	ColorGold       = lipgloss.Color("#d4a017")
	ColorRed        = lipgloss.Color("#cc3333")
	ColorWhite      = lipgloss.Color("#f0f0f0")
	ColorGray       = lipgloss.Color("#888888")
	ColorDarkGray   = lipgloss.Color("#444444")
	ColorBlack      = lipgloss.Color("#111111")
	ColorCyan       = lipgloss.Color("#00cccc")
	ColorYellow     = lipgloss.Color("#cccc00")
	ColorBrightRed  = lipgloss.Color("#ff4444")
	ColorDim        = lipgloss.Color("#666666")
	ColorHighlight  = lipgloss.Color("#ffcc00")
	ColorCardWhite  = lipgloss.Color("#ffffff")
	ColorCardBack   = lipgloss.Color("#2244aa")
	ColorSpade      = lipgloss.Color("#e0e0e0")
	ColorHeart      = lipgloss.Color("#ff3333")
	ColorDiamond    = lipgloss.Color("#ff6633")
	ColorClub       = lipgloss.Color("#99cc99")
	ColorActiveSeat = lipgloss.Color("#335533")
	ColorFolded     = lipgloss.Color("#333333")
	ColorDealer     = lipgloss.Color("#ffdd44")
	ColorAllIn      = lipgloss.Color("#ff6600")
)

// Styles — commonly used styled renderers.
var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold).
			Align(lipgloss.Center)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Align(lipgloss.Center)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkGray)

	StyleMenuTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold).
			MarginBottom(1)

	StyleMenuItem = lipgloss.NewStyle().
			Foreground(ColorWhite).
			PaddingLeft(2)

	StyleMenuItemActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorHighlight).
				PaddingLeft(2)

	StyleMenuItemDim = lipgloss.NewStyle().
				Foreground(ColorDim).
				PaddingLeft(2)

	StyleActionBar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorDarkGray).
			Padding(0, 1)

	StyleKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	StyleKeyLabel = lipgloss.NewStyle().
			Foreground(ColorGray)

	StylePot = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold)

	StyleChips = lipgloss.NewStyle().
			Foreground(ColorGold)

	StylePlayerName = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	StylePlayerNameDim = lipgloss.NewStyle().
				Foreground(ColorDim)

	StyleHumanLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	StyleDealer = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDealer)

	StyleFolded = lipgloss.NewStyle().
			Foreground(ColorDim).
			Strikethrough(true)

	StyleAllIn = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAllIn)

	StyleStatus = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorBrightRed)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorCyan)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleBold = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorDarkGray)

	StyleFooter = lipgloss.NewStyle().
			Foreground(ColorGray).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorDarkGray)

	StyleHandRank = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	StyleWinner = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold)

	StyleLoser = lipgloss.NewStyle().
			Foreground(ColorDim)

	StyleSplash = lipgloss.NewStyle().
			Foreground(ColorGold).
			Bold(true)
)

// StyleBet is used for opponent bet display.
var StyleBet = lipgloss.NewStyle().
	Foreground(ColorCyan)

// SeatStyle returns the style for a player seat panel.
func SeatStyle(active, isHuman, isFolded, isAllIn, isDealer bool) lipgloss.Style {
	base := lipgloss.NewStyle().
		Width(22).
		Padding(0, 1)
	if isFolded {
		return base.Foreground(ColorDim)
	}
	if isAllIn {
		return base.Foreground(ColorAllIn).Bold(true)
	}
	if isHuman && active {
		return base.Foreground(ColorHighlight).Bold(true)
	}
	if active {
		return base.Foreground(ColorWhite)
	}
	return base.Foreground(ColorGray)
}

// CenterH returns a style that centers content horizontally in the given width.
func CenterH(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
}

// Pad returns a string padded to the given width.
func Pad(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Render(s)
}
