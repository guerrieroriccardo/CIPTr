package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui"

	// Register resource definitions.
	_ "github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

func main() {
	apiURL := os.Getenv("CIPTR_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080/api/v1"
	}

	client := apiclient.New(apiURL)

	menu := tui.NewMenu()
	app := tui.NewApp(menu, client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
