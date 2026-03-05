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

// DeviceConnectionHandler groups all HTTP handlers for the /device-connections resource.
type DeviceConnectionHandler struct {
	db *sql.DB
}

// NewDeviceConnectionHandler creates a DeviceConnectionHandler with the given database connection.
func NewDeviceConnectionHandler(db *sql.DB) *DeviceConnectionHandler {
	return &DeviceConnectionHandler{db: db}
}

// deviceConnectionSelectSQL is the base SELECT used by every read operation.
const deviceConnectionSelectSQL = `SELECT id, interface_id, switch_port_id, patch_panel_port_id, connected_at, notes FROM device_connections`

// scanDeviceConnection reads one row into a DeviceConnection struct.
func scanDeviceConnection(row interface{ Scan(...any) error }) (models.DeviceConnection, error) {
	var dc models.DeviceConnection
	err := row.Scan(&dc.ID, &dc.InterfaceID, &dc.SwitchPortID, &dc.PatchPanelPortID, &dc.ConnectedAt, &dc.Notes)
	return dc, err
}

// List handles GET /device-connections
// Supports optional query params: ?interface_id=, ?switch_port_id=, ?patch_panel_port_id=
func (h *DeviceConnectionHandler) List(c *gin.Context) {
	query := deviceConnectionSelectSQL
	var conds []string
	var args []any
	n := 1

	if ifID := c.Query("interface_id"); ifID != "" {
		conds = append(conds, fmt.Sprintf("interface_id = $%d", n))
		args = append(args, ifID)
		n++
	}
	if spID := c.Query("switch_port_id"); spID != "" {
		conds = append(conds, fmt.Sprintf("switch_port_id = $%d", n))
		args = append(args, spID)
		n++
	}
	if ppID := c.Query("patch_panel_port_id"); ppID != "" {
		conds = append(conds, fmt.Sprintf("patch_panel_port_id = $%d", n))
		args = append(args, ppID)
		n++
	}

	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	conns := []models.DeviceConnection{}
	for rows.Next() {
		dc, err := scanDeviceConnection(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		conns = append(conns, dc)
	}

	ok(c, http.StatusOK, conns)
}

// ListByDevice handles GET /devices/:id/connections
// Returns all connections for all interfaces of the given device.
func (h *DeviceConnectionHandler) ListByDevice(c *gin.Context) {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		deviceConnectionSelectSQL+` WHERE interface_id IN (SELECT id FROM device_interfaces WHERE device_id = $1) ORDER BY id`,
		deviceID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	conns := []models.DeviceConnection{}
	for rows.Next() {
		dc, err := scanDeviceConnection(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		conns = append(conns, dc)
	}

	ok(c, http.StatusOK, conns)
}

// GetByID handles GET /device-connections/:id
// Returns 404 if the connection does not exist.
func (h *DeviceConnectionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	dc, err := scanDeviceConnection(h.db.QueryRowContext(c.Request.Context(),
		deviceConnectionSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device connection not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, dc)
}

// Create handles POST /device-connections
// interface_id is required.
func (h *DeviceConnectionHandler) Create(c *gin.Context) {
	var input models.DeviceConnectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	dc, err := scanDeviceConnection(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO device_connections (interface_id, switch_port_id, patch_panel_port_id, connected_at, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, interface_id, switch_port_id, patch_panel_port_id, connected_at, notes`,
		input.InterfaceID, input.SwitchPortID, input.PatchPanelPortID, input.ConnectedAt, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "device_connections", dc.ID, fmt.Sprintf("Created connection #%d", dc.ID))
	ok(c, http.StatusCreated, dc)
}

// Update handles PUT /device-connections/:id
// Replaces all fields. Returns 404 if the connection does not exist.
func (h *DeviceConnectionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.DeviceConnectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	dc, err := scanDeviceConnection(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE device_connections SET interface_id = $1, switch_port_id = $2, patch_panel_port_id = $3, connected_at = $4, notes = $5
		 WHERE id = $6
		 RETURNING id, interface_id, switch_port_id, patch_panel_port_id, connected_at, notes`,
		input.InterfaceID, input.SwitchPortID, input.PatchPanelPortID, input.ConnectedAt, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device connection not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "device_connections", id, fmt.Sprintf("Updated connection #%d", id))
	ok(c, http.StatusOK, dc)
}

// Delete handles DELETE /device-connections/:id
func (h *DeviceConnectionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM device_connections WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device connection not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "device_connections", id, fmt.Sprintf("Deleted connection #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
