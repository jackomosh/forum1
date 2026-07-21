package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

var _ repository.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	client *Client
}

func NewSessionRepository(client *Client) *SessionRepository {
	return &SessionRepository{client: client}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}

	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.client.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.ExpiresAt.Unix(),
		session.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ?
	`

	var session domain.Session
	var expiresAt int64
	var createdAt int64

	err := r.client.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&expiresAt,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session by id: %w", err)
	}

	session.ExpiresAt = time.Unix(expiresAt, 0)
	session.CreatedAt = time.Unix(createdAt, 0)
	return &session, nil
}

func (r *SessionRepository) GetByUserID(ctx context.Context, userID domain.UserID) ([]domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.client.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get sessions by user id: %w", err)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var session domain.Session
		var expiresAt int64
		var createdAt int64

		err := rows.Scan(&session.ID, &session.UserID, &expiresAt, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}

		session.ExpiresAt = time.Unix(expiresAt, 0)
		session.CreatedAt = time.Unix(createdAt, 0)
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	_, err := r.client.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := r.client.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete sessions by user id: %w", err)
	}
	return nil
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.client.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("cleanup expired sessions: %w", err)
	}
	return nil
}
