package service

import (
	"context"
	"testing"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestListingCRUD(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	svc := NewListingService(pool, t.TempDir())
	ctx := context.Background()

	// create a test seller
	email := "seller_" + randomHex(4) + "@example.com"
	var sellerID string
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password, name, role) VALUES ($1,'hash','Seller','seller') RETURNING id`, email,
	).Scan(&sellerID)
	if err != nil {
		t.Fatalf("insert seller: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM listings WHERE seller_id=$1", sellerID)
		pool.Exec(ctx, "DELETE FROM users WHERE id=$1", sellerID)
	})

	// create listing
	l, err := svc.Create(ctx, sellerID, model.CreateListingRequest{
		Title:       "Test iPhone",
		Description: "Great condition",
		Price:       499.99,
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if l.Title != "Test iPhone" {
		t.Fatalf("title mismatch: %s", l.Title)
	}

	// get by id (increments views)
	got, err := svc.GetByID(ctx, l.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Views != 1 {
		t.Fatalf("expected views=1, got %d", got.Views)
	}

	// search
	page, err := svc.Search(ctx, model.ListingsFilter{Query: "iPhone", Limit: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ID == l.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("listing not found in search results")
	}

	// update
	newPrice := 399.99
	updated, err := svc.Update(ctx, l.ID, sellerID, model.UpdateListingRequest{Price: &newPrice})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Price != newPrice {
		t.Fatalf("price not updated: %f", updated.Price)
	}

	// forbidden update by another user
	_, err = svc.Update(ctx, l.ID, "other-user-id", model.UpdateListingRequest{Price: &newPrice})
	if err == nil {
		t.Fatal("expected forbidden error")
	}

	// delete
	if err := svc.Delete(ctx, l.ID, sellerID, false); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
