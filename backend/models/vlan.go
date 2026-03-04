package models

// VLAN represents a subnet carved from an address block at a site.
// ID is the database primary key; VlanID is the 802.1Q VLAN number (e.g. 10, 100).
type VLAN struct {
	ID                int64   `json:"id"`
	SiteID            int64   `json:"site_id"`
	AddressBlockID    *int64  `json:"address_block_id"`     // nullable: VLAN may not belong to a block
	VlanID            int64   `json:"vlan_id"`              // the actual VLAN tag number
	Name              string  `json:"name"`
	Subnet            *string `json:"subnet"`               // e.g. "10.10.0.0/24"
	GatewayDeviceIPID *int64  `json:"gateway_device_ip_id"` // FK to device_ips
	Description       *string `json:"description"`          // nullable
}

// VLANInput is used for create and update requests.
type VLANInput struct {
	SiteID            int64   `json:"site_id" binding:"required"`
	AddressBlockID    *int64  `json:"address_block_id"`
	VlanID            int64   `json:"vlan_id" binding:"required"`
	Name              string  `json:"name"    binding:"required"`
	Subnet            *string `json:"subnet"`
	GatewayDeviceIPID *int64  `json:"gateway_device_ip_id"`
	Description       *string `json:"description"`
}
