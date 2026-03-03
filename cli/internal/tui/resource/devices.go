package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("devices", &Def{
		Name:    "Device",
		Plural:  "Devices",
		APIPath: "/devices",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Hostname", Width: 22},
			{Title: "Status", Width: 10},
			{Title: "Category", Width: 14},
			{Title: "Site", Width: 16},
			{Title: "OS", Width: 14},
		},
		ToRow: func(raw any) table.Row {
			d := raw.(*models.Device)
			return table.Row{
				fmt.Sprintf("%d", d.ID),
				d.Hostname,
				d.Status,
				CategoryName(d.CategoryID),
				SiteName(d.SiteID),
				derefStr(d.Os),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Device).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "hostname", Label: "Hostname", Required: true},
			{Key: "category_id", Label: "Category", Required: true, PickerKey: "categories"},
			{Key: "location_id", Label: "Location", PickerKey: "locations"},
			{Key: "model_id", Label: "Model", PickerKey: "device_models"},
			{Key: "dns_name", Label: "DNS Name"},
			{Key: "serial_number", Label: "Serial Number"},
			{Key: "asset_tag", Label: "Asset Tag"},
			{Key: "status", Label: "Status (active/decommissioned/storage)"},
			{Key: "is_up", Label: "Is Up (true/false)"},
			{Key: "os", Label: "OS"},
			{Key: "has_rmm", Label: "Has RMM (true/false)"},
			{Key: "has_antivirus", Label: "Has Antivirus (true/false)"},
			{Key: "supplier_id", Label: "Supplier", PickerKey: "suppliers"},
			{Key: "installation_date", Label: "Installation Date (YYYY-MM-DD)"},
			{Key: "is_reserved", Label: "Is Reserved (true/false)"},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Device
			if err := client.Get("/devices", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceInput{
				SiteID:           mustInt64(data["site_id"]),
				Hostname:         data["hostname"],
				CategoryID:       mustInt64(data["category_id"]),
				LocationID:       int64Ptr(data["location_id"]),
				ModelID:          int64Ptr(data["model_id"]),
				DnsName:          strPtr(data["dns_name"]),
				SerialNumber:     strPtr(data["serial_number"]),
				AssetTag:         strPtr(data["asset_tag"]),
				Status:           strPtr(data["status"]),
				IsUp:             boolPtr(data["is_up"]),
				Os:               strPtr(data["os"]),
				HasRmm:           boolPtr(data["has_rmm"]),
				HasAntivirus:     boolPtr(data["has_antivirus"]),
				SupplierID:       int64Ptr(data["supplier_id"]),
				InstallationDate: strPtr(data["installation_date"]),
				IsReserved:       boolPtr(data["is_reserved"]),
				Notes:            strPtr(data["notes"]),
			}
			var created models.Device
			err := client.Post("/devices", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.DeviceInput{
				SiteID:           mustInt64(data["site_id"]),
				Hostname:         data["hostname"],
				CategoryID:       mustInt64(data["category_id"]),
				LocationID:       int64Ptr(data["location_id"]),
				ModelID:          int64Ptr(data["model_id"]),
				DnsName:          strPtr(data["dns_name"]),
				SerialNumber:     strPtr(data["serial_number"]),
				AssetTag:         strPtr(data["asset_tag"]),
				Status:           strPtr(data["status"]),
				IsUp:             boolPtr(data["is_up"]),
				Os:               strPtr(data["os"]),
				HasRmm:           boolPtr(data["has_rmm"]),
				HasAntivirus:     boolPtr(data["has_antivirus"]),
				SupplierID:       int64Ptr(data["supplier_id"]),
				InstallationDate: strPtr(data["installation_date"]),
				IsReserved:       boolPtr(data["is_reserved"]),
				Notes:            strPtr(data["notes"]),
			}
			var updated models.Device
			err := client.Put("/devices/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/devices/" + id)
		},
	})
}
