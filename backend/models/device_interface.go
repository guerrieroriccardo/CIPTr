package models

// DeviceInterface represents a physical or virtual NIC on a device
// (e.g. eth0, iDRAC, WAN, LAN1).
type DeviceInterface struct {
	ID         int64   `json:"id"`
	DeviceID   int64   `json:"device_id"`
	Name       string  `json:"name"`
	MacAddress *string `json:"mac_address"`
	Notes      *string `json:"notes"`
}

// DeviceInterfaceInput is used for create and update requests.
type DeviceInterfaceInput struct {
	DeviceID   int64   `json:"device_id"   binding:"required"`
	Name       string  `json:"name"        binding:"required"`
	MacAddress *string `json:"mac_address"`
	Notes      *string `json:"notes"`
}
