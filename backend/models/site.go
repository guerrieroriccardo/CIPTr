package models

// Site represents a physical location belonging to a client.
type Site struct {
	ID        int64   `json:"id"`
	ClientID  int64   `json:"client_id"`
	Name      string  `json:"name"`
	Address   *string `json:"address"`
	Notes     *string `json:"notes"`
	CreatedAt string  `json:"created_at"`
}

// SiteInput is used for create and update requests.
type SiteInput struct {
	ClientID int64   `json:"client_id" binding:"required"`
	Name     string  `json:"name"      binding:"required"`
	Address  *string `json:"address"`
	Notes    *string `json:"notes"`
}
