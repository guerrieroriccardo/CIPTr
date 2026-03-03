package resource

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

// Field describes a single form field for create/edit.
type Field struct {
	Key       string // JSON key (e.g. "name")
	Label     string // Human label (e.g. "Name")
	Required  bool
	PickerKey string // If set, enables FK picker using this resolver key (e.g. "clients", "sites")
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

	// API operations — filled in at registration time
	List   func(client *apiclient.Client) ([]any, error)
	Create func(client *apiclient.Client, data map[string]string) (any, error)
	Update func(client *apiclient.Client, id string, data map[string]string) (any, error)
	Delete func(client *apiclient.Client, id string) error
}

// Registry maps menu keys to resource definitions.
var Registry = map[string]*Def{}

// Register adds a resource definition to the global registry.
func Register(key string, def *Def) {
	Registry[key] = def
}
