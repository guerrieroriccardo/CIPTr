package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"math/big"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/guerrieroriccardo/CIPTr/backend/db"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ciptr:ciptr@localhost:5432/ciptr?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.Open(dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("database connected")

	ensureDefaultAdmin(database)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("failed to generate JWT secret: %v", err)
		}
		jwtSecret = b
		log.Printf("WARNING: no JWT_SECRET set, generated random secret: %s", hex.EncodeToString(b))
		log.Printf("Set JWT_SECRET env var to persist tokens across restarts")
	}

	r := setupRouter(database, jwtSecret)

	log.Printf("starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// ensureDefaultAdmin creates an admin user with random credentials if no users exist.
func ensureDefaultAdmin(database *sql.DB) {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		log.Printf("WARNING: could not check users table: %v", err)
		return
	}
	if count > 0 {
		return
	}

	password := randomPassword(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash default admin password: %v", err)
	}

	_, err = database.Exec(
		`INSERT INTO users (username, password_hash, role) VALUES ($1, $2, 'admin')`,
		"admin", string(hash),
	)
	if err != nil {
		log.Printf("WARNING: could not create default admin: %v", err)
		return
	}

	log.Println("══════════════════════════════════════════════════")
	log.Println("  FIRST BOOT — default admin account created")
	log.Printf("  Username: admin")
	log.Printf("  Password: %s", password)
	log.Println("  Change this password after first login!")
	log.Println("══════════════════════════════════════════════════")
}

func randomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}
