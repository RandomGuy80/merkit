package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"marketplace/internal/model"
)

const (
	UserIDKey   = "userID"
	UserRoleKey = "userRole"
)

func Auth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		if tokenType, _ := claims["type"].(string); tokenType != "access" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		userID, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)
		if userID == "" || role == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		c.Locals(UserIDKey, userID)
		c.Locals(UserRoleKey, role)
		return c.Next()
	}
}

func RequireRole(roles ...model.Role) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[string(r)] = true
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(UserRoleKey).(string)
		if !allowed[role] {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) string {
	id, _ := c.Locals(UserIDKey).(string)
	return id
}

func GetUserRole(c *fiber.Ctx) string {
	role, _ := c.Locals(UserRoleKey).(string)
	return role
}
