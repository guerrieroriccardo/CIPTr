package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("wifi_ssids", &Def{
		Name:    "WiFi SSID",
		Plural:  "WiFi SSIDs",
		APIPath: "/wifi-ssids",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "SSID", Width: 25},
			{Title: "Auth", Width: 20},
			{Title: "VLAN", Width: 20},
			{Title: "Site", Width: 20},
			{Title: "Notes", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			w := raw.(*models.WifiSSID)
			return table.Row{
				fmt.Sprintf("%d", w.ID),
				w.SSID,
				derefStr(w.Auth),
				VLANName(w.VlanID),
				SiteName(w.SiteID),
				derefStr(w.Notes),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.WifiSSID).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "ssid", Label: "SSID", Required: true},
			{Key: "auth", Label: "Auth Protocol", PickerOptions: []string{
				"WPA2-PSK", "WPA3-SAE", "WPA2-Enterprise", "WPA3-Enterprise", "Open",
			}},
			{Key: "vlan_id", Label: "VLAN", PickerKey: "vlans"},
			{Key: "notes", Label: "Notes"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil {
				return items
			}
			if key == "vlan_id" && values["site_id"] != "" {
				siteID := mustInt64(values["site_id"])
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.VLANSite[id] == siteID {
						filtered[id] = name
					}
				}
				return filtered
			}
			return items
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.WifiSSID
			if err := client.Get("/wifi-ssids", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.WifiSSIDInput{
				SiteID: mustInt64(data["site_id"]),
				SSID:   data["ssid"],
				Auth:   strPtr(data["auth"]),
				VlanID: int64Ptr(data["vlan_id"]),
				Notes:  strPtr(data["notes"]),
			}
			var created models.WifiSSID
			err := client.Post("/wifi-ssids", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.WifiSSIDInput{
				SiteID: mustInt64(data["site_id"]),
				SSID:   data["ssid"],
				Auth:   strPtr(data["auth"]),
				VlanID: int64Ptr(data["vlan_id"]),
				Notes:  strPtr(data["notes"]),
			}
			var updated models.WifiSSID
			err := client.Put("/wifi-ssids/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/wifi-ssids/" + id)
		},
	})
}
