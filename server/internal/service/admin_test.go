package service

import (
	"context"
	"testing"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestAdminListUsersAndSetRole(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	svc := NewAdminService(pool)
	ctx := context.Background()

	// create test user
	email := "admin_test_" + randomHex(4) + "@example.com"
	var userID string
	pool.QueryRow(ctx, `INSERT INTO users (email,password,name,role) VALUES ($1,'h','U','buyer') RETURNING id`, email).Scan(&userID)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM users WHERE id=$1", userID) })

	// list users
	users, err := svc.ListUsers(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) == 0 {
		t.Fatal("expected at least one user")
	}

	// set role
	updated, err := svc.SetRole(ctx, userID, model.RoleSeller)
	if err != nil {
		t.Fatalf("set role: %v", err)
	}
	if updated.Role != model.RoleSeller {
		t.Fatalf("expected seller role, got %s", updated.Role)
	}

	// invalid role
	_, err = svc.SetRole(ctx, userID, "superuser")
	if err == nil {
		t.Fatal("expected error on invalid role")
	}

	// list orders
	orders, err := svc.ListOrders(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	_ = orders
}
