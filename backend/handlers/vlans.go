package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
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

// validateVLAN checks:
//  1. VLAN tag number is unique per site.
//  2. If subnet and address_block_id are both set, the subnet must be within the block's network.
func (h *VLANHandler) validateVLAN(ctx context.Context, input *models.VLANInput, excludeID int64) error {
	// Check VLAN tag uniqueness within the site.
	var existing int64
	err := h.db.QueryRowContext(ctx,
		`SELECT id FROM vlans WHERE site_id = $1 AND vlan_id = $2 AND id != $3 LIMIT 1`,
		input.SiteID, input.VlanID, excludeID,
	).Scan(&existing)
	if err == nil {
		return fmt.Errorf("VLAN tag %d already exists in this site", input.VlanID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Parse subnet if provided.
	var subnetNet *net.IPNet
	if input.Subnet != nil && *input.Subnet != "" {
		_, parsed, err := net.ParseCIDR(*input.Subnet)
		if err != nil {
			return fmt.Errorf("invalid subnet CIDR: %s", *input.Subnet)
		}
		subnetNet = parsed
	}

	// Check subnet uniqueness within the site.
	if subnetNet != nil {
		var overlapping string
		err := h.db.QueryRowContext(ctx,
			`SELECT subnet FROM vlans WHERE site_id = $1 AND id != $2 AND subnet && $3::cidr LIMIT 1`,
			input.SiteID, excludeID, *input.Subnet,
		).Scan(&overlapping)
		if err == nil {
			return fmt.Errorf("subnet %s overlaps with existing VLAN subnet %s", *input.Subnet, overlapping)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	// Validate subnet is within address block.
	if input.AddressBlockID != nil && subnetNet != nil {
		var blockNetwork string
		err := h.db.QueryRowContext(ctx,
			`SELECT network FROM address_blocks WHERE id = $1`, *input.AddressBlockID,
		).Scan(&blockNetwork)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("address block %d not found", *input.AddressBlockID)
		}
		if err != nil {
			return err
		}

		_, blockNet, err := net.ParseCIDR(blockNetwork)
		if err != nil {
			return fmt.Errorf("address block has invalid network: %s", blockNetwork)
		}

		// Check that the entire subnet fits within the block:
		// the block must contain both the subnet's network address and broadcast address.
		subnetStart := subnetNet.IP
		if !blockNet.Contains(subnetStart) {
			return fmt.Errorf("subnet %s is not within address block %s", *input.Subnet, blockNetwork)
		}
		// Compute broadcast (last address in the subnet).
		broadcast := make(net.IP, len(subnetStart))
		for i := range subnetStart {
			broadcast[i] = subnetStart[i] | ^subnetNet.Mask[i]
		}
		if !blockNet.Contains(broadcast) {
			return fmt.Errorf("subnet %s is not fully within address block %s", *input.Subnet, blockNetwork)
		}
	}

	// Validate DHCP range is within subnet.
	if subnetNet != nil {
		if input.DHCPStart != nil && *input.DHCPStart != "" {
			startIP := net.ParseIP(*input.DHCPStart)
			if startIP == nil {
				return fmt.Errorf("invalid DHCP start IP: %s", *input.DHCPStart)
			}
			if !subnetNet.Contains(startIP) {
				return fmt.Errorf("DHCP start %s is not within subnet %s", *input.DHCPStart, *input.Subnet)
			}
		}
		if input.DHCPEnd != nil && *input.DHCPEnd != "" {
			endIP := net.ParseIP(*input.DHCPEnd)
			if endIP == nil {
				return fmt.Errorf("invalid DHCP end IP: %s", *input.DHCPEnd)
			}
			if !subnetNet.Contains(endIP) {
				return fmt.Errorf("DHCP end %s is not within subnet %s", *input.DHCPEnd, *input.Subnet)
			}
		}
	}

	return nil
}

// vlanColumns lists the columns read by scanVLAN (used in SELECT and RETURNING).
const vlanColumns = `id, site_id, address_block_id, vlan_id, name, subnet, gateway_device_ip_id, dhcp_start, dhcp_end, description`

// vlanSelectSQL is the base SELECT used by every read operation.
const vlanSelectSQL = `SELECT ` + vlanColumns + ` FROM vlans`

// scanVLAN reads one row into a VLAN struct.
func scanVLAN(row interface{ Scan(...any) error }) (models.VLAN, error) {
	var v models.VLAN
	err := row.Scan(&v.ID, &v.SiteID, &v.AddressBlockID, &v.VlanID, &v.Name, &v.Subnet, &v.GatewayDeviceIPID, &v.DHCPStart, &v.DHCPEnd, &v.Description)
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

	if err := h.validateVLAN(c.Request.Context(), &input, 0); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	v, err := scanVLAN(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO vlans (site_id, address_block_id, vlan_id, name, subnet, gateway_device_ip_id, dhcp_start, dhcp_end, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+vlanColumns,
		input.SiteID, input.AddressBlockID, input.VlanID, input.Name,
		input.Subnet, input.GatewayDeviceIPID, input.DHCPStart, input.DHCPEnd, input.Description,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "vlans", v.ID, fmt.Sprintf("Created VLAN '%s'", v.Name))
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

	if err := h.validateVLAN(c.Request.Context(), &input, id); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	v, err := scanVLAN(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE vlans SET site_id = $1, address_block_id = $2, vlan_id = $3, name = $4,
		 subnet = $5, gateway_device_ip_id = $6, dhcp_start = $7, dhcp_end = $8, description = $9
		 WHERE id = $10
		 RETURNING `+vlanColumns,
		input.SiteID, input.AddressBlockID, input.VlanID, input.Name,
		input.Subnet, input.GatewayDeviceIPID, input.DHCPStart, input.DHCPEnd, input.Description, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("vlan not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "vlans", id, fmt.Sprintf("Updated VLAN '%s'", v.Name))
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

	logAudit(c.Request.Context(), h.db, c, "delete", "vlans", id, fmt.Sprintf("Deleted VLAN #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
