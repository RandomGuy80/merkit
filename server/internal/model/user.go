package model

import "time"

type Role string

const (
	RoleBuyer  Role = "buyer"
	RoleSeller Role = "seller"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           Role      `json:"role"`
	Avatar         *string   `json:"avatar,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	Rating         float64   `json:"rating"`
	ReviewCount    int       `json:"review_count"`
	StripeAccountID *string  `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     Role   `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type UpdateProfileRequest struct {
	Name *string `json:"name"`
	Bio  *string `json:"bio"`
}
