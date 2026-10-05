package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"marketplace/internal/model"
)

type ListingService struct {
	db        *pgxpool.Pool
	uploadDir string
}

func NewListingService(db *pgxpool.Pool, uploadDir string) *ListingService {
	return &ListingService{db: db, uploadDir: uploadDir}
}

func (s *ListingService) Create(ctx context.Context, sellerID string, req model.CreateListingRequest) (*model.Listing, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("title required")
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	var l model.Listing
	err := s.db.QueryRow(ctx,
		`INSERT INTO listings (seller_id, category_id, title, description, price, currency, location, tags)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, seller_id, category_id, title, description, price, currency, status, location, images, tags, views, created_at, updated_at`,
		sellerID, req.CategoryID, req.Title, req.Description, req.Price, req.Currency, req.Location, tagsOrEmpty(req.Tags),
	).Scan(&l.ID, &l.SellerID, &l.CategoryID, &l.Title, &l.Description, &l.Price, &l.Currency,
		&l.Status, &l.Location, &l.Images, &l.Tags, &l.Views, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	return &l, nil
}

func (s *ListingService) GetByID(ctx context.Context, id string) (*model.Listing, error) {
	var l model.Listing
	err := s.db.QueryRow(ctx,
		`UPDATE listings SET views = views + 1 WHERE id=$1
		 RETURNING id, seller_id, category_id, title, description, price, currency, status, location, images, tags, views, created_at, updated_at`,
		id,
	).Scan(&l.ID, &l.SellerID, &l.CategoryID, &l.Title, &l.Description, &l.Price, &l.Currency,
		&l.Status, &l.Location, &l.Images, &l.Tags, &l.Views, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("listing not found")
	}
	return &l, nil
}

func (s *ListingService) Update(ctx context.Context, id, sellerID string, req model.UpdateListingRequest) (*model.Listing, error) {
	existing, err := s.getByIDNoView(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.SellerID != sellerID {
		return nil, fmt.Errorf("forbidden")
	}

	if req.Title != nil {
		existing.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, fmt.Errorf("price must be positive")
		}
		existing.Price = *req.Price
	}
	if req.Location != nil {
		existing.Location = req.Location
	}
	if req.CategoryID != nil {
		existing.CategoryID = req.CategoryID
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}

	var l model.Listing
	err = s.db.QueryRow(ctx,
		`UPDATE listings SET category_id=$1, title=$2, description=$3, price=$4, location=$5, tags=$6, status=$7, updated_at=NOW()
		 WHERE id=$8
		 RETURNING id, seller_id, category_id, title, description, price, currency, status, location, images, tags, views, created_at, updated_at`,
		existing.CategoryID, existing.Title, existing.Description, existing.Price,
		existing.Location, tagsOrEmpty(existing.Tags), existing.Status, id,
	).Scan(&l.ID, &l.SellerID, &l.CategoryID, &l.Title, &l.Description, &l.Price, &l.Currency,
		&l.Status, &l.Location, &l.Images, &l.Tags, &l.Views, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update listing: %w", err)
	}
	return &l, nil
}

func (s *ListingService) Delete(ctx context.Context, id, callerID string, isAdmin bool) error {
	l, err := s.getByIDNoView(ctx, id)
	if err != nil {
		return err
	}
	if !isAdmin && l.SellerID != callerID {
		return fmt.Errorf("forbidden")
	}
	_, err = s.db.Exec(ctx, `DELETE FROM listings WHERE id=$1`, id)
	return err
}

func (s *ListingService) Search(ctx context.Context, f model.ListingsFilter) (*model.ListingsPage, error) {
	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}

	args := []any{}
	conds := []string{"l.status = 'active'"}
	i := 1

	if f.Query != "" {
		conds = append(conds, fmt.Sprintf("l.search_vec @@ plainto_tsquery('english', $%d)", i))
		args = append(args, f.Query)
		i++
	}
	if f.CategoryID != nil {
		conds = append(conds, fmt.Sprintf("l.category_id = $%d", i))
		args = append(args, *f.CategoryID)
		i++
	}
	if f.MinPrice != nil {
		conds = append(conds, fmt.Sprintf("l.price >= $%d", i))
		args = append(args, *f.MinPrice)
		i++
	}
	if f.MaxPrice != nil {
		conds = append(conds, fmt.Sprintf("l.price <= $%d", i))
		args = append(args, *f.MaxPrice)
		i++
	}
	if f.SellerID != "" {
		conds = append(conds, fmt.Sprintf("l.seller_id = $%d", i))
		args = append(args, f.SellerID)
		i++
	}
	if f.Cursor != "" {
		conds = append(conds, fmt.Sprintf("l.created_at < (SELECT created_at FROM listings WHERE id=$%d)", i))
		args = append(args, f.Cursor)
		i++
	}

	where := strings.Join(conds, " AND ")
	args = append(args, f.Limit+1)
	query := fmt.Sprintf(`
		SELECT l.id, l.seller_id, l.category_id, l.title, l.description, l.price, l.currency,
		       l.status, l.location, l.images, l.tags, l.views, l.created_at, l.updated_at
		FROM listings l
		WHERE %s
		ORDER BY l.created_at DESC
		LIMIT $%d`, where, i)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search listings: %w", err)
	}
	defer rows.Close()

	var items []model.Listing
	for rows.Next() {
		var l model.Listing
		if err := rows.Scan(&l.ID, &l.SellerID, &l.CategoryID, &l.Title, &l.Description,
			&l.Price, &l.Currency, &l.Status, &l.Location, &l.Images, &l.Tags, &l.Views,
			&l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	page := &model.ListingsPage{}
	if len(items) > f.Limit {
		page.NextCursor = items[f.Limit-1].ID
		items = items[:f.Limit]
	}
	page.Items = items
	return page, nil
}

func (s *ListingService) AddImage(ctx context.Context, listingID, sellerID string, data []byte) (string, error) {
	l, err := s.getByIDNoView(ctx, listingID)
	if err != nil {
		return "", err
	}
	if l.SellerID != sellerID {
		return "", fmt.Errorf("forbidden")
	}
	if len(l.Images) >= 5 {
		return "", fmt.Errorf("maximum 5 images per listing")
	}

	mimeType := detectMIME(data)
	ext, ok := allowedImageExt(mimeType)
	if !ok {
		return "", fmt.Errorf("unsupported image type")
	}

	filename := uuid.NewString() + ext
	imgDir := filepath.Join(s.uploadDir, "listings")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(imgDir, filename), data, 0644); err != nil {
		return "", err
	}

	url := "/uploads/listings/" + filename
	_, err = s.db.Exec(ctx,
		`UPDATE listings SET images = array_append(images, $1), updated_at=NOW() WHERE id=$2`,
		url, listingID)
	if err != nil {
		os.Remove(filepath.Join(imgDir, filename))
		return "", err
	}
	return url, nil
}

