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
type bulkDoneMsg struct{ errors []error }

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
	width  int // terminal width for input sizing
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

	// Multi-select picker state
	multiSelected map[string]bool // set of selected IDs during multi-pick

	// Bulk edit mode
	bulkItems   []any             // when non-nil, form is in bulk-edit mode
	bulkContext map[string]string // common values from selected items (for picker scoping)
	bulkDone    int               // items updated so far
	bulkErrs    []error           // errors collected during bulk update

	// Confirmation prompt (shared by bulk edit and pre-submit)
	confirming   bool            // waiting for user to type confirmation word
	confirmInput textinput.Model // text input for confirmation
	confirmMsg   string          // message to display above prompt
	confirmWord  string          // word to type (e.g. "apply", "confirm")
	confirmCmd   func() tea.Cmd  // action to run after confirmation
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
				// Check for JSON array (e.g. [1,2,3]) — format as comma-separated.
				var nums []json.Number
				if json.Unmarshal(v, &nums) == nil {
					parts := make([]string, len(nums))
					for j, n := range nums {
						parts[j] = n.String()
					}
					inputs[i].SetValue(strings.Join(parts, ","))
				} else {
					// Numeric or boolean — use raw JSON text (e.g. "42", "true").
					inputs[i].SetValue(strings.Trim(string(v), `"`))
				}
			}
		}
	}

	return form
}

// NewBulkEditForm creates a form in bulk-edit mode. All fields start empty;
// only non-empty fields are applied to every selected item on submit.
// Common values from selected items are stored as context for picker filtering.
func NewBulkEditForm(def *resource.Def, client *apiclient.Client, items []any) ResourceForm {
	inputs := make([]textinput.Model, len(def.Fields))
	for i, f := range def.Fields {
		ti := textinput.New()
		ti.Placeholder = f.Label
		ti.CharLimit = 256
		ti.Width = 40
		inputs[i] = ti
	}
	inputs[0].Focus()

	// Extract common values from selected items for picker scoping.
	// If all items share a value for a field, use it as context.
	bulkCtx := map[string]string{}
	if len(items) > 0 {
		allData := make([]map[string]string, len(items))
		for i, item := range items {
			allData[i] = extractItemData(def, item)
		}
		for _, field := range def.Fields {
			val := allData[0][field.Key]
			common := true
			for _, d := range allData[1:] {
				if d[field.Key] != val {
					common = false
					break
				}
			}
			if common && val != "" {
				bulkCtx[field.Key] = val
			}
		}
	}

	return ResourceForm{
		def:         def,
		client:      client,
		inputs:      inputs,
		bulkItems:   items,
		bulkContext:  bulkCtx,
	}
}

func (f ResourceForm) Title() string {
	if f.bulkItems != nil {
		return fmt.Sprintf("Bulk Edit %d %s", len(f.bulkItems), f.def.Plural)
	}
	if f.id == "" {
		return "New " + f.def.Name
	}
	return "Edit " + f.def.Name
}

