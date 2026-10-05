package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"marketplace/internal/model"
)

const platformFeeRate = 0.05 // 5%

type OrderService struct {
	db *pgxpool.Pool
}

func NewOrderService(db *pgxpool.Pool) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) Create(ctx context.Context, buyerID string, req model.CreateOrderRequest) (*model.Order, error) {
	var l model.Listing
	err := s.db.QueryRow(ctx,
		`SELECT id, seller_id, price, status FROM listings WHERE id=$1`,
		req.ListingID,
	).Scan(&l.ID, &l.SellerID, &l.Price, &l.Status)
	if err != nil {
		return nil, fmt.Errorf("listing not found")
	}
	if l.Status != model.ListingActive {
		return nil, fmt.Errorf("listing is not available")
	}
	if l.SellerID == buyerID {
		return nil, fmt.Errorf("cannot buy your own listing")
	}

	fee := l.Price * platformFeeRate

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var order model.Order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (listing_id, buyer_id, seller_id, amount, platform_fee)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, listing_id, buyer_id, seller_id, amount, platform_fee, status, created_at, updated_at`,
		req.ListingID, buyerID, l.SellerID, l.Price, fee,
	).Scan(&order.ID, &order.ListingID, &order.BuyerID, &order.SellerID,
		&order.Amount, &order.PlatformFee, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// mark listing as sold
	if _, err := tx.Exec(ctx,
		`UPDATE listings SET status='sold', updated_at=NOW() WHERE id=$1`, req.ListingID,
	); err != nil {
		return nil, fmt.Errorf("mark listing sold: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) GetByID(ctx context.Context, orderID, callerID string, isAdmin bool) (*model.Order, error) {
	var o model.Order
	err := s.db.QueryRow(ctx,
		`SELECT id, listing_id, buyer_id, seller_id, amount, platform_fee, status, stripe_payment_id, created_at, updated_at
		 FROM orders WHERE id=$1`, orderID,
	).Scan(&o.ID, &o.ListingID, &o.BuyerID, &o.SellerID,
		&o.Amount, &o.PlatformFee, &o.Status, &o.StripePaymentID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if !isAdmin && o.BuyerID != callerID && o.SellerID != callerID {
		return nil, fmt.Errorf("forbidden")
	}
	return &o, nil
}

func (s *OrderService) ListForUser(ctx context.Context, userID string) ([]model.Order, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, listing_id, buyer_id, seller_id, amount, platform_fee, status, stripe_payment_id, created_at, updated_at
		 FROM orders WHERE buyer_id=$1 OR seller_id=$1 ORDER BY created_at DESC`,
		userID,
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

func (s *OrderService) UpdateStatus(ctx context.Context, orderID, callerID string, isAdmin bool, req model.UpdateOrderStatusRequest) (*model.Order, error) {
	order, err := s.GetByID(ctx, orderID, callerID, isAdmin)
	if err != nil {
		return nil, err
	}

	if err := validateStatusTransition(order.Status, req.Status, callerID, order, isAdmin); err != nil {
		return nil, err
	}

	var updated model.Order
	err = s.db.QueryRow(ctx,
		`UPDATE orders SET status=$1, updated_at=NOW() WHERE id=$2
		 RETURNING id, listing_id, buyer_id, seller_id, amount, platform_fee, status, stripe_payment_id, created_at, updated_at`,
		req.Status, orderID,
	).Scan(&updated.ID, &updated.ListingID, &updated.BuyerID, &updated.SellerID,
		&updated.Amount, &updated.PlatformFee, &updated.Status, &updated.StripePaymentID, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}
	return &updated, nil
}

func validateStatusTransition(from, to model.OrderStatus, callerID string, order *model.Order, isAdmin bool) error {
	isBuyer := order.BuyerID == callerID
	isSeller := order.SellerID == callerID

	switch from {
	case model.OrderPending:
		switch to {
		case model.OrderPaid:
			// only admin; the webhook handles this in the normal flow
			if !isAdmin {
				return fmt.Errorf("forbidden")
			}
		case model.OrderCancelled:
			if !isAdmin && !isBuyer && !isSeller {
				return fmt.Errorf("forbidden")
			}
		default:
			return fmt.Errorf("invalid status transition %s -> %s", from, to)
		}
	case model.OrderPaid:
		switch to {
		case model.OrderShipped:
			if !isAdmin && !isSeller {
				return fmt.Errorf("only seller can mark order as shipped")
			}
		case model.OrderCancelled:
			if !isAdmin && !isSeller {
				return fmt.Errorf("only seller can cancel after payment")
			}
		default:
			return fmt.Errorf("invalid status transition %s -> %s", from, to)
		}
	case model.OrderShipped:
		if to != model.OrderDelivered {
			return fmt.Errorf("invalid status transition %s -> %s", from, to)
		}
		if !isAdmin && !isBuyer && !isSeller {
			return fmt.Errorf("forbidden")
		}
	case model.OrderDelivered:
		if to != model.OrderRefunded {
			return fmt.Errorf("invalid status transition %s -> %s", from, to)
		}
		if !isAdmin {
			return fmt.Errorf("only admin can issue refunds")
		}
	default:
		return fmt.Errorf("cannot change status from %s", from)
	}
	return nil
}
