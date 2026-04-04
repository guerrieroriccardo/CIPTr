package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// DeviceIPHandler groups all HTTP handlers for the /device-ips resource.
type DeviceIPHandler struct {
	db *sql.DB
}

// NewDeviceIPHandler creates a DeviceIPHandler with the given database connection.
func NewDeviceIPHandler(db *sql.DB) *DeviceIPHandler {
	return &DeviceIPHandler{db: db}
}

// deviceIPSelectSQL is the base SELECT used by every read operation.
const deviceIPSelectSQL = `SELECT id, interface_id, ip_address, vlan_id, is_primary, notes FROM device_ips`

// scanDeviceIP reads one row into a DeviceIP struct.
func scanDeviceIP(row interface{ Scan(...any) error }) (models.DeviceIP, error) {
	var d models.DeviceIP
	err := row.Scan(&d.ID, &d.InterfaceID, &d.IPAddress, &d.VlanID, &d.IsPrimary, &d.Notes)
	return d, err
}

func (h *DeviceIPHandler) validateDeviceIP(ctx context.Context, input *models.DeviceIPInput) error {
	if input.VlanID == nil {
		return nil
	}

	var subnet *string
	err := h.db.QueryRowContext(ctx,
		`SELECT subnet FROM vlans WHERE id = $1`, *input.VlanID,
	).Scan(&subnet)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("VLAN %d not found", *input.VlanID)
	}
	if err != nil {
		return err
	}

	if subnet == nil || *subnet == "" {
		return nil
	}

	_, subnetNet, err := net.ParseCIDR(*subnet)
	if err != nil {
		return fmt.Errorf("VLAN has invalid subnet: %s", *subnet)
	}

	ip := net.ParseIP(input.IPAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", input.IPAddress)
	}

	if !subnetNet.Contains(ip) {
		return fmt.Errorf("IP %s is not within VLAN subnet %s", input.IPAddress, *subnet)
	}

	return nil
}

// maybeSetVLANGateway updates the VLAN's gateway to point to the given device IP
// when the input has SetAsGateway=true and a VlanID.
func (h *DeviceIPHandler) maybeSetVLANGateway(ctx context.Context, c *gin.Context, deviceIPID int64, ipAddress string, input *models.DeviceIPInput) error {
	if input.SetAsGateway == nil || !*input.SetAsGateway || input.VlanID == nil {
		return nil
	}
	_, err := h.db.ExecContext(ctx,
		`UPDATE vlans SET gateway_device_ip_id = $1 WHERE id = $2`,
		deviceIPID, *input.VlanID,
	)
	if err != nil {
		return err
	}
	logAudit(ctx, h.db, c, "update", "vlans", *input.VlanID,
		fmt.Sprintf("Set gateway to device IP '%s' (#%d)", ipAddress, deviceIPID))
	return nil
}

// List handles GET /device-ips
// Supports optional query params: ?interface_id=, ?vlan_id=
func (h *DeviceIPHandler) List(c *gin.Context) {
	query := deviceIPSelectSQL
	var conds []string
	var args []any
	n := 1

	if ifID := c.Query("interface_id"); ifID != "" {
		conds = append(conds, fmt.Sprintf("interface_id = $%d", n))
		args = append(args, ifID)
		n++
	}
	if vlanID := c.Query("vlan_id"); vlanID != "" {
		conds = append(conds, fmt.Sprintf("vlan_id = $%d", n))
		args = append(args, vlanID)
		n++
	}

	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY ip_address"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	ips := []models.DeviceIP{}
	for rows.Next() {
		d, err := scanDeviceIP(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		ips = append(ips, d)
	}

	ok(c, http.StatusOK, ips)
}

// ListByDevice handles GET /devices/:id/ips
// Returns all IPs for all interfaces of the given device.
func (h *DeviceIPHandler) ListByDevice(c *gin.Context) {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		deviceIPSelectSQL+` WHERE interface_id IN (SELECT id FROM device_interfaces WHERE device_id = $1) ORDER BY ip_address`,
		deviceID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	ips := []models.DeviceIP{}
	for rows.Next() {
		d, err := scanDeviceIP(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		ips = append(ips, d)
	}

	ok(c, http.StatusOK, ips)
}

// GetByID handles GET /device-ips/:id
// Returns 404 if the IP does not exist.
func (h *DeviceIPHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	d, err := scanDeviceIP(h.db.QueryRowContext(c.Request.Context(),
		deviceIPSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device ip not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, d)
}

// Create handles POST /device-ips
// interface_id and ip_address are required.
func (h *DeviceIPHandler) Create(c *gin.Context) {
	var input models.DeviceIPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validateDeviceIP(c.Request.Context(), &input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	d, err := scanDeviceIP(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO device_ips (interface_id, ip_address, vlan_id, is_primary, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, interface_id, ip_address, vlan_id, is_primary, notes`,
		input.InterfaceID, input.IPAddress, input.VlanID, input.IsPrimary, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "device_ips", d.ID, fmt.Sprintf("Created device IP '%s'", d.IPAddress))
	if err := h.maybeSetVLANGateway(c.Request.Context(), c, d.ID, d.IPAddress, &input); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, http.StatusCreated, d)
}

// Update handles PUT /device-ips/:id
// Replaces all fields. Returns 404 if the IP does not exist.
func (h *DeviceIPHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.DeviceIPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validateDeviceIP(c.Request.Context(), &input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	d, err := scanDeviceIP(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE device_ips SET interface_id = $1, ip_address = $2, vlan_id = $3, is_primary = $4, notes = $5
		 WHERE id = $6
		 RETURNING id, interface_id, ip_address, vlan_id, is_primary, notes`,
		input.InterfaceID, input.IPAddress, input.VlanID, input.IsPrimary, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device ip not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "device_ips", id, fmt.Sprintf("Updated device IP '%s'", d.IPAddress))
	if err := h.maybeSetVLANGateway(c.Request.Context(), c, d.ID, d.IPAddress, &input); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, http.StatusOK, d)
}

// Delete handles DELETE /device-ips/:id
func (h *DeviceIPHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM device_ips WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device ip not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "device_ips", id, fmt.Sprintf("Deleted device IP #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
