package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("device_groups", &Def{
		Name:    "Device Group",
		Plural:  "Device Groups",
		APIPath: "/device-groups",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site", Width: 16},
			{Title: "Name", Width: 20},
			{Title: "Description", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			g := raw.(*models.DeviceGroup)
			return table.Row{
				fmt.Sprintf("%d", g.ID),
				SiteName(g.SiteID),
				g.Name,
				derefStr(g.Description),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceGroup).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "name", Label: "Name", Required: true},
			{Key: "description", Label: "Description"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.DeviceGroup
			if err := client.Get("/device-groups", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceGroupInput{
				SiteID:      mustInt64(data["site_id"]),
				Name:        data["name"],
				Description: strPtr(data["description"]),
				Notes:       strPtr(data["notes"]),
			}
			var created models.DeviceGroup
			err := client.Post("/device-groups", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.DeviceGroupInput{
				SiteID:      mustInt64(data["site_id"]),
				Name:        data["name"],
				Description: strPtr(data["description"]),
				Notes:       strPtr(data["notes"]),
			}
			var updated models.DeviceGroup
			err := client.Put("/device-groups/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/device-groups/" + id)
		},
	})
}
