package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// PatchPanelPortHandler groups all HTTP handlers for the /patch-panel-ports resource.
type PatchPanelPortHandler struct {
	db *sql.DB
}

// NewPatchPanelPortHandler creates a PatchPanelPortHandler with the given database connection.
func NewPatchPanelPortHandler(db *sql.DB) *PatchPanelPortHandler {
	return &PatchPanelPortHandler{db: db}
}

// patchPanelPortSelectSQL is the base SELECT used by every read operation.
const patchPanelPortSelectSQL = `SELECT id, patch_panel_id, port_number, port_label, notes FROM patch_panel_ports`

// scanPatchPanelPort reads one row into a PatchPanelPort struct.
func scanPatchPanelPort(row interface{ Scan(...any) error }) (models.PatchPanelPort, error) {
	var ppp models.PatchPanelPort
	err := row.Scan(&ppp.ID, &ppp.PatchPanelID, &ppp.PortNumber, &ppp.PortLabel, &ppp.Notes)
	return ppp, err
}

// List handles GET /patch-panel-ports
// Supports optional query param: ?patch_panel_id=
func (h *PatchPanelPortHandler) List(c *gin.Context) {
	query := patchPanelPortSelectSQL
	args := []any{}

	if ppID := c.Query("patch_panel_id"); ppID != "" {
		query += ` WHERE patch_panel_id = $1`
		args = append(args, ppID)
	}
	query += ` ORDER BY port_number`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	ports := []models.PatchPanelPort{}
	for rows.Next() {
		ppp, err := scanPatchPanelPort(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		ports = append(ports, ppp)
	}

	ok(c, http.StatusOK, ports)
}

// ListByPatchPanel handles GET /patch-panels/:id/ports
func (h *PatchPanelPortHandler) ListByPatchPanel(c *gin.Context) {
	ppID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid patch panel id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		patchPanelPortSelectSQL+` WHERE patch_panel_id = $1 ORDER BY port_number`, ppID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	ports := []models.PatchPanelPort{}
	for rows.Next() {
		ppp, err := scanPatchPanelPort(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		ports = append(ports, ppp)
	}

	ok(c, http.StatusOK, ports)
}

// GetByID handles GET /patch-panel-ports/:id
func (h *PatchPanelPortHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	ppp, err := scanPatchPanelPort(h.db.QueryRowContext(c.Request.Context(),
		patchPanelPortSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ppp)
}

// Create handles POST /patch-panel-ports
func (h *PatchPanelPortHandler) Create(c *gin.Context) {
	var input models.PatchPanelPortInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	ppp, err := scanPatchPanelPort(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO patch_panel_ports (patch_panel_id, port_number, port_label, notes)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, patch_panel_id, port_number, port_label, notes`,
		input.PatchPanelID, input.PortNumber, input.PortLabel, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, ppp)
}

// Update handles PUT /patch-panel-ports/:id
func (h *PatchPanelPortHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.PatchPanelPortInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	ppp, err := scanPatchPanelPort(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE patch_panel_ports SET patch_panel_id = $1, port_number = $2, port_label = $3, notes = $4
		 WHERE id = $5
		 RETURNING id, patch_panel_id, port_number, port_label, notes`,
		input.PatchPanelID, input.PortNumber, input.PortLabel, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ppp)
}

// Delete handles DELETE /patch-panel-ports/:id
func (h *PatchPanelPortHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM patch_panel_ports WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
