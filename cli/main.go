package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui"
)

func main() {
	menu := tui.NewMenu()
	app := tui.NewApp(menu)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
