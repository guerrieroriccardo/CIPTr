package models

// DeviceConnection represents a physical cable link from a device interface
// to a switch port and/or patch panel port.
type DeviceConnection struct {
	ID                int64   `json:"id"`
	InterfaceID       int64   `json:"interface_id"`
	SwitchPortID      *int64  `json:"switch_port_id"`
	PatchPanelPortID  *int64  `json:"patch_panel_port_id"`
	ConnectedAt       *string `json:"connected_at"` // DATE as string (YYYY-MM-DD)
	Notes             *string `json:"notes"`
}

// DeviceConnectionInput is used for create and update requests.
type DeviceConnectionInput struct {
	InterfaceID       int64   `json:"interface_id"        binding:"required"`
	SwitchPortID      *int64  `json:"switch_port_id"`
	PatchPanelPortID  *int64  `json:"patch_panel_port_id"`
	ConnectedAt       *string `json:"connected_at"`
	Notes             *string `json:"notes"`
}
