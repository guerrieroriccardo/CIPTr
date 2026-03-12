package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
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
const switchSelectSQL = `SELECT id, site_id, hostname, model_id, ip_address, location_id, total_ports, notes FROM switches`

// scanSwitch reads one row into a Switch struct.
func scanSwitch(row interface{ Scan(...any) error }) (models.Switch, error) {
	var s models.Switch
	err := row.Scan(&s.ID, &s.SiteID, &s.Hostname, &s.ModelID, &s.IPAddress, &s.LocationID, &s.TotalPorts, &s.Notes)
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
	query += ` ORDER BY hostname`

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
		switchSelectSQL+` WHERE site_id = $1 ORDER BY hostname`, siteID)
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

// NextName handles GET /switches/next-name?site_id=X
// Returns the next available switch name for the given site (e.g. SW001, SW002).
func (h *SwitchHandler) NextName(c *gin.Context) {
	siteIDStr := c.Query("site_id")
	if siteIDStr == "" {
		fail(c, http.StatusBadRequest, errors.New("site_id is required"))
		return
	}
	siteID, err := strconv.ParseInt(siteIDStr, 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site_id"))
		return
	}

	// Load hostname format from settings.
	format, err := GetHostnameFormat(c.Request.Context(), h.db)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	const label = "SW"

	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT hostname FROM switches WHERE site_id = $1 AND hostname LIKE $2`,
		siteID, HostnameLikePattern(label, format),
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	maxNum := MaxHostnameNumber(format.NumDigits)
	taken := make(map[int]bool)
	for rows.Next() {
		var hostname string
		if err := rows.Scan(&hostname); err != nil {
			continue
		}
		if num, ok := ParseHostnameNumber(hostname, label, format); ok && num >= 1 && num <= maxNum {
			taken[num] = true
		}
	}

	next := 0
	for i := 1; i <= maxNum; i++ {
		if !taken[i] {
			next = i
			break
		}
	}
	if next == 0 {
		fail(c, http.StatusConflict, fmt.Errorf("all %d switch names are taken", maxNum))
		return
	}

	hostname := BuildHostname(label, next, format)
	ok(c, http.StatusOK, gin.H{"hostname": hostname})
}

// Create handles POST /switches
// Automatically creates N switch_port rows based on total_ports.
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

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	s, err := scanSwitch(tx.QueryRowContext(c.Request.Context(),
		`INSERT INTO switches (site_id, hostname, model_id, ip_address, location_id, total_ports, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, site_id, hostname, model_id, ip_address, location_id, total_ports, notes`,
		input.SiteID, input.Hostname, input.ModelID, input.IPAddress, input.LocationID, totalPorts, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	// Auto-create port rows.
	if totalPorts > 0 {
		var vals []string
		var args []any
		for i := 1; i <= totalPorts; i++ {
			vals = append(vals, fmt.Sprintf("($%d, $%d)", i*2-1, i*2))
			args = append(args, s.ID, i)
		}
		_, err = tx.ExecContext(c.Request.Context(),
			`INSERT INTO switch_ports (switch_id, port_number) VALUES `+strings.Join(vals, ", "),
			args...)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "switches", s.ID, fmt.Sprintf("Created switch '%s'", s.Hostname))
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
		`UPDATE switches SET site_id = $1, hostname = $2, model_id = $3, ip_address = $4, location_id = $5, total_ports = $6, notes = $7
		 WHERE id = $8
		 RETURNING id, site_id, hostname, model_id, ip_address, location_id, total_ports, notes`,
		input.SiteID, input.Hostname, input.ModelID, input.IPAddress, input.LocationID, totalPorts, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "switches", id, fmt.Sprintf("Updated switch '%s'", s.Hostname))
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

	logAudit(c.Request.Context(), h.db, c, "delete", "switches", id, fmt.Sprintf("Deleted switch #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
