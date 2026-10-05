package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"marketplace/internal/model"
)

type CategoryService struct {
	db *pgxpool.Pool
}

func NewCategoryService(db *pgxpool.Pool) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name, slug, COALESCE(icon,'') FROM categories ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Icon); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (s *CategoryService) Create(ctx context.Context, req model.CreateCategoryRequest) (*model.Category, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Name == "" || req.Slug == "" {
		return nil, fmt.Errorf("name and slug required")
	}

	var cat model.Category
	err := s.db.QueryRow(ctx,
		`INSERT INTO categories (name, slug, icon) VALUES ($1, $2, $3)
		 RETURNING id, name, slug, COALESCE(icon,'')`,
		req.Name, req.Slug, req.Icon,
	).Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Icon)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("category already exists")
		}
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &cat, nil
}

func (s *CategoryService) Delete(ctx context.Context, id int) error {
	res, err := s.db.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}
