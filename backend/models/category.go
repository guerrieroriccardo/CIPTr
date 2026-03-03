package models

import "time"

// Category represents a device category (e.g. Server, PC, Switch, Printer).
type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CategoryInput is used for create and update requests.
type CategoryInput struct {
	Name string `json:"name" binding:"required"`
}
