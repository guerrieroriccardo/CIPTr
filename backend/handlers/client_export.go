package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	qrcode "github.com/skip2/go-qrcode"
)

// ---------------------------------------------------------------------------
// Internal structs for PDF export (not shared with models — export-only)
// ---------------------------------------------------------------------------

type exportClient struct {
	ID        int64
	Name      string
	ShortCode string
	Domain    string
	Notes     string
}

type exportSite struct {
	ID      int64
	Name    string
	Address string
	Domain  string
	Notes   string
}

type exportLocation struct {
	Name  string
	Floor string
	Notes string
}

type exportAddressBlock struct {
	Network     string
	Description string
	Notes       string
}

type exportVLAN struct {
	VlanID    int64
	Name      string
	Subnet    string
	Gateway   string
	DHCPStart string
	DHCPEnd   string
}

type exportSwitch struct {
	ID         int64
	Hostname   string
	IPAddress  string
	Model      string
	Location   string
	TotalPorts int
	Notes      string
}

type exportSwitchPort struct {
	PortNumber int
	PortLabel  string
	Speed      string
	IsUplink   bool
	Notes      string
}

type exportPatchPanel struct {
	ID         int64
	Name       string
	TotalPorts int
	Location   string
	Notes      string
}

type exportPatchPanelPort struct {
	PortNumber int
	PortLabel  string
	LinkedPort string
	Notes      string
}

type exportDevice struct {
	ID               int64
	Hostname         string
	DnsName          string
	SerialNumber     string
	AssetTag         string
	Category         string
	Status           string
	IsUp             bool
	OS               string
	HasRmm           bool
	HasAntivirus     bool
	Supplier         string
	Model            string
	Location         string
	InstallationDate string
	VmID             string
	Notes            string
}

type exportDeviceInterface struct {
	DeviceID   int64
	Name       string
	MacAddress string
}

type exportDeviceIP struct {
	InterfaceDeviceID int64
	InterfaceName     string
	IPAddress         string
	VlanName          string
	IsPrimary         bool
}

type exportDeviceConnection struct {
	InterfaceDeviceID int64
	InterfaceName     string
	SwitchPort        string
	PatchPanelPort    string
}

type exportDeviceGroup struct {
	ID          int64
	Name        string
	Description string
}

type exportDeviceGroupMember struct {
	GroupID        int64
	DeviceHostname string
}

type exportFirewallRule struct {
	Position    int
	Protocol    string
	SrcPort     string
	DstPort     string
	Src         string
	Dst         string
	Action      string
	Enabled     bool
	Description string
}

type exportBackupPolicy struct {
	Name          string
	Destination   string
	Source        string
	Enabled       bool
	ScheduleTimes string
	RetainLast    int
	RetainHourly  int
	RetainDaily   int
	RetainWeekly  int
	RetainMonthly int
	RetainYearly  int
	Notes         string
}

// ---------------------------------------------------------------------------
// PDF styling constants
// ---------------------------------------------------------------------------

const (
	pdfMargin      = 10.0
	pdfPageW       = 210.0
	pdfPageH       = 297.0
	pdfContentW    = pdfPageW - 2*pdfMargin
	pdfHeaderH     = 7.0
	pdfRowH        = 5.0
	pdfFontSize    = 8.0
	pdfTitleSize   = 14.0
	pdfSubtitleSize = 11.0
)

