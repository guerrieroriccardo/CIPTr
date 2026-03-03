package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("locations", &Def{
		Name:    "Location",
		Plural:  "Locations",
		APIPath: "/locations",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site", Width: 20},
			{Title: "Name", Width: 30},
			{Title: "Floor", Width: 10},
			{Title: "Notes", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			l := raw.(*models.Location)
			return table.Row{
				fmt.Sprintf("%d", l.ID),
				SiteName(l.SiteID),
				l.Name,
				derefStr(l.Floor),
				derefStr(l.Notes),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Location).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site ID", Required: true},
			{Key: "name", Label: "Name", Required: true},
			{Key: "floor", Label: "Floor"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Location
			if err := client.Get("/locations", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.LocationInput{
				SiteID: mustInt64(data["site_id"]),
				Name:   data["name"],
				Floor:  strPtr(data["floor"]),
				Notes:  strPtr(data["notes"]),
			}
			var created models.Location
			err := client.Post("/locations", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.LocationInput{
				SiteID: mustInt64(data["site_id"]),
				Name:   data["name"],
				Floor:  strPtr(data["floor"]),
				Notes:  strPtr(data["notes"]),
			}
			var updated models.Location
			err := client.Put("/locations/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/locations/" + id)
		},
	})
}
