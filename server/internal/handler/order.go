package handler

import (
	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	"marketplace/internal/service"
)

type OrderHandler struct {
	svc       *service.OrderService
	jwtSecret string
}

func NewOrderHandler(svc *service.OrderService, jwtSecret string) *OrderHandler {
	return &OrderHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *OrderHandler) Register(router fiber.Router) {
	auth := middleware.Auth(h.jwtSecret)

	orders := router.Group("/orders", auth)
	orders.Post("/", h.create)
	orders.Get("/", h.list)
	orders.Get("/:id", h.getByID)
	orders.Patch("/:id/status", h.updateStatus)
}

func (h *OrderHandler) create(c *fiber.Ctx) error {
	var req model.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.ListingID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "listing_id required")
	}
	order, err := h.svc.Create(c.Context(), middleware.GetUserID(c), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) list(c *fiber.Ctx) error {
	orders, err := h.svc.ListForUser(c.Context(), middleware.GetUserID(c))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch orders")
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return c.JSON(orders)
}

func (h *OrderHandler) getByID(c *fiber.Ctx) error {
	role := middleware.GetUserRole(c)
	order, err := h.svc.GetByID(c.Context(), c.Params("id"), middleware.GetUserID(c), role == string(model.RoleAdmin))
	if err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	}
	return c.JSON(order)
}

func (h *OrderHandler) updateStatus(c *fiber.Ctx) error {
	var req model.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	role := middleware.GetUserRole(c)
	order, err := h.svc.UpdateStatus(c.Context(), c.Params("id"), middleware.GetUserID(c), role == string(model.RoleAdmin), req)
	if err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(order)
}
