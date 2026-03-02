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
	DeviceType string `json:"device_type"`
	Status     string `json:"status"`
	IsUp       *bool  `json:"is_up"`

	// Software / management
	Os           *string `json:"os"`
	HasRmm       *bool   `json:"has_rmm"`
	HasAntivirus *bool   `json:"has_antivirus"`
	Supplier     *string `json:"supplier"`

	// Logistics
	InstallationDate *string `json:"installation_date"` // DATE as string (YYYY-MM-DD)
	IsReserved       *bool   `json:"is_reserved"`

	// Ticket / reason
	TicketRef *string `json:"ticket_ref"`
	Reason    *string `json:"reason"`

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

	DeviceType string  `json:"device_type" binding:"required"`
	Status     *string `json:"status"`
	IsUp       *bool   `json:"is_up"`

	Os           *string `json:"os"`
	HasRmm       *bool   `json:"has_rmm"`
	HasAntivirus *bool   `json:"has_antivirus"`
	Supplier     *string `json:"supplier"`

	InstallationDate *string `json:"installation_date"`
	IsReserved       *bool   `json:"is_reserved"`

	TicketRef *string `json:"ticket_ref"`
	Reason    *string `json:"reason"`

	Notes *string `json:"notes"`
}
