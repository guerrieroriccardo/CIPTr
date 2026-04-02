package models

// PatchPanelPort represents a single port on a patch panel (which is a device).
type PatchPanelPort struct {
	ID           int64   `json:"id"`
	DeviceID     int64   `json:"device_id"`
	PortNumber   int     `json:"port_number"`
	PortLabel    *string `json:"port_label"`
	LinkedPortID *int64  `json:"linked_port_id"`
	Notes        *string `json:"notes"`
}

// PatchPanelPortInput is used for create and update requests.
type PatchPanelPortInput struct {
	DeviceID     int64   `json:"device_id"   binding:"required"`
	PortNumber   int     `json:"port_number" binding:"required"`
	PortLabel    *string `json:"port_label"`
	LinkedPortID *int64  `json:"linked_port_id"`
	Notes        *string `json:"notes"`
}
