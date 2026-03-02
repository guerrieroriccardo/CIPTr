package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// AddressBlockHandler groups all HTTP handlers for the /address-blocks resource.
type AddressBlockHandler struct {
	db *sql.DB
}

// NewAddressBlockHandler creates an AddressBlockHandler with the given database connection.
func NewAddressBlockHandler(db *sql.DB) *AddressBlockHandler {
	return &AddressBlockHandler{db: db}
}

// addressBlockSelectSQL is the base SELECT used by every read operation.
const addressBlockSelectSQL = `SELECT id, site_id, network, description, notes FROM address_blocks`

// scanAddressBlock reads one row into an AddressBlock struct.
func scanAddressBlock(row interface{ Scan(...any) error }) (models.AddressBlock, error) {
	var ab models.AddressBlock
	err := row.Scan(&ab.ID, &ab.SiteID, &ab.Network, &ab.Description, &ab.Notes)
	return ab, err
}

// List handles GET /address-blocks
// Supports optional query param: ?site_id=1
func (h *AddressBlockHandler) List(c *gin.Context) {
	query := addressBlockSelectSQL
	args := []any{}

	if siteID := c.Query("site_id"); siteID != "" {
		query += ` WHERE site_id = $1`
		args = append(args, siteID)
	}
	query += ` ORDER BY network`

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	blocks := []models.AddressBlock{}
	for rows.Next() {
		ab, err := scanAddressBlock(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		blocks = append(blocks, ab)
	}

	ok(c, http.StatusOK, blocks)
}

// ListBySite handles GET /sites/:id/address-blocks
// Returns all address blocks for the given site, ordered by network.
func (h *AddressBlockHandler) ListBySite(c *gin.Context) {
	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid site id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		addressBlockSelectSQL+` WHERE site_id = $1 ORDER BY network`, siteID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	blocks := []models.AddressBlock{}
	for rows.Next() {
		ab, err := scanAddressBlock(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		blocks = append(blocks, ab)
	}

	ok(c, http.StatusOK, blocks)
}

// GetByID handles GET /address-blocks/:id
// Returns 404 if the address block does not exist.
func (h *AddressBlockHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	ab, err := scanAddressBlock(h.db.QueryRowContext(c.Request.Context(),
		addressBlockSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("address block not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ab)
}

// Create handles POST /address-blocks
// site_id and network are required. network must be valid CIDR notation.
func (h *AddressBlockHandler) Create(c *gin.Context) {
	var input models.AddressBlockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	ab, err := scanAddressBlock(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO address_blocks (site_id, network, description, notes) VALUES ($1, $2, $3, $4)
		 RETURNING id, site_id, network, description, notes`,
		input.SiteID, input.Network, input.Description, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, ab)
}

// Update handles PUT /address-blocks/:id
// Replaces all fields. Returns 404 if the address block does not exist.
func (h *AddressBlockHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.AddressBlockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	ab, err := scanAddressBlock(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE address_blocks SET site_id = $1, network = $2, description = $3, notes = $4 WHERE id = $5
		 RETURNING id, site_id, network, description, notes`,
		input.SiteID, input.Network, input.Description, input.Notes, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("address block not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, ab)
}

// Delete handles DELETE /address-blocks/:id
// Cascades to vlans via the DB foreign key ON DELETE CASCADE.
func (h *AddressBlockHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM address_blocks WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("address block not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
