package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID *int64    `json:"resource_id"`
	Detail     *string   `json:"detail"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditHandler struct {
	db *sql.DB
}

func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

func (h *AuditHandler) List(c *gin.Context) {
	query := `SELECT id, user_id, username, action, resource, resource_id, detail, created_at
	          FROM audit_logs WHERE 1=1`
	args := []any{}
	n := 0

	if res := c.Query("resource"); res != "" {
		n++
		query += ` AND resource = $` + strconv.Itoa(n)
		args = append(args, res)
	}
	if uid := c.Query("user_id"); uid != "" {
		n++
		query += ` AND user_id = $` + strconv.Itoa(n)
		args = append(args, uid)
	}

	query += ` ORDER BY created_at DESC`

	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}
	n++
	query += ` LIMIT $` + strconv.Itoa(n)
	args = append(args, limit)

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	logs := []AuditLog{}
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.ResourceID, &l.Detail, &l.CreatedAt); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		logs = append(logs, l)
	}

	ok(c, http.StatusOK, logs)
}

// logAudit inserts an audit log entry. Errors are silently ignored to avoid
// breaking the main operation.
func logAudit(ctx context.Context, db *sql.DB, c *gin.Context, action, resource string, resourceID int64, detail string) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	uid, _ := userID.(int64)
	uname, _ := username.(string)
	_, _ = db.ExecContext(ctx,
		`INSERT INTO audit_logs (user_id, username, action, resource, resource_id, detail)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uid, uname, action, resource, resourceID, detail)
}

// logAuditManual inserts an audit log entry with explicit user info (for pre-auth actions like login/register).
func logAuditManual(ctx context.Context, db *sql.DB, userID int64, username, action, resource string, resourceID int64, detail string) {
	var uid *int64
	if userID != 0 {
		uid = &userID
	}
	_, _ = db.ExecContext(ctx,
		`INSERT INTO audit_logs (user_id, username, action, resource, resource_id, detail)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		uid, username, action, resource, resourceID, detail)
}
