package tui

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

// globalTableSeq assigns a unique ID to each ResourceTable instance so that
// stale dataLoadedMsg from a popped table cannot contaminate a different table.
var globalTableSeq uint64

// Messages for async data loading.
type dataLoadedMsg struct {
	items []any
	seq   uint64
}

type dataErrorMsg struct {
	err error
	seq uint64
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
	seq      uint64             // unique ID for this instance; filters out stale dataLoadedMsg
	onSelect func(item any) tea.Cmd // if set, enter drills down instead of editing

	status string // flash message (e.g. export result)

	// Filtering
	filtering   bool
	filterInput textinput.Model
	filterText  string
	allRows     []table.Row // rows in current sort order
	filtered    []int       // indices into items for currently visible rows

	// Sorting
	sorting    bool // sort picker open
	sortCol    int  // -1 = none
	sortAsc    bool
	sortCursor int

	// Multi-select for bulk edit
	selected map[int]bool // indices into items
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
		seq:         atomic.AddUint64(&globalTableSeq, 1),
		sortCol:     -1,
		sortAsc:     true,
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
		rt.scaleColumns()

	case dataLoadedMsg:
		if msg.seq != rt.seq {
			return rt, nil // stale message from a different table instance
		}
		rt.items = msg.items
		rt.loaded = true
		rt.err = nil
		rt.selected = nil
		rt.allRows = make([]table.Row, len(msg.items))
		for i, item := range msg.items {
			rt.allRows[i] = rt.def.ToRow(item)
		}
		rt.applySort()
		rt.applyFilter()
		rt.table.GotoTop()

	case dataErrorMsg:
		if msg.seq != rt.seq {
			return rt, nil // stale message from a different table instance
		}
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

		// When sort picker is open, handle navigation.
		if rt.sorting {
			switch msg.String() {
			case "esc":
				rt.sorting = false
			case "up", "k":
				if rt.sortCursor > 0 {
					rt.sortCursor--
				}
			case "down", "j":
				if rt.sortCursor < len(rt.def.Columns)-1 {
					rt.sortCursor++
				}
			case "enter":
				rt.sorting = false
				if rt.sortCol == rt.sortCursor {
					rt.sortAsc = !rt.sortAsc
				} else {
					rt.sortCol = rt.sortCursor
					rt.sortAsc = true
				}
				rt.applySort()
				rt.applyFilter()
			}
			return rt, nil
		}

		switch msg.String() {
		case "esc":
			if rt.filterText != "" {
				rt.filterText = ""
				rt.filterInput.SetValue("")
				rt.applyFilter()
				return rt, nil
			}
			if len(rt.selected) > 0 {
				rt.selected = nil
				rt.applyFilter()
				return rt, nil
			}
			return rt, func() tea.Msg { return PopScreenMsg{} }
		case " ":
			if idx := rt.selectedItemIndex(); idx >= 0 {
				if rt.selected == nil {
					rt.selected = map[int]bool{}
				}
				if rt.selected[idx] {
					delete(rt.selected, idx)
					if len(rt.selected) == 0 {
						rt.selected = nil
					}
				} else {
					rt.selected[idx] = true
				}
				rt.applyFilter()
			}
			return rt, nil
		case "b":
			if rt.def.Update == nil || len(rt.selected) == 0 {
				break
			}
			var items []any
			for idx := range rt.selected {
				items = append(items, rt.items[idx])
			}
			return rt, func() tea.Msg {
				return PushScreenMsg{Screen: NewBulkEditForm(rt.def, rt.client, items)}
			}
		case "/":
			rt.filtering = true
			rt.filterInput.Focus()
			return rt, textinput.Blink
		case "s":
			rt.sorting = true
			if rt.sortCol >= 0 {
				rt.sortCursor = rt.sortCol
			} else {
				rt.sortCursor = 0
			}
			return rt, nil
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
		default:
			// Custom actions defined per resource.
			for _, action := range rt.def.CustomActions {
				if msg.String() == action.Key {
					if item := rt.selectedItem(); item != nil {
						handler := action.Handler
						client := rt.client
						return rt, func() tea.Msg {
							def, id, defaults := handler(item)
							def.Defaults = defaults
							return PushScreenMsg{Screen: NewResourceForm(def, client, id, nil)}
						}
					}
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

	sortIndicator := ""
	if rt.sortCol >= 0 && rt.sortCol < len(rt.def.Columns) {
		dir := "↑"
		if !rt.sortAsc {
			dir = "↓"
		}
		sortIndicator = fmt.Sprintf(" [%s %s]", rt.def.Columns[rt.sortCol].Title, dir)
	}
	selIndicator := ""
	if len(rt.selected) > 0 {
		selIndicator = fmt.Sprintf(" [%d selected]", len(rt.selected))
	}
	title := TitleStyle.Render(fmt.Sprintf("%s (%d)%s%s", rt.def.Plural, len(rt.filtered), sortIndicator, selIndicator))

	// Sort picker overlay.
	if rt.sorting {
		var sb strings.Builder
		sb.WriteString(title + "\n")
		sb.WriteString(HelpStyle.Render("Sort by: ↑↓ navigate • enter select • esc cancel") + "\n\n")
		for i, col := range rt.def.Columns {
			prefix := "  "
			if i == rt.sortCursor {
				prefix = "> "
			}
			suffix := ""
			if i == rt.sortCol {
				if rt.sortAsc {
					suffix = " ↑"
				} else {
					suffix = " ↓"
				}
			}
			sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, col.Title, suffix))
		}
		return sb.String()
	}

	var filterLine string
	if rt.filtering {
		filterLine = rt.filterInput.View() + "\n"
	} else if rt.filterText != "" {
		filterLine = HelpStyle.Render(fmt.Sprintf("filter: %s (esc to clear)", rt.filterText)) + "\n"
	}

	helpText := "/ filter • s sort • n new • enter edit • d delete • r refresh • esc back"
	if rt.onSelect != nil {
		helpText = "/ filter • s sort • n new • enter open • e edit • d delete • r refresh • esc back"
	}
	if rt.def.Update != nil {
		helpText = "space select • b bulk edit • " + helpText
	}
	if rt.def.ExportLabel != nil {
		helpText = "l label • " + helpText
	}
	for _, action := range rt.def.CustomActions {
		helpText = action.Key + " " + action.Label + " • " + helpText
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
	seq := rt.seq
	return func() tea.Msg {
		items, err := def.List(client)
		if err != nil {
			return dataErrorMsg{err: err, seq: seq}
		}
		return dataLoadedMsg{items: items, seq: seq}
	}
}

// applySort sorts allRows and items together by the selected column.
func (rt *ResourceTable) applySort() {
	if rt.sortCol < 0 || rt.sortCol >= len(rt.def.Columns) {
		return
	}
	col := rt.sortCol
	asc := rt.sortAsc
	// Build a combined index so we can sort items and allRows in parallel.
	n := len(rt.allRows)
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	// Stable sort to preserve API order for equal values.
	stableSort(indices, func(a, b int) bool {
		va := ""
		vb := ""
		if col < len(rt.allRows[a]) {
			va = rt.allRows[a][col]
		}
		if col < len(rt.allRows[b]) {
			vb = rt.allRows[b][col]
		}
		if asc {
			return strings.ToLower(va) < strings.ToLower(vb)
		}
		return strings.ToLower(va) > strings.ToLower(vb)
	})
	newRows := make([]table.Row, n)
	newItems := make([]any, n)
	for i, idx := range indices {
		newRows[i] = rt.allRows[idx]
		newItems[i] = rt.items[idx]
	}
	rt.allRows = newRows
	rt.items = newItems
}

// stableSort is a simple insertion sort (stable, adequate for typical list sizes).
func stableSort(indices []int, less func(a, b int) bool) {
	for i := 1; i < len(indices); i++ {
		for j := i; j > 0 && less(indices[j], indices[j-1]); j-- {
			indices[j], indices[j-1] = indices[j-1], indices[j]
		}
	}
}

// applyFilter rebuilds the table rows based on the current filterText.
func (rt *ResourceTable) applyFilter() {
	if rt.filterText == "" {
		rt.filtered = make([]int, len(rt.allRows))
		for i := range rt.allRows {
			rt.filtered[i] = i
		}
		rt.table.SetRows(rt.markSelected(rt.allRows, rt.filtered))
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
	rt.table.SetRows(rt.markSelected(rows, rt.filtered))
}

// markSelected prepends a marker to the first column of selected rows.
func (rt *ResourceTable) markSelected(rows []table.Row, filtered []int) []table.Row {
	if len(rt.selected) == 0 {
		return rows
	}
	out := make([]table.Row, len(rows))
	for i, row := range rows {
		newRow := make(table.Row, len(row))
		copy(newRow, row)
		if rt.selected[filtered[i]] {
			newRow[0] = "*" + newRow[0]
		}
		out[i] = newRow
	}
	return out
}

// selectedItem returns the original item for the currently highlighted table row,
// accounting for filtering. Returns nil if nothing is selected.
func (rt ResourceTable) selectedItem() any {
	idx := rt.selectedItemIndex()
	if idx < 0 {
		return nil
	}
	return rt.items[idx]
}

// selectedItemIndex returns the index into rt.items for the currently highlighted row.
func (rt ResourceTable) selectedItemIndex() int {
	if len(rt.filtered) == 0 {
		return -1
	}
	cursor := rt.table.Cursor()
	if cursor >= len(rt.filtered) {
		return -1
	}
	return rt.filtered[cursor]
}

// scaleColumns proportionally resizes column widths to fill the terminal width.
func (rt *ResourceTable) scaleColumns() {
	origCols := rt.def.Columns
	if len(origCols) == 0 || rt.width <= 0 {
		return
	}

	// Total original width (used as basis for proportional scaling).
	var origTotal int
	for _, c := range origCols {
		origTotal += c.Width
	}
	if origTotal == 0 {
		return
	}

	// Available width: terminal width minus column separators/padding.
	// bubbles/table uses ~3 chars per column for padding/borders.
	available := rt.width - len(origCols)*3 - 2
	if available < len(origCols)*4 {
		available = len(origCols) * 4
	}

	scaled := make([]table.Column, len(origCols))
	var assigned int
	for i, c := range origCols {
		scaled[i].Title = c.Title
		w := c.Width * available / origTotal
		if w < 4 {
			w = 4
		}
		scaled[i].Width = w
		assigned += w
	}
	// Distribute remaining space to the last column.
	if diff := available - assigned; diff > 0 {
		scaled[len(scaled)-1].Width += diff
	}

	rt.table.SetColumns(scaled)
}
