package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

type formSavedMsg struct{}
type formErrorMsg struct{ err error }
type asyncDeriveMsg struct{ values map[string]string }

// pickerItem represents one selectable entry in the FK picker.
type pickerItem struct {
	id    string
	label string
}

// ResourceForm is a generic create/edit form for any resource.
type ResourceForm struct {
	def    *resource.Def
	client *apiclient.Client
	id     string // empty for create, set for edit
	inputs []textinput.Model
	focus  int
	err    error
	saving bool
	height int // terminal height for scrolling
	scroll int // first visible field index

	// Picker state
	picking      bool
	pickerItems  []pickerItem // all items for current picker
	pickerMatch  []pickerItem // filtered items
	pickerCursor int
	pickerFilter string
	pickerScroll int // scroll offset for picker list

	// Tracks fields manually edited by the user (not auto-derived).
	manuallyEdited map[int]bool
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
		inputs[i] = ti
	}

	form := ResourceForm{
		def:    def,
		client: client,
		id:     id,
		inputs: inputs,
	}

	// Pre-fill defaults for create mode (used in browse to scope parent IDs).
	// Also advance focus past any defaulted fields.
	if item == nil && def.Defaults != nil {
		firstNonDefault := 0
		for i, f := range def.Fields {
			if v, ok := def.Defaults[f.Key]; ok {
				inputs[i].SetValue(v)
				if i == firstNonDefault {
					firstNonDefault = i + 1
				}
			}
		}
		if firstNonDefault >= len(inputs) {
			firstNonDefault = 0
		}
		form.focus = firstNonDefault
	}

	// Set focus on the correct field.
	inputs[form.focus].Focus()

	// Pre-populate for edit mode by marshaling the item to JSON and extracting
	// values by field key. This avoids fragile column-title matching.
	if item != nil {
		raw, _ := json.Marshal(item)
		var m map[string]json.RawMessage
		_ = json.Unmarshal(raw, &m)
		for i, f := range def.Fields {
			v, ok := m[f.Key]
			if !ok || string(v) == "null" {
				continue
			}
			// Try string first, then fall back to raw number/bool.
			var s string
			if json.Unmarshal(v, &s) == nil {
				inputs[i].SetValue(s)
			} else {
				// Numeric or boolean — use raw JSON text (e.g. "42", "true").
				inputs[i].SetValue(strings.Trim(string(v), `"`))
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

// visibleFields returns how many fields fit on screen.
// Each field takes 3 lines (label + input + blank), plus title (2) + help (2).
func (f ResourceForm) visibleFields() int {
	if f.height <= 0 {
		return len(f.def.Fields)
	}
	available := f.height - 4 // title + help + error margin
	perField := 3
	n := available / perField
	if n < 1 {
		n = 1
	}
	if n > len(f.def.Fields) {
		n = len(f.def.Fields)
	}
	return n
}

func (f *ResourceForm) ensureVisible() {
	vis := f.visibleFields()
	if f.focus < f.scroll {
		f.scroll = f.focus
	}
	if f.focus >= f.scroll+vis {
		f.scroll = f.focus - vis + 1
	}
}

// openPicker populates picker items from the resolver or static options and enters picker mode.
func (f *ResourceForm) openPicker() {
	field := f.def.Fields[f.focus]

	var items []pickerItem

	if len(field.PickerOptions) > 0 {
		// Static option list — id and label are the same value.
		for _, opt := range field.PickerOptions {
			items = append(items, pickerItem{id: opt, label: opt})
		}
	} else if field.PickerKey != "" && resource.Resolve != nil {
		m := resource.Resolve.Lookup(field.PickerKey)
		if m == nil {
			return
		}
		// Apply contextual filter if defined.
		if f.def.PickerFilter != nil {
			currentValues := make(map[string]string, len(f.def.Fields))
			for j, fld := range f.def.Fields {
				currentValues[fld.Key] = f.inputs[j].Value()
			}
			m = f.def.PickerFilter(field.Key, currentValues, m)
		}
		items = make([]pickerItem, 0, len(m))
		for id, name := range m {
			items = append(items, pickerItem{
				id:    fmt.Sprintf("%d", id),
				label: name,
			})
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].label < items[j].label
		})
	} else {
		return
	}

	f.picking = true
	f.pickerItems = items
	f.pickerFilter = ""
	f.pickerCursor = 0
	f.pickerScroll = 0
	f.filterPicker()
}

// filterPicker updates pickerMatch based on pickerFilter.
func (f *ResourceForm) filterPicker() {
	if f.pickerFilter == "" {
		f.pickerMatch = f.pickerItems
	} else {
		lower := strings.ToLower(f.pickerFilter)
		f.pickerMatch = nil
		for _, item := range f.pickerItems {
			if strings.Contains(strings.ToLower(item.label), lower) {
				f.pickerMatch = append(f.pickerMatch, item)
			}
		}
	}
	if f.pickerCursor >= len(f.pickerMatch) {
		f.pickerCursor = max(0, len(f.pickerMatch)-1)
	}
	f.ensurePickerVisible()
}

// pickerVisibleRows returns how many picker items fit on screen.
func (f ResourceForm) pickerVisibleRows() int {
	if f.height <= 0 {
		return 10
	}
	// Reserve: title(1) + field label(1) + filter(1) + blank(1) + help(1) = 5
	n := f.height - 5
	if n < 3 {
		n = 3
	}
	return n
}

func (f *ResourceForm) ensurePickerVisible() {
	vis := f.pickerVisibleRows()
	if f.pickerCursor < f.pickerScroll {
		f.pickerScroll = f.pickerCursor
	}
	if f.pickerCursor >= f.pickerScroll+vis {
		f.pickerScroll = f.pickerCursor - vis + 1
	}
}

func (f ResourceForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.height = msg.Height
		f.ensureVisible()
		return f, nil

	case formSavedMsg:
		return f, func() tea.Msg { return PopScreenMsg{} }

	case formErrorMsg:
		f.err = msg.err
		f.saving = false
		return f, nil

	case asyncDeriveMsg:
		for k, v := range msg.values {
			for i, field := range f.def.Fields {
				if field.Key == k && !f.manuallyEdited[i] {
					f.inputs[i].SetValue(v)
				}
			}
		}
		return f, nil

	case tea.KeyMsg:
		// Picker mode — intercept all keys.
		if f.picking {
			return f.updatePicker(msg)
		}

		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return PopScreenMsg{} }
		case "tab", "down":
			f.focus = (f.focus + 1) % len(f.inputs)
			f.ensureVisible()
			return f, f.updateFocus()
		case "shift+tab", "up":
			f.focus = (f.focus - 1 + len(f.inputs)) % len(f.inputs)
			f.ensureVisible()
			return f, f.updateFocus()
		case "enter":
			// If current field has a picker, open it.
			if f.def.Fields[f.focus].PickerKey != "" || len(f.def.Fields[f.focus].PickerOptions) > 0 {
				f.openPicker()
				return f, nil
			}
			// If on last field, submit
			if f.focus == len(f.inputs)-1 {
				return f, f.submit()
			}
			// Otherwise move to next field
			f.focus = (f.focus + 1) % len(f.inputs)
			f.ensureVisible()
			return f, f.updateFocus()
		case "ctrl+s":
			return f, f.submit()
		}
	}

	// Update the focused input.
	oldVal := f.inputs[f.focus].Value()
	var cmd tea.Cmd
	f.inputs[f.focus], cmd = f.inputs[f.focus].Update(msg)
	newVal := f.inputs[f.focus].Value()

	// Track manual edits and apply derived field values.
	if oldVal != newVal {
		if f.manuallyEdited == nil {
			f.manuallyEdited = map[int]bool{}
		}
		f.manuallyEdited[f.focus] = true

		if f.def.DeriveField != nil {
			derived := f.def.DeriveField(f.def.Fields[f.focus].Key, newVal)
			for k, v := range derived {
				for i, field := range f.def.Fields {
					if field.Key == k && !f.manuallyEdited[i] {
						f.inputs[i].SetValue(v)
					}
				}
			}
		}
	}

	return f, cmd
}

