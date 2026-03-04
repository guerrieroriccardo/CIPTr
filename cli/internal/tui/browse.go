package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/tui/resource"
)

// ---------------------------------------------------------------------------
// ScopeMenu — lightweight sub-menu for choosing a child resource type
// ---------------------------------------------------------------------------

// ScopeMenuItem is a selectable entry in a browse sub-menu.
type ScopeMenuItem struct {
	label string
	build func() Screen
}

// ScopeMenu presents a list of child resource types for a parent entity.
type ScopeMenu struct {
	title  string
	items  []ScopeMenuItem
	cursor int
}

func (sm ScopeMenu) Title() string  { return sm.title }
func (sm ScopeMenu) Init() tea.Cmd  { return nil }

func (sm ScopeMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return sm, func() tea.Msg { return PopScreenMsg{} }
		case "up", "k":
			if sm.cursor > 0 {
				sm.cursor--
			}
		case "down", "j":
			if sm.cursor < len(sm.items)-1 {
				sm.cursor++
			}
		case "enter":
			screen := sm.items[sm.cursor].build()
			return sm, func() tea.Msg {
				return PushScreenMsg{Screen: screen}
			}
		}
	}
	return sm, nil
}

func (sm ScopeMenu) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render(sm.title) + "\n\n")
	for i, item := range sm.items {
		cursor := "  "
		if i == sm.cursor {
			cursor = "> "
		}
		b.WriteString(cursor + item.label + "\n")
	}
	b.WriteString("\n" + HelpStyle.Render("↑↓/jk navigate • enter select • esc back"))
	return b.String()
}

// ---------------------------------------------------------------------------
// Scoped Def factories — shallow copies with filtered List + Defaults
// ---------------------------------------------------------------------------

