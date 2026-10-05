package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"marketplace/internal/config"
	"marketplace/internal/model"
	rstore "marketplace/internal/redis"
)

type AuthService struct {
	db    *pgxpool.Pool
	store *rstore.Store
	cfg   *config.Config
}

func NewAuthService(db *pgxpool.Pool, store *rstore.Store, cfg *config.Config) *AuthService {
	return &AuthService{db: db, store: store, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	if req.Role != model.RoleBuyer && req.Role != model.RoleSeller {
		req.Role = model.RoleBuyer
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var user model.User
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (email, password, name, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, name, role, avatar, bio, rating, review_count, created_at`,
		req.Email, string(hash), req.Name, req.Role,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.Avatar, &user.Bio, &user.Rating, &user.ReviewCount, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return s.issueTokens(ctx, &user)
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	var user model.User
	var hash string
	err := s.db.QueryRow(ctx,
		`SELECT id, email, name, role, avatar, bio, rating, review_count, created_at, password
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.Avatar, &user.Bio, &user.Rating, &user.ReviewCount, &user.CreatedAt, &hash)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return s.issueTokens(ctx, &user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	if tokenType, _ := claims["type"].(string); tokenType != "refresh" {
		return nil, fmt.Errorf("invalid token")
	}

	tokenID, _ := claims["jti"].(string)
	userID, _ := claims["sub"].(string)

	storedUID, err := s.store.GetRefreshToken(ctx, tokenID)
	if err != nil || storedUID != userID {
		return nil, fmt.Errorf("invalid token")
	}

	// rotate: delete old token
	_ = s.store.DeleteRefreshToken(ctx, tokenID)

	var user model.User
	err = s.db.QueryRow(ctx,
		`SELECT id, email, name, role, avatar, bio, rating, review_count, created_at
		 FROM users WHERE id = $1`, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.Avatar, &user.Bio, &user.Rating, &user.ReviewCount, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.issueTokens(ctx, &user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil
	}
	tokenID, _ := claims["jti"].(string)
	return s.store.DeleteRefreshToken(ctx, tokenID)
}

func (s *AuthService) issueTokens(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	accessTTL, err := time.ParseDuration(s.cfg.JWTAccessTTL)
	if err != nil {
		accessTTL = 15 * time.Minute
	}
	refreshTTL, err := time.ParseDuration(s.cfg.JWTRefreshTTL)
	if err != nil {
		refreshTTL = 168 * time.Hour
	}

	accessToken, err := s.signToken(user.ID, string(user.Role), "access", uuid.NewString(), accessTTL)
	if err != nil {
		return nil, err
	}

	refreshID := uuid.NewString()
	refreshToken, err := s.signToken(user.ID, string(user.Role), "refresh", refreshID, refreshTTL)
	if err != nil {
		return nil, err
	}

	if err := s.store.SetRefreshToken(ctx, refreshID, user.ID, refreshTTL); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) signToken(userID, role, tokenType, jti string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"type": tokenType,
		"jti":  jti,
		"exp":  time.Now().Add(ttl).Unix(),
		"iat":  time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
