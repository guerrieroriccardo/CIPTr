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
const patchPanelPortSelectSQL = `SELECT id, device_id, port_number, port_label, linked_port_id, switch_port_id, notes FROM patch_panel_ports`

// scanPatchPanelPort reads one row into a PatchPanelPort struct.
func scanPatchPanelPort(row interface{ Scan(...any) error }) (models.PatchPanelPort, error) {
	var ppp models.PatchPanelPort
	err := row.Scan(&ppp.ID, &ppp.DeviceID, &ppp.PortNumber, &ppp.PortLabel, &ppp.LinkedPortID, &ppp.SwitchPortID, &ppp.Notes)
	return ppp, err
}

// loadPPConnections enriches patch panel ports with connection info.
func (h *PatchPanelPortHandler) loadPPConnections(c *gin.Context, ports []models.PatchPanelPort) error {
	if len(ports) == 0 {
		return nil
	}

	ids := make([]int64, len(ports))
	idx := map[int64]int{}
	for i, p := range ports {
		ids[i] = p.ID
		idx[p.ID] = i
	}

	// Enrich from device_connections: which device interface is connected to this PP port
	ph, args := inPlaceholders(ids)
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT dc.patch_panel_port_id, d.hostname, di.name
		 FROM device_connections dc
		 JOIN device_interfaces di ON di.id = dc.interface_id
		 JOIN devices d ON d.id = di.device_id
		 WHERE dc.patch_panel_port_id IN `+ph, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ppID int64
		var hostname, ifName string
		if err := rows.Scan(&ppID, &hostname, &ifName); err != nil {
			return err
		}
		if i, ok := idx[ppID]; ok {
			ports[i].ConnectedDevice = &hostname
			ports[i].ConnectedInterface = &ifName
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Enrich from switch_port_id: which switch port is this PP port linked to
	for i, p := range ports {
		if p.SwitchPortID != nil {
			var hostname string
			var portNum int
			err := h.db.QueryRowContext(c.Request.Context(),
				`SELECT d.hostname, sp.port_number
				 FROM switch_ports sp
				 JOIN devices d ON d.id = sp.device_id
				 WHERE sp.id = $1`, *p.SwitchPortID,
			).Scan(&hostname, &portNum)
			if err == nil {
				ports[i].ConnectedSwitch = &hostname
				ports[i].ConnectedSwitchPort = &portNum
			}
		}
	}

	return nil
}

// List handles GET /patch-panel-ports
// Supports optional query param: ?device_id=
func (h *PatchPanelPortHandler) List(c *gin.Context) {
	query := patchPanelPortSelectSQL
	args := []any{}

	if deviceID := c.Query("device_id"); deviceID != "" {
		query += ` WHERE device_id = $1`
		args = append(args, deviceID)
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

	if err := h.loadPPConnections(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ports)
}

// ListByDevice handles GET /devices/:id/patch-panel-ports
func (h *PatchPanelPortHandler) ListByDevice(c *gin.Context) {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		patchPanelPortSelectSQL+` WHERE device_id = $1 ORDER BY port_number`, deviceID)
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

	if err := h.loadPPConnections(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
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

	ports := []models.PatchPanelPort{ppp}
	if err := h.loadPPConnections(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ports[0])
}

// Create handles POST /patch-panel-ports
func (h *PatchPanelPortHandler) Create(c *gin.Context) {
	var input models.PatchPanelPortInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// Validate: switch_port_id and linked_port_id are mutually exclusive
	if input.SwitchPortID != nil && input.LinkedPortID != nil {
		fail(c, http.StatusBadRequest, errors.New("cannot set both switch_port_id and linked_port_id"))
		return
	}

	// 1:1 validation for switch_port_id
	if input.SwitchPortID != nil {
		if taken, err := isSwitchPortTaken(c.Request.Context(), h.db, *input.SwitchPortID, 0); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		} else if taken {
			fail(c, http.StatusConflict, errors.New("switch port is already connected"))
			return
		}
	}

	ppp, err := scanPatchPanelPort(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO patch_panel_ports (device_id, port_number, port_label, linked_port_id, switch_port_id, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, device_id, port_number, port_label, linked_port_id, switch_port_id, notes`,
		input.DeviceID, input.PortNumber, input.PortLabel, input.LinkedPortID, input.SwitchPortID, input.Notes,
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

	// Validate: switch_port_id and linked_port_id are mutually exclusive
	if input.SwitchPortID != nil && input.LinkedPortID != nil {
		fail(c, http.StatusBadRequest, errors.New("cannot set both switch_port_id and linked_port_id"))
		return
	}

	// 1:1 validation for switch_port_id (exclude this PP port)
	if input.SwitchPortID != nil {
		if taken, err := isSwitchPortTakenExcludingPP(c.Request.Context(), h.db, *input.SwitchPortID, id); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		} else if taken {
			fail(c, http.StatusConflict, errors.New("switch port is already connected"))
			return
		}
	}

	// Get old linked_port_id before update so we can unlink the old partner.
	var oldLinkedPortID *int64
	_ = h.db.QueryRowContext(c.Request.Context(),
		`SELECT linked_port_id FROM patch_panel_ports WHERE id = $1`, id,
	).Scan(&oldLinkedPortID)

	ppp, err := scanPatchPanelPort(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE patch_panel_ports SET device_id = $1, port_number = $2, port_label = $3, linked_port_id = $4, switch_port_id = $5, notes = $6
		 WHERE id = $7
		 RETURNING id, device_id, port_number, port_label, linked_port_id, switch_port_id, notes`,
		input.DeviceID, input.PortNumber, input.PortLabel, input.LinkedPortID, input.SwitchPortID, input.Notes, id,
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
