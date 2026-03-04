package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("device_connections", &Def{
		Name:    "Device Connection",
		Plural:  "Device Connections",
		APIPath: "/device-connections",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Interface", Width: 36},
			{Title: "Switch Port", Width: 28},
			{Title: "Patch Port", Width: 28},
			{Title: "Connected At", Width: 12},
		},
		ToRow: func(raw any) table.Row {
			dc := raw.(*models.DeviceConnection)
			return table.Row{
				fmt.Sprintf("%d", dc.ID),
				InterfaceName(dc.InterfaceID),
				lookupOptional(func() map[int64]string { return safeLookup().SwitchPorts }, dc.SwitchPortID),
				lookupOptional(func() map[int64]string { return safeLookup().PatchPanelPorts }, dc.PatchPanelPortID),
				derefStr(dc.ConnectedAt),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceConnection).ID)
		},

		Fields: []Field{
			{Key: "interface_id", Label: "Interface", Required: true, PickerKey: "interfaces"},
			{Key: "switch_port_id", Label: "Switch Port", PickerKey: "switch_ports"},
			{Key: "patch_panel_port_id", Label: "Patch Panel Port", PickerKey: "patch_panel_ports"},
			{Key: "connected_at", Label: "Connected At (YYYY-MM-DD)"},
			{Key: "notes", Label: "Notes"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil || values["interface_id"] == "" {
				return items
			}
			ifaceID := mustInt64(values["interface_id"])

			switch key {
			case "interface_id":
				// Show only interfaces of the same device.
				deviceID, ok := Resolve.InterfaceDevice[ifaceID]
				if !ok {
					return items
				}
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.InterfaceDevice[id] == deviceID {
						filtered[id] = name
					}
				}
				return filtered

			case "switch_port_id":
				// Show only switch ports from switches at the same site.
				siteID, ok := Resolve.InterfaceSite[ifaceID]
				if !ok {
					return items
				}
				filtered := make(map[int64]string)
				for id, name := range items {
					if swID, ok2 := Resolve.SwitchPortSwitch[id]; ok2 {
						if Resolve.SwitchSite[swID] == siteID {
							filtered[id] = name
						}
					}
				}
				return filtered

			case "patch_panel_port_id":
				// Show only patch panel ports from panels at the same site.
				siteID, ok := Resolve.InterfaceSite[ifaceID]
				if !ok {
					return items
				}
				filtered := make(map[int64]string)
				for id, name := range items {
					if panelID, ok2 := Resolve.PatchPanelPortPanel[id]; ok2 {
						if Resolve.PatchPanelSite[panelID] == siteID {
							filtered[id] = name
						}
					}
				}
				return filtered
			}
			return items
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.DeviceConnection
			if err := client.Get("/device-connections", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceConnectionInput{
				InterfaceID:      mustInt64(data["interface_id"]),
				SwitchPortID:     int64Ptr(data["switch_port_id"]),
				PatchPanelPortID: int64Ptr(data["patch_panel_port_id"]),
				ConnectedAt:      strPtr(data["connected_at"]),
				Notes:            strPtr(data["notes"]),
			}
			var created models.DeviceConnection
			err := client.Post("/device-connections", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.DeviceConnectionInput{
				InterfaceID:      mustInt64(data["interface_id"]),
				SwitchPortID:     int64Ptr(data["switch_port_id"]),
				PatchPanelPortID: int64Ptr(data["patch_panel_port_id"]),
				ConnectedAt:      strPtr(data["connected_at"]),
				Notes:            strPtr(data["notes"]),
			}
			var updated models.DeviceConnection
			err := client.Put("/device-connections/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/device-connections/" + id)
		},
	})
}
