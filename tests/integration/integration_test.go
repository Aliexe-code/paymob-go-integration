package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/modules/dashboard"
	"paymob-demo/internal/modules/payment"
	"paymob-demo/internal/modules/webhook"
	"paymob-demo/internal/views"
	"paymob-demo/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp(t *testing.T) (*fiber.App, *payment.Repository) {
	// Use temporary database
	dbPath := "test_integration.db"

	repo, err := payment.NewRepository(dbPath)
	require.NoError(t, err)

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{
		PayMobBaseURL: "https://accept.paymobsolutions.com",
		ServerURL:     "http://localhost:3000",
		DemoMode:      true,
	}

	paymentService := payment.NewService(cfg)
	paymentHandler := payment.NewHandler(paymentService, repo, renderer, cfg)
	dashboardHandler := dashboard.NewHandler(repo, renderer)
	webhookHandler := webhook.NewHandler(paymentService, repo)

	app := fiber.New(fiber.Config{
		AppName: "Test App",
	})

	// API Routes
	api := app.Group("/api")
	api.Get("/health", paymentHandler.HealthCheck)
	api.Get("/benchmark", paymentHandler.Benchmark)
	api.Post("/payments", paymentHandler.InitiatePayment)
	api.Get("/dashboard", dashboardHandler.GetDashboardData)
	api.Get("/dashboard/html", dashboardHandler.GetDashboardHTML)
	api.Post("/simulate/:order_id", paymentHandler.SimulatePaymentSuccess)
	api.Post("/simulate-failure/:order_id", paymentHandler.SimulatePaymentFailure)
	api.Post("/webhook", webhookHandler.Webhook)

	// Page Routes
	app.Get("/", paymentHandler.GetPaymentPage)
	app.Get("/success", paymentHandler.PaymentSuccess)
	app.Get("/failure", paymentHandler.PaymentFailure)
	app.Get("/pay/simulate", paymentHandler.SimulatePaymentPage)

	return app, repo
}

func TestHealthCheck(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "healthy", body["status"])
	assert.Equal(t, "paymob-demo", body["service"])
}

func TestBenchmark(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/api/benchmark", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetPaymentPage(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "PayMob")
	assert.Contains(t, string(body), "Secure Payment")
	assert.Contains(t, string(body), "Payment Details")
}

func TestInitiatePayment(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	// Test valid payment request
	formData := "amount=1000&name=Test%20User&email=test@example.com&phone=%2B201000000000"
	req := httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Payment Ready!")
	assert.Contains(t, string(body), "1,000 EGP")
}

func TestInitiatePayment_InvalidAmount(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	formData := "amount=0&name=Test&email=test@example.com"
	req := httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Amount must be greater than 0")
}

func TestInitiatePayment_MissingFields(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// Empty body results in 0 amount which triggers "Amount must be greater than 0"
	assert.Contains(t, string(body), "Amount must be greater than 0")
}

func TestPaymentSuccess(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/success?order_id=test-order-123", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Payment Successful")
}

func TestPaymentFailure(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/failure?order_id=test-order-456", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Payment Failed")
}

func TestGetDashboardData(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	assert.Contains(t, data, "total_payments")
	assert.Contains(t, data, "total_amount")
	assert.Contains(t, data, "success_count")
	assert.Contains(t, data, "failed_count")
	assert.Contains(t, data, "pending_count")
	assert.Contains(t, data, "recent_payments")
}

func TestGetDashboardHTML(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/api/dashboard/html", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Total Payments")
	assert.Contains(t, string(body), "Successful")
	assert.Contains(t, string(body), "Failed")
}

func TestSimulatePaymentSuccess(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	// First, create a payment
	ctx := t.Context()
	p := &domain.Payment{
		ID:        "sim-test-1",
		OrderID:   "sim-order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
	}
	repo.Add(ctx, p)

	// Simulate success
	req := httptest.NewRequest("POST", "/api/simulate/sim-order-1", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Success")

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "sim-order-1")
	assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
	assert.Contains(t, updated.TransactionID, "SIM_")
}

func TestSimulatePaymentFailure(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	// Create a payment
	ctx := t.Context()
	p := &domain.Payment{
		ID:        "sim-test-2",
		OrderID:   "sim-order-2",
		Amount:    500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
	}
	repo.Add(ctx, p)

	// Simulate failure
	req := httptest.NewRequest("POST", "/api/simulate-failure/sim-order-2", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "sim-order-2")
	assert.Equal(t, domain.PaymentStatusFailed, updated.Status)
}

func TestSimulatePayment_NotFound(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("POST", "/api/simulate/nonexistent-order", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestWebhook(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	// Create a payment
	ctx := t.Context()
	p := &domain.Payment{
		ID:        "webhook-test-1",
		OrderID:   "webhook-order-1",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
	}
	repo.Add(ctx, p)

	// Create webhook payload
	payload := domain.WebhookPayload{
		Type: "TRANSACTION",
		Obj: domain.WebhookObject{
			ID: 12345,
			Order: domain.WebhookOrder{
				ID:              67890,
				MerchantOrderID: "webhook-order-1",
			},
			Success:     true,
			AmountCents: 100000,
			Currency:    "EGP",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "webhook-order-1")
	assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
}

func TestSimulatePaymentPage(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	// Create a payment
	ctx := t.Context()
	p := &domain.Payment{
		ID:        "simulate-page-1",
		OrderID:   "simulate-page-order-1",
		Amount:    2500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
	}
	repo.Add(ctx, p)

	req := httptest.NewRequest("GET", "/pay/simulate?order_id=simulate-page-order-1", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "DEMO MODE")
	assert.Contains(t, string(body), "2,500 EGP")
}

func TestSimulatePaymentPage_MissingOrderID(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/pay/simulate", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)
}

func TestSimulatePaymentPage_NotFound(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	req := httptest.NewRequest("GET", "/pay/simulate?order_id=nonexistent", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestFullPaymentFlow(t *testing.T) {
	app, repo := setupTestApp(t)
	defer repo.Close()
	defer os.Remove("test_integration.db")

	// 1. Get payment page
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// 2. Initiate payment
	formData := "amount=5000&name=Integration%20Test&email=integration@test.com&phone=%2B201111111111"
	req = httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = app.Test(req, -1)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	assert.Contains(t, string(body), "Payment Ready!")
	assert.Contains(t, string(body), "5,000 EGP")

	// 3. Get dashboard
	req = httptest.NewRequest("GET", "/api/dashboard", nil)
	resp, err = app.Test(req, -1)
	require.NoError(t, err)

	var dashboardData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&dashboardData)
	resp.Body.Close()

	assert.GreaterOrEqual(t, dashboardData["total_payments"].(float64), 1.0)
}

// Test utility functions in integration context
func TestUtilsInIntegration(t *testing.T) {
	assert.Equal(t, "1,000", utils.FormatAmount(1000))
	assert.Equal(t, "bg-green-500/20 text-green-400", utils.StatusClass("success"))
	assert.Equal(t, "Success", utils.StatusText("success"))
}
