package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

//go:embed schema.sql
var schemaSQL string

func Open(dbPath string) (Repository, error) {
	if dbPath == "" {
		dbPath = filepath.Join("data", "hatch.db")
	}
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("store: create db dir %s: %w", dir, err)
	}
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", dbPath, err)
	}
	// SQLite supports only one writer at a time. Using MaxOpenConns(1)
	// prevents "database is locked" errors under concurrent writes.
	// For read-heavy workloads with WAL mode, this can be increased,
	// but the single-writer constraint remains.
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0) // No limit; SQLite connections are lightweight.
	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}
	return NewSQLiteRepo(conn)
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("store/migrate: %w", err)
	}
	return nil
}
