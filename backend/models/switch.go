package models

// Switch represents a network switch deployed at a site.
type Switch struct {
	ID         int64   `json:"id"`
	SiteID     int64   `json:"site_id"`
	Name       string  `json:"name"`
	ModelID    *int64  `json:"model_id"`
	IPAddress  *string `json:"ip_address"`
	Location   *string `json:"location"`
	TotalPorts int     `json:"total_ports"`
	Notes      *string `json:"notes"`
}

// SwitchInput is used for create and update requests.
type SwitchInput struct {
	SiteID     int64   `json:"site_id"     binding:"required"`
	Name       string  `json:"name"        binding:"required"`
	ModelID    *int64  `json:"model_id"`
	IPAddress  *string `json:"ip_address"`
	Location   *string `json:"location"`
	TotalPorts *int    `json:"total_ports"`
	Notes      *string `json:"notes"`
}
