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

// VLANHandler groups all HTTP handlers for the /vlans resource.
type VLANHandler struct {
	db *sql.DB
}

// NewVLANHandler creates a VLANHandler with the given database connection.
func NewVLANHandler(db *sql.DB) *VLANHandler {
	return &VLANHandler{db: db}
}

// vlanSelectSQL is the base SELECT used by every read operation.
const vlanSelectSQL = `SELECT id, site_id, address_block_id, vlan_id, name, subnet, gateway, description FROM vlans`

// scanVLAN reads one row into a VLAN struct.
func scanVLAN(row interface{ Scan(...any) error }) (models.VLAN, error) {
	var v models.VLAN
	err := row.Scan(&v.ID, &v.SiteID, &v.AddressBlockID, &v.VlanID, &v.Name, &v.Subnet, &v.Gateway, &v.Description)
	return v, err
}

// List handles GET /vlans
// Supports optional query params: ?site_id=1 and/or ?address_block_id=2
// Both filters can be combined.
func (h *VLANHandler) List(c *gin.Context) {
	query := vlanSelectSQL
	var conds []string
	var args []any
	n := 1 // PostgreSQL placeholder counter

	if siteID := c.Query("site_id"); siteID != "" {
		conds = append(conds, fmt.Sprintf("site_id = $%d", n))
		args = append(args, siteID)
		n++
	}
	if blockID := c.Query("address_block_id"); blockID != "" {
		conds = append(conds, fmt.Sprintf("address_block_id = $%d", n))
		args = append(args, blockID)
		n++
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY vlan_id"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	vlans := []models.VLAN{}
	for rows.Next() {
		v, err := scanVLAN(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		vlans = append(vlans, v)
	}

	ok(c, http.StatusOK, vlans)
}

// ListBySite handles GET /sites/:id/vlans
// Returns all VLANs for the given site, ordered by VLAN number.
func (h *VLANHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		vlanSelectSQL+` WHERE site_id = $1 ORDER BY vlan_id`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	vlans := []models.VLAN{}
	for rows.Next() {
		v, err := scanVLAN(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		vlans = append(vlans, v)
	}

	ok(c, http.StatusOK, vlans)
}

// ListByAddressBlock handles GET /address-blocks/:id/vlans
// Returns all VLANs belonging to the given address block.
func (h *VLANHandler) ListByAddressBlock(c *gin.Context) {
	blockID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid address block id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		vlanSelectSQL+` WHERE address_block_id = $1 ORDER BY vlan_id`, blockID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	vlans := []models.VLAN{}
	for rows.Next() {
		v, err := scanVLAN(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		vlans = append(vlans, v)
	}

	ok(c, http.StatusOK, vlans)
}

// GetByID handles GET /vlans/:id
// Returns 404 if the VLAN does not exist.
func (h *VLANHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	v, err := scanVLAN(h.db.QueryRowContext(c.Request.Context(),
		vlanSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("vlan not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, v)
}

// Create handles POST /vlans
// site_id, vlan_id, and name are required.
func (h *VLANHandler) Create(c *gin.Context) {
	var input models.VLANInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	v, err := scanVLAN(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO vlans (site_id, address_block_id, vlan_id, name, subnet, gateway, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, site_id, address_block_id, vlan_id, name, subnet, gateway, description`,
		input.SiteID, input.AddressBlockID, input.VlanID, input.Name,
		input.Subnet, input.Gateway, input.Description,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, v)
}

// Update handles PUT /vlans/:id
// Replaces all fields. Returns 404 if the VLAN does not exist.
func (h *VLANHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.VLANInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	v, err := scanVLAN(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE vlans SET site_id = $1, address_block_id = $2, vlan_id = $3, name = $4,
		 subnet = $5, gateway = $6, description = $7 WHERE id = $8
		 RETURNING id, site_id, address_block_id, vlan_id, name, subnet, gateway, description`,
		input.SiteID, input.AddressBlockID, input.VlanID, input.Name,
		input.Subnet, input.Gateway, input.Description, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("vlan not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, v)
}

// Delete handles DELETE /vlans/:id
// Returns 404 if the VLAN does not exist.
func (h *VLANHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM vlans WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("vlan not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