// updatePicker handles key events while the picker is open.
func (f ResourceForm) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		f.picking = false
		return f, nil
	case "enter":
		if len(f.pickerMatch) > 0 {
			selected := f.pickerMatch[f.pickerCursor]
			changedKey := f.def.Fields[f.focus].Key
			f.inputs[f.focus].SetValue(selected.id)
			f.picking = false
			// Auto-advance to next field.
			f.focus = (f.focus + 1) % len(f.inputs)
			f.ensureVisible()
			cmds := []tea.Cmd{f.updateFocus()}
			if asyncCmd := f.fireAsyncDerive(changedKey); asyncCmd != nil {
				cmds = append(cmds, asyncCmd)
			}
			return f, tea.Batch(cmds...)
		}
		return f, nil
	case "up", "shift+tab":
		if f.pickerCursor > 0 {
			f.pickerCursor--
			f.ensurePickerVisible()
		}
		return f, nil
	case "down", "tab":
		if f.pickerCursor < len(f.pickerMatch)-1 {
			f.pickerCursor++
			f.ensurePickerVisible()
		}
		return f, nil
	case "backspace":
		if len(f.pickerFilter) > 0 {
			f.pickerFilter = f.pickerFilter[:len(f.pickerFilter)-1]
			f.filterPicker()
		}
		return f, nil
	default:
		// Type-to-filter: accept printable characters.
		r := msg.String()
		if len(r) == 1 && r[0] >= 32 && r[0] < 127 {
			f.pickerFilter += r
			f.filterPicker()
		}
		return f, nil
	}
}

var pickerSelectedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	Bold(true)

