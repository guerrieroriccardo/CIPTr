package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// DeviceHandler groups all HTTP handlers for the /devices resource.
type DeviceHandler struct {
	db *sql.DB
}

// NewDeviceHandler creates a DeviceHandler with the given database connection.
func NewDeviceHandler(db *sql.DB) *DeviceHandler {
	return &DeviceHandler{db: db}
}

// deviceSelectSQL is the base SELECT used by every read operation.
const deviceSelectSQL = `SELECT id, site_id, location_id, model_id,
	hostname, dns_name, serial_number, asset_tag,
	category_id, status, is_up,
	os, has_rmm, has_antivirus, supplier_id,
	installation_date, is_reserved,
	notes, created_at, updated_at
	FROM devices`

// scanDevice reads one row into a Device struct.
func scanDevice(row interface{ Scan(...any) error }) (models.Device, error) {
	var d models.Device
	err := row.Scan(
		&d.ID, &d.SiteID, &d.LocationID, &d.ModelID,
		&d.Hostname, &d.DnsName, &d.SerialNumber, &d.AssetTag,
		&d.CategoryID, &d.Status, &d.IsUp,
		&d.Os, &d.HasRmm, &d.HasAntivirus, &d.SupplierID,
		&d.InstallationDate, &d.IsReserved,
		&d.Notes, &d.CreatedAt, &d.UpdatedAt,
	)
	return d, err
}

func (h *DeviceHandler) validateDevice(ctx context.Context, input *models.DeviceInput, excludeID int64) error {
	var existing int64
	err := h.db.QueryRowContext(ctx,
		`SELECT id FROM devices WHERE site_id = $1 AND hostname = $2 AND id != $3 LIMIT 1`,
		input.SiteID, input.Hostname, excludeID,
	).Scan(&existing)
	if err == nil {
		return fmt.Errorf("hostname %q already exists in this site", input.Hostname)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return nil
}

// List handles GET /devices
// Supports optional query params: ?site_id=, ?status=, ?category_id=, ?search=
func (h *DeviceHandler) List(c *gin.Context) {
	query := deviceSelectSQL
	var conds []string
	var args []any
	n := 1

	if siteID := c.Query("site_id"); siteID != "" {
		conds = append(conds, fmt.Sprintf("site_id = $%d", n))
		args = append(args, siteID)
		n++
	}
	if status := c.Query("status"); status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", n))
		args = append(args, status)
		n++
	}
	if catID := c.Query("category_id"); catID != "" {
		conds = append(conds, fmt.Sprintf("category_id = $%d", n))
		args = append(args, catID)
		n++
	}
	if search := c.Query("search"); search != "" {
		conds = append(conds, fmt.Sprintf("(hostname ILIKE $%d OR dns_name ILIKE $%d)", n, n))
		args = append(args, "%"+search+"%")
		n++
	}

	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY hostname"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	devices := []models.Device{}
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		devices = append(devices, d)
	}

	ok(c, http.StatusOK, devices)
}

// ListBySite handles GET /sites/:id/devices
// Returns all devices for the given site, ordered by hostname.
func (h *DeviceHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		deviceSelectSQL+` WHERE site_id = $1 ORDER BY hostname`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	devices := []models.Device{}
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		devices = append(devices, d)
	}

	ok(c, http.StatusOK, devices)
}

// GetByID handles GET /devices/:id
// Returns 404 if the device does not exist.
func (h *DeviceHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	d, err := scanDevice(h.db.QueryRowContext(c.Request.Context(),
		deviceSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, d)
}

// Create handles POST /devices
// site_id, hostname, and category_id are required.
func (h *DeviceHandler) Create(c *gin.Context) {
	var input models.DeviceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validateDevice(c.Request.Context(), &input, 0); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	status := "active"
	if input.Status != nil {
		status = *input.Status
	}

	d, err := scanDevice(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO devices (
			site_id, location_id, model_id,
			hostname, dns_name, serial_number, asset_tag,
			category_id, status, is_up,
			os, has_rmm, has_antivirus, supplier_id,
			installation_date, is_reserved, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING id, site_id, location_id, model_id,
			hostname, dns_name, serial_number, asset_tag,
			category_id, status, is_up,
			os, has_rmm, has_antivirus, supplier_id,
			installation_date, is_reserved,
			notes, created_at, updated_at`,
		input.SiteID, input.LocationID, input.ModelID,
		input.Hostname, input.DnsName, input.SerialNumber, input.AssetTag,
		input.CategoryID, status, input.IsUp,
		input.Os, input.HasRmm, input.HasAntivirus, input.SupplierID,
		input.InstallationDate, input.IsReserved, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, d)
}

// Update handles PUT /devices/:id
// Replaces all fields. Returns 404 if the device does not exist.
func (h *DeviceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.DeviceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validateDevice(c.Request.Context(), &input, id); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	status := "active"
	if input.Status != nil {
		status = *input.Status
	}

	d, err := scanDevice(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE devices SET
			site_id = $1, location_id = $2, model_id = $3,
			hostname = $4, dns_name = $5, serial_number = $6, asset_tag = $7,
			category_id = $8, status = $9, is_up = $10,
			os = $11, has_rmm = $12, has_antivirus = $13, supplier_id = $14,
			installation_date = $15, is_reserved = $16, notes = $17
		WHERE id = $18
		RETURNING id, site_id, location_id, model_id,
			hostname, dns_name, serial_number, asset_tag,
			category_id, status, is_up,
			os, has_rmm, has_antivirus, supplier_id,
			installation_date, is_reserved,
			notes, created_at, updated_at`,
		input.SiteID, input.LocationID, input.ModelID,
		input.Hostname, input.DnsName, input.SerialNumber, input.AssetTag,
		input.CategoryID, status, input.IsUp,
		input.Os, input.HasRmm, input.HasAntivirus, input.SupplierID,
		input.InstallationDate, input.IsReserved, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, d)
}

// Delete handles DELETE /devices/:id
// Cascades to device_interfaces, device_ips, and device_connections via DB FK.
func (h *DeviceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM devices WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
