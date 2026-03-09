package resource

import (
	"fmt"
	"os"

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
			{Title: "Hostname", Width: 20},
			{Title: "Status", Width: 10},
			{Title: "Category", Width: 14},
			{Title: "Site", Width: 14},
			{Title: "Location", Width: 14},
			{Title: "OS", Width: 12},
		},
		ToRow: func(raw any) table.Row {
			d := raw.(*models.Device)
			return table.Row{
				fmt.Sprintf("%d", d.ID),
				d.Hostname,
				d.Status,
				CategoryName(d.CategoryID),
				SiteName(d.SiteID),
				LocationName(d.LocationID),
				OsName(d.OsID),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Device).ID)
		},

		Defaults: map[string]string{"status": "planned"},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "category_id", Label: "Category", Required: true, PickerKey: "categories"},
			{Key: "hostname", Label: "Hostname", Required: true},
			{Key: "location_id", Label: "Location", PickerKey: "locations"},
			{Key: "model_id", Label: "Model", PickerKey: "device_models"},
			{Key: "dns_name", Label: "DNS Name"},
			{Key: "serial_number", Label: "Serial Number"},
			{Key: "asset_tag", Label: "Asset Tag"},
			{Key: "status", Label: "Status", PickerOptions: []string{"planned", "active", "inactive", "decommissioned", "storage"}},
			{Key: "is_up", Label: "Is Up", PickerOptions: []string{"true", "false"}},
			{Key: "os_id", Label: "OS", PickerKey: "operating_systems"},
			{Key: "has_rmm", Label: "Has RMM", PickerOptions: []string{"true", "false"}},
			{Key: "has_antivirus", Label: "Has Antivirus", PickerOptions: []string{"true", "false"}},
			{Key: "supplier_id", Label: "Supplier", PickerKey: "suppliers"},
			{Key: "installation_date", Label: "Installation Date (YYYY-MM-DD)"},
			{Key: "is_reserved", Label: "Is Reserved", PickerOptions: []string{"true", "false"}},
			{Key: "notes", Label: "Notes"},
		},

		ExportLabel: func(client *apiclient.Client, id string) (string, error) {
			data, err := client.GetRaw("/devices/" + id + "/label")
			if err != nil {
				return "", err
			}
			path := fmt.Sprintf("label-device-%s.pdf", id)
			return path, os.WriteFile(path, data, 0644)
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil {
				return items
			}
			switch key {
			case "location_id":
				if values["site_id"] == "" {
					return items
				}
				siteID := mustInt64(values["site_id"])
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.LocationSite[id] == siteID {
						filtered[id] = name
					}
				}
				return filtered
			case "model_id":
				if values["category_id"] == "" {
					return items
				}
				catID := mustInt64(values["category_id"])
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.DeviceModelCategory[id] == catID {
						filtered[id] = name
					}
				}
				return filtered
			}
			return items
		},

		AsyncDerive: func(client *apiclient.Client, key string, values map[string]string) map[string]string {
			if key != "site_id" && key != "category_id" {
				return nil
			}
			if values["site_id"] == "" || values["category_id"] == "" {
				return nil
			}
			var result struct {
				Hostname string `json:"hostname"`
				DnsName  string `json:"dns_name"`
			}
			err := client.Get(fmt.Sprintf("/devices/next-hostname?site_id=%s&category_id=%s",
				values["site_id"], values["category_id"]), &result)
			if err != nil || result.Hostname == "" {
				return nil
			}
			derived := map[string]string{"hostname": result.Hostname}
			if result.DnsName != "" {
				derived["dns_name"] = result.DnsName
			}
			return derived
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
				OsID:             int64Ptr(data["os_id"]),
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
				OsID:             int64Ptr(data["os_id"]),
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
