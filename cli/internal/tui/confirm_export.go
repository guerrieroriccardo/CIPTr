package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

type exportClientSuccessMsg struct{ path string }
type exportClientErrorMsg struct{ err error }

// ConfirmExport asks y/N before downloading the full client PDF export.
type ConfirmExport struct {
	client    *apiclient.Client
	clientID  string
	name      string
	shortCode string
	err       error
	exporting bool
	done      string
}

// NewConfirmExport creates a confirmation screen for PDF export.
func NewConfirmExport(client *apiclient.Client, clientID, name, shortCode string) ConfirmExport {
	return ConfirmExport{
		client:    client,
		clientID:  clientID,
		name:      name,
		shortCode: shortCode,
	}
}

func (ce ConfirmExport) Title() string { return "Export to PDF" }
func (ce ConfirmExport) Init() tea.Cmd { return nil }

func (ce ConfirmExport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case exportClientSuccessMsg:
		ce.exporting = false
		ce.done = msg.path
		return ce, nil

	case exportClientErrorMsg:
		ce.err = msg.err
		ce.exporting = false
		return ce, nil

	case tea.KeyMsg:
		if ce.done != "" {
			return ce, func() tea.Msg { return PopScreenMsg{} }
		}
		switch msg.String() {
		case "y", "Y":
			ce.exporting = true
			apiClient := ce.client
			clientID := ce.clientID
			shortCode := ce.shortCode
			return ce, func() tea.Msg {
				data, err := apiClient.GetRaw("/clients/" + clientID + "/export")
				if err != nil {
					return exportClientErrorMsg{err: err}
				}
				filename := fmt.Sprintf("export-%s.pdf", strings.ToLower(shortCode))
				if err := os.WriteFile(filename, data, 0644); err != nil {
					return exportClientErrorMsg{err: fmt.Errorf("save file: %w", err)}
				}
				return exportClientSuccessMsg{path: filename}
			}
		case "n", "N", "esc":
			return ce, func() tea.Msg { return PopScreenMsg{} }
		}
	}
	return ce, nil
}

func (ce ConfirmExport) View() string {
	if ce.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", ce.err)) + "\n\n" +
			HelpStyle.Render("press any key to go back")
	}
	if ce.done != "" {
		return fmt.Sprintf("PDF saved to %s\n\n%s", ce.done, HelpStyle.Render("press any key to go back"))
	}
	if ce.exporting {
		return "Exporting..."
	}

	return fmt.Sprintf(
		"%s\n\nExport all data for client %q to PDF?\n\n%s",
		TitleStyle.Render("Export to PDF"),
		ce.name,
		HelpStyle.Render("y confirm • n/esc cancel"),
	)
}
