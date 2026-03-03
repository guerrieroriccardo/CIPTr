package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("switch_ports", &Def{
		Name:    "Switch Port",
		Plural:  "Switch Ports",
		APIPath: "/switch-ports",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Switch", Width: 18},
			{Title: "Port #", Width: 7},
			{Title: "Label", Width: 15},
			{Title: "Speed", Width: 10},
			{Title: "Uplink", Width: 7},
		},
		ToRow: func(raw any) table.Row {
			sp := raw.(*models.SwitchPort)
			return table.Row{
				fmt.Sprintf("%d", sp.ID),
				SwitchName(sp.SwitchID),
				fmt.Sprintf("%d", sp.PortNumber),
				derefStr(sp.PortLabel),
				derefStr(sp.Speed),
				derefBool(sp.IsUplink),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.SwitchPort).ID)
		},

		Fields: []Field{
			{Key: "switch_id", Label: "Switch ID", Required: true},
			{Key: "port_number", Label: "Port Number", Required: true},
			{Key: "port_label", Label: "Port Label"},
			{Key: "speed", Label: "Speed"},
			{Key: "is_uplink", Label: "Uplink (true/false)"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.SwitchPort
			if err := client.Get("/switch-ports", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.SwitchPortInput{
				SwitchID:   mustInt64(data["switch_id"]),
				PortNumber: mustInt(data["port_number"]),
				PortLabel:  strPtr(data["port_label"]),
				Speed:      strPtr(data["speed"]),
				IsUplink:   boolPtr(data["is_uplink"]),
				Notes:      strPtr(data["notes"]),
			}
			var created models.SwitchPort
			err := client.Post("/switch-ports", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.SwitchPortInput{
				SwitchID:   mustInt64(data["switch_id"]),
				PortNumber: mustInt(data["port_number"]),
				PortLabel:  strPtr(data["port_label"]),
				Speed:      strPtr(data["speed"]),
				IsUplink:   boolPtr(data["is_uplink"]),
				Notes:      strPtr(data["notes"]),
			}
			var updated models.SwitchPort
			err := client.Put("/switch-ports/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/switch-ports/" + id)
		},
	})
}
