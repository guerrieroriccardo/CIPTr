package tui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/auth"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

const (
	MinWidth  = 60
	MinHeight = 15
)

// PushScreenMsg tells the app to push a new screen onto the nav stack.
type PushScreenMsg struct {
	Screen Screen
}

// PopScreenMsg tells the app to pop the current screen (no data changed).
type PopScreenMsg struct{}

// MutationPopMsg tells the app to pop and refresh data (after create/update/delete).
type MutationPopMsg struct{}

// App is the root bubbletea model that manages the navigation stack.
type App struct {
	nav    NavStack
	client *apiclient.Client
	width  int
	height int
}

// NewApp creates the root application model with the given initial screen.
func NewApp(initial Screen, client *apiclient.Client) App {
	app := App{client: client}
	app.nav.Push(initial)
	return app
}

func (a App) Init() tea.Cmd {
	cmds := []tea.Cmd{a.nav.Current().Init()}
	// Only init resolver if we're already authenticated (not on login screen).
	if a.client.Token != "" {
		cmds = append(cmds, resource.InitResolver(a.client))
	}
	return tea.Batch(cmds...)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a.updateCurrent(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}

	case MenuItemSelected:
		return a.handleMenuSelection(msg.Key)

	case PushScreenMsg:
		a.nav.Push(msg.Screen)
		// Forward current window size so the new screen renders correctly.
		sizeCmd := func() tea.Msg {
			return tea.WindowSizeMsg{Width: a.width, Height: a.height}
		}
		return a, tea.Batch(msg.Screen.Init(), sizeCmd)

	case dataErrorMsg:
		if errors.Is(msg.err, apiclient.ErrForbidden) {
			return a.updateCurrent(msg)
		}
		if errors.Is(msg.err, apiclient.ErrUnauthorized) {
			return a.forceLogin()
		}
		return a.updateCurrent(msg)

	case formErrorMsg:
		if errors.Is(msg.err, apiclient.ErrForbidden) {
			return a.updateCurrent(msg)
		}
		if errors.Is(msg.err, apiclient.ErrUnauthorized) {
			return a.forceLogin()
		}
		return a.updateCurrent(msg)

	case loginSuccessMsg:
		// Replace login screen with main menu and init resolver.
		a.nav = NavStack{}
		menu := NewMenu()
		a.nav.Push(menu)
		sizeCmd := func() tea.Msg {
			return tea.WindowSizeMsg{Width: a.width, Height: a.height}
		}
		return a, tea.Batch(menu.Init(), resource.InitResolver(a.client), sizeCmd)

	case resource.ResolverReadyMsg:
		resource.Resolve = msg.R
		return a, nil

	case PopScreenMsg:
		if a.nav.Len() > 1 {
			a.nav.Pop()
			// No data changed — preserve existing screen state, just fix layout.
			return a, func() tea.Msg {
				return tea.WindowSizeMsg{Width: a.width, Height: a.height}
			}
		}

	case MutationPopMsg:
		if a.nav.Len() > 1 {
			a.nav.Pop()
			// Data changed — re-fetch table data and refresh resolver for pickers.
			return a, tea.Batch(
				a.nav.Current().Init(),
				resource.InitResolver(a.client),
			)
		}
	}

	return a.updateCurrent(msg)
}

func (a App) View() string {
	if a.width < MinWidth || a.height < MinHeight {
		msg := fmt.Sprintf("Terminal too small (%dx%d minimum)\nCurrent: %dx%d", MinWidth, MinHeight, a.width, a.height)
		style := lipgloss.NewStyle().
			Width(a.width).
			Height(a.height).
			Align(lipgloss.Center, lipgloss.Center)
		return style.Render(msg)
	}

	current := a.nav.Current()
	if current == nil {
		return ""
	}

	var out string

	// Show breadcrumb when deeper than the main menu.
	if a.nav.Len() > 1 {
		out += BreadcrumbStyle.Render(a.nav.Breadcrumb()) + "\n"
	}

	out += current.View()
	return out
}

// handleMenuSelection maps a menu key to a resource table screen.
func (a App) handleMenuSelection(key string) (tea.Model, tea.Cmd) {
	// Hierarchical browse entry point.
	if key == "browse_clients" {
		screen := NewBrowseByClientScreen(a.client)
		return a, func() tea.Msg {
			return PushScreenMsg{Screen: screen}
		}
	}

	// Logout: clear token and go back to login screen.
	if key == "logout" {
		return a.forceLogin()
	}

	// Administration submenu.
	if key == "admin" {
		screen := NewAdminMenu(a.client)
		return a, func() tea.Msg {
			return PushScreenMsg{Screen: screen}
		}
	}

	def, ok := resource.Registry[key]
	if !ok {
		// Unknown key, ignore.
		return a, nil
	}

	// For lookup tables, enter drills down to show associated entries.
	var screen ResourceTable
	switch key {
	case "categories":
		screen = NewResourceTableWithSelect(def, a.client, categoryDrillDown(a.client))
	case "suppliers":
		screen = NewResourceTableWithSelect(def, a.client, supplierDrillDown(a.client))
	case "device_models":
		screen = NewResourceTableWithSelect(def, a.client, deviceModelDrillDown(a.client))
	case "manufacturers":
		screen = NewResourceTableWithSelect(def, a.client, manufacturerDrillDown(a.client))
	case "locations":
		screen = NewResourceTableWithSelect(def, a.client, locationDrillDown(a.client))
	case "operating_systems":
		screen = NewResourceTableWithSelect(def, a.client, osDrillDown(a.client))
	default:
		screen = NewResourceTable(def, a.client)
	}
	return a, func() tea.Msg {
		return PushScreenMsg{Screen: screen}
	}
}

// forceLogin clears the expired token and replaces the nav stack with the login screen.
func (a App) forceLogin() (tea.Model, tea.Cmd) {
	a.client.Token = ""
	_ = auth.ClearToken()
	a.nav = NavStack{}
	login := NewLoginScreen(a.client)
	a.nav.Push(login)
	return a, login.Init()
}

// updateCurrent forwards a message to the current screen and returns the result.
func (a App) updateCurrent(msg tea.Msg) (tea.Model, tea.Cmd) {
	current := a.nav.Current()
	if current == nil {
		return a, nil
	}
	updated, cmd := current.Update(msg)
	a.nav.stack[len(a.nav.stack)-1] = updated.(Screen)
	return a, cmd
}
