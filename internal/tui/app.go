package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/session"
	"github.com/fabiomigueldp/ante/internal/storage"
)

// Screen identifies which screen the app is currently showing.
type Screen uint8

const (
	ScreenSplash Screen = iota
	ScreenMenu
	ScreenSetup
	ScreenLoadGame
	ScreenGame
	ScreenResults
	ScreenStats
	ScreenHistory
	ScreenReplay
	ScreenSettings
	ScreenHelp
)

// Messages used to switch screens and start games.
type (
	switchScreenMsg struct {
		screen Screen
		data   interface{}
	}
	startGameMsg struct {
		config session.Config
	}
	appErrorMsg struct {
		message string
	}
	settingsSavedMsg struct {
		config storage.Config
	}
)

// App is the root Bubble Tea model that routes between screens.
type App struct {
	screen Screen
	width  int
	height int

	splash   SplashModel
	menu     MenuModel
	setup    SetupModel
	game     GameModel
	results  ResultsModel
	stats    StatsViewModel
	history  HistoryViewModel
	replay   ReplayModel
	settings SettingsModel
	help     HelpModel
	loadGame LoadGameModel

	config     storage.Config
	lastSess   *session.Session
	lastResult string
	errorMsg   string
}

// NewApp creates the root application model.
func NewApp() App {
	cfg := storage.LoadConfig()
	audio.SetEnabled(cfg.SoundEnabled)
	audio.SetVolume(float64(cfg.SoundVolume) / 100.0)
	return App{
		screen:   ScreenSplash,
		splash:   NewSplashModel(),
		menu:     NewMenuModel(),
		setup:    NewSetupModel(cfg),
		config:   cfg,
		settings: NewSettingsModel(cfg),
		help:     NewHelpModel(),
		stats:    NewStatsViewModel(),
		history:  NewHistoryViewModel(),
		loadGame: NewLoadGameModel(),
	}
}

func (a App) Init() tea.Cmd {
	return a.splash.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case switchScreenMsg:
		return a.switchTo(msg)
	case startGameMsg:
		return a.startGame(msg)
	case appErrorMsg:
		a.errorMsg = msg.message
		if a.screen == ScreenGame {
			a.lastResult = msg.message
			return a.switchTo(switchScreenMsg{screen: ScreenResults})
		}
	case settingsSavedMsg:
		a.config = msg.config
		a.clearError()
	}

	switch a.screen {
	case ScreenSplash:
		return a, a.updateSplash(msg)
	case ScreenMenu:
		return a, a.updateMenu(msg)
	case ScreenSetup:
		return a, a.updateSetup(msg)
	case ScreenLoadGame:
		return a, a.updateLoadGame(msg)
	case ScreenGame:
		return a, a.updateGame(msg)
	case ScreenResults:
		return a, a.updateResults(msg)
	case ScreenStats:
		return a, a.updateStats(msg)
	case ScreenHistory:
		return a, a.updateHistory(msg)
	case ScreenReplay:
		return a, a.updateReplay(msg)
	case ScreenSettings:
		return a, a.updateSettings(msg)
	case ScreenHelp:
		return a, a.updateHelp(msg)
	default:
		return a, nil
	}
}

func (a App) View() string {
	base := ""
	switch a.screen {
	case ScreenSplash:
		base = a.splash.View()
	case ScreenMenu:
		base = a.menu.View()
	case ScreenSetup:
		base = a.setup.View()
	case ScreenLoadGame:
		base = a.loadGame.View()
	case ScreenGame:
		base = a.game.View()
	case ScreenResults:
		base = a.results.View()
	case ScreenStats:
		base = a.stats.View()
	case ScreenHistory:
		base = a.history.View()
	case ScreenReplay:
		base = a.replay.View()
	case ScreenSettings:
		base = a.settings.View()
	case ScreenHelp:
		base = a.help.View()
	}

	if a.errorMsg == "" {
		return base
	}

	errorLine := lipgloss.NewStyle().
		Foreground(ColorBrightRed).
		Bold(true).
		Padding(0, 1).
		Render("ERROR: " + a.errorMsg)

	if a.width > 0 {
		errorLine = lipgloss.NewStyle().Width(a.width).Render(errorLine)
	}
	return lipgloss.JoinVertical(lipgloss.Left, errorLine, base)
}

func (a *App) clearError() {
	a.errorMsg = ""
}

func (a App) switchTo(msg switchScreenMsg) (tea.Model, tea.Cmd) {
	if a.screen == ScreenGame && msg.screen != ScreenGame && a.lastSess != nil {
		a.lastSess.Stop()
	}

	a.screen = msg.screen
	a.clearError()

	var cmd tea.Cmd

	switch msg.screen {
	case ScreenSplash:
		a.splash = NewSplashModel()
		cmd = a.splash.Init()
	case ScreenMenu:
		a.menu = NewMenuModel()
	case ScreenSetup:
		a.config = storage.LoadConfig()
		a.setup = NewSetupModel(a.config)
	case ScreenLoadGame:
		a.loadGame = NewLoadGameModel()
	case ScreenResults:
		a.results = NewResultsModel(a.lastResult, a.lastSess)
	case ScreenStats:
		a.stats = NewStatsViewModel()
	case ScreenHistory:
		a.history = NewHistoryViewModel()
	case ScreenReplay:
		if record, ok := msg.data.(*session.Session); ok && record != nil && record.History != nil {
			_ = record
		}
	case ScreenSettings:
		a.config = storage.LoadConfig()
		a.settings = NewSettingsModel(a.config)
	case ScreenHelp:
		a.help = NewHelpModel()
	}

	if a.width > 0 && a.height > 0 {
		a.applyWindowSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	}

	return a, cmd
}

