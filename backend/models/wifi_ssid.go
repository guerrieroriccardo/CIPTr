package models

// WifiSSID represents a wireless network configured at a site.
type WifiSSID struct {
	ID     int64   `json:"id"`
	SiteID int64   `json:"site_id"`
	SSID   string  `json:"ssid"`
	Auth   *string `json:"auth"`
	VlanID *int64  `json:"vlan_id"`
	Notes  *string `json:"notes"`
}

// WifiSSIDInput is used for create and update requests.
type WifiSSIDInput struct {
	SiteID int64   `json:"site_id" binding:"required"`
	SSID   string  `json:"ssid"    binding:"required"`
	Auth   *string `json:"auth"`
	VlanID *int64  `json:"vlan_id"`
	Notes  *string `json:"notes"`
}
