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

// OperatingSystemHandler groups all HTTP handlers for the /operating-systems resource.
type OperatingSystemHandler struct {
	db *sql.DB
}

// NewOperatingSystemHandler creates an OperatingSystemHandler with the given database connection.
func NewOperatingSystemHandler(db *sql.DB) *OperatingSystemHandler {
	return &OperatingSystemHandler{db: db}
}

const operatingSystemSelectSQL = `SELECT id, name, created_at FROM operating_systems`

func scanOperatingSystem(row interface{ Scan(...any) error }) (models.OperatingSystem, error) {
	var os models.OperatingSystem
	err := row.Scan(&os.ID, &os.Name, &os.CreatedAt)
	return os, err
}

// List handles GET /operating-systems
func (h *OperatingSystemHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		operatingSystemSelectSQL+` ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []models.OperatingSystem{}
	for rows.Next() {
		os, err := scanOperatingSystem(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, os)
	}

	ok(c, http.StatusOK, items)
}

// GetByID handles GET /operating-systems/:id
func (h *OperatingSystemHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	os, err := scanOperatingSystem(h.db.QueryRowContext(c.Request.Context(),
		operatingSystemSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("operating system not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, os)
}

// Create handles POST /operating-systems
func (h *OperatingSystemHandler) Create(c *gin.Context) {
	var input models.OperatingSystemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	os, err := scanOperatingSystem(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO operating_systems (name) VALUES ($1)
		 RETURNING id, name, created_at`,
		input.Name,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "create", "operating_systems", os.ID, fmt.Sprintf("Created OS '%s'", os.Name))
	ok(c, http.StatusCreated, os)
}

// Update handles PUT /operating-systems/:id
func (h *OperatingSystemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.OperatingSystemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	os, err := scanOperatingSystem(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE operating_systems SET name = $1 WHERE id = $2
		 RETURNING id, name, created_at`,
		input.Name, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("operating system not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "operating_systems", id, fmt.Sprintf("Updated OS '%s'", os.Name))
	ok(c, http.StatusOK, os)
}

// Delete handles DELETE /operating-systems/:id
func (h *OperatingSystemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM operating_systems WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("operating system not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "operating_systems", id, fmt.Sprintf("Deleted OS #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
