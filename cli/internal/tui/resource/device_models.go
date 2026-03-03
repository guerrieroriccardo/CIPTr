package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("device_models", &Def{
		Name:    "Device Model",
		Plural:  "Device Models",
		APIPath: "/device-models",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Manufacturer", Width: 16},
			{Title: "Model Name", Width: 25},
			{Title: "Category", Width: 14},
			{Title: "OS Default", Width: 15},
		},
		ToRow: func(raw any) table.Row {
			dm := raw.(*models.DeviceModel)
			return table.Row{
				fmt.Sprintf("%d", dm.ID),
				ManufacturerName(dm.ManufacturerID),
				dm.ModelName,
				CategoryName(dm.CategoryID),
				derefStr(dm.OsDefault),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceModel).ID)
		},

		Fields: []Field{
			{Key: "manufacturer_id", Label: "Manufacturer ID", Required: true},
			{Key: "model_name", Label: "Model Name", Required: true},
			{Key: "category_id", Label: "Category ID", Required: true},
			{Key: "os_default", Label: "OS Default"},
			{Key: "specs", Label: "Specs"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.DeviceModel
			if err := client.Get("/device-models", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceModelInput{
				ManufacturerID: mustInt64(data["manufacturer_id"]),
				ModelName:      data["model_name"],
				CategoryID:     mustInt64(data["category_id"]),
				OsDefault:      strPtr(data["os_default"]),
				Specs:          strPtr(data["specs"]),
				Notes:          strPtr(data["notes"]),
			}
			var created models.DeviceModel
			err := client.Post("/device-models", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.DeviceModelInput{
				ManufacturerID: mustInt64(data["manufacturer_id"]),
				ModelName:      data["model_name"],
				CategoryID:     mustInt64(data["category_id"]),
				OsDefault:      strPtr(data["os_default"]),
				Specs:          strPtr(data["specs"]),
				Notes:          strPtr(data["notes"]),
			}
			var updated models.DeviceModel
			err := client.Put("/device-models/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/device-models/" + id)
		},
	})
}
