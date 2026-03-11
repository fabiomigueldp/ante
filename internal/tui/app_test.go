package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/session"
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

func TestAppSettingsSavedUpdatesRootConfig(t *testing.T) {
	a := NewApp()
	updated := storage.DefaultConfig()
	updated.SoundVolume = 40

	model, _ := a.Update(settingsSavedMsg{config: updated})
	next, ok := model.(App)
	if !ok {
		t.Fatalf("expected App, got %T", model)
	}
	if next.config.SoundVolume != 40 {
		t.Fatalf("sound volume = %d, want 40", next.config.SoundVolume)
	}
}

func TestAppHandlesWindowResizeWithoutPanic(t *testing.T) {
	a := NewApp()
	model, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if _, ok := model.(App); !ok {
		t.Fatalf("expected App, got %T", model)
	}
}

func TestAppSwitchToGameWithResumedSession(t *testing.T) {
	a := NewApp()
	a.width = 120
	a.height = 40
	resumed, err := session.New(session.Config{Mode: 0, Seats: 2, StartingStack: 10, PlayerName: "Hero", Seed: 1})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}
	model, cmd := a.switchTo(switchScreenMsg{screen: ScreenGame, data: resumed})
	next := model.(App)
	if next.lastSess != resumed {
		t.Fatal("expected resumed session to become lastSess")
	}
	if next.screen != ScreenGame {
		t.Fatalf("screen = %v, want ScreenGame", next.screen)
	}
	if cmd == nil {
		t.Fatal("expected game init command")
	}
}

func TestAppCopiesFallbackResultFromFinishedGameModel(t *testing.T) {
	a := NewApp()
	a.screen = ScreenGame
	a.game = GameModel{}
	cmd := a.updateGame(sessionDoneMsg{})
	if cmd != nil {
		t.Fatal("expected no command for sessionDone fallback")
	}
	if a.game.vm.Result != "Session ended." {
		t.Fatalf("game vm result = %q, want %q", a.game.vm.Result, "Session ended.")
	}
	if a.lastResult != "Session ended." {
		t.Fatalf("lastResult = %q, want %q", a.lastResult, "Session ended.")
	}
}
