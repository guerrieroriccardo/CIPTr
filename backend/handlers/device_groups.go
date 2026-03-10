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

// DeviceGroupHandler groups all HTTP handlers for the /device-groups resource.
type DeviceGroupHandler struct {
	db *sql.DB
}

// NewDeviceGroupHandler creates a DeviceGroupHandler with the given database connection.
func NewDeviceGroupHandler(db *sql.DB) *DeviceGroupHandler {
	return &DeviceGroupHandler{db: db}
}

const deviceGroupColumns = `id, site_id, name, description, notes, created_at`

const deviceGroupSelectSQL = `SELECT ` + deviceGroupColumns + ` FROM device_groups`

// scanDeviceGroup reads one row into a DeviceGroup struct.
func scanDeviceGroup(row interface{ Scan(...any) error }) (models.DeviceGroup, error) {
	var g models.DeviceGroup
	err := row.Scan(&g.ID, &g.SiteID, &g.Name, &g.Description, &g.Notes, &g.CreatedAt)
	return g, err
}

// List handles GET /device-groups
// Supports optional query param: ?site_id=1
func (h *DeviceGroupHandler) List(c *gin.Context) {
	query := deviceGroupSelectSQL
	var conds []string
	var args []any
	n := 1

	if siteID := c.Query("site_id"); siteID != "" {
		conds = append(conds, fmt.Sprintf("site_id = $%d", n))
		args = append(args, siteID)
		n++
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY name"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	groups := []models.DeviceGroup{}
	for rows.Next() {
		g, err := scanDeviceGroup(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		groups = append(groups, g)
	}

	ok(c, http.StatusOK, groups)
}

// ListBySite handles GET /sites/:id/device-groups
func (h *DeviceGroupHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		deviceGroupSelectSQL+` WHERE site_id = $1 ORDER BY name`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	groups := []models.DeviceGroup{}
	for rows.Next() {
		g, err := scanDeviceGroup(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		groups = append(groups, g)
	}

	ok(c, http.StatusOK, groups)
}

// GetByID handles GET /device-groups/:id
func (h *DeviceGroupHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	g, err := scanDeviceGroup(h.db.QueryRowContext(c.Request.Context(),
		deviceGroupSelectSQL+` WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device group not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, g)
}

// Create handles POST /device-groups
func (h *DeviceGroupHandler) Create(c *gin.Context) {
	var input models.DeviceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	g, err := scanDeviceGroup(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO device_groups (site_id, name, description, notes)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+deviceGroupColumns,
		input.SiteID, input.Name, input.Description, input.Notes,
	))
	if err != nil {
		if strings.Contains(err.Error(), "device_groups_site_id_name_key") {
			fail(c, http.StatusBadRequest, fmt.Errorf("group '%s' already exists in this site", input.Name))
			return
		}
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "device_groups", g.ID, fmt.Sprintf("Created device group '%s'", g.Name))
	ok(c, http.StatusCreated, g)
}

// Update handles PUT /device-groups/:id
func (h *DeviceGroupHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.DeviceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	g, err := scanDeviceGroup(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE device_groups SET site_id = $1, name = $2, description = $3, notes = $4
		 WHERE id = $5
		 RETURNING `+deviceGroupColumns,
		input.SiteID, input.Name, input.Description, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device group not found"))
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "device_groups_site_id_name_key") {
			fail(c, http.StatusBadRequest, fmt.Errorf("group '%s' already exists in this site", input.Name))
			return
		}
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "device_groups", id, fmt.Sprintf("Updated device group '%s'", g.Name))
	ok(c, http.StatusOK, g)
}

// Delete handles DELETE /device-groups/:id
func (h *DeviceGroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM device_groups WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device group not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "device_groups", id, fmt.Sprintf("Deleted device group #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
