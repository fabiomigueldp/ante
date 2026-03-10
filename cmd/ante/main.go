package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/tui"
)

func main() {
	_ = audio.Init()
	defer audio.Close()

	app := tui.NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
