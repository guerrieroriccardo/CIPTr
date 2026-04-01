package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("switches", &Def{
		Name:    "Switch",
		Plural:  "Switches",
		APIPath: "/switches",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site", Width: 16},
			{Title: "Hostname", Width: 22},
			{Title: "IP Address", Width: 16},
			{Title: "VLAN", Width: 14},
			{Title: "Ports", Width: 6},
			{Title: "Location", Width: 18},
		},
		ToRow: func(raw any) table.Row {
			s := raw.(*models.Switch)
			return table.Row{
				fmt.Sprintf("%d", s.ID),
				SiteName(s.SiteID),
				s.Hostname,
				derefStr(s.IPAddress),
				VLANName(s.VlanID),
				fmt.Sprintf("%d", s.TotalPorts),
				LocationName(s.LocationID),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.Switch).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "hostname", Label: "Hostname", Required: true},
			{Key: "model_id", Label: "Model", PickerKey: "device_models"},
			{Key: "vlan_id", Label: "VLAN", PickerKey: "vlans"},
			{Key: "ip_address", Label: "IP Address"},
			{Key: "location_id", Label: "Location", PickerKey: "locations"},
			{Key: "total_ports", Label: "Total Ports"},
			{Key: "notes", Label: "Notes"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil || values["site_id"] == "" {
				return items
			}
			siteID := mustInt64(values["site_id"])
			switch key {
			case "location_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.LocationSite[id] == siteID {
						filtered[id] = name
					}
				}
				return filtered
			case "vlan_id":
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

		FieldHint: func(key string, values map[string]string) string {
			if key == "ip_address" && values["vlan_id"] != "" && Resolve != nil {
				id := mustInt64(values["vlan_id"])
				if subnet, ok := Resolve.VLANSubnets[id]; ok {
					return "Subnet: " + subnet
				}
			}
			return ""
		},

		AsyncDerive: func(client *apiclient.Client, key string, values map[string]string) map[string]string {
			if key != "site_id" || values["site_id"] == "" {
				return nil
			}
			var result struct {
				Hostname string `json:"hostname"`
			}
			err := client.Get(fmt.Sprintf("/switches/next-name?site_id=%s", values["site_id"]), &result)
			if err != nil || result.Hostname == "" {
				return nil
			}
			return map[string]string{"hostname": result.Hostname}
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.Switch
			if err := client.Get("/switches", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.SwitchInput{
				SiteID:     mustInt64(data["site_id"]),
				Hostname:   data["hostname"],
				ModelID:    int64Ptr(data["model_id"]),
				IPAddress:  strPtr(data["ip_address"]),
				VlanID:     int64Ptr(data["vlan_id"]),
				LocationID: int64Ptr(data["location_id"]),
				TotalPorts: intPtr(data["total_ports"]),
				Notes:      strPtr(data["notes"]),
			}
			var created models.Switch
			err := client.Post("/switches", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.SwitchInput{
				SiteID:     mustInt64(data["site_id"]),
				Hostname:   data["hostname"],
				ModelID:    int64Ptr(data["model_id"]),
				IPAddress:  strPtr(data["ip_address"]),
				VlanID:     int64Ptr(data["vlan_id"]),
				LocationID: int64Ptr(data["location_id"]),
				TotalPorts: intPtr(data["total_ports"]),
				Notes:      strPtr(data["notes"]),
			}
			var updated models.Switch
			err := client.Put("/switches/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/switches/" + id)
		},
	})
}
