package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("switches", &Def{
		Name:    "Switch",
		Plural:  "Switches",
		APIPath: "/switches",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site", Width: 16},
			{Title: "Name", Width: 22},
			{Title: "IP Address", Width: 16},
			{Title: "Ports", Width: 6},
			{Title: "Location", Width: 18},
		},
		ToRow: func(raw any) table.Row {
			s := raw.(*models.Switch)
			return table.Row{
				fmt.Sprintf("%d", s.ID),
				SiteName(s.SiteID),
				s.Name,
				derefStr(s.IPAddress),
				fmt.Sprintf("%d", s.TotalPorts),
				derefStr(s.Location),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Switch).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site ID", Required: true},
			{Key: "name", Label: "Name", Required: true},
			{Key: "model_id", Label: "Model ID"},
			{Key: "ip_address", Label: "IP Address"},
			{Key: "location", Label: "Location"},
			{Key: "total_ports", Label: "Total Ports"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Switch
			if err := client.Get("/switches", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.SwitchInput{
				SiteID:     mustInt64(data["site_id"]),
				Name:       data["name"],
				ModelID:    int64Ptr(data["model_id"]),
				IPAddress:  strPtr(data["ip_address"]),
				Location:   strPtr(data["location"]),
				TotalPorts: intPtr(data["total_ports"]),
				Notes:      strPtr(data["notes"]),
			}
			var created models.Switch
			err := client.Post("/switches", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.SwitchInput{
				SiteID:     mustInt64(data["site_id"]),
				Name:       data["name"],
				ModelID:    int64Ptr(data["model_id"]),
				IPAddress:  strPtr(data["ip_address"]),
				Location:   strPtr(data["location"]),
				TotalPorts: intPtr(data["total_ports"]),
				Notes:      strPtr(data["notes"]),
			}
			var updated models.Switch
			err := client.Put("/switches/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/switches/" + id)
		},
	})
}
