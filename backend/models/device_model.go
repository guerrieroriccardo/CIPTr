package models

import "time"

// DeviceModel represents a hardware model in the catalog (e.g. "HP ProLiant DL360 Gen10").
type DeviceModel struct {
	ID             int64     `json:"id"`
	ManufacturerID int64     `json:"manufacturer_id"`
	ModelName      string    `json:"model_name"`
	CategoryID     int64     `json:"category_id"`
	OsDefaultID    *int64    `json:"os_default_id"`
	Specs          *string   `json:"specs"`
	Notes          *string   `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
}

// DeviceModelInput is used for create and update requests.
type DeviceModelInput struct {
	ManufacturerID int64   `json:"manufacturer_id" binding:"required"`
	ModelName      string  `json:"model_name"      binding:"required"`
	CategoryID     int64   `json:"category_id"     binding:"required"`
	OsDefaultID    *int64  `json:"os_default_id"`
	Specs          *string `json:"specs"`
	Notes          *string `json:"notes"`
}
