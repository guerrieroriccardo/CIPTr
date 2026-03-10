package resource

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

func init() {
	Register("device_group_members", &Def{
		Name:    "Device Group Member",
		Plural:  "Device Group Members",
		APIPath: "/device-group-members",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Group", Width: 20},
			{Title: "Device", Width: 20},
		},
		ToRow: func(raw any) table.Row {
			m := raw.(*models.DeviceGroupMember)
			return table.Row{
				fmt.Sprintf("%d", m.ID),
				DeviceGroupName(m.GroupID),
				DeviceName(m.DeviceID),
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.DeviceGroupMember).ID)
		},

		Fields: []Field{
			{Key: "group_id", Label: "Device Group", Required: true, PickerKey: "device_groups"},
			{Key: "device_id", Label: "Device", Required: true, PickerKey: "devices"},
		},

		PickerFilter: func(key string, values map[string]string, items map[int64]string) map[int64]string {
			if Resolve == nil {
				return items
			}
			switch key {
			case "device_id":
				// Filter devices to the same site as the selected group.
				if values["group_id"] == "" {
					return items
				}
				groupID := mustInt64(values["group_id"])
				groupSiteID, ok := Resolve.DeviceGroupSite[groupID]
				if !ok {
					return items
				}
				filtered := make(map[int64]string)
				for id, name := range items {
					if Resolve.DeviceSite[id] == groupSiteID {
						filtered[id] = name
					}
				}
				return filtered
			}
			return items
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.DeviceGroupMember
			if err := client.Get("/device-group-members", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.DeviceGroupMemberInput{
				GroupID:  mustInt64(data["group_id"]),
				DeviceID: mustInt64(data["device_id"]),
			}
			var created models.DeviceGroupMember
			err := client.Post("/device-group-members", input, &created)
			return &created, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/device-group-members/" + id)
		},
	})
}
