package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("clients", &Def{
		Name:    "Client",
		Plural:  "Clients",
		APIPath: "/clients",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: 30},
			{Title: "Code", Width: 10},
			{Title: "Notes", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			c := raw.(*models.Client)
			notes := ""
			if c.Notes != nil {
				notes = *c.Notes
			}
			return table.Row{
				fmt.Sprintf("%d", c.ID),
				c.Name,
				c.ShortCode,
				notes,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Client).ID)
		},

		Fields: []Field{
			{Key: "name", Label: "Name", Required: true},
			{Key: "short_code", Label: "Short Code", Required: true},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var clients []models.Client
			if err := client.Get("/clients", &clients); err != nil {
				return nil, err
			}
			result := make([]any, len(clients))
			for i := range clients {
				result[i] = &clients[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.ClientInput{
				Name:      data["name"],
				ShortCode: data["short_code"],
			}
			if v, ok := data["notes"]; ok && v != "" {
				input.Notes = &v
			}
			var created models.Client
			err := client.Post("/clients", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.ClientInput{
				Name:      data["name"],
				ShortCode: data["short_code"],
			}
			if v, ok := data["notes"]; ok && v != "" {
				input.Notes = &v
			}
			var updated models.Client
			err := client.Put("/clients/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/clients/" + id)
		},
	})
}
