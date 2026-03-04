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
			{Title: "Name", Width: 16},
			{Title: "Subnet", Width: 18},
			{Title: "Gateway", Width: 16},
			{Title: "DHCP Range", Width: 22},
		},
		ToRow: func(raw any) table.Row {
			v := raw.(*models.VLAN)
			dhcp := ""
			if v.DHCPStart != nil && *v.DHCPStart != "" {
				dhcp = *v.DHCPStart
				if v.DHCPEnd != nil && *v.DHCPEnd != "" {
					dhcp += " - " + *v.DHCPEnd
				}
			}
			return table.Row{
				fmt.Sprintf("%d", v.ID),
				SiteName(v.SiteID),
				fmt.Sprintf("%d", v.VlanID),
				v.Name,
				derefStr(v.Subnet),
				lookupOptional(func() map[int64]string { return safeLookup().DeviceIPs }, v.GatewayDeviceIPID),
				dhcp,
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
			{Key: "gateway_device_ip_id", Label: "Gateway", PickerKey: "device_ips"},
			{Key: "dhcp_start", Label: "DHCP Start"},
			{Key: "dhcp_end", Label: "DHCP End"},
			{Key: "description", Label: "Description"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil {
				return items
			}
			switch key {
			case "address_block_id":
				if values["site_id"] == "" {
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

			case "gateway_device_ip_id":
				// Show only device IPs assigned to this VLAN.
				// When editing, the VLAN ID is available via GetID context.
				// Filter by VlanID FK on device_ips matching the current VLAN's DB id.
				// We need the VLAN's own DB id — but during edit we don't have it in values.
				// Instead, filter by site: show IPs from the same site.
				if values["site_id"] == "" {
					return items
				}
				siteID := mustInt64(values["site_id"])
				filtered := make(map[int64]string)
				for id, name := range items {
					if ifaceID, ok := Resolve.DeviceIPInterface[id]; ok {
						if Resolve.InterfaceSite[ifaceID] == siteID {
							filtered[id] = name
						}
					}
				}
				return filtered
			}
			return items
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
				SiteID:            mustInt64(data["site_id"]),
				AddressBlockID:    int64Ptr(data["address_block_id"]),
				VlanID:            mustInt64(data["vlan_id"]),
				Name:              data["name"],
				Subnet:            strPtr(data["subnet"]),
				GatewayDeviceIPID: int64Ptr(data["gateway_device_ip_id"]),
				DHCPStart:         strPtr(data["dhcp_start"]),
				DHCPEnd:           strPtr(data["dhcp_end"]),
				Description:       strPtr(data["description"]),
			}
			var created models.VLAN
			err := client.Post("/vlans", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.VLANInput{
				SiteID:            mustInt64(data["site_id"]),
				AddressBlockID:    int64Ptr(data["address_block_id"]),
				VlanID:            mustInt64(data["vlan_id"]),
				Name:              data["name"],
				Subnet:            strPtr(data["subnet"]),
				GatewayDeviceIPID: int64Ptr(data["gateway_device_ip_id"]),
				DHCPStart:         strPtr(data["dhcp_start"]),
				DHCPEnd:           strPtr(data["dhcp_end"]),
				Description:       strPtr(data["description"]),
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
