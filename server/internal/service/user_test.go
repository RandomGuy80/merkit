package service

import (
	"context"
	"testing"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestGetAndUpdateUser(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	userSvc := NewUserService(pool, t.TempDir())
	ctx := context.Background()

	// create a user first via direct insert
	email := "test_user_" + randomHex(4) + "@example.com"
	var userID string
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password, name, role) VALUES ($1, 'hash', 'Test', 'buyer') RETURNING id`,
		email,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM users WHERE id=$1", userID)
	})

	// get me
	user, err := userSvc.GetMe(ctx, userID)
	if err != nil {
		t.Fatalf("get me: %v", err)
	}
	if user.Email != email {
		t.Fatalf("want email %s got %s", email, user.Email)
	}

	// update name + bio
	bio := "hello world"
	name := "Updated Name"
	updated, err := userSvc.UpdateMe(ctx, userID, model.UpdateProfileRequest{
		Name: &name,
		Bio:  &bio,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != name {
		t.Fatalf("want name %s got %s", name, updated.Name)
	}
	if updated.Bio == nil || *updated.Bio != bio {
		t.Fatal("bio not updated")
	}

	// get public
	pub, err := userSvc.GetPublic(ctx, userID)
	if err != nil {
		t.Fatalf("get public: %v", err)
	}
	if pub.ID != userID {
		t.Fatal("public profile id mismatch")
	}
}
