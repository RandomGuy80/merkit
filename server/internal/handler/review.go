package handler

import (
	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	"marketplace/internal/service"
)

type ReviewHandler struct {
	svc       *service.ReviewService
	jwtSecret string
}

func NewReviewHandler(svc *service.ReviewService, jwtSecret string) *ReviewHandler {
	return &ReviewHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *ReviewHandler) Register(router fiber.Router) {
	auth := middleware.Auth(h.jwtSecret)

	router.Post("/reviews", auth, middleware.RequireRole(model.RoleBuyer, model.RoleAdmin), h.create)
	router.Get("/users/:id/reviews", h.listForSeller)
}

func (h *ReviewHandler) create(c *fiber.Ctx) error {
	var req model.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	review, err := h.svc.Create(c.Context(), middleware.GetUserID(c), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(review)
}

func (h *ReviewHandler) listForSeller(c *fiber.Ctx) error {
	reviews, err := h.svc.ListForSeller(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch reviews")
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	return c.JSON(reviews)
}
