package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("patch_panels", &Def{
		Name:    "Patch Panel",
		Plural:  "Patch Panels",
		APIPath: "/patch-panels",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site", Width: 16},
			{Title: "Name", Width: 22},
			{Title: "Ports", Width: 6},
			{Title: "Location", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			pp := raw.(*models.PatchPanel)
			return table.Row{
				fmt.Sprintf("%d", pp.ID),
				SiteName(pp.SiteID),
				pp.Name,
				fmt.Sprintf("%d", pp.TotalPorts),
				derefStr(pp.Location),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.PatchPanel).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site ID", Required: true},
			{Key: "name", Label: "Name", Required: true},
			{Key: "total_ports", Label: "Total Ports"},
			{Key: "location", Label: "Location"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.PatchPanel
			if err := client.Get("/patch-panels", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.PatchPanelInput{
				SiteID:     mustInt64(data["site_id"]),
				Name:       data["name"],
				TotalPorts: intPtr(data["total_ports"]),
				Location:   strPtr(data["location"]),
				Notes:      strPtr(data["notes"]),
			}
			var created models.PatchPanel
			err := client.Post("/patch-panels", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.PatchPanelInput{
				SiteID:     mustInt64(data["site_id"]),
				Name:       data["name"],
				TotalPorts: intPtr(data["total_ports"]),
				Location:   strPtr(data["location"]),
				Notes:      strPtr(data["notes"]),
			}
			var updated models.PatchPanel
			err := client.Put("/patch-panels/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/patch-panels/" + id)
		},
	})
}
