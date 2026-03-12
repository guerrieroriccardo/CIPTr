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

// SettingsHandler groups HTTP handlers for the /settings resource.
type SettingsHandler struct {
	db *sql.DB
}

// NewSettingsHandler creates a SettingsHandler with the given database connection.
func NewSettingsHandler(db *sql.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// HostnameFormat holds the parsed hostname configuration.
type HostnameFormat struct {
	PrefixSource   string // "short_code" or "name"
	PrefixPosition string // "before", "after", "none"
	NumDigits      int    // 1-6
}

// settingValidation defines allowed values for known setting keys.
var settingValidation = map[string][]string{
	"hostname_prefix_source":   {"short_code", "name"},
	"hostname_prefix_position": {"before", "after", "none"},
}

// List handles GET /settings — returns all settings.
func (h *SettingsHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT key, value FROM settings ORDER BY key`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.Key, &s.Value); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		settings = append(settings, s)
	}
	ok(c, http.StatusOK, settings)
}

// GetByKey handles GET /settings/:key — returns a single setting.
func (h *SettingsHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	var s models.Setting
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT key, value FROM settings WHERE key = $1`, key,
	).Scan(&s.Key, &s.Value)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, fmt.Errorf("setting %q not found", key))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	ok(c, http.StatusOK, s)
}

// Update handles PUT /settings/:key — updates a setting value.
func (h *SettingsHandler) Update(c *gin.Context) {
	key := c.Param("key")

	var body struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	body.Value = strings.TrimSpace(body.Value)
	if body.Value == "" {
		fail(c, http.StatusBadRequest, errors.New("value is required"))
		return
	}

	// Validate known keys.
	if allowed, exists := settingValidation[key]; exists {
		valid := false
		for _, v := range allowed {
			if body.Value == v {
				valid = true
				break
			}
		}
		if !valid {
			fail(c, http.StatusBadRequest, fmt.Errorf("invalid value %q for %s, allowed: %s", body.Value, key, strings.Join(allowed, ", ")))
			return
		}
	}

	if key == "hostname_num_digits" {
		n, err := strconv.Atoi(body.Value)
		if err != nil || n < 1 || n > 6 {
			fail(c, http.StatusBadRequest, errors.New("hostname_num_digits must be an integer between 1 and 6"))
			return
		}
	}

	result, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE settings SET value = $1 WHERE key = $2`, body.Value, key)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		fail(c, http.StatusNotFound, fmt.Errorf("setting %q not found", key))
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "settings", 0, fmt.Sprintf("%s = %s", key, body.Value))

	ok(c, http.StatusOK, models.Setting{Key: key, Value: body.Value})
}

// GetHostnameFormat loads the hostname configuration from the settings table.
func GetHostnameFormat(ctx context.Context, db *sql.DB) (HostnameFormat, error) {
	format := HostnameFormat{
		PrefixSource:   "short_code",
		PrefixPosition: "before",
		NumDigits:      3,
	}

	rows, err := db.QueryContext(ctx,
		`SELECT key, value FROM settings WHERE key IN ('hostname_prefix_source', 'hostname_prefix_position', 'hostname_num_digits')`)
	if err != nil {
		return format, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return format, err
		}
		switch key {
		case "hostname_prefix_source":
			format.PrefixSource = value
		case "hostname_prefix_position":
			format.PrefixPosition = value
		case "hostname_num_digits":
			if n, err := strconv.Atoi(value); err == nil && n >= 1 && n <= 6 {
				format.NumDigits = n
			}
		}
	}
	return format, nil
}
