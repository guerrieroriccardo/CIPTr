package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// DeviceModelHandler groups all HTTP handlers for the /device-models resource.
type DeviceModelHandler struct {
	db *sql.DB
}

// NewDeviceModelHandler creates a DeviceModelHandler with the given database connection.
func NewDeviceModelHandler(db *sql.DB) *DeviceModelHandler {
	return &DeviceModelHandler{db: db}
}

// deviceModelSelectSQL is the base SELECT used by every read operation.
const deviceModelSelectSQL = `SELECT id, manufacturer_id, model_name, category_id, os_default_id, specs, notes, created_at FROM device_models`

// scanDeviceModel reads one row into a DeviceModel struct.
func scanDeviceModel(row interface{ Scan(...any) error }) (models.DeviceModel, error) {
	var dm models.DeviceModel
	err := row.Scan(&dm.ID, &dm.ManufacturerID, &dm.ModelName, &dm.CategoryID, &dm.OsDefaultID, &dm.Specs, &dm.Notes, &dm.CreatedAt)
	return dm, err
}

// List handles GET /device-models
// Supports optional query params: ?category_id=, ?manufacturer_id=
func (h *DeviceModelHandler) List(c *gin.Context) {
	query := deviceModelSelectSQL
	var conds []string
	var args []any
	n := 1

	if catID := c.Query("category_id"); catID != "" {
		conds = append(conds, "category_id = $"+strconv.Itoa(n))
		args = append(args, catID)
		n++
	}
	if mfgID := c.Query("manufacturer_id"); mfgID != "" {
		conds = append(conds, "manufacturer_id = $"+strconv.Itoa(n))
		args = append(args, mfgID)
		n++
	}

	if len(conds) > 0 {
		query += " WHERE " + conds[0]
		for _, cond := range conds[1:] {
			query += " AND " + cond
		}
	}
	query += ` ORDER BY manufacturer_id, model_name`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	models := []models.DeviceModel{}
	for rows.Next() {
		dm, err := scanDeviceModel(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		models = append(models, dm)
	}

	ok(c, http.StatusOK, models)
}

// GetByID handles GET /device-models/:id
// Returns 404 if the device model does not exist.
func (h *DeviceModelHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	dm, err := scanDeviceModel(h.db.QueryRowContext(c.Request.Context(),
		deviceModelSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device model not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, dm)
}

// Create handles POST /device-models
// manufacturer_id, model_name, and category_id are required.
func (h *DeviceModelHandler) Create(c *gin.Context) {
	var input models.DeviceModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	dm, err := scanDeviceModel(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO device_models (manufacturer_id, model_name, category_id, os_default_id, specs, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, manufacturer_id, model_name, category_id, os_default_id, specs, notes, created_at`,
		input.ManufacturerID, input.ModelName, input.CategoryID,
		input.OsDefaultID, input.Specs, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "device_models", dm.ID, fmt.Sprintf("Created device model '%s'", dm.ModelName))
	ok(c, http.StatusCreated, dm)
}

// Update handles PUT /device-models/:id
// Replaces all fields. Returns 404 if the device model does not exist.
func (h *DeviceModelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.DeviceModelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	dm, err := scanDeviceModel(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE device_models SET manufacturer_id = $1, model_name = $2, category_id = $3,
		 os_default_id = $4, specs = $5, notes = $6 WHERE id = $7
		 RETURNING id, manufacturer_id, model_name, category_id, os_default_id, specs, notes, created_at`,
		input.ManufacturerID, input.ModelName, input.CategoryID,
		input.OsDefaultID, input.Specs, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device model not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "device_models", id, fmt.Sprintf("Updated device model '%s'", dm.ModelName))
	ok(c, http.StatusOK, dm)
}

// Delete handles DELETE /device-models/:id
// Devices and switches referencing this model have their model_id set to NULL (ON DELETE SET NULL).
func (h *DeviceModelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM device_models WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device model not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "device_models", id, fmt.Sprintf("Deleted device model #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
