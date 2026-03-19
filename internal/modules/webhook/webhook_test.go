package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/modules/payment"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Webhook(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add a test payment
	repo.Add(ctx, &domain.Payment{
		ID:        "test-1",
		OrderID:   "webhook-order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	cfg := &config.Config{}
	paymentService := payment.NewService(cfg)
	handler := NewHandler(paymentService, repo)

	app.Post("/api/webhook", handler.Webhook)

	t.Run("Success", func(t *testing.T) {
		payload := domain.WebhookPayload{
			Type: "TRANSACTION",
			Obj: domain.WebhookObject{
				ID: 12345,
				Order: domain.WebhookOrder{
					ID:              100,
					MerchantOrderID: "webhook-order-1",
				},
				Success:     true,
				Pending:     false,
				AmountCents: 100000,
				Currency:    "EGP",
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Verify payment was updated
		updated, _ := repo.GetByOrderID(ctx, "webhook-order-1")
		require.NotNil(t, updated)
		assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
	})

	t.Run("PaymentNotFound", func(t *testing.T) {
		payload := domain.WebhookPayload{
			Type: "TRANSACTION",
			Obj: domain.WebhookObject{
				ID: 12346,
				Order: domain.WebhookOrder{
					ID:              101,
					MerchantOrderID: "nonexistent-order",
				},
				Success: true,
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Failure", func(t *testing.T) {
		// Add another test payment
		repo.Add(ctx, &domain.Payment{
			ID:        "test-2",
			OrderID:   "webhook-order-2",
			Amount:    500,
			Currency:  "EGP",
			Status:    domain.PaymentStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		payload := domain.WebhookPayload{
			Type: "TRANSACTION",
			Obj: domain.WebhookObject{
				ID: 12347,
				Order: domain.WebhookOrder{
					ID:              102,
					MerchantOrderID: "webhook-order-2",
				},
				Success:     false,
				Pending:     false,
				AmountCents: 50000,
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Verify payment was updated to failed
		updated, _ := repo.GetByOrderID(ctx, "webhook-order-2")
		require.NotNil(t, updated)
		assert.Equal(t, domain.PaymentStatusFailed, updated.Status)
	})

	t.Run("InvalidPayload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/webhook", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestHandler_Webhook_WithSignature(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add a test payment
	repo.Add(ctx, &domain.Payment{
		ID:        "test-3",
		OrderID:   "webhook-order-3",
		Amount:    2000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	cfg := &config.Config{
		PayMobHMACSecret: "test-secret",
	}
	paymentService := payment.NewService(cfg)
	handler := NewHandler(paymentService, repo)

	app.Post("/api/webhook", handler.Webhook)

	payload := domain.WebhookPayload{
		Type: "TRANSACTION",
		Obj: domain.WebhookObject{
			ID: 12348,
			Order: domain.WebhookOrder{
				ID:              103,
				MerchantOrderID: "webhook-order-3",
			},
			Success: true,
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Paymob-Signature", "invalid-signature")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 401 because signature doesn't match
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandler_Webhook_Pending(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add a test payment
	repo.Add(ctx, &domain.Payment{
		ID:        "test-4",
		OrderID:   "webhook-order-4",
		Amount:    3000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	cfg := &config.Config{}
	paymentService := payment.NewService(cfg)
	handler := NewHandler(paymentService, repo)

	app.Post("/api/webhook", handler.Webhook)

	payload := domain.WebhookPayload{
		Type: "TRANSACTION",
		Obj: domain.WebhookObject{
			ID: 12349,
			Order: domain.WebhookOrder{
				ID:              104,
				MerchantOrderID: "webhook-order-4",
			},
			Success: false,
			Pending: true,
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment stays pending
	updated, _ := repo.GetByOrderID(ctx, "webhook-order-4")
	require.NotNil(t, updated)
	assert.Equal(t, domain.PaymentStatusPending, updated.Status)
}
