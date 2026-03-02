package models

// AddressBlock represents an IP address range (e.g. a /20) assigned to a site.
// A site can have multiple blocks for flexibility.
type AddressBlock struct {
	ID          int64   `json:"id"`
	SiteID      int64   `json:"site_id"`
	Network     string  `json:"network"`     // e.g. "10.10.0.0/20"
	Description *string `json:"description"` // nullable
	Notes       *string `json:"notes"`       // nullable
}

// AddressBlockInput is used for create and update requests.
type AddressBlockInput struct {
	SiteID      int64   `json:"site_id" binding:"required"`
	Network     string  `json:"network" binding:"required"`
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
}
