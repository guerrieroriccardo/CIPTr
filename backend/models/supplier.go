package models

import "time"

// Supplier represents a company that supplies devices.
type Supplier struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   *string   `json:"address"`
	Phone     *string   `json:"phone"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// SupplierInput is used for create and update requests.
type SupplierInput struct {
	Name    string  `json:"name" binding:"required"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
	Email   *string `json:"email"`
}
