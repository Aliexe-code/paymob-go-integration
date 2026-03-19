package views

import (
	"testing"

	"paymob-demo/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)
	require.NotNil(t, renderer)
}

func TestRenderPaymentPage(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.PaymentPageData{
		Title: "Test Payment Page",
	}

	html, err := renderer.RenderPaymentPage(data)
	require.NoError(t, err)
	assert.Contains(t, html, "PayMob")
	assert.Contains(t, html, "Secure Payment")
	assert.Contains(t, html, "hx-post=\"/api/payments\"")
	assert.Contains(t, html, "Payment Details")
}

func TestRenderPaymentResult_Success(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.PaymentResultData{
		Success:     true,
		Amount:      1000,
		Currency:    "EGP",
		CheckoutURL: "https://pay.example.com/checkout",
		OrderID:     "order-123",
	}

	html, err := renderer.RenderPaymentResult(data)
	require.NoError(t, err)
	assert.Contains(t, html, "Payment Ready!")
	assert.Contains(t, html, "1,000 EGP")
	assert.Contains(t, html, "order-123")
}

func TestRenderPaymentResult_Failure(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.PaymentResultData{
		Success: false,
		Message: "Payment failed: insufficient funds",
	}

	html, err := renderer.RenderPaymentResult(data)
	require.NoError(t, err)
	assert.Contains(t, html, "insufficient funds")
}

func TestRenderDashboard(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.DashboardPageData{
		Title:         "Dashboard",
		TotalPayments: 100,
		TotalAmount:   "50,000",
		SuccessCount:  80,
		FailedCount:   15,
		PendingCount:  5,
		RecentPayments: []domain.PaymentTableRow{
			{OrderID: "o1", Amount: "100", Currency: "EGP", Status: "success", StatusText: "Success", CreatedAt: "2024-01-01"},
		},
	}

	html, err := renderer.RenderDashboard(data)
	require.NoError(t, err)
	assert.Contains(t, html, "Payment Dashboard")
	assert.Contains(t, html, "100")
	assert.Contains(t, html, "50,000 EGP")
}

func TestRenderDashboardHTML(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.DashboardPageData{
		TotalPayments:  50,
		TotalAmount:    "25,000",
		SuccessCount:   40,
		FailedCount:    8,
		PendingCount:   2,
		RecentPayments: []domain.PaymentTableRow{},
	}

	html, err := renderer.RenderDashboardHTML(data)
	require.NoError(t, err)
	assert.Contains(t, html, "50")
	assert.Contains(t, html, "25,000")
}

func TestRenderSuccessPage(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.ResultPageData{
		Title:   "Payment Successful",
		Message: "Your payment was completed successfully.",
		OrderID: "order-456",
	}

	html, err := renderer.RenderSuccessPage(data)
	require.NoError(t, err)
	assert.Contains(t, html, "Payment Successful")
	assert.Contains(t, html, "order-456")
	assert.Contains(t, html, "bg-gradient-to-br from-green-500")
}

func TestRenderFailurePage(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.ResultPageData{
		Title:   "Payment Failed",
		Message: "Your payment was not completed.",
		OrderID: "order-789",
	}

	html, err := renderer.RenderFailurePage(data)
	require.NoError(t, err)
	assert.Contains(t, html, "Payment Failed")
	assert.Contains(t, html, "order-789")
}

func TestRenderSimulatePage(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	data := domain.SimulatePageData{
		Amount:   500,
		Currency: "EGP",
		OrderID:  "sim-order-123",
		Status:   "Pending",
	}

	html, err := renderer.RenderSimulatePage(data)
	require.NoError(t, err)
	assert.Contains(t, html, "DEMO MODE")
	assert.Contains(t, html, "500 EGP")
	assert.Contains(t, html, "sim-order-123")
}

func TestRenderInvalidTemplate(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	_, err = renderer.Render("nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined")
}

func TestFormatAmountFunc(t *testing.T) {
	assert.Equal(t, "0", formatAmount(0))
	assert.Equal(t, "1", formatAmount(1))
	assert.Equal(t, "100", formatAmount(100))
	assert.Equal(t, "1,000", formatAmount(1000))
	assert.Equal(t, "10,000", formatAmount(10000))
	assert.Equal(t, "1,000,000", formatAmount(1000000))
}

func TestStatusClassFunc(t *testing.T) {
	assert.Equal(t, "bg-green-500/20 text-green-400", statusClass("success"))
	assert.Equal(t, "bg-red-500/20 text-red-400", statusClass("failed"))
	assert.Equal(t, "bg-yellow-500/20 text-yellow-400", statusClass("pending"))
	assert.Equal(t, "bg-yellow-500/20 text-yellow-400", statusClass("unknown"))
}

func TestStatusTextFunc(t *testing.T) {
	assert.Equal(t, "Success", statusText("success"))
	assert.Equal(t, "Failed", statusText("failed"))
	assert.Equal(t, "Pending", statusText("pending"))
	assert.Equal(t, "Cancelled", statusText("cancelled"))
	assert.Equal(t, "Pending", statusText("unknown"))
}

func TestRenderPaymentRow(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	t.Run("SuccessStatus", func(t *testing.T) {
		data := domain.PaymentTableRow{
			OrderID:    "order-123",
			Amount:     "1,000",
			Currency:   "EGP",
			Status:     "success",
			StatusText: "Success",
			CreatedAt:  "2024-01-15 10:30",
		}

		html, err := renderer.RenderPaymentRow(data)
		require.NoError(t, err)
		assert.Contains(t, html, "order-123")
		assert.Contains(t, html, "1,000 EGP")
		assert.Contains(t, html, "Success")
		assert.Contains(t, html, "✓ Completed")
	})

	t.Run("PendingStatus", func(t *testing.T) {
		data := domain.PaymentTableRow{
			OrderID:    "order-456",
			Amount:     "500",
			Currency:   "USD",
			Status:     "pending",
			StatusText: "Pending",
			CreatedAt:  "2024-01-15 11:00",
		}

		html, err := renderer.RenderPaymentRow(data)
		require.NoError(t, err)
		assert.Contains(t, html, "order-456")
		assert.Contains(t, html, "500 USD")
		assert.Contains(t, html, "hx-post=\"/api/simulate/order-456\"")
	})

	t.Run("FailedStatus", func(t *testing.T) {
		data := domain.PaymentTableRow{
			OrderID:    "order-789",
			Amount:     "2,000",
			Currency:   "EUR",
			Status:     "failed",
			StatusText: "Failed",
			CreatedAt:  "2024-01-15 12:00",
		}

		html, err := renderer.RenderPaymentRow(data)
		require.NoError(t, err)
		assert.Contains(t, html, "order-789")
		assert.Contains(t, html, "✗ Failed")
	})
}

func TestLoadTemplatesFromDir(t *testing.T) {
	t.Run("ValidDirectory", func(t *testing.T) {
		renderer, err := LoadTemplatesFromDir("templates")
		require.NoError(t, err)
		require.NotNil(t, renderer)

		// Verify it can render
		html, err := renderer.RenderPaymentPage(domain.PaymentPageData{Title: "Test"})
		require.NoError(t, err)
		assert.Contains(t, html, "PayMob")
	})

	t.Run("InvalidDirectory", func(t *testing.T) {
		_, err := LoadTemplatesFromDir("/nonexistent/path")
		assert.Error(t, err)
	})
}

func TestGetTemplateFS(t *testing.T) {
	fs := GetTemplateFS()
	require.NotNil(t, fs)

	// Verify we can read a template file
	content, err := fs.Open("templates/payment.html")
	require.NoError(t, err)
	content.Close()
}
