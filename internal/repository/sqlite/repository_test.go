package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"forum/internal/domain"
)

// getTestSchema returns the schema as a string (no file I/O)
func getTestSchema() string {
	return `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'published',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS post_categories (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		body TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'visible',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS votes (
		user_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		value INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		PRIMARY KEY (user_id, target_type, target_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
	CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
	CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
	CREATE INDEX IF NOT EXISTS idx_comments_author_id ON comments(author_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	CREATE INDEX IF NOT EXISTS idx_votes_target ON votes(target_type, target_id);
	`
}

// runTestMigrations runs migrations on a test database
func runTestMigrations(db *sql.DB) error {
	_, err := db.Exec(getTestSchema())
	return err
}

// TestNewClient tests database client creation
func TestNewClient(t *testing.T) {
	t.Log("Testing NewClient...")
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Error("Expected client to not be nil")
	}

	if client.db == nil {
		t.Error("Expected db connection to not be nil")
	}

	err = client.db.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	t.Log("NewClient test passed!")
}

// TestUserRepository_Create tests user creation
func TestUserRepository_Create(t *testing.T) {
	t.Log("Testing UserRepository.Create...")

	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if err := runTestMigrations(client.db); err != nil {
		t.Fatalf("runTestMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	user := &domain.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set, got 0")
	}
	t.Log("UserRepository.Create test passed!")
}

// TestUserRepository_GetByEmail tests retrieving a user by email
func TestUserRepository_GetByEmail(t *testing.T) {
	t.Log("Testing UserRepository.GetByEmail...")

	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if err := runTestMigrations(client.db); err != nil {
		t.Fatalf("runTestMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	user := &domain.User{
		Username:     "getuser",
		Email:        "getuser@example.com",
		PasswordHash: "hashedpassword",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.GetByEmail(context.Background(), "getuser@example.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	if found == nil {
		t.Error("Expected to find user, got nil")
	} else if found.Email != "getuser@example.com" {
		t.Errorf("Expected email 'getuser@example.com', got '%s'", found.Email)
	}
	t.Log("UserRepository.GetByEmail test passed!")
}

// TestUserRepository_ExistsByEmail tests email existence check
func TestUserRepository_ExistsByEmail(t *testing.T) {
	t.Log("Testing UserRepository.ExistsByEmail...")

	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if err := runTestMigrations(client.db); err != nil {
		t.Fatalf("runTestMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	user := &domain.User{
		Username:     "existsuser",
		Email:        "exists@example.com",
		PasswordHash: "hashedpassword",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	exists, err := repo.ExistsByEmail(context.Background(), "exists@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}

	if !exists {
		t.Error("Expected email to exist, got false")
	}

	exists, err = repo.ExistsByEmail(context.Background(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}

	if exists {
		t.Error("Expected email to not exist, got true")
	}
	t.Log("UserRepository.ExistsByEmail test passed!")
}
