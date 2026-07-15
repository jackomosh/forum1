package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/internal/domain"
)

type CommentRepository struct {
	client *Client
}

func NewCommentRepository(client *Client) *CommentRepository {
	return &CommentRepository{client: client}
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	now := time.Now()
	query := `
		INSERT INTO comments (post_id, author_id, body, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	err := r.client.db.QueryRowContext(ctx, query,
		comment.PostID,
		comment.AuthorID,
		comment.Body,
		comment.Status,
		now.Unix(), // Convert to Unix timestamp
		now.Unix(), // Convert to Unix timestamp
	).Scan(&comment.ID)

	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}

	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id domain.CommentID) (*domain.CommentWithAuthor, error) {
	query := `
		SELECT 
			c.id, c.post_id, c.author_id, c.body, c.status, c.created_at, c.updated_at,
			u.id, u.username, u.role, u.created_at
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = ?
	`

	var comment domain.Comment
	var author domain.PublicUser
	var commentCreatedAt int64
	var commentUpdatedAt int64
	var authorCreatedAt int64

	err := r.client.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorID,
		&comment.Body,
		&comment.Status,
		&commentCreatedAt,
		&commentUpdatedAt,
		&author.ID,
		&author.Username,
		&author.Role,
		&authorCreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get comment by id: %w", err)
	}

	comment.CreatedAt = time.Unix(commentCreatedAt, 0)
	comment.UpdatedAt = time.Unix(commentUpdatedAt, 0)
	author.CreatedAt = time.Unix(authorCreatedAt, 0)

	// Get stats
	stats, err := r.getCommentStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get comment stats: %w", err)
	}

	return &domain.CommentWithAuthor{
		Comment:  comment,
		Author:   author,
		Stats:    *stats,
		UserVote: domain.VoteNone,
	}, nil
}

func (r *CommentRepository) GetByPostID(ctx context.Context, postID domain.PostID, limit, offset int) ([]domain.CommentWithAuthor, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM comments 
		WHERE post_id = ? AND status = 'visible'
	`

	var total int
	err := r.client.db.QueryRowContext(ctx, countQuery, postID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}

	if limit == 0 {
		limit = 20
	}

	// Get comments with author info
	query := `
		SELECT 
			c.id, c.post_id, c.author_id, c.body, c.status, c.created_at, c.updated_at,
			u.id, u.username, u.role, u.created_at
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.post_id = ? AND c.status = 'visible'
		ORDER BY c.created_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.client.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.CommentWithAuthor
	for rows.Next() {
		var comment domain.Comment
		var author domain.PublicUser
		var commentCreatedAt int64
		var commentUpdatedAt int64
		var authorCreatedAt int64

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.Body,
			&comment.Status,
			&commentCreatedAt,
			&commentUpdatedAt,
			&author.ID,
			&author.Username,
			&author.Role,
			&authorCreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan comment: %w", err)
		}

		comment.CreatedAt = time.Unix(commentCreatedAt, 0)
		comment.UpdatedAt = time.Unix(commentUpdatedAt, 0)
		author.CreatedAt = time.Unix(authorCreatedAt, 0)

		// Get stats for this comment
		stats, err := r.getCommentStats(ctx, comment.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("get comment stats: %w", err)
		}

		comments = append(comments, domain.CommentWithAuthor{
			Comment:  comment,
			Author:   author,
			Stats:    *stats,
			UserVote: domain.VoteNone,
		})
	}

	return comments, total, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	comment.UpdatedAt = time.Now()
	query := `
		UPDATE comments
		SET body = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.client.db.ExecContext(ctx, query,
		comment.Body,
		comment.Status,
		comment.UpdatedAt.Unix(), // Convert to Unix timestamp
		comment.ID,
	)

	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id domain.CommentID) error {
	query := `DELETE FROM comments WHERE id = ?`

	result, err := r.client.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

func (r *CommentRepository) getCommentStats(ctx context.Context, commentID domain.CommentID) (*domain.CommentStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END), 0) as likes,
			COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END), 0) as dislikes
		FROM votes 
		WHERE target_type = 'comment' AND target_id = ?
	`

	var stats domain.CommentStats
	err := r.client.db.QueryRowContext(ctx, query, commentID).Scan(
		&stats.LikeCount,
		&stats.DislikeCount,
	)

	if err != nil {
		return nil, fmt.Errorf("get comment stats: %w", err)
	}

	stats.Score = stats.LikeCount - stats.DislikeCount
	return &stats, nil
}
