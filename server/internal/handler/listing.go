package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"marketplace/internal/middleware"
	"marketplace/internal/model"
	"marketplace/internal/service"
)

type ListingHandler struct {
	svc       *service.ListingService
	jwtSecret string
}

func NewListingHandler(svc *service.ListingService, jwtSecret string) *ListingHandler {
	return &ListingHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *ListingHandler) Register(router fiber.Router) {
	auth := middleware.Auth(h.jwtSecret)

	listings := router.Group("/listings")
	listings.Get("/", h.search)
	listings.Get("/:id", h.getByID)
	listings.Post("/", auth, middleware.RequireRole(model.RoleSeller, model.RoleAdmin), h.create)
	listings.Put("/:id", auth, h.update)
	listings.Delete("/:id", auth, h.delete)
	listings.Post("/:id/images", auth, h.addImage)
	listings.Delete("/:id/images/:index", auth, h.removeImage)
}

func (h *ListingHandler) search(c *fiber.Ctx) error {
	f := model.ListingsFilter{
		Query:  c.Query("q"),
		Cursor: c.Query("cursor"),
		Limit:  c.QueryInt("limit", 20),
	}
	if catStr := c.Query("category_id"); catStr != "" {
		if id, err := strconv.Atoi(catStr); err == nil {
			f.CategoryID = &id
		}
	}
	if minStr := c.Query("min_price"); minStr != "" {
		if v, err := strconv.ParseFloat(minStr, 64); err == nil {
			f.MinPrice = &v
		}
	}
	if maxStr := c.Query("max_price"); maxStr != "" {
		if v, err := strconv.ParseFloat(maxStr, 64); err == nil {
			f.MaxPrice = &v
		}
	}
	if sid := c.Query("seller_id"); sid != "" {
		f.SellerID = sid
	}

	page, err := h.svc.Search(c.Context(), f)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "search failed")
	}
	return c.JSON(page)
}

func (h *ListingHandler) getByID(c *fiber.Ctx) error {
	l, err := h.svc.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "listing not found")
	}
	return c.JSON(l)
}

func (h *ListingHandler) create(c *fiber.Ctx) error {
	var req model.CreateListingRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	l, err := h.svc.Create(c.Context(), middleware.GetUserID(c), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(l)
}

func (h *ListingHandler) update(c *fiber.Ctx) error {
	var req model.UpdateListingRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	l, err := h.svc.Update(c.Context(), c.Params("id"), middleware.GetUserID(c), req)
	if err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(l)
}

func (h *ListingHandler) delete(c *fiber.Ctx) error {
	role := middleware.GetUserRole(c)
	err := h.svc.Delete(c.Context(), c.Params("id"), middleware.GetUserID(c), role == string(model.RoleAdmin))
	if err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ListingHandler) removeImage(c *fiber.Ctx) error {
	index, err := strconv.Atoi(c.Params("index"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid index")
	}
	if err := h.svc.RemoveImage(c.Context(), c.Params("id"), middleware.GetUserID(c), index); err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ListingHandler) addImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "image file required")
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

	url, err := h.svc.AddImage(c.Context(), c.Params("id"), middleware.GetUserID(c), data)
	if err != nil {
		if err.Error() == "forbidden" {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"image_url": url})
}
