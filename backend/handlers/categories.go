package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// CategoryHandler groups all HTTP handlers for the /categories resource.
type CategoryHandler struct {
	db *sql.DB
}

// NewCategoryHandler creates a CategoryHandler with the given database connection.
func NewCategoryHandler(db *sql.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

const categorySelectSQL = `SELECT id, name, created_at FROM categories`

func scanCategory(row interface{ Scan(...any) error }) (models.Category, error) {
	var cat models.Category
	err := row.Scan(&cat.ID, &cat.Name, &cat.CreatedAt)
	return cat, err
}

// List handles GET /categories
func (h *CategoryHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		categorySelectSQL+` ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	items := []models.Category{}
	for rows.Next() {
		cat, err := scanCategory(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		items = append(items, cat)
	}

	ok(c, http.StatusOK, items)
}

// GetByID handles GET /categories/:id
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	cat, err := scanCategory(h.db.QueryRowContext(c.Request.Context(),
		categorySelectSQL+` WHERE id = $1`, id))

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("category not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, cat)
}

// Create handles POST /categories
func (h *CategoryHandler) Create(c *gin.Context) {
	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	cat, err := scanCategory(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO categories (name) VALUES ($1)
		 RETURNING id, name, created_at`,
		input.Name,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusCreated, cat)
}

// Update handles PUT /categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	cat, err := scanCategory(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE categories SET name = $1 WHERE id = $2
		 RETURNING id, name, created_at`,
		input.Name, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("category not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, cat)
}

// Delete handles DELETE /categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM categories WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("category not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
