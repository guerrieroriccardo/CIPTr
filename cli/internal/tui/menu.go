package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/version"
)

// MenuItem represents a selectable entry in the main menu.
type MenuItem struct {
	name string
	desc string
	key  string // internal key used to identify the action
}

func (i MenuItem) Title() string       { return i.name }
func (i MenuItem) Description() string { return i.desc }
func (i MenuItem) FilterValue() string { return i.name }
func (i MenuItem) Key() string         { return i.key }

// MenuItemSelected is sent when a menu item is chosen.
type MenuItemSelected struct {
	Key string
}

// Menu is the main menu screen.
type Menu struct {
	list   list.Model
	width  int
	height int
}

// NewMenu creates the main menu with hierarchical and flat entries.
func NewMenu() Menu {
	items := []list.Item{
		// Hierarchical entry point
		MenuItem{name: "Browse by Client", desc: "Navigate clients → sites → resources", key: "browse_clients"},
		// Flat resource access
		MenuItem{name: "All Clients", desc: "List all clients", key: "clients"},
		MenuItem{name: "All Sites", desc: "List all sites", key: "sites"},
		MenuItem{name: "All Locations", desc: "List all locations", key: "locations"},
		MenuItem{name: "All Address Blocks", desc: "List all address blocks", key: "address_blocks"},
		MenuItem{name: "All VLANs", desc: "List all VLANs", key: "vlans"},
		MenuItem{name: "All Manufacturers", desc: "Hardware manufacturers", key: "manufacturers"},
		MenuItem{name: "All Categories", desc: "Device categories", key: "categories"},
		MenuItem{name: "All Suppliers", desc: "Device suppliers", key: "suppliers"},
		MenuItem{name: "All Operating Systems", desc: "Operating systems catalog", key: "operating_systems"},
		MenuItem{name: "All Device Models", desc: "Hardware catalog", key: "device_models"},
		MenuItem{name: "All Devices", desc: "List all devices", key: "devices"},
		MenuItem{name: "All Device Interfaces", desc: "List all NICs", key: "device_interfaces"},
		MenuItem{name: "All Device IPs", desc: "List all IP assignments", key: "device_ips"},
		MenuItem{name: "All Device Connections", desc: "List all physical connections", key: "device_connections"},
		MenuItem{name: "All Switches", desc: "List all switches", key: "switches"},
		MenuItem{name: "All Switch Ports", desc: "List all switch ports", key: "switch_ports"},
		MenuItem{name: "All Patch Panels", desc: "List all patch panels", key: "patch_panels"},
		MenuItem{name: "All Patch Panel Ports", desc: "List all patch panel ports", key: "patch_panel_ports"},
		MenuItem{name: "All Device Groups", desc: "Named groups of devices", key: "device_groups"},
		MenuItem{name: "All Device Group Members", desc: "Devices in groups", key: "device_group_members"},
		// Administration
		MenuItem{name: "Administration", desc: "Users, audit logs, password change", key: "admin"},
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "CIPTr — Client IP Tracker"
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginLeft(1)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	return Menu{list: l}
}

func (m Menu) Title() string { return "Menu" }

func (m Menu) Init() tea.Cmd { return nil }

func (m Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-2)
		return m, nil

	case tea.KeyMsg:
		// Don't intercept keys while filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			item, ok := m.list.SelectedItem().(MenuItem)
			if ok {
				return m, func() tea.Msg {
					return MenuItemSelected{Key: item.Key()}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Menu) View() string {
	return fmt.Sprintf("%s\n%s", m.list.View(), HelpStyle.Render(fmt.Sprintf("ciptr-cli %s • q quit • / filter • enter select", version.Version)))
}
