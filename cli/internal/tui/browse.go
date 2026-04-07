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
	base.Defaults = map[string]string{"site_id": siteID, "status": "planned"}
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

func scopedDeviceGroups(siteID string) *resource.Def {
	base := *resource.Registry["device_groups"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.DeviceGroup
		if err := c.Get("/device-groups?site_id="+siteID, &items); err != nil {
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

func scopedDeviceGroupMembers(groupID string) *resource.Def {
	base := *resource.Registry["device_group_members"]
	base.Defaults = map[string]string{"group_id": groupID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.DeviceGroupMember
		if err := c.Get("/device-group-members?group_id="+groupID, &items); err != nil {
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

func scopedBackupPolicies(clientID string) *resource.Def {
	base := *resource.Registry["backup_policies"]
	base.Defaults = map[string]string{"client_id": clientID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.BackupPolicy
		if err := c.Get("/backup-policies?client_id="+clientID, &items); err != nil {
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

func scopedWifiSSIDs(siteID string) *resource.Def {
	base := *resource.Registry["wifi_ssids"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.WifiSSID
		if err := c.Get("/wifi-ssids?site_id="+siteID, &items); err != nil {
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

func scopedFirewallRules(siteID string) *resource.Def {
	base := *resource.Registry["firewall_rules"]
	base.Defaults = map[string]string{"site_id": siteID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.FirewallRule
		if err := c.Get("/firewall-rules?site_id="+siteID, &items); err != nil {
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

func scopedSwitchPorts(deviceID string) *resource.Def {
	base := *resource.Registry["switch_ports"]
	base.Defaults = map[string]string{"device_id": deviceID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.SwitchPort
		if err := c.Get("/switch-ports?device_id="+deviceID, &items); err != nil {
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

func scopedPatchPanelPorts(deviceID string) *resource.Def {
	base := *resource.Registry["patch_panel_ports"]
	base.Defaults = map[string]string{"device_id": deviceID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.PatchPanelPort
		if err := c.Get("/patch-panel-ports?device_id="+deviceID, &items); err != nil {
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
// Scoped Defs for "All" tables — show associated items on enter
// ---------------------------------------------------------------------------

func scopedDevicesByCategory(catID string) *resource.Def {
	base := *resource.Registry["devices"]
	base.Defaults = map[string]string{"category_id": catID, "status": "planned"}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Device
		if err := c.Get("/devices?category_id="+catID, &items); err != nil {
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

func scopedDevicesBySupplier(supplierID string) *resource.Def {
	base := *resource.Registry["devices"]
	base.Defaults = map[string]string{"supplier_id": supplierID, "status": "planned"}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Device
		if err := c.Get("/devices?supplier_id="+supplierID, &items); err != nil {
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

func scopedDevicesByModel(modelID string) *resource.Def {
	base := *resource.Registry["devices"]
	base.Defaults = map[string]string{"model_id": modelID, "status": "planned"}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Device
		if err := c.Get("/devices?model_id="+modelID, &items); err != nil {
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

func scopedDevicesByLocation(locationID string) *resource.Def {
	base := *resource.Registry["devices"]
	base.Defaults = map[string]string{"location_id": locationID, "status": "planned"}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Device
		if err := c.Get("/devices?location_id="+locationID, &items); err != nil {
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

func scopedDevicesByOs(osID string) *resource.Def {
	base := *resource.Registry["devices"]
	base.Defaults = map[string]string{"os_id": osID, "status": "planned"}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.Device
		if err := c.Get("/devices?os_id="+osID, &items); err != nil {
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

func scopedModelsByManufacturer(mfgID string) *resource.Def {
	base := *resource.Registry["device_models"]
	base.Defaults = map[string]string{"manufacturer_id": mfgID}
	base.List = func(c *apiclient.Client) ([]any, error) {
		var items []models.DeviceModel
		if err := c.Get("/device-models?manufacturer_id="+mfgID, &items); err != nil {
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
		menu := newClientScopeMenu(c, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: menu}
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

func deviceGroupDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		g := raw.(*models.DeviceGroup)
		groupID := fmt.Sprintf("%d", g.ID)
		screen := NewResourceTable(scopedDeviceGroupMembers(groupID), apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, g.Name + " Members"}}
		}
	}
}

// ---------------------------------------------------------------------------
// Scope menu factories
// ---------------------------------------------------------------------------

func newClientScopeMenu(client *models.Client, apiClient *apiclient.Client) ScopeMenu {
	clientID := fmt.Sprintf("%d", client.ID)
	return ScopeMenu{
		title: client.Name,
		items: []ScopeMenuItem{
			{label: "Sites", build: func() Screen {
				def := scopedSites(clientID)
				screen := NewResourceTableWithSelect(def, apiClient, siteDrillDown(apiClient))
				return titledScreen{screen, client.Name + " — Sites"}
			}},
			{label: "Backup Policies", build: func() Screen {
				return NewResourceTable(scopedBackupPolicies(clientID), apiClient)
			}},
			{label: "IP Address Space", build: func() Screen {
				return NewIPUsageScreen(apiClient, "client", "?client_id="+clientID)
			}},
			{label: "Export to PDF", build: func() Screen {
				return NewConfirmExport(apiClient, clientID, client.Name, client.ShortCode)
			}},
		},
	}
}

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
			{label: "Device Groups", build: func() Screen {
				return NewResourceTableWithSelect(scopedDeviceGroups(siteID), apiClient, deviceGroupDrillDown(apiClient))
			}},
			{label: "WiFi SSIDs", build: func() Screen {
				return NewResourceTable(scopedWifiSSIDs(siteID), apiClient)
			}},
			{label: "Firewall Rules", build: func() Screen {
				return NewResourceTable(scopedFirewallRules(siteID), apiClient)
			}},
			{label: "IP Address Space", build: func() Screen {
				return NewIPUsageScreen(apiClient, "site", "?site_id="+siteID)
			}},
		},
	}
}

func newDeviceScopeMenu(device *models.Device, apiClient *apiclient.Client) ScopeMenu {
	deviceID := fmt.Sprintf("%d", device.ID)
	items := []ScopeMenuItem{
		{label: "Interfaces", build: func() Screen {
			return NewResourceTableWithSelect(scopedInterfaces(deviceID), apiClient, interfaceDrillDown(apiClient))
		}},
	}
	// Add port management based on category port_type.
	if resource.Resolve != nil {
		portType := resource.Resolve.CategoryPortType[device.CategoryID]
		if portType == "switch" {
			items = append(items, ScopeMenuItem{label: "Switch Ports", build: func() Screen {
				return NewResourceTable(scopedSwitchPorts(deviceID), apiClient)
			}})
		} else if portType == "patch_panel" {
			items = append(items, ScopeMenuItem{label: "Patch Panel Ports", build: func() Screen {
				return NewResourceTable(scopedPatchPanelPorts(deviceID), apiClient)
			}})
		}
	}
	return ScopeMenu{
		title: device.Hostname,
		items: items,
	}
}

// ---------------------------------------------------------------------------
// OnSelect callbacks for "All" lookup tables
// ---------------------------------------------------------------------------

func categoryDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		cat := raw.(*models.Category)
		catID := fmt.Sprintf("%d", cat.ID)
		def := scopedDevicesByCategory(catID)
		screen := NewResourceTable(def, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, cat.Name + " Devices"}}
		}
	}
}

func supplierDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		sup := raw.(*models.Supplier)
		supID := fmt.Sprintf("%d", sup.ID)
		def := scopedDevicesBySupplier(supID)
		screen := NewResourceTable(def, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, sup.Name + " Devices"}}
		}
	}
}

func deviceModelDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		dm := raw.(*models.DeviceModel)
		dmID := fmt.Sprintf("%d", dm.ID)
		def := scopedDevicesByModel(dmID)
		screen := NewResourceTable(def, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, dm.ModelName + " Devices"}}
		}
	}
}

func manufacturerDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		mfg := raw.(*models.Manufacturer)
		mfgID := fmt.Sprintf("%d", mfg.ID)
		def := scopedModelsByManufacturer(mfgID)
		screen := NewResourceTableWithSelect(def, apiClient, deviceModelDrillDown(apiClient))
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, mfg.Name + " Models"}}
		}
	}
}

func osDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		os := raw.(*models.OperatingSystem)
		osID := fmt.Sprintf("%d", os.ID)
		def := scopedDevicesByOs(osID)
		screen := NewResourceTable(def, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, os.Name + " Devices"}}
		}
	}
}

func locationDrillDown(apiClient *apiclient.Client) func(any) tea.Cmd {
	return func(raw any) tea.Cmd {
		loc := raw.(*models.Location)
		locID := fmt.Sprintf("%d", loc.ID)
		def := scopedDevicesByLocation(locID)
		screen := NewResourceTable(def, apiClient)
		return func() tea.Msg {
			return PushScreenMsg{Screen: titledScreen{screen, loc.Name + " Devices"}}
		}
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
