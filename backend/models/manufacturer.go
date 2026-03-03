package models

import "time"

// Manufacturer represents a hardware manufacturer (e.g. HP, Cisco, Dell).
type Manufacturer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ManufacturerInput is used for create and update requests.
type ManufacturerInput struct {
	Name string `json:"name" binding:"required"`
}
