package webhook

import (
	"context"
	"encoding/json"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/modules/payment"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handler handles webhook HTTP requests
type Handler struct {
	service *payment.Service
	repo    *payment.Repository
}

// NewHandler creates a new webhook handler
func NewHandler(service *payment.Service, repo *payment.Repository) *Handler {
	return &Handler{
		service: service,
		repo:    repo,
	}
}

// Webhook handles PayMob webhook callbacks
func (h *Handler) Webhook(c *fiber.Ctx) error {
	// Get HMAC signature from header
	signature := c.Get("X-Paymob-Signature")
	if signature == "" {
		signature = c.Get("PAYMOB_SIGNATURE")
	}

	body := c.Body()

	// Verify webhook signature
	if !h.service.VerifyWebhookSignature(signature, body) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid webhook signature",
		})
	}

	// Parse the new PayMob webhook format
	var payload domain.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook payload"})
	}

	// Extract merchant order ID
	merchantOrderID := payload.Obj.Order.MerchantOrderID
	if merchantOrderID == "" {
		// Try legacy format
		var legacyPayload domain.LegacyWebhookPayload
		if err := json.Unmarshal(body, &legacyPayload); err == nil {
			merchantOrderID = legacyPayload.MerchantOrderID
		}
	}

	ctx := context.Background()
	payment, err := h.repo.GetByOrderID(ctx, merchantOrderID)
	if err != nil || payment == nil {
		return c.JSON(fiber.Map{"status": "ignored", "reason": "payment not found"})
	}

	// Update payment status
	if payload.Obj.Success {
		payment.Status = domain.PaymentStatusSuccess
		payment.TransactionID = strconv.Itoa(payload.Obj.ID)
	} else if payload.Obj.Pending {
		payment.Status = domain.PaymentStatusPending
	} else {
		payment.Status = domain.PaymentStatusFailed
	}

	h.repo.Update(ctx, payment)
	return c.JSON(fiber.Map{"status": "received", "payment_id": payment.ID})
}
