package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

type formSavedMsg struct{}
type formErrorMsg struct{ err error }

// ResourceForm is a generic create/edit form for any resource.
type ResourceForm struct {
	def    *resource.Def
	client *apiclient.Client
	id     string // empty for create, set for edit
	inputs []textinput.Model
	focus  int
	err    error
	saving bool
}

// NewResourceForm creates a form screen. If id is non-empty and item is
// provided, fields are pre-populated for editing.
func NewResourceForm(def *resource.Def, client *apiclient.Client, id string, item any) ResourceForm {
	inputs := make([]textinput.Model, len(def.Fields))
	for i, f := range def.Fields {
		ti := textinput.New()
		ti.Placeholder = f.Label
		if f.Required {
			ti.Placeholder += " *"
		}
		ti.CharLimit = 256
		ti.Width = 40
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	form := ResourceForm{
		def:    def,
		client: client,
		id:     id,
		inputs: inputs,
	}

	// Pre-fill defaults for create mode (used in browse to scope parent IDs).
	if item == nil && def.Defaults != nil {
		for i, f := range def.Fields {
			if v, ok := def.Defaults[f.Key]; ok {
				inputs[i].SetValue(v)
			}
		}
	}

	// Pre-populate for edit mode by reading current row values.
	if item != nil {
		row := def.ToRow(item)
		// Row has ID as first column, fields start from index 1.
		for i, f := range def.Fields {
			// Find matching column by checking field key against column titles.
			for colIdx, col := range def.Columns {
				if strings.EqualFold(col.Title, f.Label) || strings.EqualFold(col.Title, f.Key) {
					if colIdx < len(row) {
						inputs[i].SetValue(row[colIdx])
					}
					break
				}
			}
		}
	}

	return form
}

func (f ResourceForm) Title() string {
	if f.id == "" {
		return "New " + f.def.Name
	}
	return "Edit " + f.def.Name
}

func (f ResourceForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f ResourceForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case formSavedMsg:
		return f, func() tea.Msg { return PopScreenMsg{} }

	case formErrorMsg:
		f.err = msg.err
		f.saving = false
		return f, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			f.focus = (f.focus + 1) % len(f.inputs)
			return f, f.updateFocus()
		case "shift+tab", "up":
			f.focus = (f.focus - 1 + len(f.inputs)) % len(f.inputs)
			return f, f.updateFocus()
		case "enter":
			// If on last field, submit
			if f.focus == len(f.inputs)-1 {
				return f, f.submit()
			}
			// Otherwise move to next field
			f.focus = (f.focus + 1) % len(f.inputs)
			return f, f.updateFocus()
		case "ctrl+s":
			return f, f.submit()
		}
	}

	// Update the focused input.
	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	return f, cmd
}

func (f ResourceForm) View() string {
	var b strings.Builder

	action := "Create"
	if f.id != "" {
		action = "Edit"
	}
	b.WriteString(TitleStyle.Render(action+" "+f.def.Name) + "\n")

	for i, field := range f.def.Fields {
		label := field.Label
		if field.Required {
			label += " *"
		}
		cursor := "  "
		if i == f.focus {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n%s  %s\n\n", cursor, label, "  ", f.inputs[i].View()))
	}

	if f.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", f.err)) + "\n\n")
	}

	if f.saving {
		b.WriteString("Saving...\n")
	}

	b.WriteString(HelpStyle.Render("tab next • ctrl+s save • esc cancel"))

	return b.String()
}

func (f ResourceForm) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(f.inputs))
	for i := range f.inputs {
		if i == f.focus {
			cmds[i] = f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (f ResourceForm) submit() tea.Cmd {
	// Validate required fields.
	for i, field := range f.def.Fields {
		if field.Required && strings.TrimSpace(f.inputs[i].Value()) == "" {
			return func() tea.Msg {
				return formErrorMsg{err: fmt.Errorf("%s is required", field.Label)}
			}
		}
	}

	// Collect data.
	data := make(map[string]string)
	for i, field := range f.def.Fields {
		data[field.Key] = f.inputs[i].Value()
	}

	def := f.def
	client := f.client
	id := f.id

	return func() tea.Msg {
		var err error
		if id == "" {
			_, err = def.Create(client, data)
		} else {
			_, err = def.Update(client, id, data)
		}
		if err != nil {
			return formErrorMsg{err: err}
		}
		return formSavedMsg{}
	}
}
