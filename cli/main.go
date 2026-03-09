package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/auth"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/selfupdate"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/version"

	// Register resource definitions.
	_ "github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("ciptr-cli %s (%s) built %s\n", version.Version, version.Commit, version.Date)
			return
		case "update":
			selfupdate.Run()
			return
		}
	}

	apiURL := os.Getenv("CIPTR_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080/api/v1"
	}

	client := apiclient.New(apiURL)

	// Load saved token. If valid, go straight to menu; otherwise show login.
	var initial tui.Screen
	if token := auth.LoadToken(); token != "" {
		client.Token = token
		initial = tui.NewMenu()
	} else {
		initial = tui.NewLoginScreen(client)
	}
	app := tui.NewApp(initial, client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
