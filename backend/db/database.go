package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed schema.sql
var schema string

// Open opens a connection to the PostgreSQL database at the given DSN,
// applies the schema (idempotent: all statements use IF NOT EXISTS),
// and returns the connection pool.
//
// DSN format: postgres://user:password@host:5432/dbname?sslmode=disable
func Open(dsn string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if _, err := database.Exec(schema); err != nil {
		database.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return database, nil
}
