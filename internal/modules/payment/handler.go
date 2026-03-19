package payment

import (
	"context"
	"fmt"
	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/views"
	"paymob-demo/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// Handler handles payment HTTP requests
type Handler struct {
	service   *Service
	repo      *Repository
	renderer  *views.Renderer
	cfg       *config.Config
}

// NewHandler creates a new payment handler
func NewHandler(service *Service, repo *Repository, renderer *views.Renderer, cfg *config.Config) *Handler {
	return &Handler{
		service:  service,
		repo:     repo,
		renderer: renderer,
		cfg:      cfg,
	}
}

// GetPaymentPage serves the payment page
func (h *Handler) GetPaymentPage(c *fiber.Ctx) error {
	html, err := h.renderer.RenderPaymentPage(domain.PaymentPageData{
		Title: "PayMob Demo - Make Payment",
	})
	if err != nil {
		return c.Status(500).SendString("Failed to render page")
	}
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// InitiatePayment initiates a new payment
func (h *Handler) InitiatePayment(c *fiber.Ctx) error {
	var req domain.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		html, _ := h.renderer.RenderPaymentResult(domain.PaymentResultData{
			Success: false,
			Message: "Invalid request. Please fill all required fields.",
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}

	if req.Amount <= 0 {
		html, _ := h.renderer.RenderPaymentResult(domain.PaymentResultData{
			Success: false,
			Message: "Amount must be greater than 0",
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}

	ctx := context.Background()
	payment, err := h.service.InitiatePayment(ctx, req)
	if err != nil {
		html, _ := h.renderer.RenderPaymentResult(domain.PaymentResultData{
			Success: false,
			Message: fmt.Sprintf("Failed: %s", err.Error()),
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}

	if err := h.repo.Add(ctx, payment); err != nil {
		html, _ := h.renderer.RenderPaymentResult(domain.PaymentResultData{
			Success: false,
			Message: fmt.Sprintf("Failed to save payment: %s", err.Error()),
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}

	html, _ := h.renderer.RenderPaymentResult(domain.PaymentResultData{
		Success:     true,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		CheckoutURL: payment.CheckoutURL,
		OrderID:     payment.OrderID,
	})
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// PaymentSuccess handles payment redirect - checks actual status before showing result
func (h *Handler) PaymentSuccess(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		orderID = c.Query("merchant_order_id")
	}

	transactionID := c.Query("id")
	if transactionID == "" {
		transactionID = c.Query("transaction_id")
	}

	// Check for authentication failure indicators from ACE emulator
	hasError := c.Query("error") != "" ||
		c.Query("error_occured") == "true" ||
		c.Query("success") == "false" ||
		c.Query("auth_result") == "failed" ||
		c.Query("3ds_status") == "failed" ||
		c.Query("acs_result") == "N"

	fmt.Printf("Payment redirect: order_id=%s, transaction_id=%s, has_error=%v\n", orderID, transactionID, hasError)

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		// Payment not found, show generic processing page
		html, _ := h.renderer.RenderSuccessPage(domain.ResultPageData{
			Title:   "Payment Processing",
			Message: "Your payment is being processed. Please check your dashboard for updates.",
			OrderID: orderID,
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}

	// Save transaction ID if provided
	if transactionID != "" {
		payment.TransactionID = transactionID
	}

	// If we detect auth failure indicators, mark as failed immediately
	if hasError && payment.Status == domain.PaymentStatusPending {
		payment.Status = domain.PaymentStatusFailed
		fmt.Printf("PaymentSuccess: Marking payment %s as FAILED due to auth error indicators\n", orderID)
	}

	h.repo.Update(ctx, payment)

	// Show result based on actual payment status
	switch payment.Status {
	case domain.PaymentStatusSuccess:
		html, _ := h.renderer.RenderSuccessPage(domain.ResultPageData{
			Title:   "Payment Successful",
			Message: "Your payment was completed successfully.",
			OrderID: orderID,
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	case domain.PaymentStatusFailed:
		html, _ := h.renderer.RenderFailurePage(domain.ResultPageData{
			Title:   "Payment Failed",
			Message: "Your payment could not be completed. Please try again.",
			OrderID: orderID,
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	default:
		// Still pending - show processing page with auto-refresh
		html, _ := h.renderer.RenderSuccessPage(domain.ResultPageData{
			Title:   "Payment Processing",
			Message: "Your payment is being processed. This page will update shortly...",
			OrderID: orderID,
		})
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}

// PaymentFailure handles payment failure redirect - same logic as success since URL is unreliable
func (h *Handler) PaymentFailure(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		orderID = c.Query("merchant_order_id")
	}

	// Use same logic as success handler - check actual status from database
	return h.PaymentSuccess(c)
}

// SimulatePaymentPage shows a demo payment simulation page
func (h *Handler) SimulatePaymentPage(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		return c.Status(400).SendString("Order ID required")
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return c.Status(404).SendString("Payment not found")
	}

	if payment.Status == domain.PaymentStatusSuccess {
		return c.Redirect("/success?order_id=" + orderID)
	}
	if payment.Status == domain.PaymentStatusFailed {
		return c.Redirect("/failure?order_id=" + orderID)
	}

	html, _ := h.renderer.RenderSimulatePage(domain.SimulatePageData{
		Amount:   payment.Amount,
		Currency: payment.Currency,
		OrderID:  orderID,
		Status:   "Pending",
	})
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// SimulatePaymentSuccess simulates a payment success
func (h *Handler) SimulatePaymentSuccess(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(400).SendString("Order ID required")
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return c.Status(404).SendString("Payment not found")
	}

	if payment.Status == domain.PaymentStatusSuccess || payment.Status == domain.PaymentStatusFailed {
		return c.Status(400).SendString("Payment already completed")
	}

	payment.Status = domain.PaymentStatusSuccess
	payment.TransactionID = "SIM_" + safeTruncate(orderID, 8)
	h.repo.Update(ctx, payment)

	html, err := h.renderer.RenderPaymentRow(domain.PaymentTableRow{
		OrderID:    payment.OrderID,
		Amount:     utils.FormatAmount(payment.Amount),
		Currency:   payment.Currency,
		Status:     string(payment.Status),
		StatusText: utils.StatusText(string(payment.Status)),
		CreatedAt:  payment.UpdatedAt.Format("2006-01-02 15:04"),
	})
	if err != nil {
		return c.Status(500).SendString("Failed to render row")
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// SimulatePaymentFailure simulates a payment failure
func (h *Handler) SimulatePaymentFailure(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(400).SendString("Order ID required")
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return c.Status(404).SendString("Payment not found")
	}

	if payment.Status == domain.PaymentStatusSuccess || payment.Status == domain.PaymentStatusFailed {
		return c.Status(400).SendString("Payment already completed")
	}

	payment.Status = domain.PaymentStatusFailed
	h.repo.Update(ctx, payment)

	return c.JSON(fiber.Map{"status": "failed", "order_id": orderID})
}

// HealthCheck returns health status
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "paymob-demo",
	})
}

// GetPaymentStatus returns the status of a payment by order ID
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

// QueryPayMobStatus queries PayMob directly for transaction status (fallback when webhook fails)
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

	// If we have a transaction ID, query PayMob directly
	if payment.TransactionID != "" && payment.TransactionID != "SIM_"+safeTruncate(orderID, 8) {
		status, err := h.service.QueryTransactionStatus(ctx, payment.TransactionID)
		if err == nil && status != nil {
			// Update payment status if it changed
			if payment.Status != *status {
				payment.Status = *status
				h.repo.Update(ctx, payment)
				fmt.Printf("QueryPayMobStatus: Updated payment %s to %s\n", orderID, *status)
			}
			return c.JSON(fiber.Map{
				"order_id":       payment.OrderID,
				"status":         payment.Status,
				"transaction_id": payment.TransactionID,
				"source":         "paymob_api",
			})
		}
	}

	// Return current status if we couldn't query PayMob
	return c.JSON(fiber.Map{
		"order_id":       payment.OrderID,
		"status":         payment.Status,
		"transaction_id": payment.TransactionID,
		"source":         "local_db",
	})
}

// Benchmark handles performance benchmarking
func (h *Handler) Benchmark(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message":        "Performance Benchmark",
		"framework":      "Fiber",
		"concurrent_ops": "High throughput capable",
		"note":           "Fiber uses fasthttp for high performance",
	})
}

// safeTruncate safely truncates a string to max length
func safeTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
