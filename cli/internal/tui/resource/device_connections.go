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
			if Resolve == nil {
				return items
			}

			// Determine site from whichever field is already filled.
			var siteID int64
			if v := values["interface_id"]; v != "" {
				ifaceID := mustInt64(v)
				if s, ok := Resolve.InterfaceSite[ifaceID]; ok {
					siteID = s
				}
			} else if v := values["switch_port_id"]; v != "" && siteID == 0 {
				spID := mustInt64(v)
				if devID, ok := Resolve.SwitchPortDevice[spID]; ok {
					siteID = Resolve.DeviceSite[devID]
				}
			} else if v := values["patch_panel_port_id"]; v != "" && siteID == 0 {
				ppID := mustInt64(v)
				if devID, ok := Resolve.PatchPanelPortDevice[ppID]; ok {
					siteID = Resolve.DeviceSite[devID]
				}
			}

			if siteID == 0 {
				return items
			}

			switch key {
			case "interface_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if s, ok := Resolve.InterfaceSite[id]; ok && s == siteID {
						filtered[id] = name
					}
				}
				return filtered

			case "switch_port_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if devID, ok := Resolve.SwitchPortDevice[id]; ok {
						if Resolve.DeviceSite[devID] == siteID {
							filtered[id] = name
						}
					}
				}
				return filtered

			case "patch_panel_port_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if devID, ok := Resolve.PatchPanelPortDevice[id]; ok {
						if Resolve.DeviceSite[devID] == siteID {
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
