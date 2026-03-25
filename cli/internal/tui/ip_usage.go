package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// ipUsageLoadedMsg carries the API response.
type ipUsageLoadedMsg struct {
	resp models.IPUsageResponse
}

// ipUsageErrorMsg carries a fetch error.
type ipUsageErrorMsg struct {
	err error
}

// IPUsageScreen shows IP address space utilization with drill-down navigation.
type IPUsageScreen struct {
	client *apiclient.Client
	level  string // "global", "client", "site", "vlan"
	param  string // query params string, e.g. "?site_id=3"

	resp   models.IPUsageResponse
	rows   []ipUsageRow // flattened rows for the table
	table  table.Model
	loaded bool
	err    error
	width  int
	height int
}

// ipUsageRow holds a flattened row with a reference to its source node.
type ipUsageRow struct {
	node  models.IPUsageNode
	depth int // indentation level
}

// NewIPUsageScreen creates a new IP usage visualization screen.
func NewIPUsageScreen(client *apiclient.Client, level, param string) IPUsageScreen {
	cols := ipUsageColumns(level)
	t := table.New(
		table.WithColumns(cols),
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

	return IPUsageScreen{
		client: client,
		level:  level,
		param:  param,
		table:  t,
	}
}

func (s IPUsageScreen) Title() string {
	return "IP Address Space"
}

func (s IPUsageScreen) Init() tea.Cmd {
	client := s.client
	param := s.param
	return func() tea.Msg {
		resp, err := client.GetIPUsage(param)
		if err != nil {
			return ipUsageErrorMsg{err: err}
		}
		return ipUsageLoadedMsg{resp: resp}
	}
}

func (s IPUsageScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.table.SetWidth(msg.Width)
		s.table.SetHeight(msg.Height - 5) // reserve space for title + help
		s.table.SetColumns(scaleColumns(ipUsageColumns(s.level), msg.Width))
		if s.loaded {
			s.rebuildTable()
		}
		return s, nil

	case ipUsageLoadedMsg:
		s.resp = msg.resp
		s.loaded = true
		s.err = nil
		s.flattenRows()
		s.rebuildTable()
		return s, nil

	case ipUsageErrorMsg:
		s.err = msg.err
		s.loaded = true
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return PopScreenMsg{} }
		case "q":
			return s, tea.Quit
		case "enter":
			if cmd := s.drillDown(); cmd != nil {
				return s, cmd
			}
		}
	}

	var cmd tea.Cmd
	s.table, cmd = s.table.Update(msg)
	return s, cmd
}

func (s IPUsageScreen) View() string {
	if s.err != nil {
		return ErrorStyle.Render("Error: "+s.err.Error()) + "\n\n" + HelpStyle.Render("esc back")
	}
	if !s.loaded {
		return "Loading IP usage..."
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("IP Address Space") + "\n")
	b.WriteString(s.table.View())
	b.WriteString("\n")

	help := "esc back • q quit"
	if s.level != "vlan" {
		help = "enter drill down • " + help
	}
	b.WriteString(HelpStyle.Render(help))
	return b.String()
}

// flattenRows walks the response tree and produces flat rows for display.
func (s *IPUsageScreen) flattenRows() {
	s.rows = nil
	for _, item := range s.resp.Items {
		s.flattenNode(item, 0)
	}
}

func (s *IPUsageScreen) flattenNode(node models.IPUsageNode, depth int) {
	s.rows = append(s.rows, ipUsageRow{node: node, depth: depth})
	for _, child := range node.Children {
		s.flattenNode(child, depth+1)
	}
}

// rebuildTable populates the bubbles table from flattened rows.
func (s *IPUsageScreen) rebuildTable() {
	var tableRows []table.Row
	for _, r := range s.rows {
		tableRows = append(tableRows, s.rowToTableRow(r))
	}
	s.table.SetRows(tableRows)
}

func (s *IPUsageScreen) rowToTableRow(r ipUsageRow) table.Row {
	indent := strings.Repeat("  ", r.depth)
	label := indent + r.node.Label

	if r.node.Type == "ip" {
		return table.Row{label, "", "", "", ""}
	}

	totalStr := "-"
	usedStr := strconv.Itoa(r.node.UsedIPs)
	pctStr := "-"
	barStr := ""

	if r.node.TotalIPs > 0 {
		totalStr = strconv.Itoa(r.node.TotalIPs)
		pct := float64(r.node.UsedIPs) / float64(r.node.TotalIPs) * 100
		pctStr = fmt.Sprintf("%.0f%%", pct)
		barStr = usageBar(r.node.UsedIPs, r.node.TotalIPs, 20)
	}

	return table.Row{label, usedStr, totalStr, pctStr, barStr}
}

// usageBar builds an ASCII bar like [#####...........].
func usageBar(used, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := used * width / total
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat(".", width-filled) + "]"
}

// drillDown navigates into the selected row's child level.
func (s *IPUsageScreen) drillDown() tea.Cmd {
	cursor := s.table.Cursor()
	if cursor < 0 || cursor >= len(s.rows) {
		return nil
	}
	row := s.rows[cursor]
	node := row.node

	var nextLevel, nextParam string
	switch node.Type {
	case "client":
		nextLevel = "client"
		nextParam = "?client_id=" + strconv.FormatInt(node.ID, 10)
	case "site":
		nextLevel = "site"
		nextParam = "?site_id=" + strconv.FormatInt(node.ID, 10)
	case "vlan":
		nextLevel = "vlan"
		nextParam = "?vlan_id=" + strconv.FormatInt(node.ID, 10)
	default:
		return nil // address_block and ip nodes are not drillable
	}

	screen := NewIPUsageScreen(s.client, nextLevel, nextParam)
	return func() tea.Msg {
		return PushScreenMsg{Screen: screen}
	}
}

// ipUsageColumns returns table columns appropriate for the level.
func ipUsageColumns(level string) []table.Column {
	return []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Used", Width: 8},
		{Title: "Total", Width: 8},
		{Title: "%", Width: 6},
		{Title: "Usage", Width: 24},
	}
}

// scaleColumns proportionally scales column widths to fill the terminal.
func scaleColumns(cols []table.Column, termWidth int) []table.Column {
	if termWidth <= 0 {
		return cols
	}
	totalDef := 0
	for _, c := range cols {
		totalDef += c.Width
	}
	if totalDef == 0 {
		return cols
	}
	usable := termWidth - 2 // small margin
	scaled := make([]table.Column, len(cols))
	for i, c := range cols {
		scaled[i] = table.Column{
			Title: c.Title,
			Width: c.Width * usable / totalDef,
		}
	}
	return scaled
}
