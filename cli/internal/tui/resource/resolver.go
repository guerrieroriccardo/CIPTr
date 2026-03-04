package resource

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

// Resolve is the global resolver instance. Set after InitResolver completes.
var Resolve *Resolver

// Resolver caches id→name mappings for FK columns.
type Resolver struct {
	Clients        map[int64]string
	Sites          map[int64]string
	Categories     map[int64]string
	Manufacturers  map[int64]string
	Suppliers      map[int64]string
	Devices        map[int64]string
	Interfaces     map[int64]string
	Switches       map[int64]string
	PatchPanels    map[int64]string
	VLANs          map[int64]string
	AddressBlocks  map[int64]string
	DeviceModels   map[int64]string
	Locations      map[int64]string
	SwitchPorts    map[int64]string
	PatchPanelPorts map[int64]string
}

// ResolverReadyMsg is sent when all lookup data has been fetched.
type ResolverReadyMsg struct{ R *Resolver }

// InitResolver returns a tea.Cmd that fetches all lookup tables in the background.
func InitResolver(c *apiclient.Client) tea.Cmd {
	return func() tea.Msg {
		r := &Resolver{
			Clients:        make(map[int64]string),
			Sites:          make(map[int64]string),
			Categories:     make(map[int64]string),
			Manufacturers:  make(map[int64]string),
			Suppliers:      make(map[int64]string),
			Devices:        make(map[int64]string),
			Interfaces:     make(map[int64]string),
			Switches:       make(map[int64]string),
			PatchPanels:    make(map[int64]string),
			VLANs:          make(map[int64]string),
			AddressBlocks:  make(map[int64]string),
			DeviceModels:   make(map[int64]string),
			Locations:      make(map[int64]string),
			SwitchPorts:    make(map[int64]string),
			PatchPanelPorts: make(map[int64]string),
		}

		// Fetch all small lookup tables. Errors are silently ignored —
		// the resolver will just return IDs for failed lookups.
		var clients []models.Client
		if err := c.Get("/clients", &clients); err == nil {
			for _, v := range clients {
				r.Clients[v.ID] = v.Name
			}
		}

		var sites []models.Site
		if err := c.Get("/sites", &sites); err == nil {
			for _, v := range sites {
				r.Sites[v.ID] = v.Name
			}
		}

		var categories []models.Category
		if err := c.Get("/categories", &categories); err == nil {
			for _, v := range categories {
				r.Categories[v.ID] = v.Name
			}
		}

		var manufacturers []models.Manufacturer
		if err := c.Get("/manufacturers", &manufacturers); err == nil {
			for _, v := range manufacturers {
				r.Manufacturers[v.ID] = v.Name
			}
		}

		var suppliers []models.Supplier
		if err := c.Get("/suppliers", &suppliers); err == nil {
			for _, v := range suppliers {
				r.Suppliers[v.ID] = v.Name
			}
		}

		var devices []models.Device
		if err := c.Get("/devices", &devices); err == nil {
			for _, v := range devices {
				r.Devices[v.ID] = v.Hostname
			}
		}

		var ifaces []models.DeviceInterface
		if err := c.Get("/device-interfaces", &ifaces); err == nil {
			for _, v := range ifaces {
				r.Interfaces[v.ID] = v.Name
			}
		}

		var switches []models.Switch
		if err := c.Get("/switches", &switches); err == nil {
			for _, v := range switches {
				r.Switches[v.ID] = v.Name
			}
		}

		var panels []models.PatchPanel
		if err := c.Get("/patch-panels", &panels); err == nil {
			for _, v := range panels {
				r.PatchPanels[v.ID] = v.Name
			}
		}

		var vlans []models.VLAN
		if err := c.Get("/vlans", &vlans); err == nil {
			for _, v := range vlans {
				r.VLANs[v.ID] = v.Name
			}
		}

		var addressBlocks []models.AddressBlock
		if err := c.Get("/address-blocks", &addressBlocks); err == nil {
			for _, v := range addressBlocks {
				r.AddressBlocks[v.ID] = v.Network
			}
		}

		var deviceModels []models.DeviceModel
		if err := c.Get("/device-models", &deviceModels); err == nil {
			for _, v := range deviceModels {
				label := v.ModelName
				if mfr, ok := r.Manufacturers[v.ManufacturerID]; ok {
					label = mfr + " " + label
				}
				r.DeviceModels[v.ID] = label
			}
		}

		var locations []models.Location
		if err := c.Get("/locations", &locations); err == nil {
			for _, v := range locations {
				r.Locations[v.ID] = v.Name
			}
		}

		var switchPorts []models.SwitchPort
		if err := c.Get("/switch-ports", &switchPorts); err == nil {
			for _, v := range switchPorts {
				label := fmt.Sprintf("Port %d", v.PortNumber)
				if v.PortLabel != nil && *v.PortLabel != "" {
					label = *v.PortLabel
				}
				r.SwitchPorts[v.ID] = label
			}
		}

		var ppPorts []models.PatchPanelPort
		if err := c.Get("/patch-panel-ports", &ppPorts); err == nil {
			for _, v := range ppPorts {
				label := fmt.Sprintf("Port %d", v.PortNumber)
				if v.PortLabel != nil && *v.PortLabel != "" {
					label = *v.PortLabel
				}
				r.PatchPanelPorts[v.ID] = label
			}
		}

		return ResolverReadyMsg{R: r}
	}
}

