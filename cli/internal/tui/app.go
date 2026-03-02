package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// PushScreenMsg tells the app to push a new screen onto the nav stack.
type PushScreenMsg struct {
	Screen Screen
}

// PopScreenMsg tells the app to pop the current screen.
type PopScreenMsg struct{}

// App is the root bubbletea model that manages the navigation stack.
type App struct {
	nav    NavStack
	width  int
	height int
}

// NewApp creates the root application model with the given initial screen.
func NewApp(initial Screen) App {
	app := App{}
	app.nav.Push(initial)
	return app
}

func (a App) Init() tea.Cmd {
	return a.nav.Current().Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to current screen.
		return a.updateCurrent(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "esc":
			if a.nav.Len() > 1 {
				a.nav.Pop()
				return a.updateCurrent(nil)
			}
		}

	case PushScreenMsg:
		a.nav.Push(msg.Screen)
		cmd := msg.Screen.Init()
		return a, cmd

	case PopScreenMsg:
		if a.nav.Len() > 1 {
			a.nav.Pop()
			return a.updateCurrent(nil)
		}
	}

	return a.updateCurrent(msg)
}

func (a App) View() string {
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
