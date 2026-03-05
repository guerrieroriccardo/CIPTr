package resource

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

type auditLog struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID *int64    `json:"resource_id"`
	Detail     *string   `json:"detail"`
	CreatedAt  time.Time `json:"created_at"`
}

func init() {
	Register("audit_logs", &Def{
		Name:    "Audit Log",
		Plural:  "Audit Logs",
		APIPath: "/audit-logs",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Time", Width: 20},
			{Title: "User", Width: 14},
			{Title: "Action", Width: 8},
			{Title: "Resource", Width: 18},
			{Title: "Detail", Width: 40},
		},
		ToRow: func(raw any) table.Row {
			a := raw.(*auditLog)
			detail := ""
			if a.Detail != nil {
				detail = *a.Detail
			}
			return table.Row{
				fmt.Sprintf("%d", a.ID),
				a.CreatedAt.Local().Format("2006-01-02 15:04:05"),
				a.Username,
				a.Action,
				a.Resource,
				detail,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*auditLog).ID)
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var raw json.RawMessage
			if err := client.Get("/audit-logs?limit=200", &raw); err != nil {
				return nil, err
			}
			var items []auditLog
			if err := json.Unmarshal(raw, &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		// Read-only: no Create, Update, Delete
	})
}
