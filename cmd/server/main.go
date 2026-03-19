package main

import (
	"fmt"
	"log"
	"paymob-demo/internal/config"
	"paymob-demo/internal/modules/dashboard"
	"paymob-demo/internal/modules/payment"
	"paymob-demo/internal/modules/webhook"
	"paymob-demo/internal/views"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	renderer, err := views.NewRenderer()
	if err != nil {
		log.Fatalf("Failed to initialize renderer: %v", err)
	}

	repo, err := payment.NewRepository("payments.db")
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	paymentService := payment.NewService(cfg)

	paymentHandler := payment.NewHandler(paymentService, repo, renderer, cfg)
	dashboardHandler := dashboard.NewHandler(repo, renderer)
	webhookHandler := webhook.NewHandler(paymentService, repo)

	app := fiber.New(fiber.Config{
		AppName:      "PayMob Demo",
		ServerHeader: "Fiber",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Static("/", "./static")

	api := app.Group("/api")
	api.Get("/health", paymentHandler.HealthCheck)
	api.Get("/benchmark", paymentHandler.Benchmark)
	api.Post("/payments", paymentHandler.InitiatePayment)
	api.Get("/payments/status", paymentHandler.GetPaymentStatus)
	api.Get("/payments/paymob-status", paymentHandler.QueryPayMobStatus)
	api.Get("/dashboard", dashboardHandler.GetDashboardData)
	api.Get("/dashboard/html", dashboardHandler.GetDashboardHTML)
	api.Post("/simulate/:order_id", paymentHandler.SimulatePaymentSuccess)
	api.Post("/simulate-failure/:order_id", paymentHandler.SimulatePaymentFailure)
	api.Post("/webhook", webhookHandler.Webhook)

	app.Get("/pay/simulate", paymentHandler.SimulatePaymentPage)

	app.Get("/", paymentHandler.GetPaymentPage)
	app.Get("/success", paymentHandler.PaymentSuccess)
	app.Get("/failure", paymentHandler.PaymentFailure)
	app.Get("/dashboard", dashboardHandler.GetDashboard)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting PayMob Demo Server on %s", addr)

	if cfg.DemoMode {
		log.Printf("📋 MODE: DEMO (simulation only)")
	} else {
		log.Printf("💳 MODE: PRODUCTION (real PayMob payments)")
	}

	log.Printf("Dashboard available at http://localhost%s/dashboard", addr)
	log.Printf("Payment page available at http://localhost%s/", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}


