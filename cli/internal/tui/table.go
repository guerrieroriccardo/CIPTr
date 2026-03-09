package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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

type exportSuccessMsg struct {
	path string
}

type exportErrorMsg struct {
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

	status string // flash message (e.g. export result)

	// Filtering
	filtering   bool
	filterInput textinput.Model
	filterText  string
	allRows     []table.Row // unfiltered rows
	filtered    []int       // indices into items for currently visible rows
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

	fi := textinput.New()
	fi.Prompt = "/ "
	fi.CharLimit = 64

	return ResourceTable{
		def:         def,
		client:      client,
		table:       t,
		filterInput: fi,
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
		rt.allRows = make([]table.Row, len(msg.items))
		for i, item := range msg.items {
			rt.allRows[i] = rt.def.ToRow(item)
		}
		rt.applyFilter()

	case dataErrorMsg:
		rt.err = msg.err
		rt.loaded = true

	case exportSuccessMsg:
		rt.status = "Saved: " + msg.path
		return rt, nil

	case exportErrorMsg:
		rt.status = "Error: " + msg.err.Error()
		return rt, nil

	case tea.KeyMsg:
		rt.status = "" // clear flash on any key press

		// When filtering, handle filter input keys.
		if rt.filtering {
			switch msg.String() {
			case "enter", "esc":
				rt.filtering = false
				rt.filterInput.Blur()
				if msg.String() == "esc" {
					rt.filterText = ""
					rt.filterInput.SetValue("")
					rt.applyFilter()
				}
				return rt, nil
			default:
				var cmd tea.Cmd
				rt.filterInput, cmd = rt.filterInput.Update(msg)
				rt.filterText = rt.filterInput.Value()
				rt.applyFilter()
				return rt, cmd
			}
		}

		switch msg.String() {
		case "esc":
			if rt.filterText != "" {
				rt.filterText = ""
				rt.filterInput.SetValue("")
				rt.applyFilter()
				return rt, nil
			}
			return rt, func() tea.Msg { return PopScreenMsg{} }
		case "/":
			rt.filtering = true
			rt.filterInput.Focus()
			return rt, textinput.Blink
		case "n":
			if rt.def.Create == nil {
				break
			}
			return rt, func() tea.Msg {
				return PushScreenMsg{Screen: NewResourceForm(rt.def, rt.client, "", nil)}
			}
		case "enter":
			if item := rt.selectedItem(); item != nil {
				if rt.onSelect != nil {
					return rt, rt.onSelect(item)
				}
				if rt.def.Update == nil {
					break
				}
				id := rt.def.GetID(item)
				return rt, func() tea.Msg {
					return PushScreenMsg{Screen: NewResourceForm(rt.def, rt.client, id, item)}
				}
			}
		case "e":
			if rt.def.Update == nil {
				break
			}
			if item := rt.selectedItem(); item != nil {
				id := rt.def.GetID(item)
				return rt, func() tea.Msg {
					return PushScreenMsg{Screen: NewResourceForm(rt.def, rt.client, id, item)}
				}
			}
		case "d":
			if rt.def.Delete == nil {
				break
			}
			if item := rt.selectedItem(); item != nil {
				id := rt.def.GetID(item)
				row := rt.def.ToRow(item)
				name := row[1]
				return rt, func() tea.Msg {
					return PushScreenMsg{Screen: NewConfirmDelete(rt.def, rt.client, id, name)}
				}
			}
		case "r":
			return rt, rt.loadData()
		case "l":
			if rt.def.ExportLabel == nil {
				break
			}
			if item := rt.selectedItem(); item != nil {
				id := rt.def.GetID(item)
				def := rt.def
				client := rt.client
				rt.status = "Downloading label..."
				return rt, func() tea.Msg {
					path, err := def.ExportLabel(client, id)
					if err != nil {
						return exportErrorMsg{err: err}
					}
					return exportSuccessMsg{path: path}
				}
			}
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

	title := TitleStyle.Render(fmt.Sprintf("%s (%d)", rt.def.Plural, len(rt.filtered)))

	var filterLine string
	if rt.filtering {
		filterLine = rt.filterInput.View() + "\n"
	} else if rt.filterText != "" {
		filterLine = HelpStyle.Render(fmt.Sprintf("filter: %s (esc to clear)", rt.filterText)) + "\n"
	}

	helpText := "/ filter • n new • enter edit • d delete • r refresh • esc back"
	if rt.onSelect != nil {
		helpText = "/ filter • n new • enter open • e edit • d delete • r refresh • esc back"
	}
	if rt.def.ExportLabel != nil {
		helpText = "l label • " + helpText
	}
	help := HelpStyle.Render(helpText)

	var statusLine string
	if rt.status != "" {
		statusLine = HelpStyle.Render(rt.status) + "\n"
	}

	return fmt.Sprintf("%s\n%s%s\n%s%s", title, filterLine, rt.table.View(), statusLine, help)
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

// applyFilter rebuilds the table rows based on the current filterText.
func (rt *ResourceTable) applyFilter() {
	if rt.filterText == "" {
		rt.filtered = make([]int, len(rt.allRows))
		for i := range rt.allRows {
			rt.filtered[i] = i
		}
		rt.table.SetRows(rt.allRows)
		return
	}
	query := strings.ToLower(rt.filterText)
	var rows []table.Row
	rt.filtered = nil
	for i, row := range rt.allRows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), query) {
				rows = append(rows, row)
				rt.filtered = append(rt.filtered, i)
				break
			}
		}
	}
	rt.table.SetRows(rows)
}

// selectedItem returns the original item for the currently highlighted table row,
// accounting for filtering. Returns nil if nothing is selected.
func (rt ResourceTable) selectedItem() any {
	if len(rt.filtered) == 0 {
		return nil
	}
	cursor := rt.table.Cursor()
	if cursor >= len(rt.filtered) {
		return nil
	}
	return rt.items[rt.filtered[cursor]]
}
