package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPaymentStatus_Constants(t *testing.T) {
	assert.Equal(t, PaymentStatus("pending"), PaymentStatusPending)
	assert.Equal(t, PaymentStatus("success"), PaymentStatusSuccess)
	assert.Equal(t, PaymentStatus("failed"), PaymentStatusFailed)
	assert.Equal(t, PaymentStatus("cancelled"), PaymentStatusCancelled)
}

func TestPayment_Struct(t *testing.T) {
	now := time.Now()
	payment := Payment{
		ID:               "pay-123",
		OrderID:          "order-456",
		Amount:           1000,
		Currency:         "EGP",
		Status:           PaymentStatusPending,
		CheckoutURL:      "https://example.com/checkout",
		PayMobOrderID:    789,
		PayMobPaymentKey: "key-abc",
		CreatedAt:        now,
		UpdatedAt:        now,
		TransactionID:    "txn-xyz",
	}

	assert.Equal(t, "pay-123", payment.ID)
	assert.Equal(t, "order-456", payment.OrderID)
	assert.Equal(t, 1000, payment.Amount)
	assert.Equal(t, "EGP", payment.Currency)
	assert.Equal(t, PaymentStatusPending, payment.Status)
	assert.Equal(t, "https://example.com/checkout", payment.CheckoutURL)
	assert.Equal(t, 789, payment.PayMobOrderID)
	assert.Equal(t, "key-abc", payment.PayMobPaymentKey)
	assert.Equal(t, now, payment.CreatedAt)
	assert.Equal(t, now, payment.UpdatedAt)
	assert.Equal(t, "txn-xyz", payment.TransactionID)
}

func TestPaymentRequest_Struct(t *testing.T) {
	req := PaymentRequest{
		Amount:   5000,
		Currency: "USD",
		Email:    "test@example.com",
		Name:     "Test User",
		Phone:    "+1234567890",
	}

	assert.Equal(t, 5000, req.Amount)
	assert.Equal(t, "USD", req.Currency)
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "Test User", req.Name)
	assert.Equal(t, "+1234567890", req.Phone)
}

func TestPaymentResponse_Struct(t *testing.T) {
	resp := PaymentResponse{
		Success:     true,
		Message:     "Payment created",
		CheckoutURL: "https://checkout.example.com",
		OrderID:     "order-789",
	}

	assert.True(t, resp.Success)
	assert.Equal(t, "Payment created", resp.Message)
	assert.Equal(t, "https://checkout.example.com", resp.CheckoutURL)
	assert.Equal(t, "order-789", resp.OrderID)
}

