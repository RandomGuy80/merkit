package handler

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	ws "marketplace/internal/ws"
)

type WSHandler struct {
	hub       *ws.Hub
	jwtSecret string
}

func NewWSHandler(hub *ws.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret}
}

func (h *WSHandler) Register(router fiber.Router) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	router.Get("/ws", middleware.Auth(h.jwtSecret), websocket.New(h.handle))
}

func (h *WSHandler) handle(conn *websocket.Conn) {
	userID, _ := conn.Locals(middleware.UserIDKey).(string)
	c := h.hub.Register(userID, conn)
	h.hub.RunClient(c)
}
