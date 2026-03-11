package tui

import (
	"fmt"
	"strings"

	"github.com/fabiomigueldp/ante/internal/engine"

	"github.com/charmbracelet/lipgloss"
)

// Suit symbols (Unicode).
var suitSymbol = map[engine.Suit]string{
	engine.Spades:   "\u2660", // ♠
	engine.Hearts:   "\u2665", // ♥
	engine.Diamonds: "\u2666", // ♦
	engine.Clubs:    "\u2663", // ♣
}

var suitColor = map[engine.Suit]lipgloss.Color{
	engine.Spades:   ColorSpade,
	engine.Hearts:   ColorHeart,
	engine.Diamonds: ColorDiamond,
	engine.Clubs:    ColorClub,
}

// RankStr returns the display string for a rank.
func RankStr(r engine.Rank) string {
	switch r {
	case engine.Ten:
		return "T"
	case engine.Jack:
		return "J"
	case engine.Queen:
		return "Q"
	case engine.King:
		return "K"
	case engine.Ace:
		return "A"
	default:
		return fmt.Sprintf("%d", r)
	}
}

// CardStr renders a single card as a compact inline string like [A♠].
func CardStr(c engine.Card) string {
	if !isRenderableCard(c) {
		return "  "
	}
	sym := suitSymbol[c.Suit]
	rank := RankStr(c.Rank)
	style := lipgloss.NewStyle().Foreground(suitColor[c.Suit])
	return style.Render(rank + sym)
}

// CardBack renders a face-down card.
func CardBack() string {
	return lipgloss.NewStyle().Foreground(ColorCardBack).Render("░░")
}

// RenderHoleCards renders two hole cards inline.
func RenderHoleCards(cards [2]engine.Card, visible bool) string {
	if !visible {
		return fmt.Sprintf("[%s][%s]", CardBack(), CardBack())
	}
	left := CardStr(cards[0])
	right := CardStr(cards[1])
	if !isRenderableCard(cards[0]) {
		left = CardBack()
	}
	if !isRenderableCard(cards[1]) {
		right = CardBack()
	}
	return fmt.Sprintf("[%s][%s]", left, right)
}

// RenderBigCard renders a card in a larger 3-line format for the human player.
func RenderBigCard(c engine.Card) string {
	if !isRenderableCard(c) {
		return renderEmptyBigCard()
	}
	rank := RankStr(c.Rank)
	sym := suitSymbol[c.Suit]
	style := lipgloss.NewStyle().Foreground(suitColor[c.Suit])

	rr := rank
	if len(rr) == 1 {
		rr = rr + " "
	}
	top := "\u250c\u2500\u2500\u2510" // ┌──┐
	mid := "\u2502" + style.Render(rr) + "\u2502"
	sym_line := "\u2502" + style.Render(sym+" ") + "\u2502"
	bot := "\u2514\u2500\u2500\u2518" // └──┘
	return top + "\n" + mid + "\n" + sym_line + "\n" + bot
}

func isRenderableCard(c engine.Card) bool {
	return c.Rank >= engine.Two && c.Rank <= engine.Ace && c.Suit <= engine.Clubs
}

// RenderBigCards renders hole cards side by side in large format.
func RenderBigCards(cards [2]engine.Card) string {
	c1 := strings.Split(RenderBigCard(cards[0]), "\n")
	c2 := strings.Split(RenderBigCard(cards[1]), "\n")
	var lines []string
	for i := range c1 {
		lines = append(lines, c1[i]+" "+c2[i])
	}
	return strings.Join(lines, "\n")
}

// RenderEmptyCard renders a placeholder for an unrevealed community card.
func RenderEmptyCard() string {
	style := lipgloss.NewStyle().Foreground(ColorDarkGray)
	return style.Render("[  ]")
}

