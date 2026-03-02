package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ciptr/models"
)

// ClientHandler groups all HTTP handlers for the /clients resource.
//
// It holds a reference to the database connection (db), injected once at
// startup via NewClientHandler. This avoids global variables and makes the
// dependency explicit — every method that needs the DB uses h.db.
type ClientHandler struct {
	db *sql.DB
}

// NewClientHandler creates a ClientHandler with the given database connection.
// Called once in router.go:
//
//	clientHandler := handlers.NewClientHandler(database)
func NewClientHandler(db *sql.DB) *ClientHandler {
	return &ClientHandler{db: db}
}

// List handles GET /clients
// Returns all clients ordered alphabetically by name.
// Response: {"data": [...], "error": null}
func (h *ClientHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	// Initialize as empty slice (not nil) so JSON returns [] instead of null.
	clients := []models.Client{}
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
// Returns 404 if the client does not exist.
func (h *ClientHandler) GetByID(c *gin.Context) {
	// c.Param("id") reads the :id segment from the URL path (e.g. /clients/42 → "42").
	// ParseInt converts the string to a 64-bit integer; returns 400 if it's not a number.
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var cl models.Client
	// QueryRowContext returns exactly one row (or sql.ErrNoRows if not found).
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
// Expects JSON body matching models.ClientInput (name and short_code are required).
// Returns the newly created client with its assigned id and created_at.
func (h *ClientHandler) Create(c *gin.Context) {
	var input models.ClientInput
	// ShouldBindJSON decodes the JSON body and validates the `binding:"required"` tags.
	// Returns 400 automatically if required fields are missing or JSON is malformed.
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

	// LastInsertId returns the auto-incremented id assigned by SQLite.
	newID, _ := res.LastInsertId()

	// Re-fetch the full record so the response includes created_at set by the DB.
	var cl models.Client
	h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients WHERE id = ?`, newID,
	).Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt)

	ok(c, http.StatusCreated, cl)
}

// Update handles PUT /clients/:id
// Replaces all fields of the client. Returns 404 if the client does not exist.
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

	// RowsAffected returns 0 if no row matched the WHERE id = ?, meaning the client doesn't exist.
	rows, _ := res.RowsAffected()
	if rows == 0 {
		fail(c, http.StatusNotFound, errors.New("client not found"))
		return
	}

	// Re-fetch and return the updated record.
	var cl models.Client
	h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, name, short_code, notes, created_at FROM clients WHERE id = ?`, id,
	).Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt)

	ok(c, http.StatusOK, cl)
}

// Delete handles DELETE /clients/:id
// Cascades to sites (and all their children) via the DB foreign key ON DELETE CASCADE.
// Returns 404 if the client does not exist.
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
