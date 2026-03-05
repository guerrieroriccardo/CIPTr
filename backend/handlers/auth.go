package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
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
		`SELECT id, username, password_hash, is_admin, created_at FROM users WHERE username = $1`,
		input.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt)
	if err == sql.ErrNoRows {
		fail(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		fail(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	ok(c, http.StatusOK, gin.H{"token": signed})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}

	var user models.User
	err = h.db.QueryRowContext(c.Request.Context(),
		`INSERT INTO users (username, password_hash) VALUES ($1, $2)
		 RETURNING id, username, is_admin, created_at`,
		input.Username, string(hash),
	).Scan(&user.ID, &user.Username, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		fail(c, http.StatusConflict, fmt.Errorf("username already taken"))
		return
	}

	ok(c, http.StatusCreated, user)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	ok(c, http.StatusOK, gin.H{"id": userID, "username": username})
}
