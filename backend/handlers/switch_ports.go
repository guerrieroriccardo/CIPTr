package handlers

import (
	"database/sql"
	"errors"
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
const switchPortSelectSQL = `SELECT id, switch_id, port_number, port_label, speed, is_uplink, notes FROM switch_ports`

// scanSwitchPort reads one row into a SwitchPort struct.
func scanSwitchPort(row interface{ Scan(...any) error }) (models.SwitchPort, error) {
	var sp models.SwitchPort
	err := row.Scan(&sp.ID, &sp.SwitchID, &sp.PortNumber, &sp.PortLabel, &sp.Speed, &sp.IsUplink, &sp.Notes)
	return sp, err
}

// List handles GET /switch-ports
// Supports optional query param: ?switch_id=
func (h *SwitchPortHandler) List(c *gin.Context) {
	query := switchPortSelectSQL
	args := []any{}

	if switchID := c.Query("switch_id"); switchID != "" {
		query += ` WHERE switch_id = $1`
		args = append(args, switchID)
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

	ok(c, http.StatusOK, ports)
}

// ListBySwitch handles GET /switches/:id/ports
func (h *SwitchPortHandler) ListBySwitch(c *gin.Context) {
	switchID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid switch id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		switchPortSelectSQL+` WHERE switch_id = $1 ORDER BY port_number`, switchID)
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

	ok(c, http.StatusOK, sp)
}

// Create handles POST /switch-ports
func (h *SwitchPortHandler) Create(c *gin.Context) {
	var input models.SwitchPortInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	sp, err := scanSwitchPort(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO switch_ports (switch_id, port_number, port_label, speed, is_uplink, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, switch_id, port_number, port_label, speed, is_uplink, notes`,
		input.SwitchID, input.PortNumber, input.PortLabel, input.Speed, input.IsUplink, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

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
		`UPDATE switch_ports SET switch_id = $1, port_number = $2, port_label = $3, speed = $4, is_uplink = $5, notes = $6
		 WHERE id = $7
		 RETURNING id, switch_id, port_number, port_label, speed, is_uplink, notes`,
		input.SwitchID, input.PortNumber, input.PortLabel, input.Speed, input.IsUplink, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("switch port not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

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

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
