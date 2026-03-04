package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("vlans", &Def{
		Name:    "VLAN",
		Plural:  "VLANs",
		APIPath: "/vlans",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Site", Width: 16},
			{Title: "VLAN #", Width: 7},
			{Title: "Name", Width: 20},
			{Title: "Subnet", Width: 20},
			{Title: "Gateway", Width: 16},
		},
		ToRow: func(raw any) table.Row {
			v := raw.(*models.VLAN)
			return table.Row{
				fmt.Sprintf("%d", v.ID),
				SiteName(v.SiteID),
				fmt.Sprintf("%d", v.VlanID),
				v.Name,
				derefStr(v.Subnet),
				derefStr(v.Gateway),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.VLAN).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "address_block_id", Label: "Address Block", PickerKey: "address_blocks"},
			{Key: "vlan_id", Label: "VLAN Tag Number", Required: true},
			{Key: "name", Label: "Name", Required: true},
			{Key: "subnet", Label: "Subnet (CIDR)"},
			{Key: "gateway", Label: "Gateway"},
			{Key: "description", Label: "Description"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if key != "address_block_id" || values["site_id"] == "" || Resolve == nil {
				return items
			}
			siteID := mustInt64(values["site_id"])
			filtered := make(map[int64]string)
			for id, name := range items {
				if Resolve.AddressBlockSite[id] == siteID {
					filtered[id] = name
				}
			}
			return filtered
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.VLAN
			if err := client.Get("/vlans", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.VLANInput{
				SiteID:         mustInt64(data["site_id"]),
				AddressBlockID: int64Ptr(data["address_block_id"]),
				VlanID:         mustInt64(data["vlan_id"]),
				Name:           data["name"],
				Subnet:         strPtr(data["subnet"]),
				Gateway:        strPtr(data["gateway"]),
				Description:    strPtr(data["description"]),
			}
			var created models.VLAN
			err := client.Post("/vlans", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.VLANInput{
				SiteID:         mustInt64(data["site_id"]),
				AddressBlockID: int64Ptr(data["address_block_id"]),
				VlanID:         mustInt64(data["vlan_id"]),
				Name:           data["name"],
				Subnet:         strPtr(data["subnet"]),
				Gateway:        strPtr(data["gateway"]),
				Description:    strPtr(data["description"]),
			}
			var updated models.VLAN
			err := client.Put("/vlans/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/vlans/" + id)
		},
	})
}
