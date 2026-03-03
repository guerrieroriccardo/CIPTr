package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("patch_panel_ports", &Def{
		Name:    "Patch Panel Port",
		Plural:  "Patch Panel Ports",
		APIPath: "/patch-panel-ports",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Panel", Width: 18},
			{Title: "Port #", Width: 7},
			{Title: "Label", Width: 15},
			{Title: "Notes", Width: 25},
		},
		ToRow: func(raw any) table.Row {
			pp := raw.(*models.PatchPanelPort)
			return table.Row{
				fmt.Sprintf("%d", pp.ID),
				PatchPanelName(pp.PatchPanelID),
				fmt.Sprintf("%d", pp.PortNumber),
				derefStr(pp.PortLabel),
				derefStr(pp.Notes),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.PatchPanelPort).ID)
		},

		Fields: []Field{
			{Key: "patch_panel_id", Label: "Patch Panel", Required: true, PickerKey: "patch_panels"},
			{Key: "port_number", Label: "Port Number", Required: true},
			{Key: "port_label", Label: "Port Label"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.PatchPanelPort
			if err := client.Get("/patch-panel-ports", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.PatchPanelPortInput{
				PatchPanelID: mustInt64(data["patch_panel_id"]),
				PortNumber:   mustInt(data["port_number"]),
				PortLabel:    strPtr(data["port_label"]),
				Notes:        strPtr(data["notes"]),
			}
			var created models.PatchPanelPort
			err := client.Post("/patch-panel-ports", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.PatchPanelPortInput{
				PatchPanelID: mustInt64(data["patch_panel_id"]),
				PortNumber:   mustInt(data["port_number"]),
				PortLabel:    strPtr(data["port_label"]),
				Notes:        strPtr(data["notes"]),
			}
			var updated models.PatchPanelPort
			err := client.Put("/patch-panel-ports/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/patch-panel-ports/" + id)
		},
	})
}
