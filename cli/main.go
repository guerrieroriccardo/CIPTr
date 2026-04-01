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

// defaultAPIURL is the fallback API endpoint. It can be overridden at build
// time via -ldflags "-X main.defaultAPIURL=https://...".
var defaultAPIURL = "http://localhost:8080/api/v1"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("ciptr-cli %s (%s) built %s\n", version.Version, version.Commit, version.Date)
			return
		case "update":
			selfupdate.Run()
			return
		case "logout":
			if err := auth.ClearToken(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Logged out successfully.")
			return
		}
	}

	// --anon flag: skip saved token and server URL, go straight to login screen.
	// Useful for testing against a local dev backend without touching the production token.
	anon := false
	for _, arg := range os.Args[1:] {
		if arg == "--anon" {
			anon = true
			break
		}
	}

	apiURL := os.Getenv("CIPTR_API_URL")
	if apiURL == "" {
		if !anon {
			if saved := auth.LoadServerURL(); saved != "" {
				apiURL = saved + "/api/v1"
			}
		}
		if apiURL == "" {
			apiURL = defaultAPIURL
		}
	}

	client := apiclient.New(apiURL)

	// Load saved token. If valid, go straight to menu; otherwise show login.
	var initial tui.Screen
	if !anon {
		if token := auth.LoadToken(); token != "" {
			client.Token = token
			initial = tui.NewMenu()
		}
	}
	if initial == nil {
		if anon {
			initial = tui.NewAnonLoginScreen(client)
		} else {
			initial = tui.NewLoginScreen(client)
		}
	}
	var app tui.App
	if anon {
		app = tui.NewAnonApp(initial, client)
	} else {
		app = tui.NewApp(initial, client)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
