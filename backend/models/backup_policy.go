package models

import "time"

// BackupPolicy represents a backup policy assigned to a client.
type BackupPolicy struct {
	ID            int64     `json:"id"`
	ClientID      int64     `json:"client_id"`
	ClientName    string    `json:"client_name"`
	Name          string    `json:"name"`
	Destination   string    `json:"destination"`
	RetainLast    int       `json:"retain_last"`
	RetainHourly  int       `json:"retain_hourly"`
	RetainDaily   int       `json:"retain_daily"`
	RetainWeekly  int       `json:"retain_weekly"`
	RetainMonthly int       `json:"retain_monthly"`
	RetainYearly  int       `json:"retain_yearly"`
	Enabled       bool      `json:"enabled"`
	Source        *string   `json:"source"`
	Notes         *string   `json:"notes"`
	ScheduleTimes []string  `json:"schedule_times"` // "HH:MM" 24h format
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BackupPolicyInput is used for create and update requests.
type BackupPolicyInput struct {
	ClientID      int64    `json:"client_id"   binding:"required"`
	Name          string   `json:"name"        binding:"required"`
	Destination   string   `json:"destination" binding:"required"`
	RetainLast    *int     `json:"retain_last"`
	RetainHourly  *int     `json:"retain_hourly"`
	RetainDaily   *int     `json:"retain_daily"`
	RetainWeekly  *int     `json:"retain_weekly"`
	RetainMonthly *int     `json:"retain_monthly"`
	RetainYearly  *int     `json:"retain_yearly"`
	Enabled       *bool    `json:"enabled"`
	Source        *string  `json:"source"`
	Notes         *string  `json:"notes"`
	ScheduleTimes []string `json:"schedule_times"` // "HH:MM" 24h format
}
