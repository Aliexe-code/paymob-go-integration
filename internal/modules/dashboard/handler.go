package dashboard

import (
	"context"
	"paymob-demo/internal/domain"
	"paymob-demo/internal/modules/payment"
	"paymob-demo/internal/views"
	"paymob-demo/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// Handler handles dashboard HTTP requests
type Handler struct {
	repo     *payment.Repository
	renderer *views.Renderer
}

// NewHandler creates a new dashboard handler
func NewHandler(repo *payment.Repository, renderer *views.Renderer) *Handler {
	return &Handler{
		repo:     repo,
		renderer: renderer,
	}
}

// GetDashboard serves the admin dashboard
func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	ctx := context.Background()
	data, err := h.repo.GetDashboardData(ctx)
	if err != nil {
		data = &domain.DashboardData{}
	}

	pageData := domain.DashboardPageData{
		Title:         "Payment Dashboard",
		TotalPayments: data.TotalPayments,
		TotalAmount:   utils.FormatAmount(data.TotalAmount),
		SuccessCount:  data.SuccessCount,
		FailedCount:   data.FailedCount,
		PendingCount:  data.PendingCount,
		RecentPayments: convertToTableRows(data.RecentPayments),
	}

	html, err := h.renderer.RenderDashboard(pageData)
	if err != nil {
		return c.Status(500).SendString("Failed to render dashboard")
	}
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
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

// GetDashboardHTML returns dashboard HTML fragments for HTMX
func (h *Handler) GetDashboardHTML(c *fiber.Ctx) error {
	ctx := context.Background()
	data, err := h.repo.GetDashboardData(ctx)
	if err != nil {
		data = &domain.DashboardData{}
	}

	pageData := domain.DashboardPageData{
		TotalPayments:  data.TotalPayments,
		TotalAmount:    utils.FormatAmount(data.TotalAmount),
		SuccessCount:   data.SuccessCount,
		FailedCount:    data.FailedCount,
		PendingCount:   data.PendingCount,
		RecentPayments: convertToTableRows(data.RecentPayments),
	}

	html, err := h.renderer.RenderDashboardHTML(pageData)
	if err != nil {
		return c.Status(500).SendString("Failed to render dashboard")
	}
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
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
