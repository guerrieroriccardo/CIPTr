package models

import "time"

// DeviceModel represents a hardware model in the catalog (e.g. "HP ProLiant DL360 Gen10").
type DeviceModel struct {
	ID           int64     `json:"id"`
	Manufacturer string    `json:"manufacturer"`
	ModelName    string    `json:"model_name"`
	Category     string    `json:"category"`
	OsDefault    *string   `json:"os_default"`
	Specs        *string   `json:"specs"`
	Notes        *string   `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
}

// DeviceModelInput is used for create and update requests.
type DeviceModelInput struct {
	Manufacturer string  `json:"manufacturer" binding:"required"`
	ModelName    string  `json:"model_name"   binding:"required"`
	Category     string  `json:"category"     binding:"required"`
	OsDefault    *string `json:"os_default"`
	Specs        *string `json:"specs"`
	Notes        *string `json:"notes"`
}
