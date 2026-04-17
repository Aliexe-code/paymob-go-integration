package unit_test

import (
	"fmt"
	"testing"
	"time"

	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"paymob-demo/pkg/utils"

	"github.com/stretchr/testify/assert"
)

// Domain Tests

func TestPaymentStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.PaymentStatus
		expected string
	}{
		{"pending", domain.PaymentStatusPending, "pending"},
		{"success", domain.PaymentStatusSuccess, "success"},
		{"failed", domain.PaymentStatusFailed, "failed"},
		{"cancelled", domain.PaymentStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, domain.PaymentStatus(tt.expected), tt.status)
		})
	}
}

func TestPaymentStruct(t *testing.T) {
	now := time.Now()
	payment := domain.Payment{
		ID:               "test-id",
		OrderID:          "order-123",
		Amount:           1000,
		Currency:         "EGP",
		Status:           domain.PaymentStatusPending,
		CheckoutURL:      "https://example.com/pay",
		PayMobOrderID:    12345,
		PayMobPaymentKey: "key-abc",
		CreatedAt:        now,
		UpdatedAt:        now,
		TransactionID:    "tx-001",
	}

	assert.Equal(t, "test-id", payment.ID)
	assert.Equal(t, "order-123", payment.OrderID)
	assert.Equal(t, 1000, payment.Amount)
	assert.Equal(t, "EGP", payment.Currency)
	assert.Equal(t, domain.PaymentStatusPending, payment.Status)
}

func TestPaymentRequest(t *testing.T) {
	req := domain.PaymentRequest{
		Amount:   500,
		Currency: "USD",
		Email:    "test@example.com",
		Name:     "John Doe",
		Phone:    "+1234567890",
	}

	assert.Equal(t, 500, req.Amount)
	assert.Equal(t, "USD", req.Currency)
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "John Doe", req.Name)
	assert.Equal(t, "+1234567890", req.Phone)
}

func TestDashboardData(t *testing.T) {
	data := domain.DashboardData{
		TotalPayments: 100,
		TotalAmount:   50000,
		SuccessCount:  80,
		FailedCount:   15,
		PendingCount:  5,
		RecentPayments: []domain.Payment{
			{ID: "p1", OrderID: "o1", Amount: 100},
		},
	}

	assert.Equal(t, 100, data.TotalPayments)
	assert.Equal(t, 50000, data.TotalAmount)
	assert.Equal(t, 80, data.SuccessCount)
	assert.Equal(t, 15, data.FailedCount)
	assert.Equal(t, 5, data.PendingCount)
	assert.Len(t, data.RecentPayments, 1)
}

func TestWebhookPayload(t *testing.T) {
	payload := domain.WebhookPayload{
		Type: "TRANSACTION",
		Obj: domain.WebhookObject{
			ID: 12345,
			Order: domain.WebhookOrder{
				ID:              67890,
				MerchantOrderID: "merchant-order-123",
			},
			Success:     true,
			Pending:     false,
			AmountCents: 10000,
			Currency:    "EGP",
		},
	}

	assert.Equal(t, "TRANSACTION", payload.Type)
	assert.Equal(t, 12345, payload.Obj.ID)
	assert.Equal(t, "merchant-order-123", payload.Obj.Order.MerchantOrderID)
	assert.True(t, payload.Obj.Success)
	assert.False(t, payload.Obj.Pending)
}

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		msg   string
	}{
		{"PaymentNotFound", domain.ErrPaymentNotFound, "payment not found"},
		{"InvalidAmount", domain.ErrInvalidAmount, "amount must be greater than 0"},
		{"InvalidRequest", domain.ErrInvalidRequest, "invalid request payload"},
		{"PaymentAlreadyCompleted", domain.ErrPaymentAlreadyCompleted, "payment already completed"},
		{"OrderIDRequired", domain.ErrOrderIDRequired, "order ID is required"},
		{"AuthenticationFailed", domain.ErrAuthenticationFailed, "PayMob authentication failed"},
		{"OrderCreationFailed", domain.ErrOrderCreationFailed, "PayMob order creation failed"},
		{"PaymentKeyFailed", domain.ErrPaymentKeyFailed, "PayMob payment key generation failed"},
		{"InvalidSignature", domain.ErrInvalidSignature, "invalid webhook signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.msg)
		})
	}
}

// Config Tests

func TestConfigLoad(t *testing.T) {
	cfg := config.Load()
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.PayMobBaseURL)
	assert.NotEmpty(t, cfg.ServerPort)
	assert.NotEmpty(t, cfg.ServerURL)
}

func TestConfigDefaults(t *testing.T) {
	cfg := config.Load()
	assert.Equal(t, "https://accept.paymobsolutions.com", cfg.PayMobBaseURL)
	assert.Equal(t, "3000", cfg.ServerPort)
	assert.Equal(t, "http://localhost:3000", cfg.ServerURL)
	// DemoMode depends on environment variable, just check it's a valid bool
	assert.IsType(t, false, cfg.DemoMode)
}

// Utils Tests

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{1000, "1,000"},
		{10000, "10,000"},
		{100000, "100,000"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.input), func(t *testing.T) {
			result := utils.FormatAmount(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusClass(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"success", "bg-green-500/20 text-green-400"},
		{"failed", "bg-red-500/20 text-red-400"},
		{"pending", "bg-yellow-500/20 text-yellow-400"},
		{"cancelled", "bg-yellow-500/20 text-yellow-400"},
		{"unknown", "bg-yellow-500/20 text-yellow-400"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := utils.StatusClass(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusText(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"success", "Success"},
		{"failed", "Failed"},
		{"cancelled", "Cancelled"},
		{"pending", "Pending"},
		{"unknown", "Pending"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := utils.StatusText(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}