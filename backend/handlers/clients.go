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

// clientSelectSQL is the base SELECT used by every read operation.
const clientSelectSQL = `SELECT id, name, short_code, notes, created_at FROM clients`

// scanClient reads one row into a Client struct.
// Accepts both *sql.Rows and *sql.Row (both implement Scan).
func scanClient(row interface{ Scan(...any) error }) (models.Client, error) {
	var cl models.Client
	err := row.Scan(&cl.ID, &cl.Name, &cl.ShortCode, &cl.Notes, &cl.CreatedAt)
	return cl, err
}

// List handles GET /clients
// Returns all clients ordered alphabetically by name.
func (h *ClientHandler) List(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		clientSelectSQL+` ORDER BY name`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	clients := []models.Client{}
	for rows.Next() {
		cl, err := scanClient(rows)
		if err != nil {
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	cl, err := scanClient(h.db.QueryRowContext(c.Request.Context(),
		clientSelectSQL+` WHERE id = $1`, id))

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
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// INSERT ... RETURNING fetches all columns in one round-trip,
	// avoiding a separate SELECT after insert.
	cl, err := scanClient(h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO clients (name, short_code, notes) VALUES ($1, $2, $3)
		 RETURNING id, name, short_code, notes, created_at`,
		input.Name, input.ShortCode, input.Notes,
	))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

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

	// UPDATE ... RETURNING returns sql.ErrNoRows if no row matched the WHERE,
	// giving us 404 detection without a separate RowsAffected check.
	cl, err := scanClient(h.db.QueryRowContext(c.Request.Context(),
		`UPDATE clients SET name = $1, short_code = $2, notes = $3 WHERE id = $4
		 RETURNING id, name, short_code, notes, created_at`,
		input.Name, input.ShortCode, input.Notes, id,
	))
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

// Delete handles DELETE /clients/:id
// Cascades to sites (and all their children) via the DB foreign key ON DELETE CASCADE.
// Returns 404 if the client does not exist.
func (h *ClientHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	// DELETE ... RETURNING id returns sql.ErrNoRows if nothing was deleted.
	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM clients WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("client not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"deleted": true})
}
