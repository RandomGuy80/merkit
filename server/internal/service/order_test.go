package service

import (
	"context"
	"testing"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestOrderCreateAndStatus(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	svc := NewOrderService(pool)
	ctx := context.Background()

	// fixtures
	buyerEmail := "buyer_" + randomHex(4) + "@example.com"
	sellerEmail := "seller_" + randomHex(4) + "@example.com"

	var buyerID, sellerID string
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','Buyer','buyer') RETURNING id`, buyerEmail).Scan(&buyerID)
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','Seller','seller') RETURNING id`, sellerEmail).Scan(&sellerID)

	var listingID string
	pool.QueryRow(ctx, `INSERT INTO listings (seller_id,title,description,price,currency) VALUES ($1,'MacBook','desc',999,'USD') RETURNING id`, sellerID).Scan(&listingID)

	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM orders WHERE listing_id=$1", listingID)
		pool.Exec(ctx, "DELETE FROM listings WHERE id=$1", listingID)
		pool.Exec(ctx, "DELETE FROM users WHERE id IN ($1,$2)", buyerID, sellerID)
	})

	// create order
	order, err := svc.Create(ctx, buyerID, model.CreateOrderRequest{ListingID: listingID})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.Status != model.OrderPending {
		t.Fatalf("expected pending, got %s", order.Status)
	}
	if order.PlatformFee != 999*platformFeeRate {
		t.Fatalf("wrong platform fee: %f", order.PlatformFee)
	}

	// cannot buy again (listing is sold)
	_, err = svc.Create(ctx, buyerID, model.CreateOrderRequest{ListingID: listingID})
	if err == nil {
		t.Fatal("expected error buying sold listing")
	}

	// seller cannot update to shipped without paying first
	_, err = svc.UpdateStatus(ctx, order.ID, sellerID, false, model.UpdateOrderStatusRequest{Status: model.OrderShipped})
	if err == nil {
		t.Fatal("expected error: cannot ship unpaid order")
	}

	// simulate payment: manually set to paid
	pool.Exec(ctx, "UPDATE orders SET status='paid' WHERE id=$1", order.ID)

	// seller ships
	updated, err := svc.UpdateStatus(ctx, order.ID, sellerID, false, model.UpdateOrderStatusRequest{Status: model.OrderShipped})
	if err != nil {
		t.Fatalf("ship order: %v", err)
	}
	if updated.Status != model.OrderShipped {
		t.Fatalf("expected shipped, got %s", updated.Status)
	}

	// buyer marks delivered
	delivered, err := svc.UpdateStatus(ctx, order.ID, buyerID, false, model.UpdateOrderStatusRequest{Status: model.OrderDelivered})
	if err != nil {
		t.Fatalf("deliver order: %v", err)
	}
	if delivered.Status != model.OrderDelivered {
		t.Fatalf("expected delivered, got %s", delivered.Status)
	}
}