func scopedSites(clientID string) *resource.Def {
	base := *resource.Registry["sites"]
	base.Defaults = map[string]string{"client_id": clientID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Site
		if err := c.Get("/sites?client_id="+clientID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedLocations(siteID string) *resource.Def {
	base := *resource.Registry["locations"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Location
		if err := c.Get("/locations?site_id="+siteID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedAddressBlocks(siteID string) *resource.Def {
	base := *resource.Registry["address_blocks"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.AddressBlock
		if err := c.Get("/address-blocks?site_id="+siteID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedVLANs(siteID string) *resource.Def {
	base := *resource.Registry["vlans"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.VLAN
		if err := c.Get("/vlans?site_id="+siteID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedDevices(siteID string) *resource.Def {
	base := *resource.Registry["devices"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Device
		if err := c.Get("/devices?site_id="+siteID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedSwitches(siteID string) *resource.Def {
	base := *resource.Registry["switches"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Switch
		if err := c.Get("/switches?site_id="+siteID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedPatchPanels(siteID string) *resource.Def {
	base := *resource.Registry["patch_panels"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.PatchPanel
		if err := c.Get("/patch-panels?site_id="+siteID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedInterfaces(deviceID string) *resource.Def {
	base := *resource.Registry["device_interfaces"]
	base.Defaults = map[string]string{"device_id": deviceID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.DeviceInterface
		if err := c.Get("/device-interfaces?device_id="+deviceID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedDeviceIPs(ifaceID string) *resource.Def {
	base := *resource.Registry["device_ips"]
	base.Defaults = map[string]string{"interface_id": ifaceID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.DeviceIP
		if err := c.Get("/device-ips?interface_id="+ifaceID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedDeviceConnections(ifaceID string) *resource.Def {
	base := *resource.Registry["device_connections"]
	base.Defaults = map[string]string{"interface_id": ifaceID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.DeviceConnection
		if err := c.Get("/device-connections?interface_id="+ifaceID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedSwitchPorts(switchID string) *resource.Def {
	base := *resource.Registry["switch_ports"]
	base.Defaults = map[string]string{"switch_id": switchID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.SwitchPort
		if err := c.Get("/switch-ports?switch_id="+switchID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

func scopedPatchPanelPorts(panelID string) *resource.Def {
	base := *resource.Registry["patch_panel_ports"]
	base.Defaults = map[string]string{"patch_panel_id": panelID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.PatchPanelPort
		if err := c.Get("/patch-panel-ports?patch_panel_id="+panelID, &items); err != nil {
			return nil, err
		}
		result := make([]any, len(items))
		for i := range items {
			result[i] = &items[i]
		}
		return result, nil
	}
	return &base
}

// ---------------------------------------------------------------------------
// OnSelect callback factories — define what happens on enter at each level
// ---------------------------------------------------------------------------

func clientDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		c := raw.(*models.Client)
		clientID := fmt.Sprintf("%d", c.ID)
		def := scopedSites(clientID)
		screen := NewResourceTableWithSelect(def, apiClient, siteDrillDown(apiClient))
		// Override title to show client name instead of generic "Sites"
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, c.Name}}
		}
	}
}

func siteDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		site := raw.(*models.Site)
		menu := newSiteScopeMenu(site, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: menu}
		}
	}
}

func deviceDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		device := raw.(*models.Device)
		menu := newDeviceScopeMenu(device, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: menu}
		}
	}
}

func interfaceDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		iface := raw.(*models.DeviceInterface)
		ifaceID := fmt.Sprintf("%d", iface.ID)
		menu := ScopeMenu{
			title: iface.Name,
			items: []ScopeMenuItem{
				{label: "IPs", build: func() Screen {
					return NewResourceTable(scopedDeviceIPs(ifaceID), apiClient)
				}},
				{label: "Connections", build: func() Screen {
					return NewResourceTable(scopedDeviceConnections(ifaceID), apiClient)
				}},
			},
		}
		return func() tea.Msg {
			return PushScreenMsg{Screen: menu}
		}
	}
}

func switchDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		sw := raw.(*models.Switch)
		switchID := fmt.Sprintf("%d", sw.ID)
		screen := NewResourceTable(scopedSwitchPorts(switchID), apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, sw.Name}}
		}
	}
}

func patchPanelDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		pp := raw.(*models.PatchPanel)
		panelID := fmt.Sprintf("%d", pp.ID)
		screen := NewResourceTable(scopedPatchPanelPorts(panelID), apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, pp.Name}}
		}
	}
}

// ---------------------------------------------------------------------------
// Scope menu factories
// ---------------------------------------------------------------------------

func newSiteScopeMenu(site *models.Site, apiClient *apiclient.Client) ScopeMenu {
	siteID := fmt.Sprintf("%d", site.ID)
	return ScopeMenu{
		title: site.Name,
		items: []ScopeMenuItem{
			{label: "Locations", build: func() Screen {
				return NewResourceTable(scopedLocations(siteID), apiClient)
			}},
			{label: "Address Blocks", build: func() Screen {
				return NewResourceTable(scopedAddressBlocks(siteID), apiClient)
			}},
			{label: "VLANs", build: func() Screen {
				return NewResourceTable(scopedVLANs(siteID), apiClient)
			}},
			{label: "Devices", build: func() Screen {
				return NewResourceTableWithSelect(scopedDevices(siteID), apiClient, deviceDrillDown(apiClient))
			}},
			{label: "Switches", build: func() Screen {
				return NewResourceTableWithSelect(scopedSwitches(siteID), apiClient, switchDrillDown(apiClient))
			}},
			{label: "Patch Panels", build: func() Screen {
				return NewResourceTableWithSelect(scopedPatchPanels(siteID), apiClient, patchPanelDrillDown(apiClient))
			}},
		},
	}
}

func newDeviceScopeMenu(device *models.Device, apiClient *apiclient.Client) ScopeMenu {
	deviceID := fmt.Sprintf("%d", device.ID)
	return ScopeMenu{
		title: device.Hostname,
		items: []ScopeMenuItem{
			{label: "Interfaces", build: func() Screen {
				return NewResourceTableWithSelect(scopedInterfaces(deviceID), apiClient, interfaceDrillDown(apiClient))
			}},
		},
	}
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

// NewBrowseByClientScreen returns the top-level clients table for hierarchical browsing.
func NewBrowseByClientScreen(apiClient *apiclient.Client) ResourceTable {
	return NewResourceTableWithSelect(
		resource.Registry["clients"],
		apiClient,
		clientDrillDown(apiClient),
	)
}

// ---------------------------------------------------------------------------
// titledScreen wraps a Screen to override its Title() for better breadcrumbs.
// ---------------------------------------------------------------------------

type titledScreen struct {
	Screen
	title string
}

func (ts titledScreen) Title() string { return ts.title }

func (ts titledScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := ts.Screen.Update(msg)
	if s, ok := updated.(Screen); ok {
		ts.Screen = s
	}
	return ts, cmd
}
