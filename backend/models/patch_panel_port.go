package models

// PatchPanelPort represents a single port on a patch panel (which is a device).
type PatchPanelPort struct {
	ID           int64   `json:"id"`
	DeviceID     int64   `json:"device_id"`
	PortNumber   int     `json:"port_number"`
	PortLabel    *string `json:"port_label"`
	LinkedPortID *int64  `json:"linked_port_id"`
	SwitchPortID *int64  `json:"switch_port_id"`
	Notes        *string `json:"notes"`

	// Read-only enrichment (populated on GET)
	ConnectedDevice    *string `json:"connected_device,omitempty"`
	ConnectedInterface *string `json:"connected_interface,omitempty"`
	ConnectedSwitch    *string `json:"connected_switch,omitempty"`
	ConnectedSwitchPort *int   `json:"connected_switch_port,omitempty"`
}

// PatchPanelPortInput is used for create and update requests.
type PatchPanelPortInput struct {
	DeviceID     int64   `json:"device_id"   binding:"required"`
	PortNumber   int     `json:"port_number" binding:"required"`
	PortLabel    *string `json:"port_label"`
	LinkedPortID *int64  `json:"linked_port_id"`
	SwitchPortID *int64  `json:"switch_port_id"`
	Notes        *string `json:"notes"`
}
