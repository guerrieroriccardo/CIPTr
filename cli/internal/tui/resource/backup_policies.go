package resource

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/guerrieroriccardo/CIPTr/backend/models"
	"github.com/guerrieroriccardo/CIPTr/cli/internal/apiclient"
)

// retentionSummary builds a compact retention string like "D:7 W:4 M:12 Y:3".
func retentionSummary(p *models.BackupPolicy) string {
	parts := []string{}
	if p.RetainLast > 0 {
		parts = append(parts, fmt.Sprintf("L:%d", p.RetainLast))
	}
	if p.RetainHourly > 0 {
		parts = append(parts, fmt.Sprintf("H:%d", p.RetainHourly))
	}
	if p.RetainDaily > 0 {
		parts = append(parts, fmt.Sprintf("D:%d", p.RetainDaily))
	}
	if p.RetainWeekly > 0 {
		parts = append(parts, fmt.Sprintf("W:%d", p.RetainWeekly))
	}
	if p.RetainMonthly > 0 {
		parts = append(parts, fmt.Sprintf("M:%d", p.RetainMonthly))
	}
	if p.RetainYearly > 0 {
		parts = append(parts, fmt.Sprintf("Y:%d", p.RetainYearly))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

// parseScheduleTimes splits a comma-separated time string into a []string,
// trimming whitespace and ignoring empty entries.
func parseScheduleTimes(s string) []string {
	if s == "" {
		return []string{}
	}
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func init() {
	Register("backup_policies", &Def{
		Name:    "Backup Policy",
		Plural:  "Backup Policies",
		APIPath: "/backup-policies",

		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "Client", Width: 20},
			{Title: "Name", Width: 20},
			{Title: "Destination", Width: 22},
			{Title: "Source", Width: 18},
			{Title: "Schedule", Width: 14},
			{Title: "Retention", Width: 22},
			{Title: "Enabled", Width: 7},
		},
		ToRow: func(raw any) table.Row {
			p := raw.(*models.BackupPolicy)
			enabled := "yes"
			if !p.Enabled {
				enabled = "no"
			}
			schedule := strings.Join(p.ScheduleTimes, ", ")
			if schedule == "" {
				schedule = "-"
			}
			return table.Row{
				fmt.Sprintf("%d", p.ID),
				p.ClientName,
				p.Name,
				p.Destination,
				derefStr(p.Source),
				schedule,
				retentionSummary(p),
				enabled,
			}
		},
		GetID: func(raw any) string {
			return fmt.Sprintf("%d", raw.(*models.BackupPolicy).ID)
		},

		Fields: []Field{
			{Key: "client_id", Label: "Client", Required: true, PickerKey: "clients"},
			{Key: "name", Label: "Name", Required: true},
			{Key: "destination", Label: "Destination", Required: true},
			{Key: "source", Label: "Source"},
			{Key: "schedule_times", Label: "Schedule (HH:MM, comma-separated)"},
			{Key: "retain_last", Label: "Retain Last (count)"},
			{Key: "retain_hourly", Label: "Retain Hourly"},
			{Key: "retain_daily", Label: "Retain Daily"},
			{Key: "retain_weekly", Label: "Retain Weekly"},
			{Key: "retain_monthly", Label: "Retain Monthly"},
			{Key: "retain_yearly", Label: "Retain Yearly"},
			{Key: "enabled", Label: "Enabled", PickerOptions: []string{"true", "false"}},
			{Key: "notes", Label: "Notes"},
		},

		List: func(client *apiclient.Client) ([]any, error) {
			var items []models.BackupPolicy
			if err := client.Get("/backup-policies", &items); err != nil {
				return nil, err
			}
			result := make([]any, len(items))
			for i := range items {
				result[i] = &items[i]
			}
			return result, nil
		},
		Create: func(client *apiclient.Client, data map[string]string) (any, error) {
			input := models.BackupPolicyInput{
				ClientID:      mustInt64(data["client_id"]),
				Name:          data["name"],
				Destination:   data["destination"],
				Source:        strPtr(data["source"]),
				RetainLast:    intPtr(data["retain_last"]),
				RetainHourly:  intPtr(data["retain_hourly"]),
				RetainDaily:   intPtr(data["retain_daily"]),
				RetainWeekly:  intPtr(data["retain_weekly"]),
				RetainMonthly: intPtr(data["retain_monthly"]),
				RetainYearly:  intPtr(data["retain_yearly"]),
				Enabled:       boolPtr(data["enabled"]),
				Notes:         strPtr(data["notes"]),
				ScheduleTimes: parseScheduleTimes(data["schedule_times"]),
			}
			var created models.BackupPolicy
			err := client.Post("/backup-policies", input, &created)
			return &created, err
		},
		Update: func(client *apiclient.Client, id string, data map[string]string) (any, error) {
			input := models.BackupPolicyInput{
				ClientID:      mustInt64(data["client_id"]),
				Name:          data["name"],
				Destination:   data["destination"],
				Source:        strPtr(data["source"]),
				RetainLast:    intPtr(data["retain_last"]),
				RetainHourly:  intPtr(data["retain_hourly"]),
				RetainDaily:   intPtr(data["retain_daily"]),
				RetainWeekly:  intPtr(data["retain_weekly"]),
				RetainMonthly: intPtr(data["retain_monthly"]),
				RetainYearly:  intPtr(data["retain_yearly"]),
				Enabled:       boolPtr(data["enabled"]),
				Notes:         strPtr(data["notes"]),
				ScheduleTimes: parseScheduleTimes(data["schedule_times"]),
			}
			var updated models.BackupPolicy
			err := client.Put("/backup-policies/"+id, input, &updated)
			return &updated, err
		},
		Delete: func(client *apiclient.Client, id string) error {
			return client.Delete("/backup-policies/" + id)
		},
	})
}
