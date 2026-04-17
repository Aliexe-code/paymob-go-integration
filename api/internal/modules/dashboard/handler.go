//go:build !web

package dashboard

import (
	"context"
	"paymob-demo/internal/domain"
	"paymob-demo/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// Handler handles dashboard HTTP requests.
// In API-only mode, only JSON endpoints are available.
type Handler struct {
	repo domain.PaymentRepository
}

// NewHandler creates a new dashboard handler (API-only mode).
func NewHandler(repo domain.PaymentRepository) *Handler {
	return &Handler{repo: repo}
}

// GetDashboard serves the admin dashboard page.
// Returns 404 in API-only mode.
func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "web frontend not enabled (build with -tags web)",
	})
}

// GetDashboardData returns dashboard data as JSON
func (h *Handler) GetDashboardData(c *fiber.Ctx) error {
	ctx := context.Background()
	data, err := h.repo.GetDashboardData(ctx)
	if err != nil {
		data = &domain.DashboardData{}
	}

	type formattedPayment struct {
		ID        string `json:"id"`
		OrderID   string `json:"order_id"`
		Amount    string `json:"amount"`
		Currency  string `json:"currency"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}

	var formatted []formattedPayment
	for _, p := range data.RecentPayments {
		formatted = append(formatted, formattedPayment{
			ID:        p.ID,
			OrderID:   p.OrderID,
			Amount:    utils.FormatAmount(p.Amount),
			Currency:  p.Currency,
			Status:    string(p.Status),
			CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(fiber.Map{
		"total_payments":  data.TotalPayments,
		"total_amount":    utils.FormatAmount(data.TotalAmount),
		"success_count":   data.SuccessCount,
		"failed_count":    data.FailedCount,
		"pending_count":   data.PendingCount,
		"recent_payments": formatted,
	})
}

// GetDashboardHTML returns dashboard HTML fragments for HTMX.
// Returns 404 in API-only mode.
func (h *Handler) GetDashboardHTML(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "web frontend not enabled (build with -tags web)",
	})
}

func convertToTableRows(payments []domain.Payment) []domain.PaymentTableRow {
	rows := make([]domain.PaymentTableRow, 0, len(payments))
	for _, p := range payments {
		rows = append(rows, domain.PaymentTableRow{
			OrderID:    p.OrderID,
			Amount:     utils.FormatAmount(p.Amount),
			Currency:   p.Currency,
			Status:     string(p.Status),
			StatusText: utils.StatusText(string(p.Status)),
			CreatedAt:  p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return rows
}
