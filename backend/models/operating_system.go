package models

import "time"

// OperatingSystem represents an OS entry (e.g. "Windows Server 2022", "FortiOS 7.4").
type OperatingSystem struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// OperatingSystemInput is used for create and update requests.
type OperatingSystemInput struct {
	Name string `json:"name" binding:"required"`
}
