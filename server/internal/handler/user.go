package handler

import (
	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	"marketplace/internal/service"
)

type UserHandler struct {
	svc       *service.UserService
	jwtSecret string
}

func NewUserHandler(svc *service.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *UserHandler) Register(router fiber.Router) {
	auth := middleware.Auth(h.jwtSecret)

	users := router.Group("/users")
	users.Get("/me", auth, h.getMe)
	users.Put("/me", auth, h.updateMe)
	users.Post("/me/avatar", auth, h.uploadAvatar)
	users.Get("/:id", h.getPublic)
}

func (h *UserHandler) getMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetMe(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	return c.JSON(user)
}

func (h *UserHandler) updateMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req model.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	user, err := h.svc.UpdateMe(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(user)
}

func (h *UserHandler) uploadAvatar(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	file, err := c.FormFile("avatar")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "avatar file required")
	}

	f, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cannot read file")
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cannot read file")
	}

	url, err := h.svc.UploadAvatar(c.Context(), userID, data)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"avatar_url": url})
}

func (h *UserHandler) getPublic(c *fiber.Ctx) error {
	userID := c.Params("id")
	user, err := h.svc.GetPublic(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	return c.JSON(user)
}
