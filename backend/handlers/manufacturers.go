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

// ManufacturerHandler groups all HTTP handlers for the /manufacturers resource.
type ManufacturerHandler struct {
	db *sql.DB
}

// NewManufacturerHandler creates a ManufacturerHandler with the given database connection.
func NewManufacturerHandler(db *sql.DB) *ManufacturerHandler {
	return &ManufacturerHandler{db: db}
}

const manufacturerSelectSQL = `SELECT id, name, created_at FROM manufacturers`

func scanManufacturer(row interface{ Scan(...any) error }) (models.Manufacturer, error) {
	var m models.Manufacturer
	err := row.Scan(&m.ID, &m.Name, &m.CreatedAt)
	return m, err
}

// List handles GET /manufacturers
func (h *ManufacturerHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		manufacturerSelectSQL+` ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []models.Manufacturer{}
	for rows.Next() {
		m, err := scanManufacturer(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, m)
	}

	ok(c, http.StatusOK, items)
}

// GetByID handles GET /manufacturers/:id
func (h *ManufacturerHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	m, err := scanManufacturer(h.db.QueryRowContext(c.Request.Context(),
		manufacturerSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("manufacturer not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, m)
}

// Create handles POST /manufacturers
func (h *ManufacturerHandler) Create(c *gin.Context) {
	var input models.ManufacturerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	m, err := scanManufacturer(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO manufacturers (name) VALUES ($1)
		 RETURNING id, name, created_at`,
		input.Name,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "manufacturers", m.ID, fmt.Sprintf("Created manufacturer '%s'", m.Name))
	ok(c, http.StatusCreated, m)
}

// Update handles PUT /manufacturers/:id
func (h *ManufacturerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.ManufacturerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	m, err := scanManufacturer(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE manufacturers SET name = $1 WHERE id = $2
		 RETURNING id, name, created_at`,
		input.Name, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("manufacturer not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "manufacturers", id, fmt.Sprintf("Updated manufacturer '%s'", m.Name))
	ok(c, http.StatusOK, m)
}

// Delete handles DELETE /manufacturers/:id
func (h *ManufacturerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM manufacturers WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("manufacturer not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "manufacturers", id, fmt.Sprintf("Deleted manufacturer #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
