package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestAppStartGameTransitionsToGame(t *testing.T) {
	a := NewApp()
	a.screen = ScreenSetup
	a.width = 120
	a.height = 40

	msg := startGameMsg{config: NewSetupModel(storage.DefaultConfig()).startGame()().(startGameMsg).config}
	model, _ := a.Update(msg)
	next, ok := model.(App)
	if !ok {
		t.Fatalf("expected App, got %T", model)
	}
	if next.screen != ScreenGame {
		t.Fatalf("expected ScreenGame, got %v", next.screen)
	}
	if next.lastSess == nil {
		t.Fatal("expected session to be created")
	}
}

func TestAppHandlesWindowResizeWithoutPanic(t *testing.T) {
	a := NewApp()
	model, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if _, ok := model.(App); !ok {
		t.Fatalf("expected App, got %T", model)
	}
}
