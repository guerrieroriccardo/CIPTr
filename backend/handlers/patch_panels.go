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

// PatchPanelHandler groups all HTTP handlers for the /patch-panels resource.
type PatchPanelHandler struct {
	db *sql.DB
}

// NewPatchPanelHandler creates a PatchPanelHandler with the given database connection.
func NewPatchPanelHandler(db *sql.DB) *PatchPanelHandler {
	return &PatchPanelHandler{db: db}
}

// patchPanelSelectSQL is the base SELECT used by every read operation.
const patchPanelSelectSQL = `SELECT id, site_id, name, total_ports, location_id, notes FROM patch_panels`

// scanPatchPanel reads one row into a PatchPanel struct.
func scanPatchPanel(row interface{ Scan(...any) error }) (models.PatchPanel, error) {
	var pp models.PatchPanel
	err := row.Scan(&pp.ID, &pp.SiteID, &pp.Name, &pp.TotalPorts, &pp.LocationID, &pp.Notes)
	return pp, err
}

// List handles GET /patch-panels
// Supports optional query param: ?site_id=
func (h *PatchPanelHandler) List(c *gin.Context) {
	query := patchPanelSelectSQL
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

	panels := []models.PatchPanel{}
	for rows.Next() {
		pp, err := scanPatchPanel(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		panels = append(panels, pp)
	}

	ok(c, http.StatusOK, panels)
}

// ListBySite handles GET /sites/:id/patch-panels
func (h *PatchPanelHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		patchPanelSelectSQL+` WHERE site_id = $1 ORDER BY name`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	panels := []models.PatchPanel{}
	for rows.Next() {
		pp, err := scanPatchPanel(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		panels = append(panels, pp)
	}

	ok(c, http.StatusOK, panels)
}

// GetByID handles GET /patch-panels/:id
func (h *PatchPanelHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	pp, err := scanPatchPanel(h.db.QueryRowContext(c.Request.Context(),
		patchPanelSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, pp)
}

// Create handles POST /patch-panels
// Automatically creates N patch_panel_port rows based on total_ports.
func (h *PatchPanelHandler) Create(c *gin.Context) {
	var input models.PatchPanelInput
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

	pp, err := scanPatchPanel(tx.QueryRowContext(c.Request.Context(),
		`INSERT INTO patch_panels (site_id, name, total_ports, location_id, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, site_id, name, total_ports, location_id, notes`,
		input.SiteID, input.Name, totalPorts, input.LocationID, input.Notes,
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
			args = append(args, pp.ID, i)
		}
		_, err = tx.ExecContext(c.Request.Context(),
			`INSERT INTO patch_panel_ports (patch_panel_id, port_number) VALUES `+strings.Join(vals, ", "),
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

	logAudit(c.Request.Context(), h.db, c, "create", "patch_panels", pp.ID, fmt.Sprintf("Created patch panel '%s'", pp.Name))
	ok(c, http.StatusCreated, pp)
}

// Update handles PUT /patch-panels/:id
func (h *PatchPanelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.PatchPanelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	totalPorts := 24
	if input.TotalPorts != nil {
		totalPorts = *input.TotalPorts
	}

	pp, err := scanPatchPanel(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE patch_panels SET site_id = $1, name = $2, total_ports = $3, location_id = $4, notes = $5
		 WHERE id = $6
		 RETURNING id, site_id, name, total_ports, location_id, notes`,
		input.SiteID, input.Name, totalPorts, input.LocationID, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "patch_panels", id, fmt.Sprintf("Updated patch panel '%s'", pp.Name))
	ok(c, http.StatusOK, pp)
}

// Delete handles DELETE /patch-panels/:id
// Cascades to patch_panel_ports via DB FK.
func (h *PatchPanelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM patch_panels WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "patch_panels", id, fmt.Sprintf("Deleted patch panel #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