// RenderBoard renders the community cards in a single line.
func RenderBoard(board []engine.Card) string {
	cards := make([]string, 5)
	for i := range 5 {
		if i < len(board) {
			cards[i] = "[" + CardStr(board[i]) + "]"
		} else {
			cards[i] = RenderEmptyCard()
		}
	}
	// Add spacing between flop/turn/river
	if len(cards) >= 5 {
		return cards[0] + " " + cards[1] + " " + cards[2] + "  " + cards[3] + "  " + cards[4]
	}
	return strings.Join(cards, " ")
}

// RenderBoardLarge renders community cards in larger format.
func RenderBoardLarge(board []engine.Card) string {
	var parts []string
	for i := range 5 {
		if i < len(board) {
			parts = append(parts, RenderBigCard(board[i]))
		} else {
			parts = append(parts, renderEmptyBigCard())
		}
	}
	// Join side by side
	if len(parts) == 0 {
		return ""
	}
	lines := make([][]string, len(parts))
	maxLines := 0
	for i, p := range parts {
		lines[i] = strings.Split(p, "\n")
		if len(lines[i]) > maxLines {
			maxLines = len(lines[i])
		}
	}
	var result []string
	for row := range maxLines {
		var rowParts []string
		for i := range parts {
			if row < len(lines[i]) {
				rowParts = append(rowParts, lines[i][row])
			} else {
				rowParts = append(rowParts, "    ")
			}
		}
		sep := " "
		// Extra space between flop and turn, turn and river
		line := rowParts[0] + sep + rowParts[1] + sep + rowParts[2]
		if len(rowParts) > 3 {
			line += "  " + rowParts[3]
		}
		if len(rowParts) > 4 {
			line += "  " + rowParts[4]
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func renderEmptyBigCard() string {
	style := lipgloss.NewStyle().Foreground(ColorDarkGray)
	top := "\u250c\u2500\u2500\u2510"
	mid := "\u2502  \u2502"
	bot := "\u2514\u2500\u2500\u2518"
	return style.Render(top) + "\n" + style.Render(mid) + "\n" + style.Render(mid) + "\n" + style.Render(bot)
}

// ActionStr returns a human-readable string for an action type.
func ActionStr(a engine.ActionType) string {
	switch a {
	case engine.ActionFold:
		return "Fold"
	case engine.ActionCheck:
		return "Check"
	case engine.ActionCall:
		return "Call"
	case engine.ActionBet:
		return "Bet"
	case engine.ActionRaise:
		return "Raise"
	case engine.ActionAllIn:
		return "All-In"
	default:
		return "?"
	}
}

// StatusStr returns a display string for player status.
func StatusStr(s engine.PlayerStatus) string {
	switch s {
	case engine.StatusActive:
		return ""
	case engine.StatusFolded:
		return "FOLDED"
	case engine.StatusAllIn:
		return "ALL-IN"
	case engine.StatusOut:
		return "OUT"
	case engine.StatusSittingOut:
		return "SITTING OUT"
	default:
		return ""
	}
}

// StreetStr returns a display string for the current street.
func StreetStr(s engine.Street) string {
	switch s {
	case engine.StreetPreflop:
		return "Preflop"
	case engine.StreetFlop:
		return "Flop"
	case engine.StreetTurn:
		return "Turn"
	case engine.StreetRiver:
		return "River"
	default:
		return ""
	}
}

// ChipStr formats a chip amount with comma separators.
func ChipStr(amount int) string {
	if amount < 0 {
		return fmt.Sprintf("-%s", ChipStr(-amount))
	}
	s := fmt.Sprintf("%d", amount)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

// PositionStr returns a short position label.
func PositionStr(p engine.Position) string {
	switch p {
	case engine.PositionSmallBlind:
		return "SB"
	case engine.PositionBigBlind:
		return "BB"
	case engine.PositionDealer:
		return "BTN"
	case engine.PositionEarly:
		return "EP"
	case engine.PositionMiddle:
		return "MP"
	case engine.PositionLate:
		return "LP"
	default:
		return ""
	}
}
