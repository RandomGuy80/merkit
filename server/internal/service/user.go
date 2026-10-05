package service

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"marketplace/internal/model"
)

type UserService struct {
	db        *pgxpool.Pool
	uploadDir string
}

func NewUserService(db *pgxpool.Pool, uploadDir string) *UserService {
	return &UserService{db: db, uploadDir: uploadDir}
}

func (s *UserService) GetMe(ctx context.Context, userID string) (*model.User, error) {
	return s.getByID(ctx, userID)
}

func (s *UserService) GetPublic(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow(ctx,
		`SELECT id, '' AS email, name, role, avatar, bio, rating, review_count, created_at
		 FROM users WHERE id=$1`, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.Avatar, &user.Bio, &user.Rating, &user.ReviewCount, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID string, req model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.getByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		user.Name = n
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}

	err = s.db.QueryRow(ctx,
		`UPDATE users SET name=$1, bio=$2
		 WHERE id=$3
		 RETURNING id, email, name, role, avatar, bio, rating, review_count, created_at`,
		user.Name, user.Bio, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.Avatar, &user.Bio, &user.Rating, &user.ReviewCount, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty file")
	}

	mime := detectMIME(data)
	ext, ok := allowedImageExt(mime)
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", mime)
	}

	filename := uuid.NewString() + ext
	avatarDir := filepath.Join(s.uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	path := filepath.Join(avatarDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	url := "/uploads/avatars/" + filename
	_, err := s.db.Exec(ctx, "UPDATE users SET avatar=$1 WHERE id=$2", url, userID)
	if err != nil {
		os.Remove(path)
		return "", fmt.Errorf("update avatar: %w", err)
	}
	return url, nil
}

func (s *UserService) getByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, name, role, avatar, bio, rating, review_count, created_at
		 FROM users WHERE id=$1`, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role,
		&user.Avatar, &user.Bio, &user.Rating, &user.ReviewCount, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func detectMIME(data []byte) string {
	buf := data
	if len(buf) > 512 {
		buf = data[:512]
	}
	return http.DetectContentType(buf)
}

func allowedImageExt(mimeType string) (string, bool) {
	exts := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	// DetectContentType may return params like "image/jpeg; charset=utf-8"
	mt, _, _ := mime.ParseMediaType(mimeType)
	ext, ok := exts[mt]
	return ext, ok
}
