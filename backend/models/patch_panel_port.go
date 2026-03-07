package models

// PatchPanelPort represents a single port on a patch panel.
type PatchPanelPort struct {
	ID           int64   `json:"id"`
	PatchPanelID int64   `json:"patch_panel_id"`
	PortNumber   int     `json:"port_number"`
	PortLabel    *string `json:"port_label"`
	LinkedPortID *int64  `json:"linked_port_id"`
	Notes        *string `json:"notes"`
}

// PatchPanelPortInput is used for create and update requests.
type PatchPanelPortInput struct {
	PatchPanelID int64   `json:"patch_panel_id" binding:"required"`
	PortNumber   int     `json:"port_number"    binding:"required"`
	PortLabel    *string `json:"port_label"`
	LinkedPortID *int64  `json:"linked_port_id"`
	Notes        *string `json:"notes"`
}
