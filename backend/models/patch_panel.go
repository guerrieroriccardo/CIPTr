package models

// PatchPanel represents a patch panel at a site.
type PatchPanel struct {
	ID         int64   `json:"id"`
	SiteID     int64   `json:"site_id"`
	Name       string  `json:"name"`
	TotalPorts int     `json:"total_ports"`
	Location   *string `json:"location"`
	Notes      *string `json:"notes"`
}

// PatchPanelInput is used for create and update requests.
type PatchPanelInput struct {
	SiteID     int64   `json:"site_id"     binding:"required"`
	Name       string  `json:"name"        binding:"required"`
	TotalPorts *int    `json:"total_ports"`
	Location   *string `json:"location"`
	Notes      *string `json:"notes"`
}
