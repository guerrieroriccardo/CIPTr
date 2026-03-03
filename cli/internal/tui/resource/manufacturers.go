package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("manufacturers", &Def{
		Name:    "Manufacturer",
		Plural:  "Manufacturers",
		APIPath: "/manufacturers",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: 40},
		},
		ToRow: func(raw any) table.Row {
			m := raw.(*models.Manufacturer)
			return table.Row{
				fmt.Sprintf("%d", m.ID),
				m.Name,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Manufacturer).ID)
		},

		Fields: []Field{
			{Key: "name", Label: "Name", Required: true},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Manufacturer
			if err := client.Get("/manufacturers", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.ManufacturerInput{Name: data["name"]}
			var created models.Manufacturer
			err := client.Post("/manufacturers", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.ManufacturerInput{Name: data["name"]}
			var updated models.Manufacturer
			err := client.Put("/manufacturers/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/manufacturers/" + id)
		},
	})
}
