package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// LocationHandler groups all HTTP handlers for the /locations resource.
type LocationHandler struct {
	db *sql.DB
}

// NewLocationHandler creates a LocationHandler with the given database connection.
func NewLocationHandler(db *sql.DB) *LocationHandler {
	return &LocationHandler{db: db}
}

// locationSelectSQL is the base SELECT used by every read operation.
const locationSelectSQL = `SELECT id, site_id, name, floor, notes FROM locations`

// scanLocation reads one row into a Location struct.
func scanLocation(row interface{ Scan(...any) error }) (models.Location, error) {
	var l models.Location
	err := row.Scan(&l.ID, &l.SiteID, &l.Name, &l.Floor, &l.Notes)
	return l, err
}

// List handles GET /locations
// Supports optional query param: ?site_id=1
func (h *LocationHandler) List(c *gin.Context) {
	query := locationSelectSQL
	args := []any{}

	if siteID := c.Query("site_id"); siteID != "" {
		query += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	query += ` ORDER BY name`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	locations := []models.Location{}
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		locations = append(locations, l)
	}

	ok(c, http.StatusOK, locations)
}

// ListBySite handles GET /sites/:id/locations
// Returns all locations for the given site, ordered by name.
func (h *LocationHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		locationSelectSQL+` WHERE site_id = $1 ORDER BY name`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	locations := []models.Location{}
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		locations = append(locations, l)
	}

	ok(c, http.StatusOK, locations)
}

// GetByID handles GET /locations/:id
// Returns 404 if the location does not exist.
func (h *LocationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	l, err := scanLocation(h.db.QueryRowContext(c.Request.Context(),
		locationSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("location not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, l)
}

// Create handles POST /locations
// site_id and name are required.
func (h *LocationHandler) Create(c *gin.Context) {
	var input models.LocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	l, err := scanLocation(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO locations (site_id, name, floor, notes) VALUES ($1, $2, $3, $4)
		 RETURNING id, site_id, name, floor, notes`,
		input.SiteID, input.Name, input.Floor, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, l)
}

// Update handles PUT /locations/:id
// Replaces all fields. Returns 404 if the location does not exist.
func (h *LocationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.LocationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	l, err := scanLocation(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE locations SET site_id = $1, name = $2, floor = $3, notes = $4 WHERE id = $5
		 RETURNING id, site_id, name, floor, notes`,
		input.SiteID, input.Name, input.Floor, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("location not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, l)
}

// Delete handles DELETE /locations/:id
// Devices referencing this location have their location_id set to NULL (ON DELETE SET NULL).
func (h *LocationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM locations WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("location not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
