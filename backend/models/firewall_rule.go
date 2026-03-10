package models

import "time"

// FirewallRule represents a firewall rule at a site.
type FirewallRule struct {
	ID          int64     `json:"id"`
	SiteID      int64     `json:"site_id"`
	SrcDeviceID *int64    `json:"src_device_id"`
	SrcGroupID  *int64    `json:"src_group_id"`
	SrcVlanID   *int64    `json:"src_vlan_id"`
	SrcCIDR     *string   `json:"src_cidr"`
	DstDeviceID *int64    `json:"dst_device_id"`
	DstGroupID  *int64    `json:"dst_group_id"`
	DstVlanID   *int64    `json:"dst_vlan_id"`
	DstCIDR     *string   `json:"dst_cidr"`
	SrcPort     string    `json:"src_port"`
	DstPort     string    `json:"dst_port"`
	Protocol    string    `json:"protocol"`
	Action      string    `json:"action"`
	Position    int       `json:"position"`
	Enabled     bool      `json:"enabled"`
	Description *string   `json:"description"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FirewallRuleInput is used for create and update requests.
type FirewallRuleInput struct {
	SiteID      int64   `json:"site_id"  binding:"required"`
	SrcDeviceID *int64  `json:"src_device_id"`
	SrcGroupID  *int64  `json:"src_group_id"`
	SrcVlanID   *int64  `json:"src_vlan_id"`
	SrcCIDR     *string `json:"src_cidr"`
	DstDeviceID *int64  `json:"dst_device_id"`
	DstGroupID  *int64  `json:"dst_group_id"`
	DstVlanID   *int64  `json:"dst_vlan_id"`
	DstCIDR     *string `json:"dst_cidr"`
	SrcPort     *string `json:"src_port"`
	DstPort     *string `json:"dst_port"`
	Protocol    *string `json:"protocol"`
	Action      *string `json:"action"`
	Position    *int    `json:"position"`
	Enabled     *bool   `json:"enabled"`
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
}
