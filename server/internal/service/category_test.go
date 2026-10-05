package service

import (
	"context"
	"testing"

	"marketplace/internal/db"
	"marketplace/internal/model"
)

func TestCategoryListAndCreate(t *testing.T) {
	pool := db.Connect("postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable")
	t.Cleanup(func() { pool.Close() })

	svc := NewCategoryService(pool)
	ctx := context.Background()

	// list — seed data must exist
	cats, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(cats) == 0 {
		t.Fatal("expected seed categories")
	}

	// create new
	slug := "test-cat-" + randomHex(4)
	cat, err := svc.Create(ctx, model.CreateCategoryRequest{
		Name: "Test Category",
		Slug: slug,
		Icon: "🧪",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if cat.ID == 0 {
		t.Fatal("expected non-zero id")
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM categories WHERE id=$1", cat.ID)
	})

	// duplicate slug
	_, err = svc.Create(ctx, model.CreateCategoryRequest{Name: "Dup", Slug: slug})
	if err == nil {
		t.Fatal("expected error on duplicate slug")
	}

	// delete
	if err := svc.Delete(ctx, cat.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// delete non-existent
	if err := svc.Delete(ctx, 999999); err == nil {
		t.Fatal("expected error deleting non-existent")
	}
}
