package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("suppliers", &Def{
		Name:    "Supplier",
		Plural:  "Suppliers",
		APIPath: "/suppliers",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Name", Width: 25},
			{Title: "Phone", Width: 18},
			{Title: "Email", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			s := raw.(*models.Supplier)
			return table.Row{
				fmt.Sprintf("%d", s.ID),
				s.Name,
				derefStr(s.Phone),
				derefStr(s.Email),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Supplier).ID)
		},

		Fields: []Field{
			{Key: "name", Label: "Name", Required: true},
			{Key: "address", Label: "Address"},
			{Key: "phone", Label: "Phone"},
			{Key: "email", Label: "Email"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Supplier
			if err := client.Get("/suppliers", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.SupplierInput{
				Name:    data["name"],
				Address: strPtr(data["address"]),
				Phone:   strPtr(data["phone"]),
				Email:   strPtr(data["email"]),
			}
			var created models.Supplier
			err := client.Post("/suppliers", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.SupplierInput{
				Name:    data["name"],
				Address: strPtr(data["address"]),
				Phone:   strPtr(data["phone"]),
				Email:   strPtr(data["email"]),
			}
			var updated models.Supplier
			err := client.Put("/suppliers/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/suppliers/" + id)
		},
	})
}