// Name helpers — return the resolved name or fall back to the numeric ID.

func ResolveName(m map[int64]string, id int64) string {
	if Resolve != nil {
		if name, ok := m[id]; ok {
			return name
		}
	}
	return fmt.Sprintf("%d", id)
}

func ResolveOptionalName(m map[int64]string, id *int64) string {
	if id == nil {
		return ""
	}
	return ResolveName(m, *id)
}

// Convenience accessors that use the global Resolve instance.

func ClientName(id int64) string       { return lookupName(func() map[int64]string { return safeLookup().Clients }, id) }
func SiteName(id int64) string         { return lookupName(func() map[int64]string { return safeLookup().Sites }, id) }
func CategoryName(id int64) string     { return lookupName(func() map[int64]string { return safeLookup().Categories }, id) }
func ManufacturerName(id int64) string { return lookupName(func() map[int64]string { return safeLookup().Manufacturers }, id) }
func SupplierName(id *int64) string    { return lookupOptional(func() map[int64]string { return safeLookup().Suppliers }, id) }
func DeviceName(id int64) string       { return lookupName(func() map[int64]string { return safeLookup().Devices }, id) }
func InterfaceName(id int64) string    { return lookupName(func() map[int64]string { return safeLookup().Interfaces }, id) }
func SwitchName(id int64) string       { return lookupName(func() map[int64]string { return safeLookup().Switches }, id) }
func PatchPanelName(id int64) string      { return lookupName(func() map[int64]string { return safeLookup().PatchPanels }, id) }
func VLANName(id *int64) string           { return lookupOptional(func() map[int64]string { return safeLookup().VLANs }, id) }
func AddressBlockName(id *int64) string   { return lookupOptional(func() map[int64]string { return safeLookup().AddressBlocks }, id) }
func DeviceModelName(id *int64) string    { return lookupOptional(func() map[int64]string { return safeLookup().DeviceModels }, id) }
func LocationName(id *int64) string       { return lookupOptional(func() map[int64]string { return safeLookup().Locations }, id) }

// Lookup returns the resolver map for a given key string (e.g. "clients", "sites").
// Returns nil if the key is unknown or the resolver is not ready.
func (r *Resolver) Lookup(key string) map[int64]string {
	switch key {
	case "clients":
		return r.Clients
	case "sites":
		return r.Sites
	case "categories":
		return r.Categories
	case "manufacturers":
		return r.Manufacturers
	case "suppliers":
		return r.Suppliers
	case "devices":
		return r.Devices
	case "interfaces":
		return r.Interfaces
	case "switches":
		return r.Switches
	case "patch_panels":
		return r.PatchPanels
	case "vlans":
		return r.VLANs
	case "address_blocks":
		return r.AddressBlocks
	case "device_models":
		return r.DeviceModels
	case "locations":
		return r.Locations
	case "switch_ports":
		return r.SwitchPorts
	case "patch_panel_ports":
		return r.PatchPanelPorts
	}
	return nil
}

func safeLookup() *Resolver {
	if Resolve == nil {
		return &Resolver{}
	}
	return Resolve
}

func lookupName(getMap func() map[int64]string, id int64) string {
	m := getMap()
	if m != nil {
		if name, ok := m[id]; ok {
			return name
		}
	}
	return fmt.Sprintf("%d", id)
}

func lookupOptional(getMap func() map[int64]string, id *int64) string {
	if id == nil {
		return ""
	}
	return lookupName(getMap, *id)
}
