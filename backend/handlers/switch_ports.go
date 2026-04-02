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

// SwitchPortHandler groups all HTTP handlers for the /switch-ports resource.
type SwitchPortHandler struct {
	db *sql.DB
}

// NewSwitchPortHandler creates a SwitchPortHandler with the given database connection.
func NewSwitchPortHandler(db *sql.DB) *SwitchPortHandler {
	return &SwitchPortHandler{db: db}
}

// switchPortSelectSQL is the base SELECT used by every read operation.
const switchPortSelectSQL = `SELECT id, device_id, port_number, port_label, speed, is_uplink, mac_restriction, untagged_vlan_id, is_disabled, notes FROM switch_ports`

// scanSwitchPort reads one row into a SwitchPort struct (without tagged VLANs).
func scanSwitchPort(row interface{ Scan(...any) error }) (models.SwitchPort, error) {
	var sp models.SwitchPort
	err := row.Scan(&sp.ID, &sp.DeviceID, &sp.PortNumber, &sp.PortLabel, &sp.Speed, &sp.IsUplink, &sp.MacRestriction, &sp.UntaggedVlanID, &sp.IsDisabled, &sp.Notes)
	return sp, err
}

// loadTaggedVlans fetches tagged VLAN IDs for a set of switch port IDs and attaches them.
func (h *SwitchPortHandler) loadTaggedVlans(c *gin.Context, ports []models.SwitchPort) error {
	if len(ports) == 0 {
		return nil
	}
	ids := make([]int64, len(ports))
	idx := map[int64]int{} // port ID → index in slice
	for i, p := range ports {
		ids[i] = p.ID
		idx[p.ID] = i
		ports[i].TaggedVlanIDs = []int64{} // ensure non-nil
	}

	ph, args := inPlaceholders(ids)
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT switch_port_id, vlan_id FROM switch_port_tagged_vlans WHERE switch_port_id IN `+ph+` ORDER BY vlan_id`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var portID, vlanID int64
		if err := rows.Scan(&portID, &vlanID); err != nil {
			return err
		}
		if i, ok := idx[portID]; ok {
			ports[i].TaggedVlanIDs = append(ports[i].TaggedVlanIDs, vlanID)
		}
	}
	return rows.Err()
}

// syncTaggedVlans replaces the tagged VLAN set for a switch port.
func (h *SwitchPortHandler) syncTaggedVlans(c *gin.Context, portID int64, vlanIDs []int64) error {
	reqCtx := c.Request.Context()
	_, err := h.db.ExecContext(reqCtx, `DELETE FROM switch_port_tagged_vlans WHERE switch_port_id = $1`, portID)
	if err != nil {
		return err
	}
	for _, vid := range vlanIDs {
		_, err := h.db.ExecContext(reqCtx,
			`INSERT INTO switch_port_tagged_vlans (switch_port_id, vlan_id) VALUES ($1, $2)`, portID, vid)
		if err != nil {
			return err
		}
	}
	return nil
}

// loadSPConnections enriches switch ports with connection info from device_connections
// and patch_panel_ports.
func (h *SwitchPortHandler) loadSPConnections(c *gin.Context, ports []models.SwitchPort) error {
	if len(ports) == 0 {
		return nil
	}

	ids := make([]int64, len(ports))
	idx := map[int64]int{}
	for i, p := range ports {
		ids[i] = p.ID
		idx[p.ID] = i
	}

	ph, args := inPlaceholders(ids)

	// From device_connections: which device interface is connected to this switch port
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT dc.switch_port_id, d.hostname, di.name
		 FROM device_connections dc
		 JOIN device_interfaces di ON di.id = dc.interface_id
		 JOIN devices d ON d.id = di.device_id
		 WHERE dc.switch_port_id IN `+ph, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var spID int64
		var hostname, ifName string
		if err := rows.Scan(&spID, &hostname, &ifName); err != nil {
			return err
		}
		if i, ok := idx[spID]; ok {
			ports[i].ConnectedDevice = &hostname
			ports[i].ConnectedInterface = &ifName
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// From patch_panel_ports: which PP port is linked to this switch port
	rows2, err := h.db.QueryContext(c.Request.Context(),
		`SELECT ppp.switch_port_id, d.hostname, ppp.port_number
		 FROM patch_panel_ports ppp
		 JOIN devices d ON d.id = ppp.device_id
		 WHERE ppp.switch_port_id IN `+ph, args...)
	if err != nil {
		return err
	}
	defer rows2.Close()
	for rows2.Next() {
		var spID int64
		var ppName string
		var ppPortNum int
		if err := rows2.Scan(&spID, &ppName, &ppPortNum); err != nil {
			return err
		}
		if i, ok := idx[spID]; ok {
			ports[i].ConnectedPatchPanel = &ppName
			ports[i].ConnectedPatchPanelPort = &ppPortNum
		}
	}
	return rows2.Err()
}

// List handles GET /switch-ports
// Supports optional query param: ?device_id=
func (h *SwitchPortHandler) List(c *gin.Context) {
	query := switchPortSelectSQL
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

	ports := []models.SwitchPort{}
	for rows.Next() {
		sp, err := scanSwitchPort(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		ports = append(ports, sp)
	}

	if err := h.loadTaggedVlans(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.loadSPConnections(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ports)
}

// ListByDevice handles GET /devices/:id/switch-ports
func (h *SwitchPortHandler) ListByDevice(c *gin.Context) {
	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid device id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		switchPortSelectSQL+` WHERE device_id = $1 ORDER BY port_number`, deviceID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	ports := []models.SwitchPort{}
	for rows.Next() {
		sp, err := scanSwitchPort(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		ports = append(ports, sp)
	}

	if err := h.loadTaggedVlans(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.loadSPConnections(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ports)
}

// GetByID handles GET /switch-ports/:id
func (h *SwitchPortHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	sp, err := scanSwitchPort(h.db.QueryRowContext(c.Request.Context(),
		switchPortSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ports := []models.SwitchPort{sp}
	if err := h.loadTaggedVlans(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.loadSPConnections(c, ports); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ports[0])
}

// Create handles POST /switch-ports
func (h *SwitchPortHandler) Create(c *gin.Context) {
	var input models.SwitchPortInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	sp, err := scanSwitchPort(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO switch_ports (device_id, port_number, port_label, speed, is_uplink, mac_restriction, untagged_vlan_id, is_disabled, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, device_id, port_number, port_label, speed, is_uplink, mac_restriction, untagged_vlan_id, is_disabled, notes`,
		input.DeviceID, input.PortNumber, input.PortLabel, input.Speed, input.IsUplink, input.MacRestriction, input.UntaggedVlanID, input.IsDisabled, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if len(input.TaggedVlanIDs) > 0 {
		if err := h.syncTaggedVlans(c, sp.ID, input.TaggedVlanIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		sp.TaggedVlanIDs = input.TaggedVlanIDs
	} else {
		sp.TaggedVlanIDs = []int64{}
	}

	logAudit(c.Request.Context(), h.db, c, "create", "switch_ports", sp.ID, fmt.Sprintf("Created switch port #%d", sp.PortNumber))
	ok(c, http.StatusCreated, sp)
}

// Update handles PUT /switch-ports/:id
func (h *SwitchPortHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.SwitchPortInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	sp, err := scanSwitchPort(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE switch_ports SET device_id = $1, port_number = $2, port_label = $3, speed = $4, is_uplink = $5, mac_restriction = $6, untagged_vlan_id = $7, is_disabled = $8, notes = $9
		 WHERE id = $10
		 RETURNING id, device_id, port_number, port_label, speed, is_uplink, mac_restriction, untagged_vlan_id, is_disabled, notes`,
		input.DeviceID, input.PortNumber, input.PortLabel, input.Speed, input.IsUplink, input.MacRestriction, input.UntaggedVlanID, input.IsDisabled, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := h.syncTaggedVlans(c, sp.ID, input.TaggedVlanIDs); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	sp.TaggedVlanIDs = input.TaggedVlanIDs
	if sp.TaggedVlanIDs == nil {
		sp.TaggedVlanIDs = []int64{}
	}

	logAudit(c.Request.Context(), h.db, c, "update", "switch_ports", id, fmt.Sprintf("Updated switch port #%d", sp.PortNumber))
	ok(c, http.StatusOK, sp)
}

// Delete handles DELETE /switch-ports/:id
func (h *SwitchPortHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM switch_ports WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "switch_ports", id, fmt.Sprintf("Deleted switch port #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
