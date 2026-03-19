package domain

import "context"

// PaymentRepository defines the interface for payment persistence
type PaymentRepository interface {
	Add(ctx context.Context, payment *Payment) error
	Get(ctx context.Context, id string) (*Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
	GetAll(ctx context.Context) ([]*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	GetDashboardData(ctx context.Context) (*DashboardData, error)
	Close() error
}

// PaymentService defines the interface for payment business logic
type PaymentService interface {
	InitiatePayment(ctx context.Context, req PaymentRequest) (*Payment, error)
	GetCheckoutURL(paymentKey string) string
	VerifyWebhookSignature(signature string, payload []byte) bool
}

// DashboardService defines the interface for dashboard operations
type DashboardService interface {
	GetDashboardData(ctx context.Context) (*DashboardData, error)
}

// TemplateRenderer defines the interface for template rendering
type TemplateRenderer interface {
	RenderPaymentPage(data PaymentPageData) (string, error)
	RenderPaymentResult(data PaymentResultData) (string, error)
	RenderDashboard(data DashboardPageData) (string, error)
	RenderDashboardHTML(data DashboardPageData) (string, error)
	RenderSuccessPage(data ResultPageData) (string, error)
	RenderFailurePage(data ResultPageData) (string, error)
	RenderSimulatePage(data SimulatePageData) (string, error)
}

// PaymentPageData contains data for the payment page
type PaymentPageData struct {
	Title   string
	APIURL  string
}

// PaymentResultData contains data for payment result fragment
type PaymentResultData struct {
	Success     bool
	Message     string
	Amount      int
	Currency    string
	CheckoutURL string
	OrderID     string
}

// DashboardPageData contains data for the dashboard page
type DashboardPageData struct {
	Title          string
	TotalPayments  int
	TotalAmount    string
	SuccessCount   int
	FailedCount    int
	PendingCount   int
	RecentPayments []PaymentTableRow
}

// PaymentTableRow represents a row in the payments table
type PaymentTableRow struct {
	OrderID    string
	Amount     string
	Currency   string
	Status     string
	StatusText string
	CreatedAt  string
}

// ResultPageData contains data for result pages
type ResultPageData struct {
	Title   string
	Message string
	OrderID string
}

// SimulatePageData contains data for the simulate payment page
type SimulatePageData struct {
	Amount   int
	Currency string
	OrderID  string
	Status   string
}
