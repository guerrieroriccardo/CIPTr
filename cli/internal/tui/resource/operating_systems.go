package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("operating_systems", &Def{
		Name:    "Operating System",
		Plural:  "Operating Systems",
		APIPath: "/operating-systems",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			os := raw.(*models.OperatingSystem)
			return table.Row{
				fmt.Sprintf("%d", os.ID),
				os.Name,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.OperatingSystem).ID)
		},

		Fields: []Field{
			{Key: "name", Label: "Name", Required: true},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.OperatingSystem
			if err := client.Get("/operating-systems", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.OperatingSystemInput{Name: data["name"]}
			var created models.OperatingSystem
			err := client.Post("/operating-systems", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.OperatingSystemInput{Name: data["name"]}
			var updated models.OperatingSystem
			err := client.Put("/operating-systems/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/operating-systems/" + id)
		},
	})
}
