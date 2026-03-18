package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

// BackupPolicyHandler groups all HTTP handlers for the /backup-policies resource.
type BackupPolicyHandler struct {
	db *sql.DB
}

// NewBackupPolicyHandler creates a BackupPolicyHandler with the given database connection.
func NewBackupPolicyHandler(db *sql.DB) *BackupPolicyHandler {
	return &BackupPolicyHandler{db: db}
}

var timeFormatRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

const backupPolicyCoreColumns = `bp.id, bp.client_id, c.name,
	bp.name, bp.destination,
	bp.retain_last, bp.retain_hourly, bp.retain_daily,
	bp.retain_weekly, bp.retain_monthly, bp.retain_yearly,
	bp.enabled, bp.notes, bp.created_at, bp.updated_at,
	string_agg(to_char(bst.run_at, 'HH24:MI'), ',' ORDER BY bst.run_at) FILTER (WHERE bst.id IS NOT NULL)`

const backupPolicySelectSQL = `SELECT ` + backupPolicyCoreColumns + `
	FROM backup_policies bp
	JOIN clients c ON c.id = bp.client_id
	LEFT JOIN backup_schedule_times bst ON bst.policy_id = bp.id`

// scanBackupPolicy reads one row (from the JOIN query) into a BackupPolicy struct.
func scanBackupPolicy(row interface{ Scan(...any) error }) (models.BackupPolicy, error) {
	var p models.BackupPolicy
	var times *string
	err := row.Scan(
		&p.ID, &p.ClientID, &p.ClientName,
		&p.Name, &p.Destination,
		&p.RetainLast, &p.RetainHourly, &p.RetainDaily,
		&p.RetainWeekly, &p.RetainMonthly, &p.RetainYearly,
		&p.Enabled, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
		&times,
	)
	if err != nil {
		return p, err
	}
	if times != nil && *times != "" {
		p.ScheduleTimes = strings.Split(*times, ",")
	} else {
		p.ScheduleTimes = []string{}
	}
	return p, nil
}

// fetchByID re-queries a single policy with the full JOIN (used after create/update).
func (h *BackupPolicyHandler) fetchByID(ctx context.Context, id int64) (models.BackupPolicy, error) {
	return scanBackupPolicy(h.db.QueryRowContext(ctx,
		backupPolicySelectSQL+` WHERE bp.id = $1 GROUP BY bp.id, c.name`, id))
}

// validateScheduleTimes checks that all entries are valid HH:MM 24-hour times.
func validateScheduleTimes(times []string) error {
	for _, t := range times {
		if !timeFormatRegex.MatchString(t) {
			return fmt.Errorf("invalid time %q: must be HH:MM in 24-hour format (e.g. 14:00)", t)
		}
	}
	return nil
}

// replaceScheduleTimes deletes existing times for a policy and inserts the new ones.
func replaceScheduleTimes(ctx context.Context, tx *sql.Tx, policyID int64, times []string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM backup_schedule_times WHERE policy_id = $1`, policyID); err != nil {
		return err
	}
	for _, t := range times {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO backup_schedule_times (policy_id, run_at) VALUES ($1, $2)`,
			policyID, t); err != nil {
			return err
		}
	}
	return nil
}