// Export handles GET /clients/:id/export
// Returns a comprehensive A4 PDF with all client data.
func (h *ClientHandler) Export(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	ctx := c.Request.Context()

	// Fetch client
	client, err := fetchExportClient(ctx, h.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("client not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	// Fetch all related data
	sites, err := fetchExportSites(ctx, h.db, id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	siteIDs := make([]int64, len(sites))
	for i, s := range sites {
		siteIDs[i] = s.ID
	}

	// Fetch per-site data (only if there are sites)
	var (
		locationsBySite     = map[int64][]exportLocation{}
		addressBlocksBySite = map[int64][]exportAddressBlock{}
		vlansBySite         = map[int64][]exportVLAN{}
		switchesBySite      = map[int64][]exportSwitch{}
		switchPortsBySwitch = map[int64][]exportSwitchPort{}
		patchPanelsBySite   = map[int64][]exportPatchPanel{}
		ppPortsByPanel      = map[int64][]exportPatchPanelPort{}
		devicesBySite       = map[int64][]exportDevice{}
		ifacesByDevice      = map[int64][]exportDeviceInterface{}
		ipsByDevice         = map[int64][]exportDeviceIP{}
		connsByDevice       = map[int64][]exportDeviceConnection{}
		groupsBySite        = map[int64][]exportDeviceGroup{}
		membersByGroup      = map[int64][]exportDeviceGroupMember{}
		firewallsBySite     = map[int64][]exportFirewallRule{}
	)

	if len(siteIDs) > 0 {
		if locationsBySite, err = fetchExportLocations(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if addressBlocksBySite, err = fetchExportAddressBlocks(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if vlansBySite, err = fetchExportVLANs(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if switchesBySite, switchPortsBySwitch, err = fetchExportSwitches(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if patchPanelsBySite, ppPortsByPanel, err = fetchExportPatchPanels(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if devicesBySite, ifacesByDevice, ipsByDevice, connsByDevice, err = fetchExportDevices(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if groupsBySite, membersByGroup, err = fetchExportDeviceGroups(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		if firewallsBySite, err = fetchExportFirewallRules(ctx, h.db, siteIDs); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
	}

	backupPolicies, err := fetchExportBackupPolicies(ctx, h.db, id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	// Count totals for cover page
	totalDevices := 0
	totalSwitches := 0
	totalVLANs := 0
	for _, sid := range siteIDs {
		totalDevices += len(devicesBySite[sid])
		totalSwitches += len(switchesBySite[sid])
		totalVLANs += len(vlansBySite[sid])
	}

	// --- Build PDF ---
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, pdfMargin+5)

	// Header/footer on every page
	pdf.SetHeaderFuncMode(func() {
		if pdf.PageNo() == 1 {
			return // skip header on cover
		}
		pdf.SetFont("Helvetica", "I", 7)
		pdf.SetTextColor(128, 128, 128)
		pdf.SetXY(pdfMargin, 3)
		pdf.CellFormat(pdfContentW, 5, client.Name+" - Export "+time.Now().Format("2006-01-02"), "", 0, "L", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetFont("Helvetica", "I", 7)
		pdf.SetTextColor(128, 128, 128)
		pdf.SetXY(pdfMargin, pdfPageH-8)
		pdf.CellFormat(pdfContentW, 5, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	// --- Cover page ---
	pdf.AddPage()
	pdf.Ln(30)

	// QR code
	baseURL := os.Getenv("LABEL_BASE_URL")
	if baseURL == "" {
		baseURL = "https://ciptr.example.com"
	}
	qrURL := fmt.Sprintf("%s/clients/%d", baseURL, client.ID)
	qrPNG, err := qrcode.Encode(qrURL, qrcode.High, 256)
	if err == nil {
		qrReader := bytes.NewReader(qrPNG)
		opts := fpdf.ImageOptions{ImageType: "PNG"}
		pdf.RegisterImageOptionsReader("cover-qr", opts, qrReader)
		qrSize := 40.0
		pdf.ImageOptions("cover-qr", (pdfPageW-qrSize)/2, pdf.GetY(), qrSize, qrSize, false, opts, 0, "")
		pdf.Ln(qrSize + 5)
	}

	pdf.SetFont("Helvetica", "B", 24)
	pdf.CellFormat(pdfContentW, 12, client.Name, "", 1, "C", false, 0, "")
	pdf.Ln(3)
	pdf.SetFont("Helvetica", "", 14)
	pdf.CellFormat(pdfContentW, 8, "Short Code: "+client.ShortCode, "", 1, "C", false, 0, "")
	if client.Domain != "" {
		pdf.CellFormat(pdfContentW, 8, "Domain: "+client.Domain, "", 1, "C", false, 0, "")
	}
	if client.Notes != "" {
		pdf.Ln(3)
		pdf.SetFont("Helvetica", "I", 10)
		pdf.MultiCell(pdfContentW, 5, client.Notes, "", "C", false)
	}

	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(pdfContentW, 7, fmt.Sprintf("Sites: %d", len(sites)), "", 1, "C", false, 0, "")
	pdf.CellFormat(pdfContentW, 7, fmt.Sprintf("Devices: %d", totalDevices), "", 1, "C", false, 0, "")
	pdf.CellFormat(pdfContentW, 7, fmt.Sprintf("Switches: %d", totalSwitches), "", 1, "C", false, 0, "")
	pdf.CellFormat(pdfContentW, 7, fmt.Sprintf("VLANs: %d", totalVLANs), "", 1, "C", false, 0, "")
	pdf.CellFormat(pdfContentW, 7, fmt.Sprintf("Backup Policies: %d", len(backupPolicies)), "", 1, "C", false, 0, "")

	pdf.Ln(15)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.CellFormat(pdfContentW, 5, "Generated: "+time.Now().Format("2006-01-02 15:04"), "", 1, "C", false, 0, "")

	// --- Per-site sections ---
	for _, site := range sites {
		pdf.AddPage()
		pdfSectionTitle(pdf, fmt.Sprintf("Site: %s", site.Name))

		// Site details
		siteRows := [][]string{}
		if site.Address != "" {
			siteRows = append(siteRows, []string{"Address", site.Address})
		}
		if site.Domain != "" {
			siteRows = append(siteRows, []string{"Domain", site.Domain})
		}
		if site.Notes != "" {
			siteRows = append(siteRows, []string{"Notes", site.Notes})
		}
		if len(siteRows) > 0 {
			pdfTable(pdf, []string{"Property", "Value"}, []float64{40, 150}, siteRows)
		}

		// Locations
		locs := locationsBySite[site.ID]
		if len(locs) > 0 {
			pdfSubsectionTitle(pdf, "Locations")
			pdfTable(pdf, []string{"Name", "Floor", "Notes"}, []float64{60, 30, 100}, func() [][]string {
				rows := make([][]string, len(locs))
				for i, l := range locs {
					rows[i] = []string{l.Name, l.Floor, l.Notes}
				}
				return rows
			}())
		}

		// Address Blocks
		blocks := addressBlocksBySite[site.ID]
		if len(blocks) > 0 {
			pdfSubsectionTitle(pdf, "Address Blocks")
			pdfTable(pdf, []string{"Network", "Description", "Notes"}, []float64{50, 70, 70}, func() [][]string {
				rows := make([][]string, len(blocks))
				for i, b := range blocks {
					rows[i] = []string{b.Network, b.Description, b.Notes}
				}
				return rows
			}())
		}

		// VLANs
		vlans := vlansBySite[site.ID]
		if len(vlans) > 0 {
			pdfSubsectionTitle(pdf, "VLANs")
			pdfTable(pdf, []string{"VLAN ID", "Name", "Subnet", "Gateway", "DHCP Start", "DHCP End"},
				[]float64{25, 35, 40, 30, 30, 30}, func() [][]string {
					rows := make([][]string, len(vlans))
					for i, v := range vlans {
						rows[i] = []string{
							fmt.Sprintf("%d", v.VlanID), v.Name, v.Subnet,
							v.Gateway, v.DHCPStart, v.DHCPEnd,
						}
					}
					return rows
				}())
		}

		// Switches
		sws := switchesBySite[site.ID]
		for _, sw := range sws {
			pdfSubsectionTitle(pdf, fmt.Sprintf("Switch: %s", sw.Hostname))
			pdf.SetFont("Helvetica", "", pdfFontSize)
			if sw.IPAddress != "" {
				pdf.CellFormat(pdfContentW, pdfRowH, "IP: "+sw.IPAddress, "", 1, "L", false, 0, "")
			}
			if sw.Model != "" {
				pdf.CellFormat(pdfContentW, pdfRowH, "Model: "+sw.Model, "", 1, "L", false, 0, "")
			}
			if sw.Location != "" {
				pdf.CellFormat(pdfContentW, pdfRowH, "Location: "+sw.Location, "", 1, "L", false, 0, "")
			}
			pdf.CellFormat(pdfContentW, pdfRowH, fmt.Sprintf("Total Ports: %d", sw.TotalPorts), "", 1, "L", false, 0, "")
			if sw.Notes != "" {
				pdf.CellFormat(pdfContentW, pdfRowH, "Notes: "+sw.Notes, "", 1, "L", false, 0, "")
			}

			ports := switchPortsBySwitch[sw.ID]
			if len(ports) > 0 {
				pdfTable(pdf, []string{"Port #", "Label", "Speed", "Uplink", "Notes"},
					[]float64{25, 40, 35, 25, 65}, func() [][]string {
						rows := make([][]string, len(ports))
						for i, p := range ports {
							rows[i] = []string{
								fmt.Sprintf("%d", p.PortNumber), p.PortLabel, p.Speed,
								boolStr(p.IsUplink), p.Notes,
							}
						}
						return rows
					}())
			}
		}

		// Patch Panels
		panels := patchPanelsBySite[site.ID]
		for _, pp := range panels {
			pdfSubsectionTitle(pdf, fmt.Sprintf("Patch Panel: %s", pp.Name))
			pdf.SetFont("Helvetica", "", pdfFontSize)
			pdf.CellFormat(pdfContentW, pdfRowH, fmt.Sprintf("Total Ports: %d", pp.TotalPorts), "", 1, "L", false, 0, "")
			if pp.Location != "" {
				pdf.CellFormat(pdfContentW, pdfRowH, "Location: "+pp.Location, "", 1, "L", false, 0, "")
			}
			if pp.Notes != "" {
				pdf.CellFormat(pdfContentW, pdfRowH, "Notes: "+pp.Notes, "", 1, "L", false, 0, "")
			}

			ports := ppPortsByPanel[pp.ID]
			if len(ports) > 0 {
				pdfTable(pdf, []string{"Port #", "Label", "Linked To", "Notes"},
					[]float64{25, 50, 55, 60}, func() [][]string {
						rows := make([][]string, len(ports))
						for i, p := range ports {
							rows[i] = []string{
								fmt.Sprintf("%d", p.PortNumber), p.PortLabel,
								p.LinkedPort, p.Notes,
							}
						}
						return rows
					}())
			}
		}

		// Devices
		devs := devicesBySite[site.ID]
		if len(devs) > 0 {
			pdfSubsectionTitle(pdf, "Devices")
			pdfTable(pdf, []string{"Hostname", "Category", "Status", "IP", "S/N", "Model", "Location"},
				[]float64{32, 22, 22, 28, 28, 30, 28}, func() [][]string {
					rows := make([][]string, len(devs))
					for i, d := range devs {
						primaryIP := ""
						for _, ip := range ipsByDevice[d.ID] {
							if ip.IsPrimary {
								primaryIP = ip.IPAddress
								break
							}
						}
						if primaryIP == "" {
							// fallback to first IP
							if ips := ipsByDevice[d.ID]; len(ips) > 0 {
								primaryIP = ips[0].IPAddress
							}
						}
						rows[i] = []string{
							d.Hostname, d.Category, d.Status,
							primaryIP, d.SerialNumber, d.Model, d.Location,
						}
					}
					return rows
				}())

			// Per-device detail: interfaces, IPs, connections
			for _, d := range devs {
				ifaces := ifacesByDevice[d.ID]
				ips := ipsByDevice[d.ID]
				conns := connsByDevice[d.ID]
				if len(ifaces) == 0 && len(ips) == 0 && len(conns) == 0 {
					continue
				}

				pdfCheckPageBreak(pdf, 20)
				pdf.SetFont("Helvetica", "B", pdfFontSize)
				pdf.CellFormat(pdfContentW, pdfRowH, "  "+d.Hostname+" - Details", "", 1, "L", false, 0, "")

				if len(ifaces) > 0 {
					pdfTable(pdf, []string{"Interface", "MAC Address"},
						[]float64{60, 130}, func() [][]string {
							rows := make([][]string, len(ifaces))
							for i, iface := range ifaces {
								rows[i] = []string{iface.Name, iface.MacAddress}
							}
							return rows
						}())
				}
				if len(ips) > 0 {
					pdfTable(pdf, []string{"Interface", "IP Address", "VLAN", "Primary"},
						[]float64{40, 45, 50, 55}, func() [][]string {
							rows := make([][]string, len(ips))
							for i, ip := range ips {
								rows[i] = []string{ip.InterfaceName, ip.IPAddress, ip.VlanName, boolStr(ip.IsPrimary)}
							}
							return rows
						}())
				}
				if len(conns) > 0 {
					pdfTable(pdf, []string{"Interface", "Switch Port", "Patch Panel Port"},
						[]float64{50, 70, 70}, func() [][]string {
							rows := make([][]string, len(conns))
							for i, conn := range conns {
								rows[i] = []string{conn.InterfaceName, conn.SwitchPort, conn.PatchPanelPort}
							}
							return rows
						}())
				}
			}
		}

		// Device Groups
		groups := groupsBySite[site.ID]
		if len(groups) > 0 {
			pdfSubsectionTitle(pdf, "Device Groups")
			groupRows := make([][]string, len(groups))
			for i, g := range groups {
				members := membersByGroup[g.ID]
				hostnames := make([]string, len(members))
				for j, m := range members {
					hostnames[j] = m.DeviceHostname
				}
				groupRows[i] = []string{g.Name, g.Description, strings.Join(hostnames, ", ")}
			}
			pdfTable(pdf, []string{"Group", "Description", "Members"}, []float64{40, 50, 100}, groupRows)
		}

		// Firewall Rules
		fwRules := firewallsBySite[site.ID]
		if len(fwRules) > 0 {
			pdfSubsectionTitle(pdf, "Firewall Rules")
			pdfTable(pdf, []string{"#", "Proto", "Src", "SPort", "Dst", "DPort", "Action", "On"},
				[]float64{12, 18, 40, 22, 40, 22, 18, 18}, func() [][]string {
					rows := make([][]string, len(fwRules))
					for i, r := range fwRules {
						rows[i] = []string{
							fmt.Sprintf("%d", r.Position), r.Protocol,
							r.Src, r.SrcPort, r.Dst, r.DstPort,
							r.Action, boolStr(r.Enabled),
						}
					}
					return rows
				}())
		}
	}

	// --- Backup Policies (client-level) ---
	if len(backupPolicies) > 0 {
		pdf.AddPage()
		pdfSectionTitle(pdf, "Backup Policies")
		bpRows := make([][]string, len(backupPolicies))
		for i, bp := range backupPolicies {
			retention := fmt.Sprintf("L:%d H:%d D:%d W:%d M:%d Y:%d",
				bp.RetainLast, bp.RetainHourly, bp.RetainDaily, bp.RetainWeekly, bp.RetainMonthly, bp.RetainYearly)
			bpRows[i] = []string{
				bp.Name, bp.Destination, bp.Source,
				boolStr(bp.Enabled), bp.ScheduleTimes, retention, bp.Notes,
			}
		}
		pdfTable(pdf, []string{"Name", "Destination", "Source", "On", "Schedule", "Retention", "Notes"},
			[]float64{30, 30, 25, 14, 25, 40, 26}, bpRows)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		fail(c, http.StatusInternalServerError, fmt.Errorf("generate PDF: %w", err))
		return
	}

	logAudit(ctx, h.db, c, "export", "clients", id, fmt.Sprintf("Exported client '%s' to PDF", client.Name))

	filename := fmt.Sprintf("export-%s.pdf", strings.ToLower(client.ShortCode))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// ---------------------------------------------------------------------------
// PDF helpers
// ---------------------------------------------------------------------------

func pdfSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdfCheckPageBreak(pdf, 12)
	pdf.SetFillColor(40, 40, 40)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", pdfTitleSize)
	pdf.CellFormat(pdfContentW, 9, "  "+title, "", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(3)
}

func pdfSubsectionTitle(pdf *fpdf.Fpdf, title string) {
	pdfCheckPageBreak(pdf, 12)
	pdf.SetFillColor(220, 220, 220)
	pdf.SetFont("Helvetica", "B", pdfSubtitleSize)
	pdf.CellFormat(pdfContentW, 7, "  "+title, "", 1, "L", true, 0, "")
	pdf.Ln(2)
}

func pdfTable(pdf *fpdf.Fpdf, headers []string, widths []float64, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	pdfCheckPageBreak(pdf, pdfHeaderH+pdfRowH*2)

	// Header
	pdf.SetFillColor(60, 60, 60)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", pdfFontSize)
	for i, h := range headers {
		pdf.CellFormat(widths[i], pdfHeaderH, " "+h, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(0, 0, 0)

	// Rows
	pdf.SetFont("Helvetica", "", pdfFontSize)
	for r, row := range rows {
		pdfCheckPageBreak(pdf, pdfRowH)
		if r%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		for i, cell := range row {
			if i < len(widths) {
				pdf.CellFormat(widths[i], pdfRowH, " "+cell, "1", 0, "L", true, 0, "")
			}
		}
		pdf.Ln(-1)
	}
	pdf.Ln(3)
}

func pdfCheckPageBreak(pdf *fpdf.Fpdf, h float64) {
	if pdf.GetY()+h > pdfPageH-pdfMargin-5 {
		pdf.AddPage()
	}
}

func deref(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func derefInt64(n *int64) string {
	if n != nil {
		return fmt.Sprintf("%d", *n)
	}
	return ""
}

func boolStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// ---------------------------------------------------------------------------
// SQL data-fetching functions
// ---------------------------------------------------------------------------

func fetchExportClient(ctx context.Context, db *sql.DB, clientID int64) (exportClient, error) {
	var c exportClient
	var domain, notes *string
	err := db.QueryRowContext(ctx,
		`SELECT id, name, short_code, domain, notes FROM clients WHERE id = $1`, clientID,
	).Scan(&c.ID, &c.Name, &c.ShortCode, &domain, &notes)
	c.Domain = deref(domain)
	c.Notes = deref(notes)
	return c, err
}

func fetchExportSites(ctx context.Context, db *sql.DB, clientID int64) ([]exportSite, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, COALESCE(address, ''), COALESCE(domain, ''), COALESCE(notes, '')
		 FROM sites WHERE client_id = $1 ORDER BY name`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []exportSite
	for rows.Next() {
		var s exportSite
		if err := rows.Scan(&s.ID, &s.Name, &s.Address, &s.Domain, &s.Notes); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func fetchExportLocations(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportLocation, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT site_id, name, COALESCE(floor, ''), COALESCE(notes, '')
		 FROM locations WHERE site_id = ANY($1) ORDER BY name`, siteIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int64][]exportLocation{}
	for rows.Next() {
		var siteID int64
		var l exportLocation
		if err := rows.Scan(&siteID, &l.Name, &l.Floor, &l.Notes); err != nil {
			return nil, err
		}
		result[siteID] = append(result[siteID], l)
	}
	return result, rows.Err()
}

func fetchExportAddressBlocks(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportAddressBlock, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT site_id, network::text, COALESCE(description, ''), COALESCE(notes, '')
		 FROM address_blocks WHERE site_id = ANY($1) ORDER BY network`, siteIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int64][]exportAddressBlock{}
	for rows.Next() {
		var siteID int64
		var b exportAddressBlock
		if err := rows.Scan(&siteID, &b.Network, &b.Description, &b.Notes); err != nil {
			return nil, err
		}
		result[siteID] = append(result[siteID], b)
	}
	return result, rows.Err()
}

func fetchExportVLANs(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportVLAN, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT v.site_id, v.vlan_id, v.name, COALESCE(v.subnet::text, ''),
		        COALESCE(dip.ip_address::text, ''),
		        COALESCE(v.dhcp_start::text, ''), COALESCE(v.dhcp_end::text, '')
		 FROM vlans v
		 LEFT JOIN device_ips dip ON dip.id = v.gateway_device_ip_id
		 WHERE v.site_id = ANY($1) ORDER BY v.vlan_id`, siteIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int64][]exportVLAN{}
	for rows.Next() {
		var siteID int64
		var v exportVLAN
		if err := rows.Scan(&siteID, &v.VlanID, &v.Name, &v.Subnet, &v.Gateway, &v.DHCPStart, &v.DHCPEnd); err != nil {
			return nil, err
		}
		result[siteID] = append(result[siteID], v)
	}
	return result, rows.Err()
}

func fetchExportSwitches(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportSwitch, map[int64][]exportSwitchPort, error) {
	// Switches
	srows, err := db.QueryContext(ctx,
		`SELECT s.id, s.site_id, s.hostname, COALESCE(s.ip_address::text, ''),
		        COALESCE(CONCAT(m.name, ' ', dm.model_name), ''),
		        COALESCE(l.name, ''), s.total_ports, COALESCE(s.notes, '')
		 FROM switches s
		 LEFT JOIN device_models dm ON dm.id = s.model_id
		 LEFT JOIN manufacturers m ON m.id = dm.manufacturer_id
		 LEFT JOIN locations l ON l.id = s.location_id
		 WHERE s.site_id = ANY($1) ORDER BY s.hostname`, siteIDs)
	if err != nil {
		return nil, nil, err
	}
	defer srows.Close()

	switchBySite := map[int64][]exportSwitch{}
	var switchIDs []int64
	for srows.Next() {
		var siteID int64
		var sw exportSwitch
		if err := srows.Scan(&siteID, &sw.ID, &sw.Hostname, &sw.IPAddress, &sw.Model, &sw.Location, &sw.TotalPorts, &sw.Notes); err != nil {
			return nil, nil, err
		}
		switchBySite[siteID] = append(switchBySite[siteID], sw)
		switchIDs = append(switchIDs, sw.ID)
	}
	if err := srows.Err(); err != nil {
		return nil, nil, err
	}

	// Switch ports
	portsBySwitch := map[int64][]exportSwitchPort{}
	if len(switchIDs) > 0 {
		prows, err := db.QueryContext(ctx,
			`SELECT switch_id, port_number, COALESCE(port_label, ''), COALESCE(speed, ''),
			        COALESCE(is_uplink, false), COALESCE(notes, '')
			 FROM switch_ports WHERE switch_id = ANY($1) ORDER BY port_number`, switchIDs)
		if err != nil {
			return nil, nil, err
		}
		defer prows.Close()
		for prows.Next() {
			var switchID int64
			var p exportSwitchPort
			if err := prows.Scan(&switchID, &p.PortNumber, &p.PortLabel, &p.Speed, &p.IsUplink, &p.Notes); err != nil {
				return nil, nil, err
			}
			portsBySwitch[switchID] = append(portsBySwitch[switchID], p)
		}
		if err := prows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return switchBySite, portsBySwitch, nil
}

func fetchExportPatchPanels(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportPatchPanel, map[int64][]exportPatchPanelPort, error) {
	pprows, err := db.QueryContext(ctx,
		`SELECT pp.id, pp.site_id, pp.name, pp.total_ports, COALESCE(l.name, ''), COALESCE(pp.notes, '')
		 FROM patch_panels pp
		 LEFT JOIN locations l ON l.id = pp.location_id
		 WHERE pp.site_id = ANY($1) ORDER BY pp.name`, siteIDs)
	if err != nil {
		return nil, nil, err
	}
	defer pprows.Close()

	panelsBySite := map[int64][]exportPatchPanel{}
	var panelIDs []int64
	for pprows.Next() {
		var siteID int64
		var pp exportPatchPanel
		if err := pprows.Scan(&siteID, &pp.ID, &pp.Name, &pp.TotalPorts, &pp.Location, &pp.Notes); err != nil {
			return nil, nil, err
		}
		panelsBySite[siteID] = append(panelsBySite[siteID], pp)
		panelIDs = append(panelIDs, pp.ID)
	}
	if err := pprows.Err(); err != nil {
		return nil, nil, err
	}

	portsByPanel := map[int64][]exportPatchPanelPort{}
	if len(panelIDs) > 0 {
		prows, err := db.QueryContext(ctx,
			`SELECT ppp.patch_panel_id, ppp.port_number, COALESCE(ppp.port_label, ''),
			        COALESCE(lpp.name || ' #' || lppp.port_number::text, ''),
			        COALESCE(ppp.notes, '')
			 FROM patch_panel_ports ppp
			 LEFT JOIN patch_panel_ports lppp ON lppp.id = ppp.linked_port_id
			 LEFT JOIN patch_panels lpp ON lpp.id = lppp.patch_panel_id
			 WHERE ppp.patch_panel_id = ANY($1) ORDER BY ppp.port_number`, panelIDs)
		if err != nil {
			return nil, nil, err
		}
		defer prows.Close()
		for prows.Next() {
			var panelID int64
			var p exportPatchPanelPort
			if err := prows.Scan(&panelID, &p.PortNumber, &p.PortLabel, &p.LinkedPort, &p.Notes); err != nil {
				return nil, nil, err
			}
			portsByPanel[panelID] = append(portsByPanel[panelID], p)
		}
		if err := prows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return panelsBySite, portsByPanel, nil
}

func fetchExportDevices(ctx context.Context, db *sql.DB, siteIDs []int64) (
	map[int64][]exportDevice,
	map[int64][]exportDeviceInterface,
	map[int64][]exportDeviceIP,
	map[int64][]exportDeviceConnection,
	error,
) {
	drows, err := db.QueryContext(ctx,
		`SELECT d.id, d.site_id, d.hostname, COALESCE(d.dns_name, ''),
		        COALESCE(d.serial_number, ''), COALESCE(d.asset_tag, ''),
		        COALESCE(c.name, ''), d.status, COALESCE(d.is_up, false),
		        COALESCE(os.name, ''), COALESCE(d.has_rmm, false), COALESCE(d.has_antivirus, false),
		        COALESCE(sup.name, ''),
		        COALESCE(CONCAT(m.name, ' ', dm.model_name), ''),
		        COALESCE(l.name, ''),
		        COALESCE(d.installation_date::text, ''),
		        COALESCE(d.vm_id::text, ''),
		        COALESCE(d.notes, '')
		 FROM devices d
		 LEFT JOIN categories c ON c.id = d.category_id
		 LEFT JOIN operating_systems os ON os.id = d.os_id
		 LEFT JOIN suppliers sup ON sup.id = d.supplier_id
		 LEFT JOIN device_models dm ON dm.id = d.model_id
		 LEFT JOIN manufacturers m ON m.id = dm.manufacturer_id
		 LEFT JOIN locations l ON l.id = d.location_id
		 WHERE d.site_id = ANY($1) ORDER BY d.hostname`, siteIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer drows.Close()

	devicesBySite := map[int64][]exportDevice{}
	var deviceIDs []int64
	for drows.Next() {
		var siteID int64
		var d exportDevice
		if err := drows.Scan(&siteID, &d.ID, &d.Hostname, &d.DnsName,
			&d.SerialNumber, &d.AssetTag, &d.Category, &d.Status, &d.IsUp,
			&d.OS, &d.HasRmm, &d.HasAntivirus, &d.Supplier, &d.Model, &d.Location,
			&d.InstallationDate, &d.VmID, &d.Notes); err != nil {
			return nil, nil, nil, nil, err
		}
		devicesBySite[siteID] = append(devicesBySite[siteID], d)
		deviceIDs = append(deviceIDs, d.ID)
	}
	if err := drows.Err(); err != nil {
		return nil, nil, nil, nil, err
	}

	ifacesByDevice := map[int64][]exportDeviceInterface{}
	ipsByDevice := map[int64][]exportDeviceIP{}
	connsByDevice := map[int64][]exportDeviceConnection{}

	if len(deviceIDs) > 0 {
		// Interfaces
		irows, err := db.QueryContext(ctx,
			`SELECT device_id, name, COALESCE(mac_address::text, '')
			 FROM device_interfaces WHERE device_id = ANY($1) ORDER BY name`, deviceIDs)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		defer irows.Close()
		for irows.Next() {
			var iface exportDeviceInterface
			if err := irows.Scan(&iface.DeviceID, &iface.Name, &iface.MacAddress); err != nil {
				return nil, nil, nil, nil, err
			}
			ifacesByDevice[iface.DeviceID] = append(ifacesByDevice[iface.DeviceID], iface)
		}
		if err := irows.Err(); err != nil {
			return nil, nil, nil, nil, err
		}

		// IPs (via interfaces)
		iprows, err := db.QueryContext(ctx,
			`SELECT di.device_id, di.name, dip.ip_address::text,
			        COALESCE(v.name || ' (' || v.vlan_id::text || ')', ''),
			        COALESCE(dip.is_primary, false)
			 FROM device_ips dip
			 JOIN device_interfaces di ON di.id = dip.interface_id
			 LEFT JOIN vlans v ON v.id = dip.vlan_id
			 WHERE di.device_id = ANY($1) ORDER BY dip.ip_address`, deviceIDs)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		defer iprows.Close()
		for iprows.Next() {
			var ip exportDeviceIP
			if err := iprows.Scan(&ip.InterfaceDeviceID, &ip.InterfaceName, &ip.IPAddress, &ip.VlanName, &ip.IsPrimary); err != nil {
				return nil, nil, nil, nil, err
			}
			ipsByDevice[ip.InterfaceDeviceID] = append(ipsByDevice[ip.InterfaceDeviceID], ip)
		}
		if err := iprows.Err(); err != nil {
			return nil, nil, nil, nil, err
		}

		// Connections (via interfaces)
		crows, err := db.QueryContext(ctx,
			`SELECT di.device_id, di.name,
			        COALESCE(sw.hostname || ' #' || sp.port_number::text, ''),
			        COALESCE(pp.name || ' #' || ppp.port_number::text, '')
			 FROM device_connections dc
			 JOIN device_interfaces di ON di.id = dc.interface_id
			 LEFT JOIN switch_ports sp ON sp.id = dc.switch_port_id
			 LEFT JOIN switches sw ON sw.id = sp.switch_id
			 LEFT JOIN patch_panel_ports ppp ON ppp.id = dc.patch_panel_port_id
			 LEFT JOIN patch_panels pp ON pp.id = ppp.patch_panel_id
			 WHERE di.device_id = ANY($1) ORDER BY di.name`, deviceIDs)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		defer crows.Close()
		for crows.Next() {
			var conn exportDeviceConnection
			if err := crows.Scan(&conn.InterfaceDeviceID, &conn.InterfaceName, &conn.SwitchPort, &conn.PatchPanelPort); err != nil {
				return nil, nil, nil, nil, err
			}
			connsByDevice[conn.InterfaceDeviceID] = append(connsByDevice[conn.InterfaceDeviceID], conn)
		}
		if err := crows.Err(); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	return devicesBySite, ifacesByDevice, ipsByDevice, connsByDevice, nil
}

func fetchExportDeviceGroups(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportDeviceGroup, map[int64][]exportDeviceGroupMember, error) {
	grows, err := db.QueryContext(ctx,
		`SELECT id, site_id, name, COALESCE(description, '')
		 FROM device_groups WHERE site_id = ANY($1) ORDER BY name`, siteIDs)
	if err != nil {
		return nil, nil, err
	}
	defer grows.Close()

	groupsBySite := map[int64][]exportDeviceGroup{}
	var groupIDs []int64
	for grows.Next() {
		var siteID int64
		var g exportDeviceGroup
		if err := grows.Scan(&g.ID, &siteID, &g.Name, &g.Description); err != nil {
			return nil, nil, err
		}
		groupsBySite[siteID] = append(groupsBySite[siteID], g)
		groupIDs = append(groupIDs, g.ID)
	}
	if err := grows.Err(); err != nil {
		return nil, nil, err
	}

	membersByGroup := map[int64][]exportDeviceGroupMember{}
	if len(groupIDs) > 0 {
		mrows, err := db.QueryContext(ctx,
			`SELECT dgm.group_id, d.hostname
			 FROM device_group_members dgm
			 JOIN devices d ON d.id = dgm.device_id
			 WHERE dgm.group_id = ANY($1) ORDER BY d.hostname`, groupIDs)
		if err != nil {
			return nil, nil, err
		}
		defer mrows.Close()
		for mrows.Next() {
			var m exportDeviceGroupMember
			if err := mrows.Scan(&m.GroupID, &m.DeviceHostname); err != nil {
				return nil, nil, err
			}
			membersByGroup[m.GroupID] = append(membersByGroup[m.GroupID], m)
		}
		if err := mrows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return groupsBySite, membersByGroup, nil
}

func fetchExportFirewallRules(ctx context.Context, db *sql.DB, siteIDs []int64) (map[int64][]exportFirewallRule, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT fr.site_id, fr.position, fr.protocol, fr.src_port, fr.dst_port,
		        COALESCE(sd.hostname, sg.name, sv.name || ' (' || sv.vlan_id::text || ')', fr.src_cidr::text, 'any'),
		        COALESCE(dd.hostname, dg.name, dv.name || ' (' || dv.vlan_id::text || ')', fr.dst_cidr::text, 'any'),
		        fr.action, fr.enabled, COALESCE(fr.description, '')
		 FROM firewall_rules fr
		 LEFT JOIN devices sd ON sd.id = fr.src_device_id
		 LEFT JOIN device_groups sg ON sg.id = fr.src_group_id
		 LEFT JOIN vlans sv ON sv.id = fr.src_vlan_id
		 LEFT JOIN devices dd ON dd.id = fr.dst_device_id
		 LEFT JOIN device_groups dg ON dg.id = fr.dst_group_id
		 LEFT JOIN vlans dv ON dv.id = fr.dst_vlan_id
		 WHERE fr.site_id = ANY($1) ORDER BY fr.position`, siteIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[int64][]exportFirewallRule{}
	for rows.Next() {
		var siteID int64
		var r exportFirewallRule
		if err := rows.Scan(&siteID, &r.Position, &r.Protocol, &r.SrcPort, &r.DstPort,
			&r.Src, &r.Dst, &r.Action, &r.Enabled, &r.Description); err != nil {
			return nil, err
		}
		result[siteID] = append(result[siteID], r)
	}
	return result, rows.Err()
}

func fetchExportBackupPolicies(ctx context.Context, db *sql.DB, clientID int64) ([]exportBackupPolicy, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT bp.name, bp.destination, COALESCE(bp.source, ''), bp.enabled,
		        bp.retain_last, bp.retain_hourly, bp.retain_daily,
		        bp.retain_weekly, bp.retain_monthly, bp.retain_yearly,
		        COALESCE(bp.notes, ''),
		        COALESCE(
		          (SELECT string_agg(bst.run_at::text, ', ' ORDER BY bst.run_at)
		           FROM backup_schedule_times bst WHERE bst.policy_id = bp.id), '')
		 FROM backup_policies bp WHERE bp.client_id = $1 ORDER BY bp.name`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []exportBackupPolicy
	for rows.Next() {
		var bp exportBackupPolicy
		if err := rows.Scan(&bp.Name, &bp.Destination, &bp.Source, &bp.Enabled,
			&bp.RetainLast, &bp.RetainHourly, &bp.RetainDaily,
			&bp.RetainWeekly, &bp.RetainMonthly, &bp.RetainYearly,
			&bp.Notes, &bp.ScheduleTimes); err != nil {
			return nil, err
		}
		policies = append(policies, bp)
	}
	return policies, rows.Err()
}
