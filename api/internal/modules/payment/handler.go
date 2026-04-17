//go:build !web

package payment

import (
	"context"
	"fmt"
	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for payment operations.
// In API-only mode (no web build tag), all responses are JSON.
type Handler struct {
	service *Service
	repo    *Repository
	cfg     *config.Config
}

// NewHandler creates a new payment handler.
// In API-only mode, no renderer is needed.
func NewHandler(service *Service, repo *Repository, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		repo:    repo,
		cfg:     cfg,
	}
}

func (h *Handler) GetPaymentPage(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "web frontend not enabled (build with -tags web)",
	})
}

func (h *Handler) InitiatePayment(c *fiber.Ctx) error {
	var req domain.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request. Please fill all required fields.",
		})
	}

	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Amount must be greater than 0",
		})
	}

	ctx := context.Background()
	payment, err := h.service.InitiatePayment(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed: %s", err.Error()),
		})
	}

	if err := h.repo.Add(ctx, payment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save payment: %s", err.Error()),
		})
	}

	return c.JSON(domain.PaymentResponse{
		Success:     true,
		Message:     "Payment initiated",
		CheckoutURL: payment.CheckoutURL,
		OrderID:     payment.OrderID,
	})
}

func (h *Handler) PaymentSuccess(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		orderID = c.Query("merchant_order_id")
	}

	transactionID := c.Query("id")
	if transactionID == "" {
		transactionID = c.Query("transaction_id")
	}

	hasError := c.Query("error") != "" ||
		c.Query("error_occured") == "true" ||
		c.Query("success") == "false" ||
		c.Query("auth_result") == "failed" ||
		c.Query("3ds_status") == "failed" ||
		c.Query("acs_result") == "N"

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return c.JSON(fiber.Map{
			"status":   "not_found",
			"message":  "Payment not found",
			"order_id": orderID,
		})
	}

	if transactionID != "" {
		payment.TransactionID = transactionID
	}

	if hasError && payment.Status == domain.PaymentStatusPending {
		payment.Status = domain.PaymentStatusFailed
	} else if !hasError && payment.Status == domain.PaymentStatusPending {
		payment.Status = domain.PaymentStatusSuccess
	}

	h.repo.Update(ctx, payment)

	return c.JSON(fiber.Map{
		"order_id":       payment.OrderID,
		"status":         payment.Status,
		"transaction_id": payment.TransactionID,
	})
}

func (h *Handler) PaymentFailure(c *fiber.Ctx) error {
	return h.PaymentSuccess(c)
}

func (h *Handler) SimulatePaymentPage(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "web frontend not enabled (build with -tags web)",
	})
}

func (h *Handler) SimulatePaymentSuccess(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order ID required"})
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}

	if payment.Status == domain.PaymentStatusSuccess || payment.Status == domain.PaymentStatusFailed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payment already completed"})
	}

	payment.Status = domain.PaymentStatusSuccess
	payment.TransactionID = "SIM_" + safeTruncate(orderID, 8)
	h.repo.Update(ctx, payment)

	return c.JSON(fiber.Map{
		"status":     "success",
		"order_id":   orderID,
		"updated_at": payment.UpdatedAt,
	})
}

func (h *Handler) SimulatePaymentFailure(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order ID required"})
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}

	if payment.Status == domain.PaymentStatusSuccess || payment.Status == domain.PaymentStatusFailed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payment already completed"})
	}

	payment.Status = domain.PaymentStatusFailed
	h.repo.Update(ctx, payment)

	return c.JSON(fiber.Map{"status": "failed", "order_id": orderID})
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "paymob-demo",
	})
}

func (h *Handler) GetPaymentStatus(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id is required",
		})
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "payment not found",
		})
	}

	return c.JSON(fiber.Map{
		"order_id":       payment.OrderID,
		"status":         payment.Status,
		"transaction_id": payment.TransactionID,
		"amount":         payment.Amount,
		"currency":       payment.Currency,
	})
}

func (h *Handler) QueryPayMobStatus(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order_id is required",
		})
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "payment not found",
		})
	}

	if payment.TransactionID != "" && payment.TransactionID != "SIM_"+safeTruncate(orderID, 8) {
		status, err := h.service.QueryTransactionStatus(ctx, payment.TransactionID)
		if err == nil && status != nil {
			if payment.Status != *status {
				payment.Status = *status
				h.repo.Update(ctx, payment)
			}
			return c.JSON(fiber.Map{
				"order_id":       payment.OrderID,
				"status":         payment.Status,
				"transaction_id": payment.TransactionID,
				"source":         "paymob_api",
			})
		}
	}

	return c.JSON(fiber.Map{
		"order_id":       payment.OrderID,
		"status":         payment.Status,
		"transaction_id": payment.TransactionID,
		"source":         "local_db",
	})
}

func (h *Handler) Benchmark(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message":        "Performance Benchmark",
		"framework":      "Fiber",
		"concurrent_ops": "High throughput capable",
		"note":           "Fiber uses fasthttp for high performance",
	})
}

func safeTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
