package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// SiteHandler groups all HTTP handlers for the /sites resource.
//
// Same pattern as ClientHandler: the database connection is injected once
// at startup so every method can reach it via h.db.
type SiteHandler struct {
	db *sql.DB
}

// NewSiteHandler creates a SiteHandler with the given database connection.
func NewSiteHandler(db *sql.DB) *SiteHandler {
	return &SiteHandler{db: db}
}

// siteSelectSQL is the base SELECT used by every read operation.
// Defined as a constant to avoid repeating the column list.
const siteSelectSQL = `SELECT id, client_id, name, address, notes, created_at FROM sites`

// scanSite reads one row (from Query or QueryRow) into a Site struct.
//
// The parameter type accepts both *sql.Rows (from QueryContext) and
// *sql.Row (from QueryRowContext) because both implement Scan(...any).
// This avoids duplicating the column list in every handler.
func scanSite(row interface{ Scan(...any) error }) (models.Site, error) {
	var s models.Site
	err := row.Scan(&s.ID, &s.ClientID, &s.Name, &s.Address, &s.Notes, &s.CreatedAt)
	return s, err
}

// List handles GET /sites
// Supports optional query param: ?client_id=1
// Without the param, returns all sites across all clients.
func (h *SiteHandler) List(c *gin.Context) {
	query := siteSelectSQL
	args := []any{}

	// c.Query("client_id") reads the value from the URL query string.
	// e.g. GET /sites?client_id=3 → clientID = "3"
	// If the param is absent it returns "", so we only add the WHERE clause when set.
	if clientID := c.Query("client_id"); clientID != "" {
		query += ` WHERE client_id = ?`
		args = append(args, clientID)
	}
	query += ` ORDER BY name`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	sites := []models.Site{} // empty slice, not nil → JSON returns [] not null
	for rows.Next() {
		s, err := scanSite(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		sites = append(sites, s)
	}

	ok(c, http.StatusOK, sites)
}

// ListByClient handles GET /clients/:id/sites
// Convenience nested route equivalent to GET /sites?client_id=:id.
// Returns 200 with an empty list if the client exists but has no sites.
// (Does not verify that the client itself exists — a 404 for unknown clients
// is not strictly necessary here since the result is just an empty list.)
func (h *SiteHandler) ListByClient(c *gin.Context) {
	// :id here refers to the client id from the parent route /clients/:id/sites
	clientID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid client id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		siteSelectSQL+` WHERE client_id = ? ORDER BY name`, clientID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	sites := []models.Site{}
	for rows.Next() {
		s, err := scanSite(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		sites = append(sites, s)
	}

	ok(c, http.StatusOK, sites)
}

// GetByID handles GET /sites/:id
// Returns 404 if the site does not exist.
func (h *SiteHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	s, err := scanSite(h.db.QueryRowContext(c.Request.Context(),
		siteSelectSQL+` WHERE id = ?`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("site not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, s)
}

// Create handles POST /sites
// client_id and name are required. Fails with 500 if client_id does not exist
// (SQLite foreign key constraint).
func (h *SiteHandler) Create(c *gin.Context) {
	var input models.SiteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	res, err := h.db.ExecContext(c.Request.Context(),
		`INSERT INTO sites (client_id, name, address, notes) VALUES (?, ?, ?, ?)`,
		input.ClientID, input.Name, input.Address, input.Notes,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := res.LastInsertId()
	s, _ := scanSite(h.db.QueryRowContext(c.Request.Context(),
		siteSelectSQL+` WHERE id = ?`, newID))

	ok(c, http.StatusCreated, s)
}

// Update handles PUT /sites/:id
// Replaces all fields. client_id can be changed to move the site to a different client.
func (h *SiteHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.SiteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	res, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE sites SET client_id = ?, name = ?, address = ?, notes = ? WHERE id = ?`,
		input.ClientID, input.Name, input.Address, input.Notes, id,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		fail(c, http.StatusNotFound, errors.New("site not found"))
		return
	}

	s, _ := scanSite(h.db.QueryRowContext(c.Request.Context(),
		siteSelectSQL+` WHERE id = ?`, id))

	ok(c, http.StatusOK, s)
}

// Delete handles DELETE /sites/:id
// Cascades to offices, vlans, switches, patch_panels and all their children.
func (h *SiteHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	res, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM sites WHERE id = ?`, id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		fail(c, http.StatusNotFound, errors.New("site not found"))
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
