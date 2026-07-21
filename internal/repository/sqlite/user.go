package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	client *Client
}

func NewUserRepository(client *Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now()
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	err := r.client.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		now.Unix(), // Convert to Unix timestamp
		now.Unix(), // Convert to Unix timestamp
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`

	var user domain.User
	var createdAt int64
	var updatedAt int64

	err := r.client.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = ?
	`

	var user domain.User
	var createdAt int64
	var updatedAt int64

	err := r.client.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users WHERE username = ?
	`

	var user domain.User
	var createdAt int64
	var updatedAt int64

	err := r.client.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	user.CreatedAt = time.Unix(createdAt, 0)
	user.UpdatedAt = time.Unix(updatedAt, 0)

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users
		SET username = ?, email = ?, password_hash = ?, role = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.client.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.UpdatedAt.Unix(), // Convert to Unix timestamp
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`
	var exists bool
	err := r.client.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`
	var exists bool
	err := r.client.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check username exists: %w", err)
	}
	return exists, nil
}
