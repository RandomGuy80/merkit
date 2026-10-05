package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	"marketplace/internal/service"
)

type CategoryHandler struct {
	svc       *service.CategoryService
	jwtSecret string
}

func NewCategoryHandler(svc *service.CategoryService, jwtSecret string) *CategoryHandler {
	return &CategoryHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *CategoryHandler) Register(router fiber.Router) {
	router.Get("/categories", h.list)

	admin := router.Group("/admin",
		middleware.Auth(h.jwtSecret),
		middleware.RequireRole(model.RoleAdmin),
	)
	admin.Post("/categories", h.create)
	admin.Delete("/categories/:id", h.delete)
}

func (h *CategoryHandler) list(c *fiber.Ctx) error {
	cats, err := h.svc.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch categories")
	}
	return c.JSON(cats)
}

func (h *CategoryHandler) create(c *fiber.Ctx) error {
	var req model.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	cat, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(cat)
}

func (h *CategoryHandler) delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
