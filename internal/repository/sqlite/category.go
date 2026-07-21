package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	client *Client
}

func NewCategoryRepository(client *Client) *CategoryRepository {
	return &CategoryRepository{client: client}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.client.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at
		FROM categories
		WHERE id = ?
	`

	category, err := scanCategory(r.client.db.QueryRowContext(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get category by id: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) GetByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error) {
	postRepo := NewPostRepository(r.client)
	return postRepo.GetCategoriesByPostID(ctx, postID)
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	now := time.Now().Unix()
	query := `
		INSERT INTO categories (name, slug, description, created_at)
		VALUES (?, ?, ?, ?)
		RETURNING id
	`

	err := r.client.db.QueryRowContext(ctx, query,
		category.Name,
		category.Slug,
		category.Description,
		now,
	).Scan(&category.ID)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	category.CreatedAt = time.Unix(now, 0)
	return nil
}

type categoryScanner interface {
	Scan(dest ...interface{}) error
}

func scanCategory(scanner categoryScanner) (domain.Category, error) {
	var category domain.Category
	var createdAt int64

	err := scanner.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&createdAt,
	)
	if err != nil {
		return domain.Category{}, err
	}

	category.CreatedAt = time.Unix(createdAt, 0)
	return category, nil
}
