package models

// Client represents a managed client company.
type Client struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	ShortCode string  `json:"short_code"`
	Notes     *string `json:"notes"`      // nullable
	CreatedAt string  `json:"created_at"` // stored as text in SQLite
}

// ClientInput is used for create and update requests.
type ClientInput struct {
	Name      string  `json:"name"       binding:"required"`
	ShortCode string  `json:"short_code" binding:"required"`
	Notes     *string `json:"notes"`
}
