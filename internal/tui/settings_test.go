package tui

import (
	"fmt"
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

func TestSettingsVolumePreviewAndVisibility(t *testing.T) {
	prevEnabled := setAudioEnabled
	prevVolume := setAudioVolume
	prevPlay := playSettingsSound
	defer func() {
		setAudioEnabled = prevEnabled
		setAudioVolume = prevVolume
		playSettingsSound = prevPlay
	}()

	enabledCalls := []bool{}
	volumeCalls := []int{}
	playCount := 0
	setAudioEnabled = func(on bool) { enabledCalls = append(enabledCalls, on) }
	setAudioVolume = func(v int) { volumeCalls = append(volumeCalls, v) }
	playSettingsSound = func() { playCount++ }

	cfg := storage.DefaultConfig()
	m := NewSettingsModel(cfg)
	m.cursor = sfVolume
	m.adjustField(1)

	if m.config.SoundVolume != 80 {
		t.Fatalf("sound volume = %d, want 80", m.config.SoundVolume)
	}
	if len(enabledCalls) != 1 || !enabledCalls[0] {
		t.Fatalf("enabled calls = %v, want [true]", enabledCalls)
	}
	if len(volumeCalls) != 1 || volumeCalls[0] != 80 {
		t.Fatalf("volume calls = %v, want [80]", volumeCalls)
	}
	if playCount != 1 {
		t.Fatalf("play count = %d, want 1", playCount)
	}
	if fields := m.visibleFields(); len(fields) < 3 || fields[2] != sfVolume {
		t.Fatalf("expected volume field to be visible, got %v", fields)
	}
}

func TestSettingsEscapeRestoresAudioPreview(t *testing.T) {
	prevEnabled := setAudioEnabled
	prevVolume := setAudioVolume
	prevPlay := playSettingsSound
	defer func() {
		setAudioEnabled = prevEnabled
		setAudioVolume = prevVolume
		playSettingsSound = prevPlay
	}()

	volumeCalls := []int{}
	enabledCalls := []bool{}
	setAudioEnabled = func(on bool) { enabledCalls = append(enabledCalls, on) }
	setAudioVolume = func(v int) { volumeCalls = append(volumeCalls, v) }
	playSettingsSound = func() {}

	cfg := storage.DefaultConfig()
	cfg.SoundVolume = 70
	m := NewSettingsModel(cfg)
	m.cursor = sfVolume
	m.adjustField(-2)

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if _, ok := model.(SettingsModel); !ok {
		t.Fatalf("expected SettingsModel, got %T", model)
	}
	if cmd == nil {
		t.Fatal("expected escape command")
	}
	msg := cmd()
	if _, ok := msg.(switchScreenMsg); !ok {
		t.Fatalf("expected switchScreenMsg, got %T", msg)
	}
	if enabledCalls[len(enabledCalls)-1] != cfg.SoundEnabled {
		t.Fatalf("last enabled call = %v, want %v", enabledCalls[len(enabledCalls)-1], cfg.SoundEnabled)
	}
	if volumeCalls[len(volumeCalls)-1] != cfg.SoundVolume {
		t.Fatalf("last volume call = %d, want %d", volumeCalls[len(volumeCalls)-1], cfg.SoundVolume)
	}
}

func TestSettingsSaveErrorReturnsAppError(t *testing.T) {
	prevSave := saveSettingsConfig
	defer func() { saveSettingsConfig = prevSave }()
	saveSettingsConfig = func(storage.Config) error { return fmt.Errorf("disk full") }

	m := NewSettingsModel(storage.DefaultConfig())
	m.cursor = sfSave
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := model.(SettingsModel); !ok {
		t.Fatalf("expected SettingsModel, got %T", model)
	}
	if cmd == nil {
		t.Fatal("expected save error command")
	}
	msg := cmd()
	errMsg, ok := msg.(appErrorMsg)
	if !ok {
		t.Fatalf("expected appErrorMsg, got %T", msg)
	}
	if errMsg.message == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestSettingsSaveReturnsSettingsSavedMsg(t *testing.T) {
	prevSave := saveSettingsConfig
	defer func() { saveSettingsConfig = prevSave }()
	saveSettingsConfig = func(storage.Config) error { return nil }

	m := NewSettingsModel(storage.DefaultConfig())
	model, cmd := m.saveConfig()
	next, ok := model.(SettingsModel)
	if !ok {
		t.Fatalf("expected SettingsModel, got %T", model)
	}
	if !next.saved {
		t.Fatal("expected saved flag")
	}
	msg := cmd()
	if _, ok := msg.(settingsSavedMsg); !ok {
		t.Fatalf("expected settingsSavedMsg, got %T", msg)
	}
}
