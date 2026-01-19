package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	db *DB
}

func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, slug, parent_id, description, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.ParentID,
		category.Description,
		category.ImageURL,
	).Scan(&category.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	query := `SELECT id, name, slug, parent_id, description, image_url, created_at FROM categories WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	category := &domain.Category{}
	err := q.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.Description,
		&category.ImageURL,
		&category.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	query := `SELECT id, name, slug, parent_id, description, image_url, created_at FROM categories WHERE slug = $1`

	q := r.db.GetQuerier(ctx)
	category := &domain.Category{}
	err := q.QueryRow(ctx, query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.Description,
		&category.ImageURL,
		&category.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = $2, slug = $3, parent_id = $4, description = $5, image_url = $6
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.ParentID,
		category.Description,
		category.ImageURL,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT id, name, slug, parent_id, description, image_url, created_at FROM categories ORDER BY name`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.ParentID,
			&category.Description,
			&category.ImageURL,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) GetWithAuctionCounts(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT c.id, c.name, c.slug, c.parent_id, c.description, c.image_url, c.created_at,
		       COUNT(a.id) as auction_count
		FROM categories c
		LEFT JOIN auctions a ON c.id = a.category_id AND a.status = 'active'
		GROUP BY c.id
		ORDER BY c.name`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories with counts: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.ParentID,
			&category.Description,
			&category.ImageURL,
			&category.CreatedAt,
			&category.AuctionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}
