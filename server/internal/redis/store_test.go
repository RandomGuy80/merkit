package redis

import (
	"context"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	c := Connect("redis://localhost:6379")
	t.Cleanup(func() { c.FlushDB(context.Background()) })
	return NewStore(c)
}

func TestRefreshToken(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.SetRefreshToken(ctx, "tok1", "user-123", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}

	uid, err := s.GetRefreshToken(ctx, "tok1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if uid != "user-123" {
		t.Fatalf("want user-123, got %s", uid)
	}

	if err := s.DeleteRefreshToken(ctx, "tok1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := s.GetRefreshToken(ctx, "tok1"); err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestRateLimit(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	key := "ratelimit:test:1.2.3.4"

	for i := 0; i < 3; i++ {
		ok, _, err := s.RateLimit(ctx, key, 3, time.Minute)
		if err != nil {
			t.Fatalf("rate limit error: %v", err)
		}
		if !ok {
			t.Fatalf("expected allowed on attempt %d", i+1)
		}
	}

	ok, remaining, err := s.RateLimit(ctx, key, 3, time.Minute)
	if err != nil {
		t.Fatalf("rate limit error: %v", err)
	}
	if ok {
		t.Fatal("expected blocked on 4th attempt")
	}
	if remaining != 0 {
		t.Fatalf("want remaining=0, got %d", remaining)
	}
}
