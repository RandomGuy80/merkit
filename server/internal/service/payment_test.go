package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestWebhookInvalidSignature(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	svc := NewPaymentService(pool, "sk_test_dummy", "whsec_test")
	ctx := context.Background()

	err := svc.HandleWebhook(ctx, []byte(`{"type":"checkout.session.completed"}`), "invalid-sig")
	if err == nil {
		t.Fatal("expected signature verification error")
	}
}

func TestWebhookValidSignature(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	secret := "whsec_testsecret1234567890abcdef"
	svc := NewPaymentService(pool, "sk_test_dummy", secret)
	ctx := context.Background()

	// create a test order with a stripe session id
	sellerEmail := "pay_seller_" + randomHex(4) + "@example.com"
	buyerEmail := "pay_buyer_" + randomHex(4) + "@example.com"
	var sellerID, buyerID, listingID, orderID string
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','S','seller') RETURNING id`, sellerEmail).Scan(&sellerID)
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','B','buyer') RETURNING id`, buyerEmail).Scan(&buyerID)
	pool.QueryRow(ctx, `INSERT INTO listings (seller_id,title,description,price,currency) VALUES ($1,'Item','d',100,'USD') RETURNING id`, sellerID).Scan(&listingID)
	pool.QueryRow(ctx, `INSERT INTO orders (listing_id,buyer_id,seller_id,amount,platform_fee,stripe_payment_id) VALUES ($1,$2,$3,100,5,'cs_test_session123') RETURNING id`,
		listingID, buyerID, sellerID).Scan(&orderID)

	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM orders WHERE id=$1", orderID)
		pool.Exec(ctx, "DELETE FROM listings WHERE id=$1", listingID)
		pool.Exec(ctx, "DELETE FROM users WHERE id IN ($1,$2)", sellerID, buyerID)
	})

	// build valid stripe webhook payload + signature
	payload := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_session123"}}}`
	ts := time.Now().Unix()
	signed := fmt.Sprintf("%d.%s", ts, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	sig := fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))

	if err := svc.HandleWebhook(ctx, []byte(payload), sig); err != nil {
		t.Fatalf("valid webhook failed: %v", err)
	}

	// verify order status changed to paid
	var status model.OrderStatus
	pool.QueryRow(ctx, "SELECT status FROM orders WHERE id=$1", orderID).Scan(&status)
	if status != model.OrderPaid {
		t.Fatalf("expected paid, got %s", status)
	}
}
