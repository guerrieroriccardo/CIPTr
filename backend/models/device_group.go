package models

import "time"

// DeviceGroup represents a named group of devices at a site.
type DeviceGroup struct {
	ID          int64     `json:"id"`
	SiteID      int64     `json:"site_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

// DeviceGroupInput is used for create and update requests.
type DeviceGroupInput struct {
	SiteID      int64   `json:"site_id"  binding:"required"`
	Name        string  `json:"name"     binding:"required"`
	Description *string `json:"description"`
	Notes       *string `json:"notes"`
}

// DeviceGroupMember represents a device's membership in a group.
type DeviceGroupMember struct {
	ID       int64 `json:"id"`
	GroupID  int64 `json:"group_id"`
	DeviceID int64 `json:"device_id"`
}

// DeviceGroupMemberInput is used for adding a device to a group.
type DeviceGroupMemberInput struct {
	GroupID  int64 `json:"group_id"  binding:"required"`
	DeviceID int64 `json:"device_id" binding:"required"`
}
