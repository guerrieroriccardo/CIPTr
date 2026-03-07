package models

// SwitchPort represents a single port on a network switch.
type SwitchPort struct {
	ID             int64   `json:"id"`
	SwitchID       int64   `json:"switch_id"`
	PortNumber     int     `json:"port_number"`
	PortLabel      *string `json:"port_label"`
	Speed          *string `json:"speed"`
	IsUplink       *bool   `json:"is_uplink"`
	MacRestriction *string `json:"mac_restriction"`
	Notes          *string `json:"notes"`
}

// SwitchPortInput is used for create and update requests.
type SwitchPortInput struct {
	SwitchID       int64   `json:"switch_id"   binding:"required"`
	PortNumber     int     `json:"port_number" binding:"required"`
	PortLabel      *string `json:"port_label"`
	Speed          *string `json:"speed"`
	IsUplink       *bool   `json:"is_uplink"`
	MacRestriction *string `json:"mac_restriction"`
	Notes          *string `json:"notes"`
}
