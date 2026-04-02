package models

// SwitchPort represents a single port on a network switch (which is a device).
type SwitchPort struct {
	ID             int64   `json:"id"`
	DeviceID       int64   `json:"device_id"`
	PortNumber     int     `json:"port_number"`
	PortLabel      *string `json:"port_label"`
	Speed          *string `json:"speed"`
	IsUplink       *bool   `json:"is_uplink"`
	MacRestriction *string `json:"mac_restriction"`
	Notes          *string `json:"notes"`
}

// SwitchPortInput is used for create and update requests.
type SwitchPortInput struct {
	DeviceID       int64   `json:"device_id"   binding:"required"`
	PortNumber     int     `json:"port_number" binding:"required"`
	PortLabel      *string `json:"port_label"`
	Speed          *string `json:"speed"`
	IsUplink       *bool   `json:"is_uplink"`
	MacRestriction *string `json:"mac_restriction"`
	Notes          *string `json:"notes"`
}
