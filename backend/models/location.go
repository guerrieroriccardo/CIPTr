package models

// Location represents a physical area within a site (e.g. a room, floor, closet).
type Location struct {
	ID     int64   `json:"id"`
	SiteID int64   `json:"site_id"`
	Name   string  `json:"name"`
	Floor  *string `json:"floor"`
	Notes  *string `json:"notes"`
}

// LocationInput is used for create and update requests.
type LocationInput struct {
	SiteID int64   `json:"site_id" binding:"required"`
	Name   string  `json:"name"    binding:"required"`
	Floor  *string `json:"floor"`
	Notes  *string `json:"notes"`
}
