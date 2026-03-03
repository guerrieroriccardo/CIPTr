package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("sites", &Def{
		Name:    "Site",
		Plural:  "Sites",
		APIPath: "/sites",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Client ID", Width: 10},
			{Title: "Name", Width: 30},
			{Title: "Address", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			s := raw.(*models.Site)
			return table.Row{
				fmt.Sprintf("%d", s.ID),
				fmt.Sprintf("%d", s.ClientID),
				s.Name,
				derefStr(s.Address),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Site).ID)
		},

		Fields: []Field{
			{Key: "client_id", Label: "Client ID", Required: true},
			{Key: "name", Label: "Name", Required: true},
			{Key: "address", Label: "Address"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Site
			if err := client.Get("/sites", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.SiteInput{
				ClientID: mustInt64(data["client_id"]),
				Name:     data["name"],
				Address:  strPtr(data["address"]),
				Notes:    strPtr(data["notes"]),
			}
			var created models.Site
			err := client.Post("/sites", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.SiteInput{
				ClientID: mustInt64(data["client_id"]),
				Name:     data["name"],
				Address:  strPtr(data["address"]),
				Notes:    strPtr(data["notes"]),
			}
			var updated models.Site
			err := client.Put("/sites/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/sites/" + id)
		},
	})
}
