package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestSettingsNameEditBackspaceDoesNotPanic(t *testing.T) {
	m := NewSettingsModel(storage.DefaultConfig())
	m.cursor = sfPlayerName
	m.nameEdit = true

	for range 20 {
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		next, ok := model.(SettingsModel)
		if !ok {
			t.Fatalf("expected SettingsModel, got %T", model)
		}
		m = next
	}

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	next, ok := model.(SettingsModel)
	if !ok {
		t.Fatalf("expected SettingsModel after typing, got %T", model)
	}
	if next.config.PlayerName != "b" {
		t.Fatalf("expected typed player name, got %q", next.config.PlayerName)
	}
}