func (f ResourceForm) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	// Fire AsyncDerive for any pre-filled defaults so derived fields get populated.
	if f.id == "" && f.def.AsyncDerive != nil && f.def.Defaults != nil {
		for _, field := range f.def.Fields {
			if _, ok := f.def.Defaults[field.Key]; ok {
				if cmd := f.fireAsyncDerive(field.Key); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}
	return tea.Batch(cmds...)
}

// activeFieldIndices returns the indices of fields currently visible (not hidden).
func (f ResourceForm) activeFieldIndices() []int {
	values := f.currentValues()
	var indices []int
	for i, field := range f.def.Fields {
		if field.Hidden == nil || !field.Hidden(values) {
			indices = append(indices, i)
		}
	}
	return indices
}

// focusPosition returns the position of f.focus within the active field list.
func (f ResourceForm) focusPosition(active []int) int {
	for i, idx := range active {
		if idx == f.focus {
			return i
		}
	}
	return 0
}

// visibleFields returns how many fields fit on screen.
// Each field takes 3 lines (label + input + blank), plus title (2) + help (2).
func (f ResourceForm) visibleFields() int {
	active := f.activeFieldIndices()
	n := len(active)
	if f.height > 0 {
		available := f.height - 4 // title + help + error margin
		perField := 3
		max := available / perField
		if max < 1 {
			max = 1
		}
		if n > max {
			n = max
		}
	}
	return n
}

func (f *ResourceForm) ensureVisible() {
	active := f.activeFieldIndices()
	if len(active) == 0 {
		return
	}
	vis := f.visibleFields()
	pos := f.focusPosition(active)
	if pos < f.scroll {
		f.scroll = pos
	}
	if pos >= f.scroll+vis {
		f.scroll = pos - vis + 1
	}
}

// currentValues returns the current form values, using bulkContext as fallback
// for empty fields in bulk-edit mode (so pickers are properly scoped).
func (f *ResourceForm) currentValues() map[string]string {
	values := make(map[string]string, len(f.def.Fields))
	for j, fld := range f.def.Fields {
		v := f.inputs[j].Value()
		if v == "" && f.bulkContext != nil {
			v = f.bulkContext[fld.Key]
		}
		values[fld.Key] = v
	}
	return values
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
			currentValues := f.currentValues()
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
	} else if field.PickerFunc != nil {
		currentValues := f.currentValues()
		entries := field.PickerFunc(currentValues)
		for _, e := range entries {
			items = append(items, pickerItem{id: e.Value, label: e.Label})
		}
	} else {
		return
	}

	// Add a "None" option at the top for optional single-select fields to allow clearing the value.
	if !field.Required && !field.MultiSelect {
		items = append([]pickerItem{{id: "", label: "(None)"}}, items...)
	}

	f.picking = true
	f.pickerItems = items
	f.pickerFilter = ""
	f.pickerCursor = 0
	f.pickerScroll = 0

	// Initialize multi-select state from current value.
	if field.MultiSelect {
		f.multiSelected = map[string]bool{}
		for _, id := range strings.Split(f.inputs[f.focus].Value(), ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				f.multiSelected[id] = true
			}
		}
	} else {
		f.multiSelected = nil
	}

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
		f.width = msg.Width
		f.height = msg.Height
		inputWidth := min(60, msg.Width-10)
		if inputWidth < 20 {
			inputWidth = 20
		}
		for i := range f.inputs {
			f.inputs[i].Width = inputWidth
		}
		f.ensureVisible()
		return f, nil

	case formSavedMsg:
		return f, func() tea.Msg { return MutationPopMsg{} }

	case bulkDoneMsg:
		if len(msg.errors) > 0 {
			f.err = fmt.Errorf("%d/%d updates failed: %v", len(msg.errors), len(f.bulkItems), msg.errors[0])
			f.saving = false
			return f, nil
		}
		return f, func() tea.Msg { return MutationPopMsg{} }

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
		// Confirmation mode — intercept all keys.
		if f.confirming {
			switch msg.String() {
			case "esc":
				f.confirming = false
				return f, nil
			case "enter":
				if f.confirmInput.Value() == f.confirmWord {
					f.confirming = false
					f.saving = true
					return f, f.confirmCmd()
				}
				f.err = fmt.Errorf("type '%s' to confirm or esc to cancel", f.confirmWord)
				return f, nil
			default:
				var cmd tea.Cmd
				f.confirmInput, cmd = f.confirmInput.Update(msg)
				return f, cmd
			}
		}

		// Picker mode — intercept all keys.
		if f.picking {
			return f.updatePicker(msg)
		}

		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return PopScreenMsg{} }
		case "tab", "down":
			active := f.activeFieldIndices()
			if len(active) > 0 {
				pos := (f.focusPosition(active) + 1) % len(active)
				f.focus = active[pos]
			}
			f.ensureVisible()
			return f, f.updateFocus()
		case "shift+tab", "up":
			active := f.activeFieldIndices()
			if len(active) > 0 {
				pos := (f.focusPosition(active) - 1 + len(active)) % len(active)
				f.focus = active[pos]
			}
			f.ensureVisible()
			return f, f.updateFocus()
		case "enter":
			// If current field has a picker, open it.
			if f.def.Fields[f.focus].PickerKey != "" || len(f.def.Fields[f.focus].PickerOptions) > 0 || f.def.Fields[f.focus].PickerFunc != nil || f.def.Fields[f.focus].MultiSelect {
				f.openPicker()
				return f, nil
			}
			// If on last active field, submit
			active := f.activeFieldIndices()
			if len(active) == 0 || f.focusPosition(active) == len(active)-1 {
				return f, f.submit()
			}
			// Otherwise move to next active field
			pos := (f.focusPosition(active) + 1) % len(active)
			f.focus = active[pos]
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
	case "ctrl+n":
		// Create new entry for this picker's resource inline.
		field := f.def.Fields[f.focus]
		if field.PickerKey == "" {
			return f, nil
		}
		subDef, ok := resource.Registry[field.PickerKey]
		if !ok || subDef.Create == nil {
			return f, nil
		}
		// Inherit site_id from the parent form so the sub-form is pre-scoped.
		scopedDef := *subDef
		for i, fld := range f.def.Fields {
			if fld.Key == "site_id" {
				if v := f.inputs[i].Value(); v != "" {
					scopedDef.Defaults = map[string]string{"site_id": v}
				}
				break
			}
		}
		f.picking = false
		return f, func() tea.Msg {
			return PushScreenMsg{Screen: NewResourceForm(&scopedDef, f.client, "", nil)}
		}
	case " ":
		// Multi-select: toggle selection with space.
		if f.multiSelected != nil && len(f.pickerMatch) > 0 {
			item := f.pickerMatch[f.pickerCursor]
			if f.multiSelected[item.id] {
				delete(f.multiSelected, item.id)
			} else {
				f.multiSelected[item.id] = true
			}
			return f, nil
		}
		return f, nil
	case "enter":
		if f.multiSelected != nil {
			// Multi-select: confirm selection.
			var ids []string
			// Preserve order from pickerItems (not pickerMatch).
			for _, item := range f.pickerItems {
				if f.multiSelected[item.id] {
					ids = append(ids, item.id)
				}
			}
			changedKey := f.def.Fields[f.focus].Key
			f.inputs[f.focus].SetValue(strings.Join(ids, ","))
			f.picking = false
			f.multiSelected = nil
			f.focus = (f.focus + 1) % len(f.inputs)
			f.ensureVisible()
			cmds := []tea.Cmd{f.updateFocus()}
			if asyncCmd := f.fireAsyncDerive(changedKey); asyncCmd != nil {
				cmds = append(cmds, asyncCmd)
			}
			return f, tea.Batch(cmds...)
		}
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

	if f.bulkItems != nil {
		b.WriteString(TitleStyle.Render(fmt.Sprintf("Bulk Edit %d %s", len(f.bulkItems), f.def.Plural)) + "\n")
		b.WriteString(HelpStyle.Render("Empty fields will not be changed") + "\n")
	} else {
		action := "Create"
		if f.id != "" {
			action = "Edit"
		}
		b.WriteString(TitleStyle.Render(action+" "+f.def.Name) + "\n")
	}

	active := f.activeFieldIndices()
	vis := f.visibleFields()
	if f.scroll > len(active) {
		f.scroll = 0
	}
	end := f.scroll + vis
	if end > len(active) {
		end = len(active)
	}

	pickerHintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Collect current values for FieldHint callback.
	var currentValues map[string]string
	if f.def.FieldHint != nil {
		currentValues = f.currentValues()
	}

	for _, i := range active[f.scroll:end] {
		field := f.def.Fields[i]
		label := field.Label
		if field.Required {
			label += " *"
		}
		if field.PickerKey != "" || len(field.PickerOptions) > 0 || field.PickerFunc != nil || field.MultiSelect {
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
				if field.MultiSelect {
					// Multi-select: resolve all comma-separated IDs.
					var names []string
					for _, part := range strings.Split(f.inputs[i].Value(), ",") {
						part = strings.TrimSpace(part)
						if part == "" {
							continue
						}
						for id, name := range m {
							if fmt.Sprintf("%d", id) == part {
								names = append(names, name)
								break
							}
						}
					}
					if len(names) > 0 {
						resolvedName = strings.Join(names, ", ")
					}
				} else {
					for id, name := range m {
						if fmt.Sprintf("%d", id) == f.inputs[i].Value() {
							resolvedName = name
							break
						}
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

	// Show scroll indicator if not all active fields are visible.
	if len(active) > vis {
		b.WriteString(HelpStyle.Render(fmt.Sprintf("  [%d/%d fields]", f.focusPosition(active)+1, len(active))) + "\n")
	}

	if f.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", f.err)) + "\n\n")
	}

	if f.confirming {
		b.WriteString(fmt.Sprintf("\n%s Type '%s' to confirm:\n", f.confirmMsg, f.confirmWord))
		b.WriteString("  " + f.confirmInput.View() + "\n")
	}

	if f.saving {
		if f.bulkItems != nil {
			b.WriteString(fmt.Sprintf("Updating %d items...\n", len(f.bulkItems)))
		} else {
			b.WriteString("Saving...\n")
		}
	}

	// Build help text — show picker hint for FK fields.
	helpParts := []string{"tab next"}
	if f.def.Fields[f.focus].PickerKey != "" || len(f.def.Fields[f.focus].PickerOptions) > 0 || f.def.Fields[f.focus].PickerFunc != nil || f.def.Fields[f.focus].MultiSelect {
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
			prefix := "  "
			if f.multiSelected != nil {
				if f.multiSelected[item.id] {
					prefix = "[x] "
				} else {
					prefix = "[ ] "
				}
			}
			if i == f.pickerCursor {
				line := pickerSelectedStyle.Render("> " + prefix + item.label)
				b.WriteString(line + "\n")
			} else {
				b.WriteString("  " + prefix + item.label + "\n")
			}
		}
		if len(f.pickerMatch) > vis {
			b.WriteString(HelpStyle.Render(fmt.Sprintf("  [%d/%d]", f.pickerCursor+1, len(f.pickerMatch))) + "\n")
		}
	}

	var helpParts []string
	if f.multiSelected != nil {
		helpParts = []string{"↑↓ navigate", "space toggle", "enter confirm", "esc cancel"}
	} else {
		helpParts = []string{"↑↓ navigate", "enter select"}
		if field.PickerKey != "" {
			if subDef, ok := resource.Registry[field.PickerKey]; ok && subDef.Create != nil {
				helpParts = append(helpParts, "ctrl+n create new")
			}
		}
		helpParts = append(helpParts, "esc cancel")
	}
	b.WriteString("\n" + HelpStyle.Render(strings.Join(helpParts, " • ")))
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

func (f *ResourceForm) enterConfirm(msg, word string, action func() tea.Cmd) tea.Cmd {
	f.confirming = true
	f.confirmMsg = msg
	f.confirmWord = word
	f.confirmCmd = action
	f.err = nil
	f.confirmInput = textinput.New()
	f.confirmInput.Placeholder = "type '" + word + "' to confirm"
	f.confirmInput.CharLimit = len(word) + 5
	f.confirmInput.Width = 30
	f.confirmInput.Focus()
	return textinput.Blink
}

func (f *ResourceForm) enterBulkConfirm() tea.Cmd {
	// Validate that at least one field is non-empty before asking confirmation.
	hasChange := false
	for i := range f.def.Fields {
		if strings.TrimSpace(f.inputs[i].Value()) != "" {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return func() tea.Msg {
			return formErrorMsg{err: fmt.Errorf("no fields to update")}
		}
	}
	return f.enterConfirm(
		fmt.Sprintf("Apply changes to %d items?", len(f.bulkItems)),
		"apply",
		func() tea.Cmd { return f.submitBulk() },
	)
}

func (f *ResourceForm) submit() tea.Cmd {
	if f.bulkItems != nil {
		return f.enterBulkConfirm()
	}

	active := f.activeFieldIndices()
	activeSet := make(map[int]bool, len(active))
	for _, i := range active {
		activeSet[i] = true
	}

	// Validate required fields (only active ones).
	for i, field := range f.def.Fields {
		if activeSet[i] && field.Required && strings.TrimSpace(f.inputs[i].Value()) == "" {
			return func() tea.Msg {
				return formErrorMsg{err: fmt.Errorf("%s is required", field.Label)}
			}
		}
	}

	// Collect data. Hidden fields are sent as empty so the backend clears them.
	data := make(map[string]string)
	for i, field := range f.def.Fields {
		if activeSet[i] {
			data[field.Key] = f.inputs[i].Value()
		} else {
			data[field.Key] = ""
		}
	}

	// Check PreSubmit for confirmation prompt.
	if f.def.PreSubmit != nil {
		if msg := f.def.PreSubmit(data); msg != "" {
			return f.enterConfirm(msg, "confirm", func() tea.Cmd {
				return f.doSubmit(data)
			})
		}
	}

	return f.doSubmit(data)
}

func (f ResourceForm) doSubmit(data map[string]string) tea.Cmd {
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

// submitBulk applies non-empty fields to all selected items.
func (f ResourceForm) submitBulk() tea.Cmd {
	// Collect only non-empty fields as the changes to apply.
	changes := make(map[string]string)
	for i, field := range f.def.Fields {
		v := strings.TrimSpace(f.inputs[i].Value())
		if v != "" {
			changes[field.Key] = f.inputs[i].Value()
		}
	}
	if len(changes) == 0 {
		return func() tea.Msg {
			return formErrorMsg{err: fmt.Errorf("no fields to update")}
		}
	}

	def := f.def
	client := f.client
	items := f.bulkItems

	return func() tea.Msg {
		var errs []error
		for _, item := range items {
			// Extract current values from the item.
			data := extractItemData(def, item)
			// Overlay the user's changes.
			for k, v := range changes {
				data[k] = v
			}
			id := def.GetID(item)
			if _, err := def.Update(client, id, data); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", id, err))
			}
		}
		return bulkDoneMsg{errors: errs}
	}
}

// extractItemData marshals an item to JSON and extracts field values as strings,
// using the same logic as form pre-population.
func extractItemData(def *resource.Def, item any) map[string]string {
	data := make(map[string]string)
	raw, _ := json.Marshal(item)
	var m map[string]json.RawMessage
	_ = json.Unmarshal(raw, &m)
	for _, field := range def.Fields {
		v, ok := m[field.Key]
		if !ok || string(v) == "null" {
			data[field.Key] = ""
			continue
		}
		var s string
		if json.Unmarshal(v, &s) == nil {
			data[field.Key] = s
		} else {
			var nums []json.Number
			if json.Unmarshal(v, &nums) == nil {
				parts := make([]string, len(nums))
				for j, n := range nums {
					parts[j] = n.String()
				}
				data[field.Key] = strings.Join(parts, ",")
			} else {
				data[field.Key] = strings.Trim(string(v), `"`)
			}
		}
	}
	return data
}
