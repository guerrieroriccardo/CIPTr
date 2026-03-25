package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

type deleteSuccessMsg struct{}
type deleteErrorMsg struct{ err error }

// ConfirmDelete asks the user to type the resource name to confirm deletion.
type ConfirmDelete struct {
	def      *resource.Def
	client   *apiclient.Client
	id       string
	name     string
	input    textinput.Model
	err      error
	deleting bool
}

// NewConfirmDelete creates a confirmation screen.
func NewConfirmDelete(def *resource.Def, client *apiclient.Client, id, name string) ConfirmDelete {
	ti := textinput.New()
	ti.Placeholder = name
	ti.Focus()
	ti.CharLimit = 256

	return ConfirmDelete{
		def:    def,
		client: client,
		id:     id,
		name:   name,
		input:  ti,
	}
}

func (cd ConfirmDelete) Title() string {
	return "Delete " + cd.def.Name
}

func (cd ConfirmDelete) Init() tea.Cmd {
	return textinput.Blink
}

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
		case "esc":
			return cd, func() tea.Msg { return PopScreenMsg{} }
		case "enter":
			if strings.EqualFold(strings.TrimSpace(cd.input.Value()), cd.name) {
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
			}
			cd.err = fmt.Errorf("name does not match — type %q to confirm", cd.name)
			return cd, nil
		}
	}

	var cmd tea.Cmd
	cd.input, cmd = cd.input.Update(msg)
	// Clear error on new input
	if cd.err != nil {
		cd.err = nil
	}
	return cd, cmd
}

func (cd ConfirmDelete) View() string {
	if cd.deleting {
		return "Deleting..."
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("Delete "+cd.def.Name) + "\n\n")
	b.WriteString(fmt.Sprintf("Are you sure you want to delete %s %q?\n", cd.def.Name, cd.name))
	b.WriteString(fmt.Sprintf("Type %q to confirm:\n\n", cd.name))
	b.WriteString(cd.input.View())
	b.WriteString("\n")

	if cd.err != nil {
		b.WriteString("\n" + ErrorStyle.Render(cd.err.Error()) + "\n")
	}

	b.WriteString("\n" + HelpStyle.Render("enter confirm • esc cancel"))
	return b.String()
}