func (s *ListingService) RemoveImage(ctx context.Context, listingID, sellerID string, index int) error {
	l, err := s.getByIDNoView(ctx, listingID)
	if err != nil {
		return err
	}
	if l.SellerID != sellerID {
		return fmt.Errorf("forbidden")
	}
	if index < 0 || index >= len(l.Images) {
		return fmt.Errorf("image index out of range")
	}

	imgPath := l.Images[index]
	newImages := append(l.Images[:index], l.Images[index+1:]...)

	_, err = s.db.Exec(ctx,
		`UPDATE listings SET images=$1, updated_at=NOW() WHERE id=$2`,
		newImages, listingID)
	if err != nil {
		return err
	}

	// delete file from disk
	filename := filepath.Base(imgPath)
	_ = os.Remove(filepath.Join(s.uploadDir, "listings", filename))
	return nil
}

func (s *ListingService) getByIDNoView(ctx context.Context, id string) (*model.Listing, error) {
	var l model.Listing
	err := s.db.QueryRow(ctx,
		`SELECT id, seller_id, category_id, title, description, price, currency, status, location, images, tags, views, created_at, updated_at
		 FROM listings WHERE id=$1`, id,
	).Scan(&l.ID, &l.SellerID, &l.CategoryID, &l.Title, &l.Description, &l.Price, &l.Currency,
		&l.Status, &l.Location, &l.Images, &l.Tags, &l.Views, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("listing not found")
	}
	return &l, nil
}

func tagsOrEmpty(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}
