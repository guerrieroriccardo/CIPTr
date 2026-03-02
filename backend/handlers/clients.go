package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// ClientHandler handles all /clients routes.
// It holds a reference to the database, injected at startup.
type ClientHandler struct {
	db *sql.DB
}

func NewClientHandler(db *sql.DB) *ClientHandler {
	return &ClientHandler{db: db}
}

// List handles GET /clients
// Returns all clients ordered by name.
func (h *ClientHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	clients := []models.Client{} // initialize as empty slice, not nil (avoids JSON "null")
	for rows.Next() {
		var cl models.Client
		if err := rows.Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		clients = append(clients, cl)
	}

	ok(c, http.StatusOK, clients)
}

// GetByID handles GET /clients/:id
func (h *ClientHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var cl models.Client
	err = h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients WHERE id = ?`, id,
	).Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("client not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, cl)
}

// Create handles POST /clients
func (h *ClientHandler) Create(c *gin.Context) {
	var input models.ClientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	res, err := h.db.ExecContext(c.Request.Context(),
		`INSERT INTO clients (name, short_code, notes) VALUES (?, ?, ?)`,
		input.Name, input.ShortCode, input.Notes,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	newID, _ := res.LastInsertId()

	// Return the newly created record
	var cl models.Client
	h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients WHERE id = ?`, newID,
	).Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt)

	ok(c, http.StatusCreated, cl)
}

// Update handles PUT /clients/:id
func (h *ClientHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.ClientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	res, err := h.db.ExecContext(c.Request.Context(),
		`UPDATE clients SET name = ?, short_code = ?, notes = ? WHERE id = ?`,
		input.Name, input.ShortCode, input.Notes, id,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		fail(c, http.StatusNotFound, errors.New("client not found"))
		return
	}

	// Return the updated record
	var cl models.Client
	h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients WHERE id = ?`, id,
	).Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt)

	ok(c, http.StatusOK, cl)
}

// Delete handles DELETE /clients/:id
func (h *ClientHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	res, err := h.db.ExecContext(c.Request.Context(),
		`DELETE FROM clients WHERE id = ?`, id,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		fail(c, http.StatusNotFound, errors.New("client not found"))
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
