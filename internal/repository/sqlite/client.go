package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type ClientOptions struct {
	Path            string
	ForeignKeys     bool
	BusyTimeoutMS   int
	MaxOpenConns    int
	MaxIdleConns    int
	SchemaPath      string
	RunMigrations   bool
	MigrationsTable string
	ConnectionName  string
}

type MigrationRecord struct {
	ID        int64
	Name      string
	Checksum  string
	AppliedAt string
}

type Client struct {
	db *sql.DB
}

func NewClient(dbPath string) (*Client, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) DB() *sql.DB {
	return c.db
}
