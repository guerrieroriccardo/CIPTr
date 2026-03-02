package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// SiteHandler handles all /sites and /clients/:id/sites routes.
type SiteHandler struct {
	db *sql.DB
}

func NewSiteHandler(db *sql.DB) *SiteHandler {
	return &SiteHandler{db: db}
}

const siteSelectSQL = `SELECT id, client_id, name, address, notes, created_at FROM sites`

func scanSite(row interface{ Scan(...any) error }) (models.Site, error) {
	var s models.Site
	err := row.Scan(&s.ID, &s.ClientID, &s.Name, &s.Address, &s.Notes, &s.CreatedAt)
	return s, err
}

// List handles GET /sites
// Supports optional query param: ?client_id=1
func (h *SiteHandler) List(c *gin.Context) {
	query := siteSelectSQL
	args := []any{}

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

// ListByClient handles GET /clients/:id/sites
// Convenience nested route — returns sites belonging to a specific client.
func (h *SiteHandler) ListByClient(c *gin.Context) {
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
