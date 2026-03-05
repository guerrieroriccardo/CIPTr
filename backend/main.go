package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

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
