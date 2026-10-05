package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"marketplace/internal/config"
	"marketplace/internal/db"
	"marketplace/internal/handler"
	rdb "marketplace/internal/redis"
	"marketplace/internal/service"
	"marketplace/internal/ws"
)

func main() {
	cfg := config.Load()

	pool := db.Connect(cfg.DBURL)
	defer pool.Close()

	redisClient := rdb.Connect(cfg.RedisURL)
	defer redisClient.Close()

	if err := os.MkdirAll(filepath.Clean(cfg.UploadDir), 0755); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}

	maxMB, err := strconv.Atoi(cfg.MaxUploadMB)
	if err != nil || maxMB <= 0 {
		maxMB = 10
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
		BodyLimit:    maxMB * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} | ${method} ${path} | ${status} | ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization",
		AllowCredentials: true,
	}))

	app.Static("/uploads", cfg.UploadDir)

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	store := rdb.NewStore(redisClient)

	authSvc := service.NewAuthService(pool, store, cfg)
	handler.NewAuthHandler(authSvc, store, cfg.JWTSecret).Register(api)

	userSvc := service.NewUserService(pool, cfg.UploadDir)
	handler.NewUserHandler(userSvc, cfg.JWTSecret).Register(api)

	catSvc := service.NewCategoryService(pool)
	handler.NewCategoryHandler(catSvc, cfg.JWTSecret).Register(api)

	listingSvc := service.NewListingService(pool, cfg.UploadDir)
	handler.NewListingHandler(listingSvc, cfg.JWTSecret).Register(api)

	orderSvc := service.NewOrderService(pool)
	handler.NewOrderHandler(orderSvc, cfg.JWTSecret).Register(api)

	paymentSvc := service.NewPaymentService(pool, cfg.StripeSecret, cfg.StripeWebhook)
	handler.NewPaymentHandler(paymentSvc, cfg.JWTSecret, cfg.FrontendURL).Register(api)

	reviewSvc := service.NewReviewService(pool)
	handler.NewReviewHandler(reviewSvc, cfg.JWTSecret).Register(api)

	hub := ws.NewHub()
	handler.NewWSHandler(hub, cfg.JWTSecret).Register(api)

	adminSvc := service.NewAdminService(pool)
	handler.NewAdminHandler(adminSvc, cfg.JWTSecret).Register(api)

	log.Printf("server listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"

	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
		msg = fe.Message
	}

	return c.Status(code).JSON(fiber.Map{"error": msg})
}
