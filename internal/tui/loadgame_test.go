package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"
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

func TestLoadGameEnterLoadsResumedSession(t *testing.T) {
	prev := resumeSavedSession
	defer func() { resumeSavedSession = prev }()
	resumed, err := session.New(session.Config{Mode: 0, Seats: 2, StartingStack: 10, PlayerName: "Hero", Seed: 1})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}
	resumeSavedSession = func(slot int) (*session.Session, error) {
		if slot != 2 {
			return nil, fmt.Errorf("unexpected slot %d", slot)
		}
		return resumed, nil
	}
	m := LoadGameModel{saves: []storage.SaveInfo{{Slot: 1, Empty: true}, {Slot: 2, Name: "Resume Me"}}, cursor: 1}

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := model.(LoadGameModel)
	if next.errorMsg != "" {
		t.Fatalf("unexpected error message: %q", next.errorMsg)
	}
	msg := cmd()
	switchMsg, ok := msg.(switchScreenMsg)
	if !ok {
		t.Fatalf("expected switchScreenMsg, got %T", msg)
	}
	if switchMsg.screen != ScreenGame {
		t.Fatalf("screen = %v, want ScreenGame", switchMsg.screen)
	}
	if switchMsg.data != resumed {
		t.Fatal("expected resumed session to be forwarded in switchScreenMsg")
	}
}

func TestLoadGameEnterShowsLoadErrorOnScreen(t *testing.T) {
	prev := resumeSavedSession
	defer func() { resumeSavedSession = prev }()
	resumeSavedSession = func(int) (*session.Session, error) {
		return nil, fmt.Errorf("bad save artifact")
	}
	m := LoadGameModel{saves: []storage.SaveInfo{{Slot: 1, Name: "Broken"}}, cursor: 0}

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := model.(LoadGameModel)
	if next.errorMsg != "bad save artifact" {
		t.Fatalf("errorMsg = %q, want bad save artifact", next.errorMsg)
	}
	if cmd != nil {
		t.Fatal("expected no navigation command on load error")
	}
}
