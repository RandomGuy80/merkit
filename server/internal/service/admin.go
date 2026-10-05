package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"marketplace/internal/model"
)

type AdminService struct {
	db *pgxpool.Pool
}

func NewAdminService(db *pgxpool.Pool) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]model.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, email, name, role, avatar, bio, rating, review_count, created_at
		 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role,
			&u.Avatar, &u.Bio, &u.Rating, &u.ReviewCount, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *AdminService) SetRole(ctx context.Context, userID string, role model.Role) (*model.User, error) {
	if role != model.RoleBuyer && role != model.RoleSeller && role != model.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}
	var u model.User
	err := s.db.QueryRow(ctx,
		`UPDATE users SET role=$1 WHERE id=$2
		 RETURNING id, email, name, role, avatar, bio, rating, review_count, created_at`,
		role, userID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role,
		&u.Avatar, &u.Bio, &u.Rating, &u.ReviewCount, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &u, nil
}

func (s *AdminService) DeleteListing(ctx context.Context, listingID string) error {
	res, err := s.db.Exec(ctx, `DELETE FROM listings WHERE id=$1`, listingID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("listing not found")
	}
	return nil
}

func (s *AdminService) ListOrders(ctx context.Context, limit, offset int) ([]model.Order, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, listing_id, buyer_id, seller_id, amount, platform_fee, status, stripe_payment_id, created_at, updated_at
		 FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.ListingID, &o.BuyerID, &o.SellerID,
			&o.Amount, &o.PlatformFee, &o.Status, &o.StripePaymentID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
