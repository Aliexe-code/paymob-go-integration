package dashboard

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/modules/payment"
	"paymob-demo/internal/views"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetDashboard(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add some test data
	repo.Add(ctx, &domain.Payment{
		ID:        "test-1",
		OrderID:   "order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusSuccess,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	repo.Add(ctx, &domain.Payment{
		ID:        "test-2",
		OrderID:   "order-2",
		Amount:    500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusFailed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	handler := NewHandler(repo, renderer)
	app.Get("/dashboard", handler.GetDashboard)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
}

func TestHandler_GetDashboardData(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add some test data
	repo.Add(ctx, &domain.Payment{
		ID:        "test-1",
		OrderID:   "order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusSuccess,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	repo.Add(ctx, &domain.Payment{
		ID:        "test-2",
		OrderID:   "order-2",
		Amount:    500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusFailed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	repo.Add(ctx, &domain.Payment{
		ID:        "test-3",
		OrderID:   "order-3",
		Amount:    750,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	handler := NewHandler(repo, renderer)
	app.Get("/api/dashboard", handler.GetDashboardData)

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	assert.Equal(t, float64(3), data["total_payments"])
	assert.Equal(t, float64(1), data["success_count"])
	assert.Equal(t, float64(1), data["failed_count"])
	assert.Equal(t, float64(1), data["pending_count"])
}

func TestHandler_GetDashboardHTML(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add some test data
	repo.Add(ctx, &domain.Payment{
		ID:        "test-1",
		OrderID:   "order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusSuccess,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	handler := NewHandler(repo, renderer)
	app.Get("/api/dashboard/html", handler.GetDashboardHTML)

	req := httptest.NewRequest("GET", "/api/dashboard/html", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
}

func TestHandler_SimulatePaymentSuccess(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add a pending payment
	repo.Add(ctx, &domain.Payment{
		ID:        "test-1",
		OrderID:   "order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	paymentService := payment.NewService(cfg)
	paymentHandler := payment.NewHandler(paymentService, repo, renderer, cfg)

	app.Post("/api/simulate/:order_id", paymentHandler.SimulatePaymentSuccess)

	req := httptest.NewRequest("POST", "/api/simulate/order-1", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "order-1")
	require.NotNil(t, updated)
	assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
}

func TestHandler_SimulatePaymentFailure(t *testing.T) {
	app := fiber.New()
	repo, cleanup := payment.NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	// Add a pending payment
	repo.Add(ctx, &domain.Payment{
		ID:        "test-2",
		OrderID:   "order-2",
		Amount:    500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	paymentService := payment.NewService(cfg)
	paymentHandler := payment.NewHandler(paymentService, repo, renderer, cfg)

	app.Post("/api/simulate-failure/:order_id", paymentHandler.SimulatePaymentFailure)

	req := httptest.NewRequest("POST", "/api/simulate-failure/order-2", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "order-2")
	require.NotNil(t, updated)
	assert.Equal(t, domain.PaymentStatusFailed, updated.Status)
}