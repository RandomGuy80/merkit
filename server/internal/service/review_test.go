package service

import (
	"context"
	"testing"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestReviewCreate(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	svc := NewReviewService(pool)
	ctx := context.Background()

	// fixtures
	buyerEmail := "rev_buyer_" + randomHex(4) + "@example.com"
	sellerEmail := "rev_seller_" + randomHex(4) + "@example.com"
	var buyerID, sellerID, listingID, orderID string
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','Buyer','buyer') RETURNING id`, buyerEmail).Scan(&buyerID)
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','Seller','seller') RETURNING id`, sellerEmail).Scan(&sellerID)
	pool.QueryRow(ctx, `INSERT INTO listings (seller_id,title,description,price,currency,status) VALUES ($1,'Laptop','d',500,'USD','sold') RETURNING id`, sellerID).Scan(&listingID)
	pool.QueryRow(ctx, `INSERT INTO orders (listing_id,buyer_id,seller_id,amount,platform_fee,status) VALUES ($1,$2,$3,500,25,'delivered') RETURNING id`,
		listingID, buyerID, sellerID).Scan(&orderID)

	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM reviews WHERE order_id=$1", orderID)
		pool.Exec(ctx, "DELETE FROM orders WHERE id=$1", orderID)
		pool.Exec(ctx, "DELETE FROM listings WHERE id=$1", listingID)
		pool.Exec(ctx, "DELETE FROM users WHERE id IN ($1,$2)", buyerID, sellerID)
	})

	// valid review
	review, err := svc.Create(ctx, buyerID, model.CreateReviewRequest{
		OrderID: orderID,
		Rating:  5,
		Body:    "Great seller!",
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	if review.Rating != 5 {
		t.Fatalf("expected rating 5, got %d", review.Rating)
	}

	// duplicate review
	_, err = svc.Create(ctx, buyerID, model.CreateReviewRequest{OrderID: orderID, Rating: 4})
	if err == nil {
		t.Fatal("expected error on duplicate review")
	}

	// invalid rating
	_, err = svc.Create(ctx, buyerID, model.CreateReviewRequest{OrderID: orderID, Rating: 6})
	if err == nil {
		t.Fatal("expected error on invalid rating")
	}

	// check seller avg rating updated
	var rating float64
	var count int
	pool.QueryRow(ctx, "SELECT rating, review_count FROM users WHERE id=$1", sellerID).Scan(&rating, &count)
	if count != 1 {
		t.Fatalf("expected review_count=1, got %d", count)
	}
	if rating != 5 {
		t.Fatalf("expected avg rating=5, got %f", rating)
	}

	// list reviews for seller
	reviews, err := svc.ListForSeller(ctx, sellerID)
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
}
