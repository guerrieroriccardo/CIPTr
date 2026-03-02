package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// DeviceInterfaceHandler groups all HTTP handlers for the /device-interfaces resource.
type DeviceInterfaceHandler struct {
	db *sql.DB
}

// NewDeviceInterfaceHandler creates a DeviceInterfaceHandler with the given database connection.
func NewDeviceInterfaceHandler(db *sql.DB) *DeviceInterfaceHandler {
	return &DeviceInterfaceHandler{db: db}
}

// deviceInterfaceSelectSQL is the base SELECT used by every read operation.
const deviceInterfaceSelectSQL = `SELECT id, device_id, name, mac_address, notes FROM device_interfaces`

// scanDeviceInterface reads one row into a DeviceInterface struct.
func scanDeviceInterface(row interface{ Scan(...any) error }) (models.DeviceInterface, error) {
	var di models.DeviceInterface
	err := row.Scan(&di.ID, &di.DeviceID, &di.Name, &di.MacAddress, &di.Notes)
	return di, err
}

// List handles GET /device-interfaces
// Supports optional query param: ?device_id=
func (h *DeviceInterfaceHandler) List(c *gin.Context) {
	query := deviceInterfaceSelectSQL
	args := []any{}

	if deviceID := c.Query("device_id"); deviceID != "" {
		query += ` WHERE device_id = $1`
		args = append(args, deviceID)
	}
	query += ` ORDER BY name`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	interfaces := []models.DeviceInterface{}
	for rows.Next() {
		di, err := scanDeviceInterface(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		interfaces = append(interfaces, di)
	}

	ok(c, http.StatusOK, interfaces)
}

// ListByDevice handles GET /devices/:id/interfaces
// Returns all interfaces for the given device, ordered by name.
func (h *DeviceInterfaceHandler) ListByDevice(c *gin.Context) {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		deviceInterfaceSelectSQL+` WHERE device_id = $1 ORDER BY name`, deviceID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	interfaces := []models.DeviceInterface{}
	for rows.Next() {
		di, err := scanDeviceInterface(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		interfaces = append(interfaces, di)
	}

	ok(c, http.StatusOK, interfaces)
}

// GetByID handles GET /device-interfaces/:id
// Returns 404 if the interface does not exist.
func (h *DeviceInterfaceHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	di, err := scanDeviceInterface(h.db.QueryRowContext(c.Request.Context(),
		deviceInterfaceSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device interface not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, di)
}

// Create handles POST /device-interfaces
// device_id and name are required.
func (h *DeviceInterfaceHandler) Create(c *gin.Context) {
	var input models.DeviceInterfaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	di, err := scanDeviceInterface(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO device_interfaces (device_id, name, mac_address, notes)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, device_id, name, mac_address, notes`,
		input.DeviceID, input.Name, input.MacAddress, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, di)
}

// Update handles PUT /device-interfaces/:id
// Replaces all fields. Returns 404 if the interface does not exist.
func (h *DeviceInterfaceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.DeviceInterfaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	di, err := scanDeviceInterface(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE device_interfaces SET device_id = $1, name = $2, mac_address = $3, notes = $4
		 WHERE id = $5
		 RETURNING id, device_id, name, mac_address, notes`,
		input.DeviceID, input.Name, input.MacAddress, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device interface not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, di)
}

// Delete handles DELETE /device-interfaces/:id
// Cascades to device_ips and device_connections via DB FK.
func (h *DeviceInterfaceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM device_interfaces WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device interface not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
