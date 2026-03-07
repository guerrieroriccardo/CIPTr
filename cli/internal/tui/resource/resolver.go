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
	SwitchPorts     map[int64]string
	PatchPanelPorts map[int64]string
	VLANSubnets     map[int64]string // VLAN ID → subnet CIDR (for hints)

	// Reverse lookups for contextual filtering.
	InterfaceSite      map[int64]int64  // interface ID → site ID (via device)
	InterfaceDevice    map[int64]int64  // interface ID → device ID
	VLANSite           map[int64]int64  // VLAN ID → site ID
	DeviceSite         map[int64]int64  // device ID → site ID
	AddressBlockSite   map[int64]int64  // address block ID → site ID
	SiteClient         map[int64]int64  // site ID → client ID
	ClientShortCode    map[int64]string // client ID → short code
	SwitchSite         map[int64]int64  // switch ID → site ID
	SwitchPortSwitch   map[int64]int64  // switch port ID → switch ID
	PatchPanelSite     map[int64]int64  // patch panel ID → site ID
	PatchPanelPortPanel map[int64]int64 // patch panel port ID → panel ID
	LocationSite        map[int64]int64  // location ID → site ID
	DeviceIPs           map[int64]string // device IP ID → "10.0.0.1 (hostname - eth0)"
	DeviceIPVLAN        map[int64]int64  // device IP ID → VLAN ID
	DeviceIPInterface   map[int64]int64  // device IP ID → interface ID
	DeviceModelCategory map[int64]int64  // device model ID → category ID
	InterfaceMAC        map[int64]string // interface ID → MAC address
	UsedMACs            map[string]bool  // MAC addresses already restricted on a switch port
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
			SwitchPorts:     make(map[int64]string),
			PatchPanelPorts: make(map[int64]string),
			VLANSubnets:     make(map[int64]string),
			InterfaceSite:      make(map[int64]int64),
			InterfaceDevice:    make(map[int64]int64),
			VLANSite:           make(map[int64]int64),
			DeviceSite:         make(map[int64]int64),
			AddressBlockSite:   make(map[int64]int64),
			SiteClient:         make(map[int64]int64),
			ClientShortCode:    make(map[int64]string),
			SwitchSite:         make(map[int64]int64),
			SwitchPortSwitch:   make(map[int64]int64),
			PatchPanelSite:     make(map[int64]int64),
			PatchPanelPortPanel: make(map[int64]int64),
			LocationSite:        make(map[int64]int64),
			DeviceIPs:           make(map[int64]string),
			DeviceIPVLAN:        make(map[int64]int64),
			DeviceIPInterface:   make(map[int64]int64),
			DeviceModelCategory: make(map[int64]int64),
			InterfaceMAC:        make(map[int64]string),
			UsedMACs:            make(map[string]bool),
		}

		// Fetch all small lookup tables. Errors are silently ignored —
		// the resolver will just return IDs for failed lookups.
		var clients []models.Client
		if err := c.Get("/clients", &clients); err == nil {
			for _, v := range clients {
				r.Clients[v.ID] = v.Name
				r.ClientShortCode[v.ID] = v.ShortCode
			}
		}

		var sites []models.Site
		if err := c.Get("/sites", &sites); err == nil {
			for _, v := range sites {
				r.Sites[v.ID] = v.Name
				r.SiteClient[v.ID] = v.ClientID
			}
		}

		var categories []models.Category
		if err := c.Get("/categories", &categories); err == nil {
			for _, v := range categories {
				label := v.ShortCode + " - " + v.Name
				r.Categories[v.ID] = label
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
				r.DeviceSite[v.ID] = v.SiteID
			}
		}

		var ifaces []models.DeviceInterface
		if err := c.Get("/device-interfaces", &ifaces); err == nil {
			for _, v := range ifaces {
				// Build label: CLIENTSHORTCODE-SITE-HOSTNAME - ifaceName
				prefix := ""
				if siteID, ok := r.DeviceSite[v.DeviceID]; ok {
					r.InterfaceSite[v.ID] = siteID
					if clientID, ok2 := r.SiteClient[siteID]; ok2 {
						if code, ok3 := r.ClientShortCode[clientID]; ok3 {
							prefix = code
						}
					}
					if siteName, ok2 := r.Sites[siteID]; ok2 {
						if prefix != "" {
							prefix += "-"
						}
						prefix += siteName
					}
				}
				if hostname, ok := r.Devices[v.DeviceID]; ok {
					if prefix != "" {
						prefix += "-"
					}
					prefix += hostname
				}
				label := v.Name
				if prefix != "" {
					label = prefix + " - " + label
				}
				r.Interfaces[v.ID] = label
				r.InterfaceDevice[v.ID] = v.DeviceID
				if v.MacAddress != nil && *v.MacAddress != "" {
					r.InterfaceMAC[v.ID] = *v.MacAddress
				}
			}
		}

		var switches []models.Switch
		if err := c.Get("/switches", &switches); err == nil {
			for _, v := range switches {
				r.Switches[v.ID] = v.Hostname
				r.SwitchSite[v.ID] = v.SiteID
			}
		}

		var panels []models.PatchPanel
		if err := c.Get("/patch-panels", &panels); err == nil {
			for _, v := range panels {
				r.PatchPanels[v.ID] = v.Name
				r.PatchPanelSite[v.ID] = v.SiteID
			}
		}

		var vlans []models.VLAN
		if err := c.Get("/vlans", &vlans); err == nil {
			for _, v := range vlans {
				r.VLANs[v.ID] = v.Name
				r.VLANSite[v.ID] = v.SiteID
				if v.Subnet != nil && *v.Subnet != "" {
					r.VLANSubnets[v.ID] = *v.Subnet
				}
			}
		}

		var addressBlocks []models.AddressBlock
		if err := c.Get("/address-blocks", &addressBlocks); err == nil {
			for _, v := range addressBlocks {
				r.AddressBlocks[v.ID] = v.Network
				r.AddressBlockSite[v.ID] = v.SiteID
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
				r.DeviceModelCategory[v.ID] = v.CategoryID
			}
		}

		var locations []models.Location
		if err := c.Get("/locations", &locations); err == nil {
			for _, v := range locations {
				r.Locations[v.ID] = v.Name
				r.LocationSite[v.ID] = v.SiteID
			}
		}

		var switchPorts []models.SwitchPort
		if err := c.Get("/switch-ports", &switchPorts); err == nil {
			for _, v := range switchPorts {
				label := fmt.Sprintf("Port %d", v.PortNumber)
				if v.PortLabel != nil && *v.PortLabel != "" {
					label = *v.PortLabel
				}
				if swName, ok := r.Switches[v.SwitchID]; ok {
					label = swName + " - " + label
				}
				r.SwitchPorts[v.ID] = label
				r.SwitchPortSwitch[v.ID] = v.SwitchID
				if v.MacRestriction != nil && *v.MacRestriction != "" {
					r.UsedMACs[*v.MacRestriction] = true
				}
			}
		}

		var ppPorts []models.PatchPanelPort
		if err := c.Get("/patch-panel-ports", &ppPorts); err == nil {
			for _, v := range ppPorts {
				label := fmt.Sprintf("Port %d", v.PortNumber)
				if v.PortLabel != nil && *v.PortLabel != "" {
					label = *v.PortLabel
				}
				if panelName, ok := r.PatchPanels[v.PatchPanelID]; ok {
					label = panelName + " - " + label
				}
				r.PatchPanelPorts[v.ID] = label
				r.PatchPanelPortPanel[v.ID] = v.PatchPanelID
			}
		}

		var deviceIPs []models.DeviceIP
		if err := c.Get("/device-ips", &deviceIPs); err == nil {
			for _, v := range deviceIPs {
				label := v.IPAddress
				if ifaceName, ok := r.Interfaces[v.InterfaceID]; ok {
					label = v.IPAddress + " (" + ifaceName + ")"
				}
				r.DeviceIPs[v.ID] = label
				r.DeviceIPInterface[v.ID] = v.InterfaceID
				if v.VlanID != nil {
					r.DeviceIPVLAN[v.ID] = *v.VlanID
				}
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
	case "device_ips":
		return r.DeviceIPs
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
