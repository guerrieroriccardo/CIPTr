package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// FirewallRuleHandler groups all HTTP handlers for the /firewall-rules resource.
type FirewallRuleHandler struct {
	db *sql.DB
}

// NewFirewallRuleHandler creates a FirewallRuleHandler with the given database connection.
func NewFirewallRuleHandler(db *sql.DB) *FirewallRuleHandler {
	return &FirewallRuleHandler{db: db}
}

const firewallRuleColumns = `id, site_id,
	src_device_id, src_group_id, src_vlan_id, src_cidr,
	dst_device_id, dst_group_id, dst_vlan_id, dst_cidr,
	src_port, dst_port, protocol, action, position, enabled,
	description, notes, created_at, updated_at`

const firewallRuleSelectSQL = `SELECT ` + firewallRuleColumns + ` FROM firewall_rules`

// scanFirewallRule reads one row into a FirewallRule struct.
func scanFirewallRule(row interface{ Scan(...any) error }) (models.FirewallRule, error) {
	var r models.FirewallRule
	err := row.Scan(
		&r.ID, &r.SiteID,
		&r.SrcDeviceID, &r.SrcGroupID, &r.SrcVlanID, &r.SrcCIDR,
		&r.DstDeviceID, &r.DstGroupID, &r.DstVlanID, &r.DstCIDR,
		&r.SrcPort, &r.DstPort, &r.Protocol, &r.Action, &r.Position, &r.Enabled,
		&r.Description, &r.Notes, &r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

var validProtocols = map[string]bool{"tcp": true, "udp": true, "both": true, "icmp": true, "any": true}
var validActions = map[string]bool{"allow": true, "deny": true}

// validateFirewallRule checks that at most one src/dst endpoint is set and that
// all FK-referenced entities belong to the rule's site.
func (h *FirewallRuleHandler) validateFirewallRule(c *gin.Context, input *models.FirewallRuleInput) error {
	// Check single src endpoint.
	srcCount := countNonNil(input.SrcDeviceID, input.SrcGroupID, input.SrcVlanID, input.SrcCIDR)
	if srcCount > 1 {
		return errors.New("at most one source endpoint (device, group, vlan, cidr) may be set")
	}

	// Check single dst endpoint.
	dstCount := countNonNil(input.DstDeviceID, input.DstGroupID, input.DstVlanID, input.DstCIDR)
	if dstCount > 1 {
		return errors.New("at most one destination endpoint (device, group, vlan, cidr) may be set")
	}

	// Validate protocol.
	if input.Protocol != nil && !validProtocols[*input.Protocol] {
		return fmt.Errorf("invalid protocol '%s'; must be one of: tcp, udp, both, icmp, any", *input.Protocol)
	}

	// Validate action.
	if input.Action != nil && !validActions[*input.Action] {
		return fmt.Errorf("invalid action '%s'; must be 'allow' or 'deny'", *input.Action)
	}

	// Validate FK entities belong to the rule's site.
	reqCtx := c.Request.Context()

	if input.SrcDeviceID != nil {
		if err := h.checkSite(reqCtx, "devices", *input.SrcDeviceID, input.SiteID, "source device"); err != nil {
			return err
		}
	}
	if input.SrcGroupID != nil {
		if err := h.checkSite(reqCtx, "device_groups", *input.SrcGroupID, input.SiteID, "source group"); err != nil {
			return err
		}
	}
	if input.SrcVlanID != nil {
		if err := h.checkSite(reqCtx, "vlans", *input.SrcVlanID, input.SiteID, "source VLAN"); err != nil {
			return err
		}
	}
	if input.DstDeviceID != nil {
		if err := h.checkSite(reqCtx, "devices", *input.DstDeviceID, input.SiteID, "destination device"); err != nil {
			return err
		}
	}
	if input.DstGroupID != nil {
		if err := h.checkSite(reqCtx, "device_groups", *input.DstGroupID, input.SiteID, "destination group"); err != nil {
			return err
		}
	}
	if input.DstVlanID != nil {
		if err := h.checkSite(reqCtx, "vlans", *input.DstVlanID, input.SiteID, "destination VLAN"); err != nil {
			return err
		}
	}

	return nil
}

// checkSite verifies that the entity in the given table belongs to the expected site.
func (h *FirewallRuleHandler) checkSite(ctx context.Context, table string, entityID, expectedSiteID int64, label string) error {
	var siteID int64
	err := h.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT site_id FROM %s WHERE id = $1`, table), entityID,
	).Scan(&siteID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s not found", label)
	}
	if err != nil {
		return err
	}
	if siteID != expectedSiteID {
		return fmt.Errorf("%s must belong to the rule's site", label)
	}
	return nil
}

// countNonNil counts how many of the given interface pointers are non-nil.
func countNonNil(ptrs ...any) int {
	count := 0
	for _, p := range ptrs {
		switch v := p.(type) {
		case *int64:
			if v != nil {
				count++
			}
		case *string:
			if v != nil {
				count++
			}
		}
	}
	return count
}

// List handles GET /firewall-rules
// Supports optional query param: ?site_id=
func (h *FirewallRuleHandler) List(c *gin.Context) {
	query := firewallRuleSelectSQL
	var conds []string
	var args []any
	n := 1

	if siteID := c.Query("site_id"); siteID != "" {
		conds = append(conds, fmt.Sprintf("site_id = $%d", n))
		args = append(args, siteID)
		n++
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY position"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	rules := []models.FirewallRule{}
	for rows.Next() {
		r, err := scanFirewallRule(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		rules = append(rules, r)
	}

	ok(c, http.StatusOK, rules)
}

// ListBySite handles GET /sites/:id/firewall-rules
func (h *FirewallRuleHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		firewallRuleSelectSQL+` WHERE site_id = $1 ORDER BY position`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	rules := []models.FirewallRule{}
	for rows.Next() {
		r, err := scanFirewallRule(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		rules = append(rules, r)
	}

	ok(c, http.StatusOK, rules)
}

// GetByID handles GET /firewall-rules/:id
func (h *FirewallRuleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	r, err := scanFirewallRule(h.db.QueryRowContext(c.Request.Context(),
		firewallRuleSelectSQL+` WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("firewall rule not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, r)
}

// Create handles POST /firewall-rules
func (h *FirewallRuleHandler) Create(c *gin.Context) {
	var input models.FirewallRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validateFirewallRule(c, &input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// Auto-assign position if not provided.
	if input.Position == nil {
		var pos int
		h.db.QueryRowContext(c.Request.Context(),
			`SELECT COALESCE(MAX(position), 0) + 1 FROM firewall_rules WHERE site_id = $1`,
			input.SiteID,
		).Scan(&pos)
		input.Position = &pos
	}

	// Default values.
	srcPort := "*"
	if input.SrcPort != nil {
		srcPort = *input.SrcPort
	}
	dstPort := "*"
	if input.DstPort != nil {
		dstPort = *input.DstPort
	}
	protocol := "any"
	if input.Protocol != nil {
		protocol = *input.Protocol
	}
	action := "allow"
	if input.Action != nil {
		action = *input.Action
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	r, err := scanFirewallRule(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO firewall_rules (
			site_id,
			src_device_id, src_group_id, src_vlan_id, src_cidr,
			dst_device_id, dst_group_id, dst_vlan_id, dst_cidr,
			src_port, dst_port, protocol, action, position, enabled,
			description, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING `+firewallRuleColumns,
		input.SiteID,
		input.SrcDeviceID, input.SrcGroupID, input.SrcVlanID, input.SrcCIDR,
		input.DstDeviceID, input.DstGroupID, input.DstVlanID, input.DstCIDR,
		srcPort, dstPort, protocol, action, *input.Position, enabled,
		input.Description, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "firewall_rules", r.ID,
		fmt.Sprintf("Created firewall rule #%d (pos %d, %s)", r.ID, r.Position, r.Action))
	ok(c, http.StatusCreated, r)
}

// Update handles PUT /firewall-rules/:id
func (h *FirewallRuleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.FirewallRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if err := h.validateFirewallRule(c, &input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// Default values.
	srcPort := "*"
	if input.SrcPort != nil {
		srcPort = *input.SrcPort
	}
	dstPort := "*"
	if input.DstPort != nil {
		dstPort = *input.DstPort
	}
	protocol := "any"
	if input.Protocol != nil {
		protocol = *input.Protocol
	}
	action := "allow"
	if input.Action != nil {
		action = *input.Action
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	position := 0
	if input.Position != nil {
		position = *input.Position
	}

	r, err := scanFirewallRule(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE firewall_rules SET
			site_id = $1,
			src_device_id = $2, src_group_id = $3, src_vlan_id = $4, src_cidr = $5,
			dst_device_id = $6, dst_group_id = $7, dst_vlan_id = $8, dst_cidr = $9,
			src_port = $10, dst_port = $11, protocol = $12, action = $13,
			position = $14, enabled = $15,
			description = $16, notes = $17,
			updated_at = NOW()
		WHERE id = $18
		RETURNING `+firewallRuleColumns,
		input.SiteID,
		input.SrcDeviceID, input.SrcGroupID, input.SrcVlanID, input.SrcCIDR,
		input.DstDeviceID, input.DstGroupID, input.DstVlanID, input.DstCIDR,
		srcPort, dstPort, protocol, action,
		position, enabled,
		input.Description, input.Notes,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("firewall rule not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "firewall_rules", id,
		fmt.Sprintf("Updated firewall rule #%d (pos %d, %s)", id, r.Position, r.Action))
	ok(c, http.StatusOK, r)
}

// Delete handles DELETE /firewall-rules/:id
func (h *FirewallRuleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM firewall_rules WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("firewall rule not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "firewall_rules", id,
		fmt.Sprintf("Deleted firewall rule #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
