package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

type VoteRow struct {
	UserID    int64
	Target    string
	TargetID  int64
	Value     int
	CreatedAt string
	UpdatedAt string
}

var _ repository.VoteRepository = (*VoteRepository)(nil)

// VoteRepository handles vote persistence
type VoteRepository struct {
	client *Client
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(client *Client) *VoteRepository {
	return &VoteRepository{client: client}
}

// AddVote adds or updates a vote (UPSERT pattern)
func (r *VoteRepository) AddVote(ctx context.Context, vote *domain.Vote) error {
	now := time.Now().Unix()
	query := `
		INSERT INTO votes (user_id, target_type, target_id, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, target_type, target_id) 
		DO UPDATE SET value = ?, updated_at = ?
	`

	_, err := r.client.db.ExecContext(
		ctx, query,
		vote.UserID,
		string(vote.Target),
		vote.TargetID,
		vote.Value,
		now,
		now,
		vote.Value,
		now,
	)
	if err != nil {
		return fmt.Errorf("add vote: %w", err)
	}

	return nil
}

// GetVote retrieves a user's vote on a specific target
func (r *VoteRepository) GetVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) (*domain.Vote, error) {
	query := `
		SELECT user_id, target_type, target_id, value, created_at, updated_at
		FROM votes
		WHERE user_id = ? AND target_type = ? AND target_id = ?
	`

	var vote domain.Vote
	var targetType string
	var createdAt int64
	var updatedAt int64

	err := r.client.db.QueryRowContext(ctx, query, userID, string(target), targetID).Scan(
		&vote.UserID,
		&targetType,
		&vote.TargetID,
		&vote.Value,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get vote: %w", err)
	}

	vote.Target = domain.VoteTarget(targetType)
	vote.CreatedAt = time.Unix(createdAt, 0)
	vote.UpdatedAt = time.Unix(updatedAt, 0)
	return &vote, nil
}

// RemoveVote deletes a vote
func (r *VoteRepository) RemoveVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) error {
	query := `
		DELETE FROM votes
		WHERE user_id = ? AND target_type = ? AND target_id = ?
	`

	result, err := r.client.db.ExecContext(ctx, query, userID, string(target), targetID)
	if err != nil {
		return fmt.Errorf("remove vote: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("vote not found")
	}

	return nil
}

// GetPostStats retrieves aggregated vote statistics for a post
func (r *VoteRepository) GetPostStats(ctx context.Context, postID domain.PostID) (*domain.PostStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END), 0) as likes,
			COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END), 0) as dislikes
		FROM votes 
		WHERE target_type = 'post' AND target_id = ?
	`

	var stats domain.PostStats
	err := r.client.db.QueryRowContext(ctx, query, postID).Scan(
		&stats.LikeCount,
		&stats.DislikeCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get post stats: %w", err)
	}

	stats.Score = stats.LikeCount - stats.DislikeCount
	return &stats, nil
}

// GetCommentStats retrieves aggregated vote statistics for a comment
func (r *VoteRepository) GetCommentStats(ctx context.Context, commentID domain.CommentID) (*domain.CommentStats, error) {
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

// GetUserVotedPostIDs retrieves all post IDs that a user has liked
func (r *VoteRepository) GetUserVotedPostIDs(ctx context.Context, userID domain.UserID) ([]domain.PostID, error) {
	query := `
		SELECT target_id
		FROM votes
		WHERE user_id = ? AND target_type = 'post' AND value = 1
	`

	rows, err := r.client.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user voted posts: %w", err)
	}
	defer rows.Close()

	var postIDs []domain.PostID
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("scan post id: %w", err)
		}
		postIDs = append(postIDs, domain.PostID(id))
	}

	return postIDs, nil
}
