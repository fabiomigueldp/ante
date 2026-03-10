package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestSetupModelStartGameUsesStackValue(t *testing.T) {
	m := NewSetupModel(storage.DefaultConfig())
	cmd := m.startGame()
	msg := cmd()
	start, ok := msg.(startGameMsg)
	if !ok {
		t.Fatalf("expected startGameMsg, got %T", msg)
	}
	if start.config.StartingStack != storage.DefaultConfig().StartingStack {
		t.Fatalf("expected starting stack %d, got %d", storage.DefaultConfig().StartingStack, start.config.StartingStack)
	}
}

func TestSetupNameEditBackspaceDoesNotPanic(t *testing.T) {
	m := NewSetupModel(storage.DefaultConfig())
	m.cursor = fieldName
	m.nameEdit = true

	for range 20 {
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		next, ok := model.(SetupModel)
		if !ok {
			t.Fatalf("expected SetupModel, got %T", model)
		}
		m = next
	}

	if m.name != "" {
		t.Fatalf("expected empty name after deletes, got %q", m.name)
	}

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	next, ok := model.(SetupModel)
	if !ok {
		t.Fatalf("expected SetupModel after typing, got %T", model)
	}
	if next.name != "a" {
		t.Fatalf("expected typed name to be preserved, got %q", next.name)
	}
}

func TestSetupStackAlwaysUsesAllowedValues(t *testing.T) {
	m := NewSetupModel(storage.DefaultConfig())
	m.cursor = fieldStack
	allowed := map[int]bool{50: true, 100: true, 200: true}

	for range 20 {
		m.adjustField(1)
		if !allowed[m.stack] {
			t.Fatalf("unexpected stack value %d", m.stack)
		}
	}
}
