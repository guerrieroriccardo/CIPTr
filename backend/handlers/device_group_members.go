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

// DeviceGroupMemberHandler groups all HTTP handlers for the /device-group-members resource.
type DeviceGroupMemberHandler struct {
	db *sql.DB
}

// NewDeviceGroupMemberHandler creates a DeviceGroupMemberHandler with the given database connection.
func NewDeviceGroupMemberHandler(db *sql.DB) *DeviceGroupMemberHandler {
	return &DeviceGroupMemberHandler{db: db}
}

const deviceGroupMemberColumns = `id, group_id, device_id`

const deviceGroupMemberSelectSQL = `SELECT ` + deviceGroupMemberColumns + ` FROM device_group_members`

// scanDeviceGroupMember reads one row into a DeviceGroupMember struct.
func scanDeviceGroupMember(row interface{ Scan(...any) error }) (models.DeviceGroupMember, error) {
	var m models.DeviceGroupMember
	err := row.Scan(&m.ID, &m.GroupID, &m.DeviceID)
	return m, err
}

// List handles GET /device-group-members
// Supports optional query param: ?group_id=1
func (h *DeviceGroupMemberHandler) List(c *gin.Context) {
	query := deviceGroupMemberSelectSQL
	var conds []string
	var args []any
	n := 1

	if groupID := c.Query("group_id"); groupID != "" {
		conds = append(conds, fmt.Sprintf("group_id = $%d", n))
		args = append(args, groupID)
		n++
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	members := []models.DeviceGroupMember{}
	for rows.Next() {
		m, err := scanDeviceGroupMember(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		members = append(members, m)
	}

	ok(c, http.StatusOK, members)
}

// ListByGroup handles GET /device-groups/:id/members
func (h *DeviceGroupMemberHandler) ListByGroup(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid group id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		deviceGroupMemberSelectSQL+` WHERE group_id = $1 ORDER BY id`, groupID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	members := []models.DeviceGroupMember{}
	for rows.Next() {
		m, err := scanDeviceGroupMember(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		members = append(members, m)
	}

	ok(c, http.StatusOK, members)
}

// Create handles POST /device-group-members
// Validates that device and group belong to the same site.
func (h *DeviceGroupMemberHandler) Create(c *gin.Context) {
	var input models.DeviceGroupMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// Validate same-site constraint.
	var groupSiteID, deviceSiteID int64
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT site_id FROM device_groups WHERE id = $1`, input.GroupID,
	).Scan(&groupSiteID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusBadRequest, errors.New("device group not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	err = h.db.QueryRowContext(c.Request.Context(),
		`SELECT site_id FROM devices WHERE id = $1`, input.DeviceID,
	).Scan(&deviceSiteID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusBadRequest, errors.New("device not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if groupSiteID != deviceSiteID {
		fail(c, http.StatusBadRequest, errors.New("device and group must belong to the same site"))
		return
	}

	m, err := scanDeviceGroupMember(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO device_group_members (group_id, device_id)
		 VALUES ($1, $2)
		 RETURNING `+deviceGroupMemberColumns,
		input.GroupID, input.DeviceID,
	))
	if err != nil {
		if strings.Contains(err.Error(), "device_group_members_group_id_device_id_key") {
			fail(c, http.StatusBadRequest, errors.New("device is already a member of this group"))
			return
		}
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "device_group_members", m.ID,
		fmt.Sprintf("Added device #%d to group #%d", input.DeviceID, input.GroupID))
	ok(c, http.StatusCreated, m)
}

// Delete handles DELETE /device-group-members/:id
func (h *DeviceGroupMemberHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM device_group_members WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("device group member not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "device_group_members", id,
		fmt.Sprintf("Removed device group member #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