func (f ResourceForm) View() string {
	// If picker is open, render picker view instead of normal form.
	if f.picking {
		return f.viewPicker()
	}

	var b strings.Builder

	action := "Create"
	if f.id != "" {
		action = "Edit"
	}
	b.WriteString(TitleStyle.Render(action+" "+f.def.Name) + "\n")

	vis := f.visibleFields()
	end := f.scroll + vis
	if end > len(f.def.Fields) {
		end = len(f.def.Fields)
	}

	pickerHintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Collect current values for FieldHint callback.
	var currentValues map[string]string
	if f.def.FieldHint != nil {
		currentValues = make(map[string]string, len(f.def.Fields))
		for j, fld := range f.def.Fields {
			currentValues[fld.Key] = f.inputs[j].Value()
		}
	}

	for i := f.scroll; i < end; i++ {
		field := f.def.Fields[i]
		label := field.Label
		if field.Required {
			label += " *"
		}
		if field.PickerKey != "" || len(field.PickerOptions) > 0 {
			label += " " + pickerHintStyle.Render("[enter to pick]")
		}
		if f.def.FieldHint != nil {
			if hint := f.def.FieldHint(field.Key, currentValues); hint != "" {
				label += " " + pickerHintStyle.Render("["+hint+"]")
			}
		}
		// For picker fields with a value, show resolved name as display with ID hint.
		resolvedName := ""
		if field.PickerKey != "" && f.inputs[i].Value() != "" && resource.Resolve != nil {
			if m := resource.Resolve.Lookup(field.PickerKey); m != nil {
				for id, name := range m {
					if fmt.Sprintf("%d", id) == f.inputs[i].Value() {
						resolvedName = name
						break
					}
				}
			}
		}
		cursor := "  "
		if i == f.focus {
			cursor = "> "
		}
		if resolvedName != "" {
			b.WriteString(fmt.Sprintf("%s%s\n  %s (ID: %s)\n\n", cursor, label, resolvedName, f.inputs[i].Value()))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n  %s\n\n", cursor, label, f.inputs[i].View()))
		}
	}

	// Show scroll indicator if not all fields are visible.
	if len(f.def.Fields) > vis {
		b.WriteString(HelpStyle.Render(fmt.Sprintf("  [%d/%d fields]", f.focus+1, len(f.def.Fields))) + "\n")
	}

	if f.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", f.err)) + "\n\n")
	}

	if f.saving {
		b.WriteString("Saving...\n")
	}

	// Build help text — show picker hint for FK fields.
	helpParts := []string{"tab next"}
	if f.def.Fields[f.focus].PickerKey != "" || len(f.def.Fields[f.focus].PickerOptions) > 0 {
		helpParts = append(helpParts, "enter pick")
	}
	helpParts = append(helpParts, "ctrl+s save", "esc cancel")
	b.WriteString(HelpStyle.Render(strings.Join(helpParts, " • ")))

	return b.String()
}

// viewPicker renders the full-screen picker overlay.
func (f ResourceForm) viewPicker() string {
	var b strings.Builder
	field := f.def.Fields[f.focus]

	b.WriteString(TitleStyle.Render("Pick "+field.Label) + "\n")
	if f.pickerFilter != "" {
		b.WriteString(HelpStyle.Render("Filter: "+f.pickerFilter) + "\n")
	} else {
		b.WriteString(HelpStyle.Render("Type to filter...") + "\n")
	}

	if len(f.pickerMatch) == 0 {
		b.WriteString("\n  (no matches)\n")
	} else {
		vis := f.pickerVisibleRows()
		end := f.pickerScroll + vis
		if end > len(f.pickerMatch) {
			end = len(f.pickerMatch)
		}
		for i := f.pickerScroll; i < end; i++ {
			item := f.pickerMatch[i]
			line := "  " + item.label
			if i == f.pickerCursor {
				line = pickerSelectedStyle.Render("> " + item.label)
			}
			b.WriteString(line + "\n")
		}
		if len(f.pickerMatch) > vis {
			b.WriteString(HelpStyle.Render(fmt.Sprintf("  [%d/%d]", f.pickerCursor+1, len(f.pickerMatch))) + "\n")
		}
	}

	b.WriteString("\n" + HelpStyle.Render("↑↓ navigate • enter select • esc cancel"))
	return b.String()
}

// fireAsyncDerive returns a tea.Cmd that calls AsyncDerive if defined.
func (f ResourceForm) fireAsyncDerive(changedKey string) tea.Cmd {
	if f.def.AsyncDerive == nil || f.id != "" {
		return nil
	}
	values := make(map[string]string, len(f.def.Fields))
	for i, fld := range f.def.Fields {
		values[fld.Key] = f.inputs[i].Value()
	}
	def := f.def
	client := f.client
	return func() tea.Msg {
		derived := def.AsyncDerive(client, changedKey, values)
		if len(derived) == 0 {
			return nil
		}
		return asyncDeriveMsg{values: derived}
	}
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
