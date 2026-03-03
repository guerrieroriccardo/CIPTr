package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("device_interfaces", &Def{
		Name:    "Device Interface",
		Plural:  "Device Interfaces",
		APIPath: "/device-interfaces",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Device", Width: 20},
			{Title: "Name", Width: 20},
			{Title: "MAC Address", Width: 20},
			{Title: "Notes", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			di := raw.(*models.DeviceInterface)
			return table.Row{
				fmt.Sprintf("%d", di.ID),
				DeviceName(di.DeviceID),
				di.Name,
				derefStr(di.MacAddress),
				derefStr(di.Notes),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceInterface).ID)
		},

		Fields: []Field{
			{Key: "device_id", Label: "Device ID", Required: true},
			{Key: "name", Label: "Name", Required: true},
			{Key: "mac_address", Label: "MAC Address"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.DeviceInterface
			if err := client.Get("/device-interfaces", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceInterfaceInput{
				DeviceID:   mustInt64(data["device_id"]),
				Name:       data["name"],
				MacAddress: strPtr(data["mac_address"]),
				Notes:      strPtr(data["notes"]),
			}
			var created models.DeviceInterface
			err := client.Post("/device-interfaces", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.DeviceInterfaceInput{
				DeviceID:   mustInt64(data["device_id"]),
				Name:       data["name"],
				MacAddress: strPtr(data["mac_address"]),
				Notes:      strPtr(data["notes"]),
			}
			var updated models.DeviceInterface
			err := client.Put("/device-interfaces/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/device-interfaces/" + id)
		},
	})
}
