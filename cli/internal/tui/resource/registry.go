package resource

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

// PickerEntry represents a single option in a dynamic picker.
type PickerEntry struct {
	Value string // stored value when selected
	Label string // display label
}

// Field describes a single form field for create/edit.
type Field struct {
	Key           string   // JSON key (e.g. "name")
	Label         string   // Human label (e.g. "Name")
	Required      bool
	PickerKey     string   // If set, enables FK picker using this resolver key (e.g. "clients", "sites")
	PickerOptions []string // If set, enables picker with static options (e.g. ["active", "planned"])
	PickerFunc    func(values map[string]string) []PickerEntry // Dynamic picker items (value differs from label)
	MultiSelect   bool                                        // If true, picker allows multiple selections (stored as comma-separated IDs)
	Hidden        func(values map[string]string) bool         // If returns true, field is hidden from the form
}

// Def defines how to list, display, and manage a resource in the TUI.
type Def struct {
	// Display
	Name    string // singular (e.g. "Client")
	Plural  string // plural (e.g. "Clients")
	APIPath string // e.g. "/clients"

	// Table columns and row extraction
	Columns []table.Column
	ToRow   func(raw any) table.Row // converts one item to a table row
	GetID   func(raw any) string    // extracts the ID as string

	// Form fields for create/edit
	Fields []Field

	// Defaults are pre-filled values for the create form (used in browse mode).
	// Keys match Field.Key values.
	Defaults map[string]string

	// DeriveField is called when a field value changes. Given the changed
	// field key and its new value, it returns derived values for other fields.
	// Example: typing a client name auto-generates the short_code.
	DeriveField func(key, value string) map[string]string

	// FieldHint returns an optional hint string to display next to a field label.
	// Called with the field key and current form values (key→value map).
	// Example: showing "Subnet: 10.10.0.0/24" next to IP address when a VLAN is selected.
	FieldHint func(key string, values map[string]string) string

	// PickerFilter narrows picker items based on current form values.
	// Called with the field key, current form values, and all picker items.
	// Returns a filtered subset. If nil, all items are shown.
	PickerFilter func(key string, values map[string]string, items map[int64]string) map[int64]string

	// AsyncDerive is called when a field value changes. Given the changed
	// field key and current form values, it may call the API and return
	// derived values for other fields. Called in a tea.Cmd (background).
	AsyncDerive func(client *apiclient.Client, key string, values map[string]string) map[string]string

	// PreSubmit is called before the form submits. If it returns a non-empty
	// string, the form enters a confirmation prompt showing that message.
	// The user must type "confirm" to proceed. Called with current form values.
	PreSubmit func(values map[string]string) string

	// ExportLabel downloads a label PDF for the selected item. Returns saved file path.
	ExportLabel func(client *apiclient.Client, id string) (string, error)

	// CustomActions maps a keybind (e.g. "c") to an action on the selected item.
	// Each action has a Label for the help bar and a Handler that receives the
	// selected item and returns the Def + pre-filled defaults for a form to push.
	CustomActions []CustomAction

	// API operations — filled in at registration time
	List   func(client *apiclient.Client) ([]any, error)
	Create func(client *apiclient.Client, data map[string]string) (any, error)
	Update func(client *apiclient.Client, id string, data map[string]string) (any, error)
	Delete func(client *apiclient.Client, id string) error
}

// CustomAction defines a keybind-triggered action on a table item.
type CustomAction struct {
	Key     string // keybind (e.g. "c")
	Label   string // help text (e.g. "connect")
	Handler func(raw any) (def *Def, id string, defaults map[string]string) // returns form config
}

// Registry maps menu keys to resource definitions.
var Registry = map[string]*Def{}

// Register adds a resource definition to the global registry.
func Register(key string, def *Def) {
	Registry[key] = def
}