// List handles GET /backup-policies
// Supports optional query param: ?client_id=
func (h *BackupPolicyHandler) List(c *gin.Context) {
	query := backupPolicySelectSQL
	var args []any
	n := 1

	if clientID := c.Query("client_id"); clientID != "" {
		query += fmt.Sprintf(" WHERE bp.client_id = $%d", n)
		args = append(args, clientID)
		n++
	}
	query += " GROUP BY bp.id, c.name ORDER BY bp.id"

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	policies := []models.BackupPolicy{}
	for rows.Next() {
		p, err := scanBackupPolicy(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		policies = append(policies, p)
	}

	ok(c, http.StatusOK, policies)
}

// ListByClient handles GET /clients/:id/backup-policies
func (h *BackupPolicyHandler) ListByClient(c *gin.Context) {
	clientID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid client id"))
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		backupPolicySelectSQL+` WHERE bp.client_id = $1 GROUP BY bp.id, c.name ORDER BY bp.id`, clientID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	policies := []models.BackupPolicy{}
	for rows.Next() {
		p, err := scanBackupPolicy(rows)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		policies = append(policies, p)
	}

	ok(c, http.StatusOK, policies)
}

// GetByID handles GET /backup-policies/:id
func (h *BackupPolicyHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	p, err := h.fetchByID(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("backup policy not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, p)
}

// Create handles POST /backup-policies
func (h *BackupPolicyHandler) Create(c *gin.Context) {
	var input models.BackupPolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err := validateScheduleTimes(input.ScheduleTimes); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// Apply defaults.
	retainLast := 0
	if input.RetainLast != nil {
		retainLast = *input.RetainLast
	}
	retainHourly := 0
	if input.RetainHourly != nil {
		retainHourly = *input.RetainHourly
	}
	retainDaily := 7
	if input.RetainDaily != nil {
		retainDaily = *input.RetainDaily
	}
	retainWeekly := 4
	if input.RetainWeekly != nil {
		retainWeekly = *input.RetainWeekly
	}
	retainMonthly := 12
	if input.RetainMonthly != nil {
		retainMonthly = *input.RetainMonthly
	}
	retainYearly := 3
	if input.RetainYearly != nil {
		retainYearly = *input.RetainYearly
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	ctx := c.Request.Context()
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback() //nolint:errcheck

	var policyID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO backup_policies
			(client_id, name, destination,
			 retain_last, retain_hourly, retain_daily,
			 retain_weekly, retain_monthly, retain_yearly,
			 enabled, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		input.ClientID, input.Name, input.Destination,
		retainLast, retainHourly, retainDaily,
		retainWeekly, retainMonthly, retainYearly,
		enabled, input.Notes,
	).Scan(&policyID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := replaceScheduleTimes(ctx, tx, policyID, input.ScheduleTimes); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	p, err := h.fetchByID(ctx, policyID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(ctx, h.db, c, "create", "backup_policies", policyID,
		fmt.Sprintf("Created backup policy %q for client %d", p.Name, p.ClientID))
	ok(c, http.StatusCreated, p)
}

// Update handles PUT /backup-policies/:id
func (h *BackupPolicyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var input models.BackupPolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err := validateScheduleTimes(input.ScheduleTimes); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	// Apply defaults.
	retainLast := 0
	if input.RetainLast != nil {
		retainLast = *input.RetainLast
	}
	retainHourly := 0
	if input.RetainHourly != nil {
		retainHourly = *input.RetainHourly
	}
	retainDaily := 7
	if input.RetainDaily != nil {
		retainDaily = *input.RetainDaily
	}
	retainWeekly := 4
	if input.RetainWeekly != nil {
		retainWeekly = *input.RetainWeekly
	}
	retainMonthly := 12
	if input.RetainMonthly != nil {
		retainMonthly = *input.RetainMonthly
	}
	retainYearly := 3
	if input.RetainYearly != nil {
		retainYearly = *input.RetainYearly
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	ctx := c.Request.Context()
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback() //nolint:errcheck

	var updatedID int64
	err = tx.QueryRowContext(ctx,
		`UPDATE backup_policies SET
			client_id = $1, name = $2, destination = $3,
			retain_last = $4, retain_hourly = $5, retain_daily = $6,
			retain_weekly = $7, retain_monthly = $8, retain_yearly = $9,
			enabled = $10, notes = $11,
			updated_at = NOW()
		 WHERE id = $12
		 RETURNING id`,
		input.ClientID, input.Name, input.Destination,
		retainLast, retainHourly, retainDaily,
		retainWeekly, retainMonthly, retainYearly,
		enabled, input.Notes, id,
	).Scan(&updatedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("backup policy not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := replaceScheduleTimes(ctx, tx, id, input.ScheduleTimes); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	p, err := h.fetchByID(ctx, id)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(ctx, h.db, c, "update", "backup_policies", id,
		fmt.Sprintf("Updated backup policy %q (client %d)", p.Name, p.ClientID))
	ok(c, http.StatusOK, p)
}

// Delete handles DELETE /backup-policies/:id
func (h *BackupPolicyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	var deletedID int64
	err = h.db.QueryRowContext(c.Request.Context(),
		`DELETE FROM backup_policies WHERE id = $1 RETURNING id`, id,
	).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusNotFound, errors.New("backup policy not found"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "delete", "backup_policies", id,
		fmt.Sprintf("Deleted backup policy #%d", id))
	ok(c, http.StatusOK, gin.H{"deleted": true})
}
