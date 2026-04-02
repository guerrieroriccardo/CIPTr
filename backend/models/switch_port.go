package models

// SwitchPort represents a single port on a network switch (which is a device).
type SwitchPort struct {
	ID              int64   `json:"id"`
	DeviceID        int64   `json:"device_id"`
	PortNumber      int     `json:"port_number"`
	PortLabel       *string `json:"port_label"`
	Speed           *string `json:"speed"`
	IsUplink        *bool   `json:"is_uplink"`
	MacRestriction  *string `json:"mac_restriction"`
	UntaggedVlanID  *int64  `json:"untagged_vlan_id"`
	TaggedVlanIDs   []int64 `json:"tagged_vlan_ids"`
	IsDisabled      *bool   `json:"is_disabled"`
	Notes           *string `json:"notes"`

	// Read-only enrichment (populated on GET)
	ConnectedDevice        *string `json:"connected_device,omitempty"`
	ConnectedInterface     *string `json:"connected_interface,omitempty"`
	ConnectedPatchPanel    *string `json:"connected_patch_panel,omitempty"`
	ConnectedPatchPanelPort *int   `json:"connected_patch_panel_port,omitempty"`
}

// SwitchPortInput is used for create and update requests.
type SwitchPortInput struct {
	DeviceID        int64   `json:"device_id"   binding:"required"`
	PortNumber      int     `json:"port_number" binding:"required"`
	PortLabel       *string `json:"port_label"`
	Speed           *string `json:"speed"`
	IsUplink        *bool   `json:"is_uplink"`
	MacRestriction  *string `json:"mac_restriction"`
	UntaggedVlanID  *int64  `json:"untagged_vlan_id"`
	TaggedVlanIDs   []int64 `json:"tagged_vlan_ids"`
	IsDisabled      *bool   `json:"is_disabled"`
	Notes           *string `json:"notes"`
}
