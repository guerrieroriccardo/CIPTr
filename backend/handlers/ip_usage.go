package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// IPUsageHandler provides the aggregated IP utilization endpoint.
type IPUsageHandler struct {
	db *sql.DB
}

// NewIPUsageHandler creates an IPUsageHandler with the given database connection.
func NewIPUsageHandler(db *sql.DB) *IPUsageHandler {
	return &IPUsageHandler{db: db}
}

// subnetSize returns the number of usable host addresses for a CIDR string.
func subnetSize(cidr string) int {
	if cidr == "" {
		return 0
	}
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0
	}
	ones, bits := ipNet.Mask.Size()
	if bits == 0 {
		return 0
	}
	total := int(math.Pow(2, float64(bits-ones)))
	if bits-ones <= 1 {
		return total // /31 and /32: all addresses usable (RFC 3021)
	}
	return total - 2
}

// GetUsage handles GET /ip-usage
// Query params (use at most one): vlan_id, site_id, client_id.
func (h *IPUsageHandler) GetUsage(c *gin.Context) {
	ctx := c.Request.Context()

	if v := c.Query("vlan_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			fail(c, http.StatusBadRequest, fmt.Errorf("invalid vlan_id"))
			return
		}
		h.vlanLevel(c, ctx, id)
		return
	}
	if v := c.Query("site_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			fail(c, http.StatusBadRequest, fmt.Errorf("invalid site_id"))
			return
		}
		h.siteLevel(c, ctx, id)
		return
	}
	if v := c.Query("client_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			fail(c, http.StatusBadRequest, fmt.Errorf("invalid client_id"))
			return
		}
		h.clientLevel(c, ctx, id)
		return
	}
	h.globalLevel(c, ctx)
}

// ---------- VLAN level ----------

