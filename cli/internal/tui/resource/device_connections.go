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
			{Title: "Interface", Width: 14},
			{Title: "Switch Port", Width: 12},
			{Title: "Patch Port", Width: 12},
			{Title: "Connected At", Width: 12},
		},
		ToRow: func(raw any) table.Row {
			dc := raw.(*models.DeviceConnection)
			return table.Row{
				fmt.Sprintf("%d", dc.ID),
				InterfaceName(dc.InterfaceID),
				derefInt64(dc.SwitchPortID),
				derefInt64(dc.PatchPanelPortID),
				derefStr(dc.ConnectedAt),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceConnection).ID)
		},

		Fields: []Field{
			{Key: "interface_id", Label: "Interface ID", Required: true},
			{Key: "switch_port_id", Label: "Switch Port ID"},
			{Key: "patch_panel_port_id", Label: "Patch Panel Port ID"},
			{Key: "connected_at", Label: "Connected At (YYYY-MM-DD)"},
			{Key: "notes", Label: "Notes"},
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
