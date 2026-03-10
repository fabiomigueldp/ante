package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadGameEmptyListNavigationDoesNotPanic(t *testing.T) {
	m := LoadGameModel{saves: nil}
	keys := []tea.KeyMsg{
		{Type: tea.KeyUp},
		{Type: tea.KeyDown},
		{Type: tea.KeyEnter},
		{Type: tea.KeyRunes, Runes: []rune{'d'}},
	}

	for _, key := range keys {
		model, _ := m.Update(key)
		next, ok := model.(LoadGameModel)
		if !ok {
			t.Fatalf("expected LoadGameModel, got %T", model)
		}
		m = next
	}
}
