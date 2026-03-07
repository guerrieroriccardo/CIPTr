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

// PatchPanelPortHandler groups all HTTP handlers for the /patch-panel-ports resource.
type PatchPanelPortHandler struct {
	db *sql.DB
}

// NewPatchPanelPortHandler creates a PatchPanelPortHandler with the given database connection.
func NewPatchPanelPortHandler(db *sql.DB) *PatchPanelPortHandler {
	return &PatchPanelPortHandler{db: db}
}

// patchPanelPortSelectSQL is the base SELECT used by every read operation.
const patchPanelPortSelectSQL = `SELECT id, patch_panel_id, port_number, port_label, linked_port_id, notes FROM patch_panel_ports`

// scanPatchPanelPort reads one row into a PatchPanelPort struct.
func scanPatchPanelPort(row interface{ Scan(...any) error }) (models.PatchPanelPort, error) {
	var ppp models.PatchPanelPort
	err := row.Scan(&ppp.ID, &ppp.PatchPanelID, &ppp.PortNumber, &ppp.PortLabel, &ppp.LinkedPortID, &ppp.Notes)
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
		`INSERT INTO patch_panel_ports (patch_panel_id, port_number, port_label, linked_port_id, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, patch_panel_id, port_number, port_label, linked_port_id, notes`,
		input.PatchPanelID, input.PortNumber, input.PortLabel, input.LinkedPortID, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	// Maintain bidirectional link: if we linked to another port, link it back.
	if ppp.LinkedPortID != nil {
		h.db.ExecContext(c.Request.Context(),
			`UPDATE patch_panel_ports SET linked_port_id = $1 WHERE id = $2`,
			ppp.ID, *ppp.LinkedPortID)
	}

	logAudit(c.Request.Context(), h.db, c, "create", "patch_panel_ports", ppp.ID, fmt.Sprintf("Created patch panel port #%d", ppp.PortNumber))
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

	// Get old linked_port_id before update so we can unlink the old partner.
	var oldLinkedPortID *int64
	_ = h.db.QueryRowContext(c.Request.Context(),
		`SELECT linked_port_id FROM patch_panel_ports WHERE id = $1`, id,
	).Scan(&oldLinkedPortID)

	ppp, err := scanPatchPanelPort(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE patch_panel_ports SET patch_panel_id = $1, port_number = $2, port_label = $3, linked_port_id = $4, notes = $5
		 WHERE id = $6
		 RETURNING id, patch_panel_id, port_number, port_label, linked_port_id, notes`,
		input.PatchPanelID, input.PortNumber, input.PortLabel, input.LinkedPortID, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("patch panel port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ctx := c.Request.Context()
	// Unlink old partner if it changed.
	if oldLinkedPortID != nil && (ppp.LinkedPortID == nil || *oldLinkedPortID != *ppp.LinkedPortID) {
		h.db.ExecContext(ctx,
			`UPDATE patch_panel_ports SET linked_port_id = NULL WHERE id = $1 AND linked_port_id = $2`,
			*oldLinkedPortID, id)
	}
	// Link new partner back.
	if ppp.LinkedPortID != nil {
		h.db.ExecContext(ctx,
			`UPDATE patch_panel_ports SET linked_port_id = $1 WHERE id = $2`,
			ppp.ID, *ppp.LinkedPortID)
	}

	logAudit(c.Request.Context(), h.db, c, "update", "patch_panel_ports", id, fmt.Sprintf("Updated patch panel port #%d", ppp.PortNumber))
	ok(c, http.StatusOK, ppp)
}

// Delete handles DELETE /patch-panel-ports/:id
func (h *PatchPanelPortHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	// Unlink partner before deleting (ON DELETE SET NULL handles the FK,
	// but we also need to clear the partner's linked_port_id).
	h.db.ExecContext(c.Request.Context(),
		`UPDATE patch_panel_ports SET linked_port_id = NULL WHERE linked_port_id = $1`, id)

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

	logAudit(c.Request.Context(), h.db, c, "delete", "patch_panel_ports", id, fmt.Sprintf("Deleted patch panel port #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
