package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

type deleteSuccessMsg struct{}
type deleteErrorMsg struct{ err error }

// ConfirmDelete asks the user to confirm deletion of a resource.
type ConfirmDelete struct {
	def      *resource.Def
	client   *apiclient.Client
	id       string
	name     string
	err      error
	deleting bool
}

// NewConfirmDelete creates a confirmation screen.
func NewConfirmDelete(def *resource.Def, client *apiclient.Client, id, name string) ConfirmDelete {
	return ConfirmDelete{
		def:    def,
		client: client,
		id:     id,
		name:   name,
	}
}

func (cd ConfirmDelete) Title() string {
	return "Delete " + cd.def.Name
}

func (cd ConfirmDelete) Init() tea.Cmd { return nil }

func (cd ConfirmDelete) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deleteSuccessMsg:
		return cd, func() tea.Msg { return MutationPopMsg{} }

	case deleteErrorMsg:
		cd.err = msg.err
		cd.deleting = false
		return cd, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			cd.deleting = true
			def := cd.def
			client := cd.client
			id := cd.id
			return cd, func() tea.Msg {
				if err := def.Delete(client, id); err != nil {
					return deleteErrorMsg{err: err}
				}
				return deleteSuccessMsg{}
			}
		case "n", "N", "esc":
			return cd, func() tea.Msg { return PopScreenMsg{} }
		}
	}

	return cd, nil
}

func (cd ConfirmDelete) View() string {
	if cd.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", cd.err)) + "\n\n" +
			HelpStyle.Render("esc back")
	}
	if cd.deleting {
		return "Deleting..."
	}

	return fmt.Sprintf(
		"%s\n\nAre you sure you want to delete %s %q?\n\n%s",
		TitleStyle.Render("Delete "+cd.def.Name),
		cd.def.Name,
		cd.name,
		HelpStyle.Render("y confirm • n/esc cancel"),
	)
}
