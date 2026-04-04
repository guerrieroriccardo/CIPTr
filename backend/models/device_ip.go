package models

// DeviceIP represents an IP address assigned to a device interface.
type DeviceIP struct {
	ID          int64   `json:"id"`
	InterfaceID int64   `json:"interface_id"`
	IPAddress   string  `json:"ip_address"`
	VlanID      *int64  `json:"vlan_id"`
	IsPrimary   *bool   `json:"is_primary"`
	Notes       *string `json:"notes"`
}

// DeviceIPInput is used for create and update requests.
type DeviceIPInput struct {
	InterfaceID  int64   `json:"interface_id" binding:"required"`
	IPAddress    string  `json:"ip_address"   binding:"required"`
	VlanID       *int64  `json:"vlan_id"`
	IsPrimary    *bool   `json:"is_primary"`
	Notes        *string `json:"notes"`
	SetAsGateway *bool   `json:"set_as_gateway"`
}
