package tui

import "github.com/charmbracelet/lipgloss"

// Seat geometry constants.
const (
	SeatTotalWidth   = 22 // outer width including padding
	SeatContentWidth = 20 // SeatTotalWidth minus Padding(0,1)
	SeatHeight       = 5  // fixed line count per seat block
	SeatGap          = 3  // horizontal space between seats
)

// Color palette — premium poker room: dark base, refined green accents.
var (
	ColorGreen      = lipgloss.Color("#1a5c2a")
	ColorDarkGreen  = lipgloss.Color("#0f3318")
	ColorFelt       = lipgloss.Color("#14472a")
	ColorGold       = lipgloss.Color("#c9a84c")
	ColorRed        = lipgloss.Color("#9b2d30")
	ColorWhite      = lipgloss.Color("#e8e4dc")
	ColorGray       = lipgloss.Color("#8a8a82")
	ColorDarkGray   = lipgloss.Color("#3a3a36")
	ColorBlack      = lipgloss.Color("#0e0e0c")
	ColorCyan       = lipgloss.Color("#5ba4a4")
	ColorYellow     = lipgloss.Color("#b8a44c")
	ColorBrightRed  = lipgloss.Color("#c44040")
	ColorDim        = lipgloss.Color("#5a6a5a")
	ColorHighlight  = lipgloss.Color("#dfc867")
	ColorCardWhite  = lipgloss.Color("#f5f3ee")
	ColorCardBack   = lipgloss.Color("#1b3a6b")
	ColorSpade      = lipgloss.Color("#d4d0c8")
	ColorHeart      = lipgloss.Color("#c44040")
	ColorDiamond    = lipgloss.Color("#c47040")
	ColorClub       = lipgloss.Color("#6ea87a")
	ColorActiveSeat = lipgloss.Color("#253d2a")
	ColorFolded     = lipgloss.Color("#2a2a26")
	ColorDealer     = lipgloss.Color("#dfc867")
	ColorAllIn      = lipgloss.Color("#cc6633")
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
			Border(lipgloss.NormalBorder()).
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
		Width(SeatTotalWidth).
		Height(SeatHeight).
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
