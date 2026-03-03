package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("categories", &Def{
		Name:    "Category",
		Plural:  "Categories",
		APIPath: "/categories",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: 40},
		},
		ToRow: func(raw any) table.Row {
			c := raw.(*models.Category)
			return table.Row{
				fmt.Sprintf("%d", c.ID),
				c.Name,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Category).ID)
		},

		Fields: []Field{
			{Key: "name", Label: "Name", Required: true},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Category
			if err := client.Get("/categories", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.CategoryInput{Name: data["name"]}
			var created models.Category
			err := client.Post("/categories", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.CategoryInput{Name: data["name"]}
			var updated models.Category
			err := client.Put("/categories/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/categories/" + id)
		},
	})
}
