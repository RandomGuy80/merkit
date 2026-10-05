package handler

import (
	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/service"
)

type PaymentHandler struct {
	svc        *service.PaymentService
	jwtSecret  string
	successURL string
	cancelURL  string
}

func NewPaymentHandler(svc *service.PaymentService, jwtSecret, baseURL string) *PaymentHandler {
	return &PaymentHandler{
		svc:        svc,
		jwtSecret:  jwtSecret,
		successURL: baseURL + "/orders",
		cancelURL:  baseURL + "/orders",
	}
}

func (h *PaymentHandler) Register(router fiber.Router) {
	auth := middleware.Auth(h.jwtSecret)

	payments := router.Group("/payments")
	payments.Post("/checkout/:order_id", auth, h.checkout)
	payments.Post("/webhook", h.webhook)
}

func (h *PaymentHandler) checkout(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	url, err := h.svc.CreateCheckoutSession(
		c.Context(), orderID, middleware.GetUserID(c),
		h.successURL, h.cancelURL,
	)
	if err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"checkout_url": url})
}

func (h *PaymentHandler) webhook(c *fiber.Ctx) error {
	sig := c.Get("Stripe-Signature")
	if err := h.svc.HandleWebhook(c.Context(), c.Body(), sig); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusOK)
}
