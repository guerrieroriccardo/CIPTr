package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("device_ips", &Def{
		Name:    "Device IP",
		Plural:  "Device IPs",
		APIPath: "/device-ips",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Interface", Width: 36},
			{Title: "IP Address", Width: 18},
			{Title: "VLAN", Width: 14},
			{Title: "Primary", Width: 8},
			{Title: "Notes", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			ip := raw.(*models.DeviceIP)
			return table.Row{
				fmt.Sprintf("%d", ip.ID),
				InterfaceName(ip.InterfaceID),
				ip.IPAddress,
				VLANName(ip.VlanID),
				derefBool(ip.IsPrimary),
				derefStr(ip.Notes),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceIP).ID)
		},

		Fields: []Field{
			{Key: "interface_id", Label: "Interface", Required: true, PickerKey: "interfaces"},
			{Key: "vlan_id", Label: "VLAN", PickerKey: "vlans"},
			{Key: "ip_address", Label: "IP Address", Required: true},
			{Key: "is_primary", Label: "Primary", PickerOptions: []string{"true", "false"}},
			{Key: "set_as_gateway", Label: "Set as Gateway", PickerOptions: []string{"true", "false"},
				Hidden: func(values map[string]string) bool { return values["vlan_id"] == "" }},
			{Key: "notes", Label: "Notes"},
		},

		PreSubmit: func(values map[string]string) string {
			if values["set_as_gateway"] != "true" || values["vlan_id"] == "" || Resolve == nil {
				return ""
			}
			vlanID := mustInt64(values["vlan_id"])
			existingGW, ok := Resolve.VLANGateway[vlanID]
			if !ok || existingGW == 0 {
				return ""
			}
			gwLabel := fmt.Sprintf("%d", existingGW)
			if name, ok := Resolve.DeviceIPs[existingGW]; ok {
				gwLabel = name
			}
			vlanName := fmt.Sprintf("%d", vlanID)
			if name, ok := Resolve.VLANs[vlanID]; ok {
				vlanName = name
			}
			return fmt.Sprintf("VLAN '%s' already has gateway '%s'. This will override it.", vlanName, gwLabel)
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if key != "vlan_id" || values["interface_id"] == "" || Resolve == nil {
				return items
			}
			ifaceID := mustInt64(values["interface_id"])
			siteID, ok := Resolve.InterfaceSite[ifaceID]
			if !ok {
				return items
			}
			filtered := make(map[int64]string)
			for id, name := range items {
				if Resolve.VLANSite[id] == siteID {
					filtered[id] = name
				}
			}
			return filtered
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

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.DeviceIP
			if err := client.Get("/device-ips", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceIPInput{
				InterfaceID:  mustInt64(data["interface_id"]),
				IPAddress:    data["ip_address"],
				VlanID:       int64Ptr(data["vlan_id"]),
				IsPrimary:    boolPtr(data["is_primary"]),
				SetAsGateway: boolPtr(data["set_as_gateway"]),
				Notes:        strPtr(data["notes"]),
			}
			var created models.DeviceIP
			err := client.Post("/device-ips", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.DeviceIPInput{
				InterfaceID:  mustInt64(data["interface_id"]),
				IPAddress:    data["ip_address"],
				VlanID:       int64Ptr(data["vlan_id"]),
				IsPrimary:    boolPtr(data["is_primary"]),
				SetAsGateway: boolPtr(data["set_as_gateway"]),
				Notes:        strPtr(data["notes"]),
			}
			var updated models.DeviceIP
			err := client.Put("/device-ips/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/device-ips/" + id)
		},
	})
}
