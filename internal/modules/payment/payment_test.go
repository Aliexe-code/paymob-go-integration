package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/views"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_All(t *testing.T) {
	repo, cleanup := NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		payment := &domain.Payment{
			ID:        "test-1",
			OrderID:   "order-1",
			Amount:    1000,
			Currency:  "EGP",
			Status:    domain.PaymentStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Add(ctx, payment)
		require.NoError(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		result, err := repo.Get(ctx, "test-1")
		require.NoError(t, err)
		assert.Equal(t, "test-1", result.ID)
		assert.Equal(t, 1000, result.Amount)
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		_, err := repo.Get(ctx, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("GetByOrderID", func(t *testing.T) {
		result, err := repo.GetByOrderID(ctx, "order-1")
		require.NoError(t, err)
		assert.Equal(t, "order-1", result.OrderID)
	})

	t.Run("GetByOrderID_NotFound", func(t *testing.T) {
		_, err := repo.GetByOrderID(ctx, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("GetAll", func(t *testing.T) {
		all, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 1)
	})

	t.Run("Update", func(t *testing.T) {
		payment, _ := repo.Get(ctx, "test-1")
		payment.Status = domain.PaymentStatusSuccess
		err := repo.Update(ctx, payment)
		require.NoError(t, err)

		updated, _ := repo.Get(ctx, "test-1")
		assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
	})

	t.Run("GetDashboardData", func(t *testing.T) {
		data, err := repo.GetDashboardData(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, data.TotalPayments, 1)
		assert.GreaterOrEqual(t, data.TotalAmount, 1000)
	})

	t.Run("Close", func(t *testing.T) {
		// Test that Close doesn't panic - create a new repo for this
		repo2, cleanup2 := NewTestRepository()
		err := repo2.Close()
		require.NoError(t, err)
		cleanup2()
	})
}

func TestService_InitiatePayment(t *testing.T) {
	cfg := &config.Config{
		PayMobBaseURL: "https://accept.paymobsolutions.com",
		ServerURL:     "http://localhost:3000",
		DemoMode:      true,
	}

	service := NewService(cfg)

	t.Run("Success", func(t *testing.T) {
		req := domain.PaymentRequest{
			Amount:   1000,
			Currency: "EGP",
			Name:     "Test User",
			Email:    "test@example.com",
		}

		ctx := context.Background()
		payment, err := service.InitiatePayment(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, payment.ID)
		assert.NotEmpty(t, payment.OrderID)
		assert.Equal(t, 1000, payment.Amount)
		assert.Equal(t, domain.PaymentStatusPending, payment.Status)
		assert.NotEmpty(t, payment.CheckoutURL)
	})

	t.Run("EmptyCurrencyDefaultsToEGP", func(t *testing.T) {
		req := domain.PaymentRequest{
			Amount: 500,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		ctx := context.Background()
		payment, err := service.InitiatePayment(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "EGP", payment.Currency)
	})
}

func TestService_GetCheckoutURL(t *testing.T) {
	cfg := &config.Config{
		PayMobBaseURL:  "https://accept.paymobsolutions.com",
		PayMobIframeID: "12345",
	}

	service := NewService(cfg)

	url := service.GetCheckoutURL("test-token")
	assert.Contains(t, url, "accept.paymobsolutions.com")
	assert.Contains(t, url, "12345")
	assert.Contains(t, url, "test-token")
}

func TestService_VerifyWebhookSignature(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		signature string
		payload   []byte
		expected  bool
	}{
		{
			name:      "NoSecret_SkipVerification",
			secret:    "",
			signature: "anything",
			payload:   []byte(`{"test": "data"}`),
			expected:  true,
		},
		{
			name:      "InvalidSignature",
			secret:    "test-secret",
			signature: "a8b2c3d4e5f6",
			payload:   []byte(`{"test": "data"}`),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{PayMobHMACSecret: tt.secret}
			service := NewService(cfg)
			result := service.VerifyWebhookSignature(tt.signature, tt.payload)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandler_GetPaymentPage(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Get("/", handler.GetPaymentPage)

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
}

func TestHandler_InitiatePayment(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Post("/api/payments", handler.InitiatePayment)

	t.Run("Success", func(t *testing.T) {
		body := "amount=1000&name=Test&email=test@example.com&phone=+20123456789"
		req := httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
	})

	t.Run("InvalidAmount", func(t *testing.T) {
		body := "amount=0&name=Test&email=test@example.com"
		req := httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("NegativeAmount", func(t *testing.T) {
		body := "amount=-100&name=Test&email=test@example.com"
		req := httptest.NewRequest("POST", "/api/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestHandler_PaymentSuccess(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	// Add a test payment first
	ctx := context.Background()
	payment := &domain.Payment{
		ID:        "test-id",
		OrderID:   "test-order-123",
		Amount:    1000,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Add(ctx, payment)

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Get("/success", handler.PaymentSuccess)

	req := httptest.NewRequest("GET", "/success?order_id=test-order-123&transaction_id=txn-123", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "test-order-123")
	require.NotNil(t, updated)
	assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
	assert.Equal(t, "txn-123", updated.TransactionID)
}

func TestHandler_PaymentFailure(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	// Add a test payment first
	ctx := context.Background()
	payment := &domain.Payment{
		ID:        "test-id",
		OrderID:   "test-order-456",
		Amount:    500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Add(ctx, payment)

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Get("/failure", handler.PaymentFailure)

	req := httptest.NewRequest("GET", "/failure?order_id=test-order-456", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify payment was updated
	updated, _ := repo.GetByOrderID(ctx, "test-order-456")
	require.NotNil(t, updated)
	assert.Equal(t, domain.PaymentStatusFailed, updated.Status)
}

func TestHandler_SimulatePayment(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	// Add a test payment first
	ctx := context.Background()
	payment := &domain.Payment{
		ID:        "test-id",
		OrderID:   "sim-order-123",
		Amount:    2500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Add(ctx, payment)

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Post("/api/simulate/:order_id", handler.SimulatePaymentSuccess)
	app.Post("/api/simulate-failure/:order_id", handler.SimulatePaymentFailure)

	t.Run("SimulateSuccess", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/simulate/sim-order-123", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		updated, _ := repo.GetByOrderID(ctx, "sim-order-123")
		require.NotNil(t, updated)
		assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
	})
}

func TestHandler_HealthCheck(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Get("/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "healthy", result["status"])
}

func TestHandler_Benchmark(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Get("/benchmark", handler.Benchmark)

	req := httptest.NewRequest("GET", "/benchmark", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandler_SimulatePaymentPage(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	// Add a test payment first
	ctx := context.Background()
	payment := &domain.Payment{
		ID:        "test-id",
		OrderID:   "simpage-123",
		Amount:    1500,
		Currency:  "EGP",
		Status:    domain.PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Add(ctx, payment)

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Get("/pay/simulate", handler.SimulatePaymentPage)

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/pay/simulate?order_id=simpage-123", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("MissingOrderID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/pay/simulate", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/pay/simulate?order_id=nonexistent", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestHandler_SimulatePaymentFailure(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Post("/api/simulate-failure/:order_id", handler.SimulatePaymentFailure)

	t.Run("Success", func(t *testing.T) {
		// Add a pending payment
		payment := &domain.Payment{
			ID:        "test-id",
			OrderID:   "fail-order-123",
			Amount:    2500,
			Currency:  "EGP",
			Status:    domain.PaymentStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Add(ctx, payment)

		req := httptest.NewRequest("POST", "/api/simulate-failure/fail-order-123", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Verify payment was marked as failed
		updated, _ := repo.GetByOrderID(ctx, "fail-order-123")
		require.NotNil(t, updated)
		assert.Equal(t, domain.PaymentStatusFailed, updated.Status)

		// Verify JSON response
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, "failed", result["status"])
		assert.Equal(t, "fail-order-123", result["order_id"])
	})

	t.Run("MissingOrderID", func(t *testing.T) {
		// Empty order_id results in 404 (route doesn't match)
		req := httptest.NewRequest("POST", "/api/simulate-failure/", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/simulate-failure/nonexistent", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("AlreadyCompleted", func(t *testing.T) {
		// Add an already completed payment
		payment := &domain.Payment{
			ID:        "test-id-2",
			OrderID:   "completed-order",
			Amount:    1000,
			Currency:  "EGP",
			Status:    domain.PaymentStatusSuccess,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Add(ctx, payment)

		req := httptest.NewRequest("POST", "/api/simulate-failure/completed-order", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})
}

func TestHandler_SimulatePaymentSuccess_ErrorCases(t *testing.T) {
	app := fiber.New()
	repo, cleanup := NewTestRepository()
	defer cleanup()

	ctx := context.Background()

	renderer, err := views.NewRenderer()
	require.NoError(t, err)

	cfg := &config.Config{DemoMode: true}
	service := NewService(cfg)
	handler := NewHandler(service, repo, renderer, cfg)

	app.Post("/api/simulate/:order_id", handler.SimulatePaymentSuccess)

	t.Run("MissingOrderID", func(t *testing.T) {
		// Empty order_id results in 404 (route doesn't match)
		req := httptest.NewRequest("POST", "/api/simulate/", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/simulate/nonexistent", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("AlreadySuccess", func(t *testing.T) {
		// Add an already successful payment
		payment := &domain.Payment{
			ID:        "test-id",
			OrderID:   "already-success",
			Amount:    1000,
			Currency:  "EGP",
			Status:    domain.PaymentStatusSuccess,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Add(ctx, payment)

		req := httptest.NewRequest("POST", "/api/simulate/already-success", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("AlreadyFailed", func(t *testing.T) {
		// Add an already failed payment
		payment := &domain.Payment{
			ID:        "test-id-2",
			OrderID:   "already-failed",
			Amount:    1000,
			Currency:  "EGP",
			Status:    domain.PaymentStatusFailed,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Add(ctx, payment)

		req := httptest.NewRequest("POST", "/api/simulate/already-failed", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("ShortOrderID", func(t *testing.T) {
		// Test safeTruncate with short order ID (prevents panic)
		payment := &domain.Payment{
			ID:        "test-id-3",
			OrderID:   "short", // Less than 8 chars
			Amount:    1000,
			Currency:  "EGP",
			Status:    domain.PaymentStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.Add(ctx, payment)

		req := httptest.NewRequest("POST", "/api/simulate/short", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		// Verify transaction ID was created without panic
		updated, _ := repo.GetByOrderID(ctx, "short")
		require.NotNil(t, updated)
		assert.Equal(t, "SIM_short", updated.TransactionID)
	})
}

func TestSafeTruncate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		maxLen  int
		expected string
	}{
		{"ShorterThanMax", "hello", 10, "hello"},
		{"ExactLength", "hello", 5, "hello"},
		{"LongerThanMax", "helloworld", 5, "hello"},
		{"Empty", "", 5, ""},
		{"SingleChar", "a", 1, "a"},
		{"Unicode", "héllo", 3, "hé"}, // safeTruncate works on bytes, 'hé' = 3 bytes in UTF-8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeTruncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRealPaymentFlow tests the actual PayMob API integration with mocked server
func TestRealPaymentFlow(t *testing.T) {
	// Create a mock PayMob server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/api/auth/tokens":
			// Authenticate endpoint
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token": "mock_auth_token_12345",
			})

		case "/api/ecommerce/orders":
			// Create order endpoint
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{
				"id": 123456,
			})

		case "/api/acceptance/payment_keys":
			// Get payment key endpoint
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token": "mock_payment_key_abcdef",
			})

		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		PayMobAPIKey:       "test_api_key",
		PayMobIntegrationID: "123456",
		PayMobMerchantID:   "789012",
		PayMobBaseURL:      mockServer.URL,
		ServerURL:          "http://localhost:8080",
		DemoMode:           false, // Important: test real flow
	}

	// Create service with mock server URL
	service := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, mockServer.URL)

	t.Run("FullPaymentFlow_Success", func(t *testing.T) {
		req := domain.PaymentRequest{
			Amount:   1000,
			Currency: "EGP",
			Name:     "Test User",
			Email:    "test@example.com",
			Phone:    "+201000000000",
		}

		payment, err := service.InitiatePayment(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, payment.ID)
		assert.NotEmpty(t, payment.OrderID)
		assert.Equal(t, 1000, payment.Amount)
		assert.Equal(t, "EGP", payment.Currency)
		assert.Equal(t, domain.PaymentStatusPending, payment.Status)
		assert.Equal(t, 123456, payment.PayMobOrderID)
		assert.Equal(t, "mock_payment_key_abcdef", payment.PayMobPaymentKey)
		assert.Contains(t, payment.CheckoutURL, "mock_payment_key_abcdef")
	})

	t.Run("Authenticate_Failure_InvalidCredentials", func(t *testing.T) {
		// Create server that returns 401 for auth
		authFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/auth/tokens" {
				http.Error(w, `{"detail": "Invalid credentials"}`, http.StatusUnauthorized)
				return
			}
		}))
		defer authFailServer.Close()

		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, authFailServer.URL)
		req := domain.PaymentRequest{
			Amount: 1000,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		_, err := svc.InitiatePayment(context.Background(), req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	})

	t.Run("CreateOrder_Failure", func(t *testing.T) {
		// Create server that fails on order creation
		orderFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/auth/tokens":
				json.NewEncoder(w).Encode(map[string]string{"token": "valid_token"})
			case "/api/ecommerce/orders":
				http.Error(w, `{"detail": "Invalid merchant"}`, http.StatusBadRequest)
			}
		}))
		defer orderFailServer.Close()

		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, orderFailServer.URL)
		req := domain.PaymentRequest{
			Amount: 1000,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		_, err := svc.InitiatePayment(context.Background(), req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrderCreationFailed)
	})

	t.Run("GetPaymentKey_Failure", func(t *testing.T) {
		// Create server that fails on payment key
		keyFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/auth/tokens":
				json.NewEncoder(w).Encode(map[string]string{"token": "valid_token"})
			case "/api/ecommerce/orders":
				json.NewEncoder(w).Encode(map[string]int{"id": 123456})
			case "/api/acceptance/payment_keys":
				http.Error(w, `{"detail": "Invalid billing data"}`, http.StatusBadRequest)
			}
		}))
		defer keyFailServer.Close()

		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, keyFailServer.URL)
		req := domain.PaymentRequest{
			Amount: 1000,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		_, err := svc.InitiatePayment(context.Background(), req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrPaymentKeyFailed)
	})

	t.Run("NetworkError_Authenticate", func(t *testing.T) {
		// Create server that closes connection
		networkErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
		}))
		defer networkErrorServer.Close()

		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 1 * time.Second}, networkErrorServer.URL)
		req := domain.PaymentRequest{
			Amount: 1000,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		_, err := svc.InitiatePayment(context.Background(), req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	})

	t.Run("EmptyAuthToken_Response", func(t *testing.T) {
		emptyTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"token": ""})
		}))
		defer emptyTokenServer.Close()

		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, emptyTokenServer.URL)
		req := domain.PaymentRequest{
			Amount: 1000,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		_, err := svc.InitiatePayment(context.Background(), req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	})

	t.Run("InvalidJSON_Response", func(t *testing.T) {
		invalidJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid json`))
		}))
		defer invalidJSONServer.Close()

		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, invalidJSONServer.URL)
		req := domain.PaymentRequest{
			Amount: 1000,
			Name:   "Test User",
			Email:  "test@example.com",
		}

		_, err := svc.InitiatePayment(context.Background(), req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
	})

	t.Run("GetCheckoutURL", func(t *testing.T) {
		url := service.GetCheckoutURL("test_payment_key")
		assert.Contains(t, url, mockServer.URL)
		assert.Contains(t, url, "test_payment_key")
		assert.Contains(t, url, "/api/acceptance/iframes/")
	})
}

// TestPaymentMethods_Individual tests each API method in isolation
func TestPaymentMethods_Individual(t *testing.T) {
	t.Run("Authenticate_Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/auth/tokens":
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify request body contains API key
				var body map[string]string
				json.NewDecoder(r.Body).Decode(&body)
				assert.Equal(t, "test_key", body["api_key"])

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"token": "auth_token_123"})
			case "/api/ecommerce/orders":
				json.NewEncoder(w).Encode(map[string]int{"id": 123})
			case "/api/acceptance/payment_keys":
				json.NewEncoder(w).Encode(map[string]string{"token": "payment_key"})
			}
		}))
		defer server.Close()

		cfg := &config.Config{
			PayMobAPIKey:        "test_key",
			PayMobIntegrationID: "123",
			PayMobMerchantID:    "456",
		}
		svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)

		// Test the full flow which includes authentication
		req := domain.PaymentRequest{
			Amount: 100,
			Name:   "Test",
			Email:  "test@test.com",
		}

		payment, err := svc.InitiatePayment(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, payment)
	})

	t.Run("CreateOrder_WithDifferentCurrencies", func(t *testing.T) {
		currencies := []string{"EGP", "USD", "EUR"}

		for _, currency := range currencies {
			t.Run(currency, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/auth/tokens":
						json.NewEncoder(w).Encode(map[string]string{"token": "auth_token"})
					case "/api/ecommerce/orders":
						var body map[string]interface{}
						json.NewDecoder(r.Body).Decode(&body)
						assert.Equal(t, currency, body["currency"])
						// Amount should be in cents (multiplied by 100)
						assert.Equal(t, float64(5000), body["amount_cents"]) // 50 * 100
						json.NewEncoder(w).Encode(map[string]int{"id": 999})
					case "/api/acceptance/payment_keys":
						json.NewEncoder(w).Encode(map[string]string{"token": "key"})
					}
				}))
				defer server.Close()

				cfg := &config.Config{
					PayMobAPIKey:        "test_key",
					PayMobMerchantID:    "123",
					PayMobIntegrationID: "456",
				}
				svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)

				req := domain.PaymentRequest{
					Amount:   50,
					Currency: currency,
					Name:     "Test",
					Email:    "test@test.com",
				}

							payment, err := svc.InitiatePayment(context.Background(), req)
							require.NoError(t, err)
							assert.Equal(t, 999, payment.PayMobOrderID)
						})
					}
				})
				}
				
				// TestRealPaymentFlow_EdgeCases tests additional edge cases for real payment flow
				func TestRealPaymentFlow_EdgeCases(t *testing.T) {
					t.Run("CreateOrder_InvalidJSONResponse", func(t *testing.T) {
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							switch r.URL.Path {
							case "/api/auth/tokens":
								json.NewEncoder(w).Encode(map[string]string{"token": "valid_token"})
							case "/api/ecommerce/orders":
								w.Header().Set("Content-Type", "application/json")
								w.Write([]byte(`{invalid json response`))
							}
						}))
						defer server.Close()
				
						cfg := &config.Config{
							PayMobAPIKey:        "test_key",
							PayMobMerchantID:    "123",
							PayMobIntegrationID: "456",
							DemoMode:            false,
						}
						svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)
				
						req := domain.PaymentRequest{
							Amount: 100,
							Name:   "Test",
							Email:  "test@test.com",
						}
				
						_, err := svc.InitiatePayment(context.Background(), req)
						assert.Error(t, err)
						assert.ErrorIs(t, err, domain.ErrOrderCreationFailed)
					})
				
					t.Run("GetPaymentKey_InvalidJSONResponse", func(t *testing.T) {
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							switch r.URL.Path {
							case "/api/auth/tokens":
								json.NewEncoder(w).Encode(map[string]string{"token": "valid_token"})
							case "/api/ecommerce/orders":
								json.NewEncoder(w).Encode(map[string]int{"id": 123456})
							case "/api/acceptance/payment_keys":
								w.Header().Set("Content-Type", "application/json")
								w.Write([]byte(`{invalid json response`))
							}
						}))
						defer server.Close()
				
						cfg := &config.Config{
							PayMobAPIKey:        "test_key",
							PayMobMerchantID:    "123",
							PayMobIntegrationID: "456",
							DemoMode:            false,
						}
						svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)
				
						req := domain.PaymentRequest{
							Amount: 100,
							Name:   "Test",
							Email:  "test@test.com",
						}
				
						_, err := svc.InitiatePayment(context.Background(), req)
						assert.Error(t, err)
						assert.ErrorIs(t, err, domain.ErrPaymentKeyFailed)
					})
				
						t.Run("CreateOrder_EmptyResponseBody", func(t *testing.T) {
							server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								switch r.URL.Path {
								case "/api/auth/tokens":
									json.NewEncoder(w).Encode(map[string]string{"token": "valid_token"})
								case "/api/ecommerce/orders":
									w.Header().Set("Content-Type", "application/json")
									w.Write([]byte(`{}`)) // Empty response - ID will be 0
								case "/api/acceptance/payment_keys":
									// This will be called with order_id=0 and should fail
									http.Error(w, "Invalid order", http.StatusBadRequest)
								}
							}))
							defer server.Close()
					
							cfg := &config.Config{
								PayMobAPIKey:        "test_key",
								PayMobMerchantID:    "123",
								PayMobIntegrationID: "456",
								DemoMode:            false,
							}
							svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)
					
							req := domain.PaymentRequest{
								Amount: 100,
								Name:   "Test",
								Email:  "test@test.com",
							}
					
							_, err := svc.InitiatePayment(context.Background(), req)
							assert.Error(t, err)
							// The flow continues with order_id=0 and fails at payment key step
							assert.ErrorIs(t, err, domain.ErrPaymentKeyFailed)
						})				
					t.Run("GetPaymentKey_EmptyTokenResponse", func(t *testing.T) {
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							switch r.URL.Path {
							case "/api/auth/tokens":
								json.NewEncoder(w).Encode(map[string]string{"token": "valid_token"})
							case "/api/ecommerce/orders":
								json.NewEncoder(w).Encode(map[string]int{"id": 123456})
							case "/api/acceptance/payment_keys":
								w.Header().Set("Content-Type", "application/json")
								json.NewEncoder(w).Encode(map[string]string{"token": ""}) // Empty token
							}
						}))
						defer server.Close()
				
						cfg := &config.Config{
							PayMobAPIKey:        "test_key",
							PayMobMerchantID:    "123",
							PayMobIntegrationID: "456",
							DemoMode:            false,
						}
						svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)
				
						req := domain.PaymentRequest{
							Amount: 100,
							Name:   "Test",
							Email:  "test@test.com",
						}
				
						_, err := svc.InitiatePayment(context.Background(), req)
						assert.Error(t, err)
						assert.ErrorIs(t, err, domain.ErrPaymentKeyFailed)
					})
				
					t.Run("InitiatePayment_HTTPTimeout", func(t *testing.T) {
						// Create a server that never responds
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							time.Sleep(5 * time.Second) // Longer than client timeout
						}))
						defer server.Close()
				
						cfg := &config.Config{
							PayMobAPIKey:        "test_key",
							PayMobMerchantID:    "123",
							PayMobIntegrationID: "456",
							DemoMode:            false,
						}
						// Use a very short timeout
						svc := NewServiceWithClient(cfg, &http.Client{Timeout: 100 * time.Millisecond}, server.URL)
				
						req := domain.PaymentRequest{
							Amount: 100,
							Name:   "Test",
							Email:  "test@test.com",
						}
				
						_, err := svc.InitiatePayment(context.Background(), req)
						assert.Error(t, err)
						assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
					})
				
					t.Run("ServerReturns500Status", func(t *testing.T) {
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						}))
						defer server.Close()
				
						cfg := &config.Config{
							PayMobAPIKey:        "test_key",
							PayMobMerchantID:    "123",
							PayMobIntegrationID: "456",
							DemoMode:            false,
						}
						svc := NewServiceWithClient(cfg, &http.Client{Timeout: 10 * time.Second}, server.URL)
				
						req := domain.PaymentRequest{
							Amount: 100,
							Name:   "Test",
							Email:  "test@test.com",
						}
				
						_, err := svc.InitiatePayment(context.Background(), req)
						assert.Error(t, err)
						assert.ErrorIs(t, err, domain.ErrAuthenticationFailed)
					})
				}