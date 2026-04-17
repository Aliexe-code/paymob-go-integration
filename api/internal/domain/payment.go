package domain

import "time"

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSuccess   PaymentStatus = "success"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment represents a payment transaction
type Payment struct {
	ID               string        `json:"id"`
	OrderID          string        `json:"order_id"`
	Amount           int           `json:"amount"`
	Currency         string        `json:"currency"`
	Status           PaymentStatus `json:"status"`
	CheckoutURL      string        `json:"checkout_url,omitempty"`
	PayMobOrderID    int           `json:"paymob_order_id,omitempty"`
	PayMobPaymentKey string        `json:"paymob_payment_key,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	TransactionID    string        `json:"transaction_id,omitempty"`
}

// PaymentRequest represents a request to initiate a payment
type PaymentRequest struct {
	Amount   int    `json:"amount" form:"amount"`
	Currency string `json:"currency" form:"currency"`
	Email    string `json:"email" form:"email"`
	Name     string `json:"name" form:"name"`
	Phone    string `json:"phone" form:"phone"`
}

// PaymentResponse represents the response after initiating a payment
type PaymentResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	OrderID     string `json:"order_id,omitempty"`
}

// DashboardData represents data for the dashboard
type DashboardData struct {
	TotalPayments  int       `json:"total_payments"`
	TotalAmount    int       `json:"total_amount"`
	SuccessCount   int       `json:"success_count"`
	FailedCount    int       `json:"failed_count"`
	PendingCount   int       `json:"pending_count"`
	RecentPayments []Payment `json:"recent_payments"`
}

// WebhookPayload represents the payload received from PayMob webhook
type WebhookPayload struct {
	Type string         `json:"type"`
	Obj  WebhookObject `json:"obj"`
}

// WebhookObject contains the webhook event details
type WebhookObject struct {
	ID           int               `json:"id"`
	Order        WebhookOrder     `json:"order"`
	Success      bool             `json:"success"`
	Pending      bool             `json:"pending"`
	IsRefund     bool             `json:"is_refund"`
	SourceData   WebhookSourceData `json:"source_data"`
	AmountCents  int              `json:"amount_cents"`
	Currency     string           `json:"currency"`
	CreatedAt    string           `json:"created_at"`
	ErrorMessage string           `json:"error_message"`
}

// WebhookOrder contains order details from webhook
type WebhookOrder struct {
	ID              int    `json:"id"`
	MerchantOrderID string `json:"merchant_order_id"`
}

// WebhookSourceData contains payment source information
type WebhookSourceData struct {
	Type     string `json:"type"`
	SubType  string `json:"sub_type"`
	Pan      string `json:"pan"`
	CardType string `json:"card_type"`
}

// LegacyWebhookPayload for backward compatibility with older PayMob format
type LegacyWebhookPayload struct {
	OrderID         int    `json:"order_id"`
	AmountCents     int    `json:"amount_cents"`
	MerchantOrderID string `json:"merchant_order_id"`
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	TransactionID   int    `json:"transaction_id"`
	ErrorOccured    bool   `json:"error_occured"`
}

// PayMobAuthResponse represents PayMob authentication response
type PayMobAuthResponse struct {
	Token string `json:"token"`
}

// PayMobOrderResponse represents PayMob order registration response
type PayMobOrderResponse struct {
	ID int `json:"id"`
}

// PayMobPaymentKeyResponse represents PayMob payment key response
type PayMobPaymentKeyResponse struct {
	Token string `json:"token"`
}
