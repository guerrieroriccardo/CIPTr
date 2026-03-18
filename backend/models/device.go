package models

import "time"

// Device represents a deployed device at a site (PC, server, printer, etc.).
type Device struct {
	ID         int64  `json:"id"`
	SiteID     int64  `json:"site_id"`
	LocationID *int64 `json:"location_id"`
	ModelID    *int64 `json:"model_id"`

	// Identification
	Hostname     string  `json:"hostname"`
	DnsName      *string `json:"dns_name"`
	SerialNumber *string `json:"serial_number"`
	AssetTag     *string `json:"asset_tag"`

	// Type and status
	CategoryID int64  `json:"category_id"`
	Status     string `json:"status"`
	IsUp       *bool  `json:"is_up"`

	// Software / management
	OsID         *int64 `json:"os_id"`
	HasRmm       *bool  `json:"has_rmm"`
	HasAntivirus *bool  `json:"has_antivirus"`
	SupplierID   *int64 `json:"supplier_id"`

	// Logistics
	InstallationDate *string `json:"installation_date"` // DATE as string (YYYY-MM-DD)
	IsReserved       *bool   `json:"is_reserved"`

	VmID      *int64    `json:"vm_id"`
	Notes     *string   `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceInput is used for create and update requests.
type DeviceInput struct {
	SiteID     int64  `json:"site_id"     binding:"required"`
	LocationID *int64 `json:"location_id"`
	ModelID    *int64 `json:"model_id"`

	Hostname     string  `json:"hostname"    binding:"required"`
	DnsName      *string `json:"dns_name"`
	SerialNumber *string `json:"serial_number"`
	AssetTag     *string `json:"asset_tag"`

	CategoryID int64   `json:"category_id" binding:"required"`
	Status     *string `json:"status"`
	IsUp       *bool   `json:"is_up"`

	OsID         *int64 `json:"os_id"`
	HasRmm       *bool  `json:"has_rmm"`
	HasAntivirus *bool  `json:"has_antivirus"`
	SupplierID   *int64 `json:"supplier_id"`

	InstallationDate *string `json:"installation_date"`
	IsReserved       *bool   `json:"is_reserved"`

	VmID  *int64  `json:"vm_id"`
	Notes *string `json:"notes"`
}
