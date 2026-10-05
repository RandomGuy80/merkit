package handler

import (
	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	"marketplace/internal/service"
)

type AdminHandler struct {
	svc       *service.AdminService
	jwtSecret string
}

func NewAdminHandler(svc *service.AdminService, jwtSecret string) *AdminHandler {
	return &AdminHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *AdminHandler) Register(router fiber.Router) {
	admin := router.Group("/admin",
		middleware.Auth(h.jwtSecret),
		middleware.RequireRole(model.RoleAdmin),
	)
	admin.Get("/users", h.listUsers)
	admin.Patch("/users/:id/role", h.setRole)
	admin.Delete("/listings/:id", h.deleteListing)
	admin.Get("/orders", h.listOrders)
}

func (h *AdminHandler) listUsers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	users, err := h.svc.ListUsers(c.Context(), limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch users")
	}
	if users == nil {
		users = []model.User{}
	}
	return c.JSON(users)
}

func (h *AdminHandler) setRole(c *fiber.Ctx) error {
	var body struct {
		Role model.Role `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	user, err := h.svc.SetRole(c.Context(), c.Params("id"), body.Role)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(user)
}

func (h *AdminHandler) deleteListing(c *fiber.Ctx) error {
	if err := h.svc.DeleteListing(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) listOrders(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	orders, err := h.svc.ListOrders(c.Context(), limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch orders")
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return c.JSON(orders)
}
