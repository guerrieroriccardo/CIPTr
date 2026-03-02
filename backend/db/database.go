package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Open opens (or creates) the SQLite database at the given path,
// applies the schema, and returns the connection.
func Open(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite performs best with a single writer connection.
	database.SetMaxOpenConns(1)

	if err := applySchema(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return database, nil
}

func applySchema(database *sql.DB) error {
	// Enable foreign keys and WAL mode on every new connection.
	if _, err := database.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL;"); err != nil {
		return err
	}
	if _, err := database.Exec(schema); err != nil {
		return err
	}
	return nil
}
