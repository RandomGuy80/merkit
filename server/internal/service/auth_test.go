package service

import (
	"context"
	"testing"

	"marketplace/internal/config"
	"marketplace/internal/db"
	rdb "marketplace/internal/redis"
	"marketplace/internal/model"
)

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()
	cfg := &config.Config{
		DBURL:         "postgres://marketplace:secret@localhost:5432/marketplace?sslmode=disable",
		RedisURL:      "redis://localhost:6379",
		JWTSecret:     "test-secret-key-min-32-chars-ok!!",
		JWTAccessTTL:  "15m",
		JWTRefreshTTL: "168h",
	}
	pool := db.Connect(cfg.DBURL)
	t.Cleanup(func() { pool.Close() })

	redisClient := rdb.Connect(cfg.RedisURL)
	t.Cleanup(func() {
		redisClient.FlushDB(context.Background())
		redisClient.Close()
	})

	store := rdb.NewStore(redisClient)
	return NewAuthService(pool, store, cfg)
}

func cleanupUser(t *testing.T, svc *AuthService, email string) {
	t.Helper()
	t.Cleanup(func() {
		svc.db.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
	})
}

func TestRegisterAndLogin(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	email := "test_auth_" + randomHex(4) + "@example.com"
	cleanupUser(t, svc, email)

	// register
	resp, err := svc.Register(ctx, model.RegisterRequest{
		Email:    email,
		Password: "password123",
		Name:     "Test User",
		Role:     model.RoleBuyer,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if resp.User.Email != email {
		t.Fatalf("want email %s, got %s", email, resp.User.Email)
	}

	// duplicate email
	_, err = svc.Register(ctx, model.RegisterRequest{
		Email:    email,
		Password: "password123",
		Name:     "Dup User",
		Role:     model.RoleBuyer,
	})
	if err == nil {
		t.Fatal("expected error on duplicate email")
	}

	// login
	loginResp, err := svc.Login(ctx, model.LoginRequest{
		Email:    email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginResp.AccessToken == "" {
		t.Fatal("expected access token on login")
	}

	// wrong password
	_, err = svc.Login(ctx, model.LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error on wrong password")
	}
}

func TestAccessTokenRejectedAsRefresh(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	email := "test_tokentype_" + randomHex(4) + "@example.com"
	cleanupUser(t, svc, email)

	resp, err := svc.Register(ctx, model.RegisterRequest{
		Email:    email,
		Password: "password123",
		Name:     "Token Type User",
		Role:     model.RoleBuyer,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = svc.Refresh(ctx, resp.AccessToken)
	if err == nil {
		t.Fatal("expected error: access token must not be accepted as refresh token")
	}
}

func TestRefreshAndLogout(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	email := "test_refresh_" + randomHex(4) + "@example.com"
	cleanupUser(t, svc, email)

	resp, err := svc.Register(ctx, model.RegisterRequest{
		Email:    email,
		Password: "password123",
		Name:     "Refresh User",
		Role:     model.RoleBuyer,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// refresh → new tokens
	refreshed, err := svc.Refresh(ctx, resp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.AccessToken == resp.AccessToken {
		t.Fatal("expected new access token after refresh")
	}

	// old refresh token should be invalid (rotated)
	_, err = svc.Refresh(ctx, resp.RefreshToken)
	if err == nil {
		t.Fatal("expected error on reused refresh token")
	}

	// logout invalidates new token
	if err := svc.Logout(ctx, refreshed.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	_, err = svc.Refresh(ctx, refreshed.RefreshToken)
	if err == nil {
		t.Fatal("expected error after logout")
	}
}
