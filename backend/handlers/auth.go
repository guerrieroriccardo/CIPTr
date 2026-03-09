package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/guerrieroriccardo/CIPTr/backend/models"
)

type AuthHandler struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewAuthHandler(db *sql.DB, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	var user models.User
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, username, password_hash, role, created_at FROM users WHERE username = $1`,
		input.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		logAuditManual(c.Request.Context(), h.db, 0, input.Username, "login_failed", "users", 0, "Invalid credentials (unknown user)")
		fail(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if user.Role == "guest" {
		fail(c, http.StatusUnauthorized, fmt.Errorf("guest accounts must use /guest-login"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		logAuditManual(c.Request.Context(), h.db, user.ID, user.Username, "login_failed", "users", user.ID, "Invalid credentials (wrong password)")
		fail(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAuditManual(c.Request.Context(), h.db, user.ID, user.Username, "login", "users", user.ID, fmt.Sprintf("User '%s' logged in", user.Username))
	ok(c, http.StatusOK, gin.H{"token": signed})
}

func (h *AuthHandler) GuestLogin(c *gin.Context) {
	var input models.GuestLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	var user models.User
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, username, role, created_at FROM users WHERE username = $1 AND role = 'guest'`,
		input.Username,
	).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		logAuditManual(c.Request.Context(), h.db, 0, input.Username, "guest_login_failed", "users", 0, "Invalid guest credentials")
		fail(c, http.StatusUnauthorized, fmt.Errorf("invalid guest credentials"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAuditManual(c.Request.Context(), h.db, user.ID, user.Username, "guest_login", "users", user.ID, fmt.Sprintf("Guest '%s' logged in", user.Username))
	ok(c, http.StatusOK, gin.H{"token": signed})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	role := input.Role
	if role == "" {
		role = "technician"
	}
	validRoles := map[string]bool{"admin": true, "technician": true, "viewer": true, "guest": true}
	if !validRoles[role] {
		fail(c, http.StatusBadRequest, fmt.Errorf("invalid role: %s", role))
		return
	}

	if role == "guest" {
		// Guest accounts have no password.
		var user models.User
		err := h.db.QueryRowContext(c.Request.Context(),
			`INSERT INTO users (username, role) VALUES ($1, $2)
			 RETURNING id, username, role, created_at`,
			input.Username, role,
		).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
		if err != nil {
			fail(c, http.StatusConflict, fmt.Errorf("username already taken"))
			return
		}
		logAudit(c.Request.Context(), h.db, c, "register", "users", user.ID, fmt.Sprintf("Created guest user '%s'", user.Username))
		ok(c, http.StatusCreated, user)
		return
	}

	if input.Password == "" {
		fail(c, http.StatusBadRequest, fmt.Errorf("password is required for non-guest accounts"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	var user models.User
	err = h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, username, role, created_at`,
		input.Username, string(hash), role,
	).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		fail(c, http.StatusConflict, fmt.Errorf("username already taken"))
		return
	}

	logAudit(c.Request.Context(), h.db, c, "register", "users", user.ID, fmt.Sprintf("Created user '%s' with role '%s'", user.Username, user.Role))
	ok(c, http.StatusCreated, user)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	ok(c, http.StatusOK, gin.H{"id": userID, "username": username, "role": role})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var input models.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	userID, _ := c.Get("user_id")

	var hash string
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT password_hash FROM users WHERE id = $1`, userID,
	).Scan(&hash)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.OldPassword)); err != nil {
		fail(c, http.StatusUnauthorized, fmt.Errorf("old password is incorrect"))
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	_, err = h.db.ExecContext(c.Request.Context(),
		`UPDATE users SET password_hash = $1 WHERE id = $2`, string(newHash), userID,
	)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "change_password", "users", userID.(int64), "Password changed")
	ok(c, http.StatusOK, gin.H{"message": "password changed"})
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(),
		`SELECT id, username, role, created_at FROM users ORDER BY id`)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		users = append(users, u)
	}
	ok(c, http.StatusOK, users)
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	if input.Role != "" {
		validRoles := map[string]bool{"admin": true, "technician": true, "viewer": true, "guest": true}
		if !validRoles[input.Role] {
			fail(c, http.StatusBadRequest, fmt.Errorf("invalid role: %s", input.Role))
			return
		}
	}

	// Build dynamic UPDATE.
	sets := []string{}
	args := []any{}
	i := 1
	if input.Username != "" {
		sets = append(sets, fmt.Sprintf("username = $%d", i))
		args = append(args, input.Username)
		i++
	}
	if input.Role != "" {
		sets = append(sets, fmt.Sprintf("role = $%d", i))
		args = append(args, input.Role)
		i++
	}
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			fail(c, http.StatusInternalServerError, err)
			return
		}
		sets = append(sets, fmt.Sprintf("password_hash = $%d", i))
		args = append(args, string(hash))
		i++
	}
	if input.Role == "guest" {
		sets = append(sets, fmt.Sprintf("password_hash = $%d", i))
		args = append(args, nil)
		i++
	}

	if len(sets) == 0 {
		fail(c, http.StatusBadRequest, fmt.Errorf("no fields to update"))
		return
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d RETURNING id, username, role, created_at",
		strings.Join(sets, ", "), i,
	)

	var user models.User
	err := h.db.QueryRowContext(c.Request.Context(), query, args...).
		Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	logAudit(c.Request.Context(), h.db, c, "update", "users", user.ID, fmt.Sprintf("Updated user '%s' (role: %s)", user.Username, user.Role))
	ok(c, http.StatusOK, user)
}
