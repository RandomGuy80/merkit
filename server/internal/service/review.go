package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"marketplace/internal/model"
)

type ReviewService struct {
	db *pgxpool.Pool
}

func NewReviewService(db *pgxpool.Pool) *ReviewService {
	return &ReviewService{db: db}
}

func (s *ReviewService) Create(ctx context.Context, reviewerID string, req model.CreateReviewRequest) (*model.Review, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	// only buyer of a delivered order can leave a review
	var order model.Order
	err := s.db.QueryRow(ctx,
		`SELECT id, buyer_id, seller_id, status FROM orders WHERE id=$1`,
		req.OrderID,
	).Scan(&order.ID, &order.BuyerID, &order.SellerID, &order.Status)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.BuyerID != reviewerID {
		return nil, fmt.Errorf("only the buyer can leave a review")
	}
	if order.Status != model.OrderDelivered {
		return nil, fmt.Errorf("can only review after delivery")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var review model.Review
	err = tx.QueryRow(ctx,
		`INSERT INTO reviews (order_id, reviewer_id, seller_id, rating, body)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, order_id, reviewer_id, seller_id, rating, COALESCE(body,''), created_at`,
		req.OrderID, reviewerID, order.SellerID, req.Rating, req.Body,
	).Scan(&review.ID, &review.OrderID, &review.ReviewerID, &review.SellerID,
		&review.Rating, &review.Body, &review.CreatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, fmt.Errorf("review already submitted for this order")
		}
		return nil, fmt.Errorf("create review: %w", err)
	}

	// update seller avg rating atomically
	if _, err := tx.Exec(ctx, `
		UPDATE users SET
			rating = (SELECT ROUND(AVG(rating)::numeric, 2) FROM reviews WHERE seller_id=$1),
			review_count = (SELECT COUNT(*) FROM reviews WHERE seller_id=$1)
		WHERE id=$1`, order.SellerID,
	); err != nil {
		return nil, fmt.Errorf("update rating: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *ReviewService) ListForSeller(ctx context.Context, sellerID string) ([]model.Review, error) {
	rows, err := s.db.Query(ctx,
		`SELECT r.id, r.order_id, r.reviewer_id, r.seller_id, r.rating, COALESCE(r.body,''), r.created_at,
		        u.name, u.avatar
		 FROM reviews r
		 JOIN users u ON u.id = r.reviewer_id
		 WHERE r.seller_id=$1 ORDER BY r.created_at DESC`,
		sellerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var r model.Review
		u := &model.User{}
		if err := rows.Scan(&r.ID, &r.OrderID, &r.ReviewerID, &r.SellerID,
			&r.Rating, &r.Body, &r.CreatedAt, &u.Name, &u.Avatar); err != nil {
			return nil, err
		}
		r.Reviewer = u
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return len(msg) > 0 && (errContains(msg, "unique") || errContains(msg, "duplicate"))
}

func errContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
