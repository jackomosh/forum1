package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"forum/internal/domain"
)

type PostRepository struct {
	client *Client
}

func NewPostRepository(client *Client) *PostRepository {
	return &PostRepository{client: client}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post, categoryIDs []domain.CategoryID) error {
	tx, err := r.client.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	query := `
		INSERT INTO posts (author_id, title, body, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	err = tx.QueryRowContext(ctx, query,
		post.AuthorID,
		post.Title,
		post.Body,
		post.Status,
		now.Unix(), // Convert to Unix timestamp
		now.Unix(), // Convert to Unix timestamp
	).Scan(&post.ID)
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	// Add categories
	if len(categoryIDs) > 0 {
		valueStrings := make([]string, 0, len(categoryIDs))
		valueArgs := make([]interface{}, 0, len(categoryIDs)*2)

		for _, catID := range categoryIDs {
			valueStrings = append(valueStrings, "(?, ?)")
			valueArgs = append(valueArgs, post.ID, catID)
		}

		insertQuery := fmt.Sprintf(
			"INSERT INTO post_categories (post_id, category_id) VALUES %s",
			strings.Join(valueStrings, ", "),
		)

		_, err = tx.ExecContext(ctx, insertQuery, valueArgs...)
		if err != nil {
			return fmt.Errorf("add categories: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id domain.PostID) (*domain.PostWithAuthor, error) {
	query := `
		SELECT 
			p.id, p.title, p.body, p.status, p.created_at, p.updated_at,
			u.id, u.username, u.role, u.created_at
		FROM posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = ?
	`

	var post domain.Post
	var author domain.PublicUser
	var postCreatedAt int64
	var postUpdatedAt int64
	var authorCreatedAt int64

	err := r.client.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Body,
		&post.Status,
		&postCreatedAt,
		&postUpdatedAt,
		&author.ID,
		&author.Username,
		&author.Role,
		&authorCreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get post by id: %w", err)
	}

	post.CreatedAt = time.Unix(postCreatedAt, 0)
	post.UpdatedAt = time.Unix(postUpdatedAt, 0)
	author.CreatedAt = time.Unix(authorCreatedAt, 0)

	// Get categories
	categories, err := r.GetCategoriesByPostID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}

	// Get stats
	stats, err := r.getPostStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}

	return &domain.PostWithAuthor{
		Post:       post,
		Author:     author,
		Categories: categories,
		Stats:      *stats,
		UserVote:   domain.VoteNone,
	}, nil
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	post.UpdatedAt = time.Now()
	query := `
		UPDATE posts
		SET title = ?, body = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.client.db.ExecContext(ctx, query,
		post.Title,
		post.Body,
		post.Status,
		post.UpdatedAt.Unix(), // Convert to Unix timestamp
		post.ID,
	)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id domain.PostID) error {
	query := `DELETE FROM posts WHERE id = ?`
	result, err := r.client.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *PostRepository) List(ctx context.Context, filter domain.PostFilter) ([]domain.PostWithAuthor, int, int, error) {
	// Build base query with filters
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "p.status = 'published'")

	if filter.AuthorID > 0 {
		conditions = append(conditions, "p.author_id = ?")
		args = append(args, filter.AuthorID)
	}

	if filter.ViewerID > 0 && filter.Kind == domain.PostFilterLiked {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM votes v 
				WHERE v.target_type = 'post' 
				AND v.target_id = p.id 
				AND v.user_id = ? 
				AND v.value = 1
			)
		`)
		args = append(args, filter.ViewerID)
	}

	if filter.ViewerID > 0 && filter.Kind == domain.PostFilterCreated {
		conditions = append(conditions, "p.author_id = ?")
		args = append(args, filter.ViewerID)
	}

	if filter.CategoryID > 0 {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM post_categories pc 
				WHERE pc.post_id = p.id 
				AND pc.category_id = ?
			)
		`)
		args = append(args, filter.CategoryID)
	}

	if filter.Search != "" {
		conditions = append(conditions, "(p.title LIKE ? OR p.body LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM posts p
		%s
	`, whereClause)

	var total int
	err := r.client.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("count posts: %w", err)
	}

	// Order by
	orderBy := "p.created_at DESC"
	switch filter.Sort {
	case domain.SortOldest:
		orderBy = "p.created_at ASC"
	case domain.SortMostLiked:
		orderBy = "like_count DESC"
	case domain.SortMostComment:
		orderBy = "comment_count DESC"
	}

	// Get posts with stats
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.title, p.body, p.status, p.created_at, p.updated_at,
			u.id, u.username, u.role, u.created_at,
			COALESCE((
				SELECT COUNT(*) FROM votes v 
				WHERE v.target_type = 'post' AND v.target_id = p.id AND v.value = 1
			), 0) as like_count,
			COALESCE((
				SELECT COUNT(*) FROM votes v 
				WHERE v.target_type = 'post' AND v.target_id = p.id AND v.value = -1
			), 0) as dislike_count,
			COALESCE((
				SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.status = 'visible'
			), 0) as comment_count
		FROM posts p
		JOIN users u ON p.author_id = u.id
		%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	limit := 20
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	queryArgs := append(args, limit, offset)
	rows, err := r.client.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []domain.PostWithAuthor
	for rows.Next() {
		var p domain.Post
		var author domain.PublicUser
		var stats domain.PostStats
		var pCreatedAt, pUpdatedAt int64
		var authorCreatedAt int64

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Body,
			&p.Status,
			&pCreatedAt,
			&pUpdatedAt,
			&author.ID,
			&author.Username,
			&author.Role,
			&authorCreatedAt,
			&stats.LikeCount,
			&stats.DislikeCount,
			&stats.CommentCount,
		)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("scan post: %w", err)
		}

		p.CreatedAt = time.Unix(pCreatedAt, 0)
		p.UpdatedAt = time.Unix(pUpdatedAt, 0)
		author.CreatedAt = time.Unix(authorCreatedAt, 0)

		// Get categories for this post
		categories, err := r.GetCategoriesByPostID(ctx, p.ID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("get categories for post: %w", err)
		}

		stats.Score = stats.LikeCount - stats.DislikeCount

		posts = append(posts, domain.PostWithAuthor{
			Post:       p,
			Author:     author,
			Categories: categories,
			Stats:      stats,
			UserVote:   domain.VoteNone,
		})
	}

	return posts, total, limit, nil
}

func (r *PostRepository) GetCategoriesByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error) {
	query := `
		SELECT c.id, c.name, c.slug, c.description, c.created_at
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`

	rows, err := r.client.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var cat domain.Category
		var createdAt int64
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.Description,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cat.CreatedAt = time.Unix(createdAt, 0)
		categories = append(categories, cat)
	}

	return categories, nil
}

func (r *PostRepository) getPostStats(ctx context.Context, postID domain.PostID) (*domain.PostStats, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END), 0) as likes,
			COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END), 0) as dislikes,
			COALESCE((SELECT COUNT(*) FROM comments WHERE post_id = ? AND status = 'visible'), 0) as comments
		FROM votes 
		WHERE target_type = 'post' AND target_id = ?
	`

	var stats domain.PostStats
	err := r.client.db.QueryRowContext(ctx, query, postID, postID).Scan(
		&stats.LikeCount,
		&stats.DislikeCount,
		&stats.CommentCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get post stats: %w", err)
	}

	stats.Score = stats.LikeCount - stats.DislikeCount
	return &stats, nil
}
