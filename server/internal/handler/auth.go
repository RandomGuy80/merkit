package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	rdb "marketplace/internal/redis"
	"marketplace/internal/service"
)

type AuthHandler struct {
	svc       *service.AuthService
	store     *rdb.Store
	jwtSecret string
}

func NewAuthHandler(svc *service.AuthService, store *rdb.Store, jwtSecret string) *AuthHandler {
	return &AuthHandler{svc: svc, store: store, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Register(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/register", h.register)
	auth.Post("/login", h.login)
	auth.Post("/refresh", h.refresh)
	auth.Post("/logout", middleware.Auth(h.jwtSecret), h.logout)
}

func (h *AuthHandler) register(c *fiber.Ctx) error {
	ip := c.IP()
	if ok, _, err := h.store.RateLimit(c.Context(), "ratelimit:register:"+ip, 5, 10*time.Minute); err == nil && !ok {
		return fiber.NewError(fiber.StatusTooManyRequests, "too many requests")
	}

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := validateRegister(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.Register(c.Context(), req)
	if err != nil {
		if isDuplicateEmail(err) {
			return fiber.NewError(fiber.StatusConflict, "email already in use")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "registration failed")
	}

	setRefreshCookie(c, resp.RefreshToken)
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *AuthHandler) login(c *fiber.Ctx) error {
	ip := c.IP()
	if ok, _, err := h.store.RateLimit(c.Context(), "ratelimit:login:"+ip, 10, time.Minute); err == nil && !ok {
		return fiber.NewError(fiber.StatusTooManyRequests, "too many requests")
	}

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password required")
	}

	resp, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	setRefreshCookie(c, resp.RefreshToken)
	return c.JSON(resp)
}

func (h *AuthHandler) refresh(c *fiber.Ctx) error {
	token := c.Cookies("refresh_token")
	if token == "" {
		// also accept from body
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		token = body.RefreshToken
	}
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "refresh token required")
	}

	resp, err := h.svc.Refresh(c.Context(), token)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
	}

	setRefreshCookie(c, resp.RefreshToken)
	return c.JSON(resp)
}

func (h *AuthHandler) logout(c *fiber.Ctx) error {
	token := c.Cookies("refresh_token")
	if token == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		token = body.RefreshToken
	}
	_ = h.svc.Logout(c.Context(), token)
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", MaxAge: -1})
	return c.JSON(fiber.Map{"message": "logged out"})
}

func setRefreshCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   false, // set true in production
		SameSite: "Strict",
		MaxAge:   7 * 24 * 60 * 60,
		Path:     "/api/v1/auth",
	})
}

func validateRegister(req model.RegisterRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(req.Password) > 72 {
		return fmt.Errorf("password too long")
	}
	if req.Name == "" {
		return fmt.Errorf("name required")
	}
	return nil
}

func isDuplicateEmail(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique")
}
