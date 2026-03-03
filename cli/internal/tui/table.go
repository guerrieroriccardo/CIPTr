package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

// Messages for async data loading.
type dataLoadedMsg struct {
	items []any
}

type dataErrorMsg struct {
	err error
}

// ResourceTable is a generic list screen for any resource.
type ResourceTable struct {
	def      *resource.Def
	client   *apiclient.Client
	table    table.Model
	items    []any // raw items from API
	err      error
	loaded   bool
	width    int
	height   int
	onSelect func(item any) tea.Cmd // if set, enter drills down instead of editing
}

// NewResourceTable creates a table screen for the given resource definition.
func NewResourceTable(def *resource.Def, client *apiclient.Client) ResourceTable {
	t := table.New(
		table.WithColumns(def.Columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return ResourceTable{
		def:    def,
		client: client,
		table:  t,
	}
}

// NewResourceTableWithSelect creates a browse-mode table where enter drills down.
func NewResourceTableWithSelect(def *resource.Def, client *apiclient.Client, onSelect func(item any) tea.Cmd) ResourceTable {
	rt := NewResourceTable(def, client)
	rt.onSelect = onSelect
	return rt
}

func (rt ResourceTable) Title() string {
	return rt.def.Plural
}

func (rt ResourceTable) Init() tea.Cmd {
	return rt.loadData()
}

func (rt ResourceTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rt.width = msg.Width
		rt.height = msg.Height
		rt.table.SetHeight(msg.Height - 8)
		rt.table.SetWidth(msg.Width)

	case dataLoadedMsg:
		rt.items = msg.items
		rt.loaded = true
		rt.err = nil
		rows := make([]table.Row, len(msg.items))
		for i, item := range msg.items {
			rows[i] = rt.def.ToRow(item)
		}
		rt.table.SetRows(rows)

	case dataErrorMsg:
		rt.err = msg.err
		rt.loaded = true

	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			// New: push create form
			return rt, func() tea.Msg {
				return PushScreenMsg{Screen: NewResourceForm(rt.def, rt.client, "", nil)}
			}
		case "enter":
			// Drill-down (if onSelect is set) or edit selected item
			if len(rt.items) > 0 {
				idx := rt.table.Cursor()
				if idx < len(rt.items) {
					item := rt.items[idx]
					if rt.onSelect != nil {
						return rt, rt.onSelect(item)
					}
					id := rt.def.GetID(item)
					return rt, func() tea.Msg {
						return PushScreenMsg{Screen: NewResourceForm(rt.def, rt.client, id, item)}
					}
				}
			}
		case "e":
			// Edit selected item (always available, needed in browse mode)
			if len(rt.items) > 0 {
				idx := rt.table.Cursor()
				if idx < len(rt.items) {
					item := rt.items[idx]
					id := rt.def.GetID(item)
					return rt, func() tea.Msg {
						return PushScreenMsg{Screen: NewResourceForm(rt.def, rt.client, id, item)}
					}
				}
			}
		case "d":
			// Delete selected item
			if len(rt.items) > 0 {
				idx := rt.table.Cursor()
				if idx < len(rt.items) {
					item := rt.items[idx]
					id := rt.def.GetID(item)
					row := rt.def.ToRow(item)
					name := row[1] // second column is typically the name
					return rt, func() tea.Msg {
						return PushScreenMsg{Screen: NewConfirmDelete(rt.def, rt.client, id, name)}
					}
				}
			}
		case "r":
			// Refresh
			return rt, rt.loadData()
		}
	}

	var cmd tea.Cmd
	rt.table, cmd = rt.table.Update(msg)
	return rt, cmd
}

func (rt ResourceTable) View() string {
	if rt.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", rt.err)) + "\n\n" +
			HelpStyle.Render("r refresh • esc back")
	}
	if !rt.loaded {
		return "Loading " + rt.def.Plural + "..."
	}

	title := TitleStyle.Render(fmt.Sprintf("%s (%d)", rt.def.Plural, len(rt.items)))
	helpText := "n new • enter edit • d delete • r refresh • esc back"
	if rt.onSelect != nil {
		helpText = "n new • enter open • e edit • d delete • r refresh • esc back"
	}
	help := HelpStyle.Render(helpText)

	return fmt.Sprintf("%s\n%s\n%s", title, rt.table.View(), help)
}

func (rt ResourceTable) loadData() tea.Cmd {
	def := rt.def
	client := rt.client
	return func() tea.Msg {
		items, err := def.List(client)
		if err != nil {
			return dataErrorMsg{err: err}
		}
		return dataLoadedMsg{items: items}
	}
}
