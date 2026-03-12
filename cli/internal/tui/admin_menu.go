package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

// AdminMenu is the administration submenu screen.
type AdminMenu struct {
	list   list.Model
	client *apiclient.Client
	width  int
	height int
}

func NewAdminMenu(client *apiclient.Client) AdminMenu {
	items := []list.Item{
		MenuItem{name: "Users", desc: "Manage user accounts (admin only)", key: "users"},
		MenuItem{name: "Audit Logs", desc: "View recent activity log (admin only)", key: "audit_logs"},
		MenuItem{name: "Settings", desc: "Hostname nomenclature and system config (admin only)", key: "settings"},
		MenuItem{name: "Change Password", desc: "Change your own password", key: "change_password"},
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Administration"
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginLeft(1)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	return AdminMenu{list: l, client: client}
}

func (m AdminMenu) Title() string { return "Administration" }

func (m AdminMenu) Init() tea.Cmd { return nil }

func (m AdminMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return PopScreenMsg{} }
		case "enter":
			item, ok := m.list.SelectedItem().(MenuItem)
			if !ok {
				return m, nil
			}
			return m, m.handleSelection(item.Key())
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AdminMenu) View() string {
	return fmt.Sprintf("%s\n%s", m.list.View(), HelpStyle.Render("esc: back • enter: select"))
}

func (m AdminMenu) handleSelection(key string) tea.Cmd {
	switch key {
	case "change_password":
		screen := NewChangePasswordScreen(m.client)
		return func() tea.Msg { return PushScreenMsg{Screen: screen} }
	case "users":
		def, ok := resource.Registry["users"]
		if !ok {
			return nil
		}
		screen := NewResourceTable(def, m.client)
		return func() tea.Msg { return PushScreenMsg{Screen: screen} }
	case "audit_logs":
		def, ok := resource.Registry["audit_logs"]
		if !ok {
			return nil
		}
		screen := NewResourceTable(def, m.client)
		return func() tea.Msg { return PushScreenMsg{Screen: screen} }
	case "settings":
		screen := NewSettingsScreen(m.client)
		return func() tea.Msg { return PushScreenMsg{Screen: screen} }
	}
	return nil
}
