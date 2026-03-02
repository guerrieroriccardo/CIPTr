package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// SwitchHandler groups all HTTP handlers for the /switches resource.
type SwitchHandler struct {
	db *sql.DB
}

// NewSwitchHandler creates a SwitchHandler with the given database connection.
func NewSwitchHandler(db *sql.DB) *SwitchHandler {
	return &SwitchHandler{db: db}
}

// switchSelectSQL is the base SELECT used by every read operation.
const switchSelectSQL = `SELECT id, site_id, name, model_id, ip_address, location, total_ports, notes FROM switches`

// scanSwitch reads one row into a Switch struct.
func scanSwitch(row interface{ Scan(...any) error }) (models.Switch, error) {
	var s models.Switch
	err := row.Scan(&s.ID, &s.SiteID, &s.Name, &s.ModelID, &s.IPAddress, &s.Location, &s.TotalPorts, &s.Notes)
	return s, err
}

// List handles GET /switches
// Supports optional query param: ?site_id=
func (h *SwitchHandler) List(c *gin.Context) {
	query := switchSelectSQL
	args := []any{}

	if siteID := c.Query("site_id"); siteID != "" {
		query += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	query += ` ORDER BY name`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	switches := []models.Switch{}
	for rows.Next() {
		s, err := scanSwitch(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		switches = append(switches, s)
	}

	ok(c, http.StatusOK, switches)
}

// ListBySite handles GET /sites/:id/switches
func (h *SwitchHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		switchSelectSQL+` WHERE site_id = $1 ORDER BY name`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	switches := []models.Switch{}
	for rows.Next() {
		s, err := scanSwitch(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		switches = append(switches, s)
	}

	ok(c, http.StatusOK, switches)
}

// GetByID handles GET /switches/:id
func (h *SwitchHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	s, err := scanSwitch(h.db.QueryRowContext(c.Request.Context(),
		switchSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, s)
}

// Create handles POST /switches
func (h *SwitchHandler) Create(c *gin.Context) {
	var input models.SwitchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	totalPorts := 24
	if input.TotalPorts != nil {
		totalPorts = *input.TotalPorts
	}

	s, err := scanSwitch(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO switches (site_id, name, model_id, ip_address, location, total_ports, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, site_id, name, model_id, ip_address, location, total_ports, notes`,
		input.SiteID, input.Name, input.ModelID, input.IPAddress, input.Location, totalPorts, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, s)
}

// Update handles PUT /switches/:id
func (h *SwitchHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.SwitchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	totalPorts := 24
	if input.TotalPorts != nil {
		totalPorts = *input.TotalPorts
	}

	s, err := scanSwitch(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE switches SET site_id = $1, name = $2, model_id = $3, ip_address = $4, location = $5, total_ports = $6, notes = $7
		 WHERE id = $8
		 RETURNING id, site_id, name, model_id, ip_address, location, total_ports, notes`,
		input.SiteID, input.Name, input.ModelID, input.IPAddress, input.Location, totalPorts, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, s)
}

// Delete handles DELETE /switches/:id
// Cascades to switch_ports via DB FK.
func (h *SwitchHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM switches WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
