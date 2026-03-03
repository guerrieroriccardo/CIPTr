package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("address_blocks", &Def{
		Name:    "Address Block",
		Plural:  "Address Blocks",
		APIPath: "/address-blocks",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site ID", Width: 8},
			{Title: "Network", Width: 20},
			{Title: "Description", Width: 30},
		},
		ToRow: func(raw any) table.Row {
			ab := raw.(*models.AddressBlock)
			return table.Row{
				fmt.Sprintf("%d", ab.ID),
				fmt.Sprintf("%d", ab.SiteID),
				ab.Network,
				derefStr(ab.Description),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.AddressBlock).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site ID", Required: true},
			{Key: "network", Label: "Network (CIDR)", Required: true},
			{Key: "description", Label: "Description"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.AddressBlock
			if err := client.Get("/address-blocks", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.AddressBlockInput{
				SiteID:      mustInt64(data["site_id"]),
				Network:     data["network"],
				Description: strPtr(data["description"]),
				Notes:       strPtr(data["notes"]),
			}
			var created models.AddressBlock
			err := client.Post("/address-blocks", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.AddressBlockInput{
				SiteID:      mustInt64(data["site_id"]),
				Network:     data["network"],
				Description: strPtr(data["description"]),
				Notes:       strPtr(data["notes"]),
			}
			var updated models.AddressBlock
			err := client.Put("/address-blocks/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/address-blocks/" + id)
		},
	})
}
