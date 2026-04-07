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

type WifiSSIDHandler struct {
	db *sql.DB
}

func NewWifiSSIDHandler(db *sql.DB) *WifiSSIDHandler {
	return &WifiSSIDHandler{db: db}
}

const wifiSSIDSelectSQL = `SELECT id, site_id, ssid, auth, vlan_id, notes FROM wifi_ssids`

func scanWifiSSID(row interface{ Scan(...any) error }) (models.WifiSSID, error) {
	var w models.WifiSSID
	err := row.Scan(&w.ID, &w.SiteID, &w.SSID, &w.Auth, &w.VlanID, &w.Notes)
	return w, err
}

// List handles GET /wifi-ssids
func (h *WifiSSIDHandler) List(c *gin.Context) {
	query := wifiSSIDSelectSQL
	args := []any{}

	if siteID := c.Query("site_id"); siteID != "" {
		query += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	query += ` ORDER BY ssid`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []models.WifiSSID{}
	for rows.Next() {
		w, err := scanWifiSSID(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, w)
	}

	ok(c, http.StatusOK, items)
}

// ListBySite handles GET /sites/:id/wifi-ssids
func (h *WifiSSIDHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		wifiSSIDSelectSQL+` WHERE site_id = $1 ORDER BY ssid`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []models.WifiSSID{}
	for rows.Next() {
		w, err := scanWifiSSID(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, w)
	}

	ok(c, http.StatusOK, items)
}

// GetByID handles GET /wifi-ssids/:id
func (h *WifiSSIDHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	w, err := scanWifiSSID(h.db.QueryRowContext(c.Request.Context(),
		wifiSSIDSelectSQL+` WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("wifi ssid not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, w)
}

// Create handles POST /wifi-ssids
func (h *WifiSSIDHandler) Create(c *gin.Context) {
	var input models.WifiSSIDInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	w, err := scanWifiSSID(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO wifi_ssids (site_id, ssid, auth, vlan_id, notes) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, site_id, ssid, auth, vlan_id, notes`,
		input.SiteID, input.SSID, input.Auth, input.VlanID, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "wifi_ssids", w.ID, fmt.Sprintf("Created WiFi SSID '%s'", w.SSID))
	ok(c, http.StatusCreated, w)
}

// Update handles PUT /wifi-ssids/:id
func (h *WifiSSIDHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.WifiSSIDInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	w, err := scanWifiSSID(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE wifi_ssids SET site_id = $1, ssid = $2, auth = $3, vlan_id = $4, notes = $5 WHERE id = $6
		 RETURNING id, site_id, ssid, auth, vlan_id, notes`,
		input.SiteID, input.SSID, input.Auth, input.VlanID, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("wifi ssid not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "wifi_ssids", id, fmt.Sprintf("Updated WiFi SSID '%s'", w.SSID))
	ok(c, http.StatusOK, w)
}

// Delete handles DELETE /wifi-ssids/:id
func (h *WifiSSIDHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM wifi_ssids WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("wifi ssid not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "wifi_ssids", id, fmt.Sprintf("Deleted WiFi SSID #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