func (a App) startGame(msg startGameMsg) (tea.Model, tea.Cmd) {
	sess, err := session.New(msg.config)
	if err != nil {
		a.errorMsg = fmt.Sprintf("Unable to start game: %v", err)
		a.screen = ScreenSetup
		return a, nil
	}

	a.lastSess = sess
	a.lastResult = ""
	a.screen = ScreenGame
	a.game = NewGameModel(sess, a.config.ShowPotOdds)
	a.clearError()

	var cmd tea.Cmd
	if a.width > 0 && a.height > 0 {
		a.applyWindowSize(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	}

	return a, tea.Batch(cmd, a.game.Init())
}

func (a *App) updateSplash(msg tea.Msg) tea.Cmd {
	m, cmd := a.splash.Update(msg)
	if next, ok := m.(SplashModel); ok {
		a.splash = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid splash model type %T", m)
	return nil
}

func (a *App) updateMenu(msg tea.Msg) tea.Cmd {
	m, cmd := a.menu.Update(msg)
	if next, ok := m.(MenuModel); ok {
		a.menu = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid menu model type %T", m)
	return nil
}

func (a *App) updateSetup(msg tea.Msg) tea.Cmd {
	m, cmd := a.setup.Update(msg)
	if next, ok := m.(SetupModel); ok {
		a.setup = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid setup model type %T", m)
	return nil
}

func (a *App) updateLoadGame(msg tea.Msg) tea.Cmd {
	m, cmd := a.loadGame.Update(msg)
	if next, ok := m.(LoadGameModel); ok {
		a.loadGame = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid load-game model type %T", m)
	return nil
}

func (a *App) updateGame(msg tea.Msg) tea.Cmd {
	m, cmd := a.game.Update(msg)
	if next, ok := m.(GameModel); ok {
		a.game = next
		if ev, ok := msg.(sessionEventMsg); ok {
			actual := session.SessionEvent(ev)
			if actual.Type == "session_error" {
				a.errorMsg = actual.Message
				a.lastResult = actual.Message
				return func() tea.Msg { return switchScreenMsg{screen: ScreenResults} }
			}
		}
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid game model type %T", m)
	return nil
}

func (a *App) updateResults(msg tea.Msg) tea.Cmd {
	m, cmd := a.results.Update(msg)
	if next, ok := m.(ResultsModel); ok {
		a.results = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid results model type %T", m)
	return nil
}

func (a *App) updateStats(msg tea.Msg) tea.Cmd {
	m, cmd := a.stats.Update(msg)
	if next, ok := m.(StatsViewModel); ok {
		a.stats = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid stats model type %T", m)
	return nil
}

func (a *App) updateHistory(msg tea.Msg) tea.Cmd {
	m, cmd := a.history.Update(msg)
	if next, ok := m.(HistoryViewModel); ok {
		a.history = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid history model type %T", m)
	return nil
}

func (a *App) updateReplay(msg tea.Msg) tea.Cmd {
	m, cmd := a.replay.Update(msg)
	if next, ok := m.(ReplayModel); ok {
		a.replay = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid replay model type %T", m)
	return nil
}

func (a *App) updateSettings(msg tea.Msg) tea.Cmd {
	m, cmd := a.settings.Update(msg)
	if next, ok := m.(SettingsModel); ok {
		a.settings = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid settings model type %T", m)
	return nil
}

func (a *App) updateHelp(msg tea.Msg) tea.Cmd {
	m, cmd := a.help.Update(msg)
	if next, ok := m.(HelpModel); ok {
		a.help = next
		return cmd
	}
	a.errorMsg = fmt.Sprintf("invalid help model type %T", m)
	return nil
}

func (a *App) applyWindowSize(msg tea.WindowSizeMsg) {
	switch a.screen {
	case ScreenSplash:
		a.updateSplash(msg)
	case ScreenMenu:
		a.updateMenu(msg)
	case ScreenSetup:
		a.updateSetup(msg)
	case ScreenLoadGame:
		a.updateLoadGame(msg)
	case ScreenGame:
		a.updateGame(msg)
	case ScreenResults:
		a.updateResults(msg)
	case ScreenStats:
		a.updateStats(msg)
	case ScreenHistory:
		a.updateHistory(msg)
	case ScreenReplay:
		a.updateReplay(msg)
	case ScreenSettings:
		a.updateSettings(msg)
	case ScreenHelp:
		a.updateHelp(msg)
	}
}
