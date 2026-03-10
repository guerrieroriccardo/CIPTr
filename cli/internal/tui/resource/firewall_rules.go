package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

// endpointSummary returns a human-readable label for a firewall rule endpoint.
func endpointSummary(deviceID, groupID, vlanID *int64, cidr *string) string {
	if deviceID != nil {
		return DeviceName(*deviceID)
	}
	if groupID != nil {
		return DeviceGroupName(*groupID)
	}
	if vlanID != nil {
		return VLANName(vlanID)
	}
	if cidr != nil {
		return *cidr
	}
	return "any"
}

func init() {
	Register("firewall_rules", &Def{
		Name:    "Firewall Rule",
		Plural:  "Firewall Rules",
		APIPath: "/firewall-rules",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Pos", Width: 4},
			{Title: "Action", Width: 6},
			{Title: "Source", Width: 22},
			{Title: "Destination", Width: 22},
			{Title: "Src Port", Width: 8},
			{Title: "Dst Port", Width: 8},
			{Title: "Proto", Width: 5},
			{Title: "Enabled", Width: 7},
		},
		ToRow: func(raw any) table.Row {
			r := raw.(*models.FirewallRule)
			enabled := "yes"
			if !r.Enabled {
				enabled = "no"
			}
			return table.Row{
				fmt.Sprintf("%d", r.ID),
				fmt.Sprintf("%d", r.Position),
				r.Action,
				endpointSummary(r.SrcDeviceID, r.SrcGroupID, r.SrcVlanID, r.SrcCIDR),
				endpointSummary(r.DstDeviceID, r.DstGroupID, r.DstVlanID, r.DstCIDR),
				r.SrcPort,
				r.DstPort,
				r.Protocol,
				enabled,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.FirewallRule).ID)
		},

		Fields: []Field{
			{Key: "site_id", Label: "Site", Required: true, PickerKey: "sites"},
			{Key: "src_device_id", Label: "Src Device", PickerKey: "devices"},
			{Key: "src_group_id", Label: "Src Group", PickerKey: "device_groups"},
			{Key: "src_vlan_id", Label: "Src VLAN", PickerKey: "vlans"},
			{Key: "src_cidr", Label: "Src CIDR"},
			{Key: "dst_device_id", Label: "Dst Device", PickerKey: "devices"},
			{Key: "dst_group_id", Label: "Dst Group", PickerKey: "device_groups"},
			{Key: "dst_vlan_id", Label: "Dst VLAN", PickerKey: "vlans"},
			{Key: "dst_cidr", Label: "Dst CIDR"},
			{Key: "src_port", Label: "Src Port"},
			{Key: "dst_port", Label: "Dst Port"},
			{Key: "protocol", Label: "Protocol", PickerOptions: []string{"any", "tcp", "udp", "both", "icmp"}},
			{Key: "action", Label: "Action", PickerOptions: []string{"allow", "deny"}},
			{Key: "position", Label: "Position"},
			{Key: "enabled", Label: "Enabled", PickerOptions: []string{"true", "false"}},
			{Key: "description", Label: "Description"},
			{Key: "notes", Label: "Notes"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil {
				return items
			}
			if values["site_id"] == "" {
				return items
			}
			siteID := mustInt64(values["site_id"])
			switch key {
			case "src_device_id", "dst_device_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.DeviceSite[id] == siteID {
						filtered[id] = name
					}
				}
				return filtered
			case "src_group_id", "dst_group_id":
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.DeviceGroupSite[id] == siteID {
						filtered[id] = name
					}
				}
				return filtered
			case "src_vlan_id", "dst_vlan_id":
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
			var items []models.FirewallRule
			if err := client.Get("/firewall-rules", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.FirewallRuleInput{
				SiteID:      mustInt64(data["site_id"]),
				SrcDeviceID: int64Ptr(data["src_device_id"]),
				SrcGroupID:  int64Ptr(data["src_group_id"]),
				SrcVlanID:   int64Ptr(data["src_vlan_id"]),
				SrcCIDR:     strPtr(data["src_cidr"]),
				DstDeviceID: int64Ptr(data["dst_device_id"]),
				DstGroupID:  int64Ptr(data["dst_group_id"]),
				DstVlanID:   int64Ptr(data["dst_vlan_id"]),
				DstCIDR:     strPtr(data["dst_cidr"]),
				SrcPort:     strPtr(data["src_port"]),
				DstPort:     strPtr(data["dst_port"]),
				Protocol:    strPtr(data["protocol"]),
				Action:      strPtr(data["action"]),
				Position:    intPtr(data["position"]),
				Enabled:     boolPtr(data["enabled"]),
				Description: strPtr(data["description"]),
				Notes:       strPtr(data["notes"]),
			}
			var created models.FirewallRule
			err := client.Post("/firewall-rules", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.FirewallRuleInput{
				SiteID:      mustInt64(data["site_id"]),
				SrcDeviceID: int64Ptr(data["src_device_id"]),
				SrcGroupID:  int64Ptr(data["src_group_id"]),
				SrcVlanID:   int64Ptr(data["src_vlan_id"]),
				SrcCIDR:     strPtr(data["src_cidr"]),
				DstDeviceID: int64Ptr(data["dst_device_id"]),
				DstGroupID:  int64Ptr(data["dst_group_id"]),
				DstVlanID:   int64Ptr(data["dst_vlan_id"]),
				DstCIDR:     strPtr(data["dst_cidr"]),
				SrcPort:     strPtr(data["src_port"]),
				DstPort:     strPtr(data["dst_port"]),
				Protocol:    strPtr(data["protocol"]),
				Action:      strPtr(data["action"]),
				Position:    intPtr(data["position"]),
				Enabled:     boolPtr(data["enabled"]),
				Description: strPtr(data["description"]),
				Notes:       strPtr(data["notes"]),
			}
			var updated models.FirewallRule
			err := client.Put("/firewall-rules/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/firewall-rules/" + id)
		},
	})
}