func (h *IPUsageHandler) vlanLevel(c *gin.Context, ctx context.Context, vlanID int64) {
	var vlanTag int64
	var vlanName string
	var subnet sql.NullString
	err := h.db.QueryRowContext(ctx,
		`SELECT vlan_id, name, subnet FROM vlans WHERE id = $1`, vlanID,
	).Scan(&vlanTag, &vlanName, &subnet)
	if err != nil {
		fail(c, http.StatusNotFound, fmt.Errorf("vlan not found"))
		return
	}

	total := 0
	if subnet.Valid {
		total = subnetSize(subnet.String)
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT dip.id, dip.ip_address, COALESCE(dip.is_primary, false),
		       di.name, d.hostname
		FROM device_ips dip
		JOIN device_interfaces di ON di.id = dip.interface_id
		JOIN devices d ON d.id = di.device_id
		WHERE dip.vlan_id = $1
		ORDER BY dip.ip_address::inet
	`, vlanID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var children []models.IPUsageNode
	for rows.Next() {
		var id int64
		var ip, iface, hostname string
		var primary bool
		if err := rows.Scan(&id, &ip, &primary, &iface, &hostname); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		label := ip + " - " + hostname + " (" + iface + ")"
		if primary {
			label += " *"
		}
		children = append(children, models.IPUsageNode{
			ID:    id,
			Label: label,
			Type:  "ip",
		})
	}

	label := fmt.Sprintf("VLAN %d - %s", vlanTag, vlanName)
	if subnet.Valid {
		label += " (" + subnet.String + ")"
	}

	ok(c, http.StatusOK, models.IPUsageResponse{
		Level: "vlan",
		Items: []models.IPUsageNode{{
			ID:       vlanID,
			Label:    label,
			Type:     "vlan",
			TotalIPs: total,
			UsedIPs:  len(children),
			Children: children,
		}},
	})
}

// ---------- Site level ----------

func (h *IPUsageHandler) siteLevel(c *gin.Context, ctx context.Context, siteID int64) {
	var siteName string
	if err := h.db.QueryRowContext(ctx, `SELECT name FROM sites WHERE id = $1`, siteID).Scan(&siteName); err != nil {
		fail(c, http.StatusNotFound, fmt.Errorf("site not found"))
		return
	}

	blocks, err := h.fetchBlocks(ctx, `WHERE ab.site_id = $1`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	vlans, err := h.fetchVLANUsage(ctx, `WHERE v.site_id = $1`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	blockNodes := buildBlockTree(blocks, vlans)
	siteNode := models.IPUsageNode{
		ID:       siteID,
		Label:    siteName,
		Type:     "site",
		Children: blockNodes,
	}
	for _, bn := range blockNodes {
		siteNode.TotalIPs += bn.TotalIPs
		siteNode.UsedIPs += bn.UsedIPs
	}

	ok(c, http.StatusOK, models.IPUsageResponse{
		Level: "site",
		Items: []models.IPUsageNode{siteNode},
	})
}

// ---------- Client level ----------

func (h *IPUsageHandler) clientLevel(c *gin.Context, ctx context.Context, clientID int64) {
	var clientName string
	if err := h.db.QueryRowContext(ctx, `SELECT name FROM clients WHERE id = $1`, clientID).Scan(&clientName); err != nil {
		fail(c, http.StatusNotFound, fmt.Errorf("client not found"))
		return
	}

	siteRows, err := h.db.QueryContext(ctx, `SELECT id, name FROM sites WHERE client_id = $1 ORDER BY name`, clientID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer siteRows.Close()

	type siteInfo struct {
		id   int64
		name string
	}
	var sites []siteInfo
	for siteRows.Next() {
		var s siteInfo
		if err := siteRows.Scan(&s.id, &s.name); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		sites = append(sites, s)
	}

	blocks, err := h.fetchBlocks(ctx,
		`WHERE ab.site_id IN (SELECT id FROM sites WHERE client_id = $1)`, clientID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	vlans, err := h.fetchVLANUsage(ctx,
		`WHERE v.site_id IN (SELECT id FROM sites WHERE client_id = $1)`, clientID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	blocksBySite := map[int64][]blockInfo{}
	for _, b := range blocks {
		blocksBySite[b.siteID] = append(blocksBySite[b.siteID], b)
	}
	vlansBySite := map[int64][]vlanUsage{}
	for _, v := range vlans {
		vlansBySite[v.siteID] = append(vlansBySite[v.siteID], v)
	}

	clientNode := models.IPUsageNode{
		ID:    clientID,
		Label: clientName,
		Type:  "client",
	}
	for _, s := range sites {
		blockNodes := buildBlockTree(blocksBySite[s.id], vlansBySite[s.id])
		siteNode := models.IPUsageNode{
			ID:       s.id,
			Label:    s.name,
			Type:     "site",
			Children: blockNodes,
		}
		for _, bn := range blockNodes {
			siteNode.TotalIPs += bn.TotalIPs
			siteNode.UsedIPs += bn.UsedIPs
		}
		clientNode.TotalIPs += siteNode.TotalIPs
		clientNode.UsedIPs += siteNode.UsedIPs
		clientNode.Children = append(clientNode.Children, siteNode)
	}

	ok(c, http.StatusOK, models.IPUsageResponse{
		Level: "client",
		Items: []models.IPUsageNode{clientNode},
	})
}

// ---------- Global level ----------

func (h *IPUsageHandler) globalLevel(c *gin.Context, ctx context.Context) {
	clientRows, err := h.db.QueryContext(ctx, `SELECT id, name FROM clients ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer clientRows.Close()

	type clientInfo struct {
		id   int64
		name string
	}
	var clients []clientInfo
	for clientRows.Next() {
		var cl clientInfo
		if err := clientRows.Scan(&cl.id, &cl.name); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		clients = append(clients, cl)
	}

	blocks, err := h.fetchBlocks(ctx, "", 0)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	vlans, err := h.fetchVLANUsage(ctx, "", 0)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	blocksBySite := map[int64][]blockInfo{}
	for _, b := range blocks {
		blocksBySite[b.siteID] = append(blocksBySite[b.siteID], b)
	}
	vlansBySite := map[int64][]vlanUsage{}
	for _, v := range vlans {
		vlansBySite[v.siteID] = append(vlansBySite[v.siteID], v)
	}

	siteRows, err := h.db.QueryContext(ctx, `SELECT id, client_id, name FROM sites ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer siteRows.Close()

	type siteInfo struct {
		id       int64
		clientID int64
		name     string
	}
	sitesByClient := map[int64][]siteInfo{}
	for siteRows.Next() {
		var s siteInfo
		if err := siteRows.Scan(&s.id, &s.clientID, &s.name); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		sitesByClient[s.clientID] = append(sitesByClient[s.clientID], s)
	}

	var items []models.IPUsageNode
	for _, cl := range clients {
		clientNode := models.IPUsageNode{
			ID:    cl.id,
			Label: cl.name,
			Type:  "client",
		}
		for _, s := range sitesByClient[cl.id] {
			blockNodes := buildBlockTree(blocksBySite[s.id], vlansBySite[s.id])
			siteNode := models.IPUsageNode{
				ID:       s.id,
				Label:    s.name,
				Type:     "site",
				Children: blockNodes,
			}
			for _, bn := range blockNodes {
				siteNode.TotalIPs += bn.TotalIPs
				siteNode.UsedIPs += bn.UsedIPs
			}
			clientNode.TotalIPs += siteNode.TotalIPs
			clientNode.UsedIPs += siteNode.UsedIPs
			clientNode.Children = append(clientNode.Children, siteNode)
		}
		items = append(items, clientNode)
	}

	ok(c, http.StatusOK, models.IPUsageResponse{
		Level: "global",
		Items: items,
	})
}

// ---------- Shared helpers ----------

type blockInfo struct {
	id      int64
	siteID  int64
	network string
}

type vlanUsage struct {
	id             int64
	siteID         int64
	addressBlockID sql.NullInt64
	vlanTag        int64
	name           string
	subnet         sql.NullString
	usedIPs        int
}

// fetchBlocks fetches address blocks. Pass empty whereClause for no filter.
func (h *IPUsageHandler) fetchBlocks(ctx context.Context, whereClause string, arg int64) ([]blockInfo, error) {
	q := `SELECT ab.id, ab.site_id, ab.network FROM address_blocks ab ` + whereClause + ` ORDER BY ab.network`
	var rows *sql.Rows
	var err error
	if whereClause == "" {
		rows, err = h.db.QueryContext(ctx, q)
	} else {
		rows, err = h.db.QueryContext(ctx, q, arg)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []blockInfo
	for rows.Next() {
		var b blockInfo
		if err := rows.Scan(&b.id, &b.siteID, &b.network); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

// fetchVLANUsage fetches VLANs with their used IP count. Pass empty whereClause for no filter.
func (h *IPUsageHandler) fetchVLANUsage(ctx context.Context, whereClause string, arg int64) ([]vlanUsage, error) {
	q := `
		SELECT v.id, v.site_id, v.address_block_id, v.vlan_id, v.name, v.subnet,
		       COUNT(dip.id) AS used_ips
		FROM vlans v
		LEFT JOIN device_ips dip ON dip.vlan_id = v.id
		` + whereClause + `
		GROUP BY v.id, v.site_id, v.address_block_id, v.vlan_id, v.name, v.subnet
		ORDER BY v.vlan_id
	`
	var rows *sql.Rows
	var err error
	if whereClause == "" {
		rows, err = h.db.QueryContext(ctx, q)
	} else {
		rows, err = h.db.QueryContext(ctx, q, arg)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []vlanUsage
	for rows.Next() {
		var v vlanUsage
		if err := rows.Scan(&v.id, &v.siteID, &v.addressBlockID, &v.vlanTag, &v.name, &v.subnet, &v.usedIPs); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

// buildBlockTree groups VLANs under their address blocks.
// VLANs without an address_block_id go under a synthetic "Unassigned" node.
func buildBlockTree(blocks []blockInfo, vlans []vlanUsage) []models.IPUsageNode {
	blockMap := map[int64]*models.IPUsageNode{}
	for _, b := range blocks {
		blockMap[b.id] = &models.IPUsageNode{
			ID:       b.id,
			Label:    b.network,
			Type:     "address_block",
			TotalIPs: subnetSize(b.network),
		}
	}

	var unassigned []models.IPUsageNode

	for _, v := range vlans {
		total := 0
		if v.subnet.Valid {
			total = subnetSize(v.subnet.String)
		}
		label := fmt.Sprintf("VLAN %d - %s", v.vlanTag, v.name)
		if v.subnet.Valid {
			label += " (" + v.subnet.String + ")"
		}
		vNode := models.IPUsageNode{
			ID:       v.id,
			Label:    label,
			Type:     "vlan",
			TotalIPs: total,
			UsedIPs:  v.usedIPs,
		}

		if v.addressBlockID.Valid {
			if bn, ok := blockMap[v.addressBlockID.Int64]; ok {
				bn.UsedIPs += v.usedIPs
				bn.Children = append(bn.Children, vNode)
				continue
			}
		}
		unassigned = append(unassigned, vNode)
	}

	var result []models.IPUsageNode
	for _, b := range blocks {
		if bn, ok := blockMap[b.id]; ok {
			result = append(result, *bn)
		}
	}
	if len(unassigned) > 0 {
		unNode := models.IPUsageNode{
			Label:    "Unassigned VLANs",
			Type:     "address_block",
			Children: unassigned,
		}
		for _, u := range unassigned {
			unNode.TotalIPs += u.TotalIPs
			unNode.UsedIPs += u.UsedIPs
		}
		result = append(result, unNode)
	}

	return result
}
