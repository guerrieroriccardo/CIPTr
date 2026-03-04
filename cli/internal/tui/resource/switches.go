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
				LocationName(s.LocationID),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Switch).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "name", Label: "Name", Required: true},
			{Key: "model_id", Label: "Model", PickerKey: "device_models"},
			{Key: "ip_address", Label: "IP Address"},
			{Key: "location_id", Label: "Location", PickerKey: "locations"},
			{Key: "total_ports", Label: "Total Ports"},
			{Key: "notes", Label: "Notes"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil || values["site_id"] == "" {
				return items
			}
			siteID := mustInt64(values["site_id"])
			switch key {
			case "location_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.LocationSite[id] == siteID {
						filtered[id] = name
					}
				}
				return filtered
			}
			return items
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
				LocationID: int64Ptr(data["location_id"]),
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
				LocationID: int64Ptr(data["location_id"]),
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
