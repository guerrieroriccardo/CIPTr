package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// SupplierHandler groups all HTTP handlers for the /suppliers resource.
type SupplierHandler struct {
	db *sql.DB
}

// NewSupplierHandler creates a SupplierHandler with the given database connection.
func NewSupplierHandler(db *sql.DB) *SupplierHandler {
	return &SupplierHandler{db: db}
}

const supplierSelectSQL = `SELECT id, name, address, phone, email, created_at FROM suppliers`

func scanSupplier(row interface{ Scan(...any) error }) (models.Supplier, error) {
	var s models.Supplier
	err := row.Scan(&s.ID, &s.Name, &s.Address, &s.Phone, &s.Email, &s.CreatedAt)
	return s, err
}

// List handles GET /suppliers
func (h *SupplierHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		supplierSelectSQL+` ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []models.Supplier{}
	for rows.Next() {
		s, err := scanSupplier(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, s)
	}

	ok(c, http.StatusOK, items)
}

// GetByID handles GET /suppliers/:id
func (h *SupplierHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	s, err := scanSupplier(h.db.QueryRowContext(c.Request.Context(),
		supplierSelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("supplier not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, s)
}

// Create handles POST /suppliers
func (h *SupplierHandler) Create(c *gin.Context) {
	var input models.SupplierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	s, err := scanSupplier(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO suppliers (name, address, phone, email) VALUES ($1, $2, $3, $4)
		 RETURNING id, name, address, phone, email, created_at`,
		input.Name, input.Address, input.Phone, input.Email,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, s)
}

// Update handles PUT /suppliers/:id
func (h *SupplierHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.SupplierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	s, err := scanSupplier(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE suppliers SET name = $1, address = $2, phone = $3, email = $4 WHERE id = $5
		 RETURNING id, name, address, phone, email, created_at`,
		input.Name, input.Address, input.Phone, input.Email, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("supplier not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, s)
}

// Delete handles DELETE /suppliers/:id
func (h *SupplierHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM suppliers WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("supplier not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