func TestDashboardData_Struct(t *testing.T) {
	now := time.Now()
	data := DashboardData{
		TotalPayments: 100,
		TotalAmount:   50000,
		SuccessCount:  80,
		FailedCount:   15,
		PendingCount:  5,
		RecentPayments: []Payment{
			{
				ID:       "pay-1",
				OrderID:  "order-1",
				Amount:   1000,
				Currency: "EGP",
				Status:   PaymentStatusSuccess,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	assert.Equal(t, 100, data.TotalPayments)
	assert.Equal(t, 50000, data.TotalAmount)
	assert.Equal(t, 80, data.SuccessCount)
	assert.Equal(t, 15, data.FailedCount)
	assert.Equal(t, 5, data.PendingCount)
	assert.Len(t, data.RecentPayments, 1)
}

func TestWebhookPayload_Struct(t *testing.T) {
	payload := WebhookPayload{
		Type: "TRANSACTION",
		Obj: WebhookObject{
			ID:      12345,
			Success: true,
			Pending: false,
			IsRefund: false,
			Order: WebhookOrder{
				ID:              67890,
				MerchantOrderID: "merchant-order-123",
			},
			SourceData: WebhookSourceData{
				Type:     "card",
				SubType:  "visa",
				Pan:      "1234",
				CardType: "credit",
			},
			AmountCents:  100000,
			Currency:     "EGP",
			CreatedAt:    "2024-01-15T10:30:00Z",
			ErrorMessage: "",
		},
	}

	assert.Equal(t, "TRANSACTION", payload.Type)
	assert.Equal(t, 12345, payload.Obj.ID)
	assert.True(t, payload.Obj.Success)
	assert.Equal(t, "merchant-order-123", payload.Obj.Order.MerchantOrderID)
	assert.Equal(t, "card", payload.Obj.SourceData.Type)
	assert.Equal(t, 100000, payload.Obj.AmountCents)
}

func TestLegacyWebhookPayload_Struct(t *testing.T) {
	payload := LegacyWebhookPayload{
		OrderID:         12345,
		AmountCents:     100000,
		MerchantOrderID: "merchant-order-123",
		Success:         true,
		Message:         "Payment successful",
		TransactionID:   67890,
		ErrorOccured:    false,
	}

	assert.Equal(t, 12345, payload.OrderID)
	assert.Equal(t, 100000, payload.AmountCents)
	assert.Equal(t, "merchant-order-123", payload.MerchantOrderID)
	assert.True(t, payload.Success)
	assert.Equal(t, "Payment successful", payload.Message)
	assert.Equal(t, 67890, payload.TransactionID)
	assert.False(t, payload.ErrorOccured)
}

func TestPayMobAuthResponse_Struct(t *testing.T) {
	resp := PayMobAuthResponse{Token: "auth-token-123"}
	assert.Equal(t, "auth-token-123", resp.Token)
}

func TestPayMobOrderResponse_Struct(t *testing.T) {
	resp := PayMobOrderResponse{ID: 12345}
	assert.Equal(t, 12345, resp.ID)
}

func TestPayMobPaymentKeyResponse_Struct(t *testing.T) {
	resp := PayMobPaymentKeyResponse{Token: "payment-key-456"}
	assert.Equal(t, "payment-key-456", resp.Token)
}

func TestPaymentPageData_Struct(t *testing.T) {
	data := PaymentPageData{
		Title:  "Test Payment",
		APIURL: "https://api.example.com",
	}
	assert.Equal(t, "Test Payment", data.Title)
	assert.Equal(t, "https://api.example.com", data.APIURL)
}

func TestPaymentResultData_Struct(t *testing.T) {
	data := PaymentResultData{
		Success:     true,
		Message:     "Success",
		Amount:      1000,
		Currency:    "EGP",
		CheckoutURL: "https://checkout.example.com",
		OrderID:     "order-123",
	}
	assert.True(t, data.Success)
	assert.Equal(t, 1000, data.Amount)
}

func TestDashboardPageData_Struct(t *testing.T) {
	data := DashboardPageData{
		Title:          "Dashboard",
		TotalPayments:  100,
		TotalAmount:    "50,000",
		SuccessCount:   80,
		FailedCount:    15,
		PendingCount:   5,
		RecentPayments: []PaymentTableRow{},
	}
	assert.Equal(t, "Dashboard", data.Title)
	assert.Equal(t, 100, data.TotalPayments)
}

func TestPaymentTableRow_Struct(t *testing.T) {
	row := PaymentTableRow{
		OrderID:    "order-123",
		Amount:     "1,000",
		Currency:   "EGP",
		Status:     "success",
		StatusText: "Success",
		CreatedAt:  "2024-01-15",
	}
	assert.Equal(t, "order-123", row.OrderID)
	assert.Equal(t, "success", row.Status)
}

func TestResultPageData_Struct(t *testing.T) {
	data := ResultPageData{
		Title:   "Success",
		Message: "Payment completed",
		OrderID: "order-123",
	}
	assert.Equal(t, "Success", data.Title)
}

func TestSimulatePageData_Struct(t *testing.T) {
	data := SimulatePageData{
		Amount:   1000,
		Currency: "EGP",
		OrderID:  "order-123",
		Status:   "Pending",
	}
	assert.Equal(t, 1000, data.Amount)
	assert.Equal(t, "Pending", data.Status)
}
