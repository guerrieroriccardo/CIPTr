package resource

import (
	"fmt"
	"strings"

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
			{Title: "ID", Width: 5},
			{Title: "Switch", Width: 16},
			{Title: "#", Width: 4},
			{Title: "Label", Width: 12},
			{Title: "Speed", Width: 8},
			{Title: "Up", Width: 4},
			{Title: "Dis", Width: 4},
			{Title: "Untagged", Width: 14},
			{Title: "Tagged", Width: 20},
			{Title: "MAC Restrict", Width: 18},
		},
		ToRow: func(raw any) table.Row {
			sp := raw.(*models.SwitchPort)
			// Resolve tagged VLAN names.
			var taggedNames []string
			for _, vid := range sp.TaggedVlanIDs {
				taggedNames = append(taggedNames, VLANName(&vid))
			}
			return table.Row{
				fmt.Sprintf("%d", sp.ID),
				DeviceName(sp.DeviceID),
				fmt.Sprintf("%d", sp.PortNumber),
				derefStr(sp.PortLabel),
				derefStr(sp.Speed),
				derefBool(sp.IsUplink),
				derefBool(sp.IsDisabled),
				VLANName(sp.UntaggedVlanID),
				strings.Join(taggedNames, ", "),
				derefStr(sp.MacRestriction),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.SwitchPort).ID)
		},

		Fields: []Field{
			{Key: "device_id", Label: "Device", Required: true, PickerKey: "devices"},
			{Key: "port_number", Label: "Port Number", Required: true},
			{Key: "port_label", Label: "Port Label"},
			{Key: "speed", Label: "Speed"},
			{Key: "is_uplink", Label: "Uplink", PickerOptions: []string{"true", "false"}},
			{Key: "is_disabled", Label: "Disabled", PickerOptions: []string{"true", "false"}},
			{Key: "untagged_vlan_id", Label: "Untagged VLAN", PickerKey: "vlans"},
			{Key: "tagged_vlan_ids", Label: "Tagged VLANs (comma-sep IDs)"},
			{Key: "mac_restriction", Label: "MAC Restriction", PickerFunc: func(values map[string]string) []PickerEntry {
				if Resolve == nil {
					return nil
				}
				// Determine site from the selected device.
				var siteID int64
				if deviceID := mustInt64(values["device_id"]); deviceID != 0 {
					siteID = Resolve.DeviceSite[deviceID]
				}
				// Allow the current value through (editing keeps own MAC).
				currentMAC := values["mac_restriction"]
				var entries []PickerEntry
				for ifaceID, mac := range Resolve.InterfaceMAC {
					// Filter by site if we know it.
					if siteID != 0 {
						if ifaceSite, ok := Resolve.InterfaceSite[ifaceID]; ok && ifaceSite != siteID {
							continue
						}
					}
					// Skip MACs already used on another port.
					if mac != currentMAC && Resolve.UsedMACs[mac] {
						continue
					}
					label := mac
					if name, ok := Resolve.Interfaces[ifaceID]; ok {
						label = mac + " (" + name + ")"
					}
					entries = append(entries, PickerEntry{Value: mac, Label: label})
				}
				return entries
			}},
			{Key: "notes", Label: "Notes"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil || values["device_id"] == "" {
				return items
			}
			deviceID := mustInt64(values["device_id"])
			siteID := Resolve.DeviceSite[deviceID]
			if siteID == 0 {
				return items
			}
			switch key {
			case "untagged_vlan_id":
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
				DeviceID:       mustInt64(data["device_id"]),
				PortNumber:     mustInt(data["port_number"]),
				PortLabel:      strPtr(data["port_label"]),
				Speed:          strPtr(data["speed"]),
				IsUplink:       boolPtr(data["is_uplink"]),
				IsDisabled:     boolPtr(data["is_disabled"]),
				MacRestriction: strPtr(data["mac_restriction"]),
				UntaggedVlanID: int64Ptr(data["untagged_vlan_id"]),
				TaggedVlanIDs:  parseInt64Slice(data["tagged_vlan_ids"]),
				Notes:          strPtr(data["notes"]),
			}
			var created models.SwitchPort
			err := client.Post("/switch-ports", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.SwitchPortInput{
				DeviceID:       mustInt64(data["device_id"]),
				PortNumber:     mustInt(data["port_number"]),
				PortLabel:      strPtr(data["port_label"]),
				Speed:          strPtr(data["speed"]),
				IsUplink:       boolPtr(data["is_uplink"]),
				IsDisabled:     boolPtr(data["is_disabled"]),
				MacRestriction: strPtr(data["mac_restriction"]),
				UntaggedVlanID: int64Ptr(data["untagged_vlan_id"]),
				TaggedVlanIDs:  parseInt64Slice(data["tagged_vlan_ids"]),
				Notes:          strPtr(data["notes"]),
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
