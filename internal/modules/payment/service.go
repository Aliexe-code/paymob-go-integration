package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"paymob-demo/internal/config"
	"paymob-demo/internal/domain"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Service implements PaymentService for PayMob integration
type Service struct {
	cfg        *config.Config
	httpClient *http.Client
	baseURL    string // Configurable for testing
}

// NewService creates a new payment service
func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: cfg.PayMobBaseURL,
	}
}

// NewServiceWithClient creates a service with custom HTTP client and base URL (for testing)
func NewServiceWithClient(cfg *config.Config, httpClient *http.Client, baseURL string) *Service {
	return &Service{
		cfg:        cfg,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// API response types
type authResponse struct {
	Token string `json:"token"`
}

type orderResponse struct {
	ID int `json:"id"`
}

type paymentKeyResponse struct {
	Token string `json:"token"`
}

// InitiatePayment creates a payment order and gets payment key
func (s *Service) InitiatePayment(ctx context.Context, req domain.PaymentRequest) (*domain.Payment, error) {
	orderID := uuid.New().String()
	amount := req.Amount
	currency := req.Currency
	if currency == "" {
		currency = "EGP"
	}

	// Demo mode
	if s.cfg.DemoMode || s.cfg.PayMobAPIKey == "" {
		return s.createDemoPayment(orderID, amount, currency, req), nil
	}

	// Step 1: Authenticate
	authToken, err := s.authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrAuthenticationFailed, err)
	}

	// Step 2: Create Order
	paymobOrderID, err := s.createOrder(ctx, authToken, orderID, amount*100, currency)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOrderCreationFailed, err)
	}

	// Step 3: Get Payment Key
	paymentKey, err := s.getPaymentKey(ctx, authToken, paymobOrderID, amount*100, currency, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPaymentKeyFailed, err)
	}

	// Build checkout URL
	checkoutURL := s.GetCheckoutURL(paymentKey)

	return &domain.Payment{
		ID:               uuid.New().String(),
		OrderID:          orderID,
		Amount:           amount,
		Currency:         currency,
		Status:           domain.PaymentStatusPending,
		CheckoutURL:      checkoutURL,
		PayMobOrderID:    paymobOrderID,
		PayMobPaymentKey: paymentKey,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

// GetCheckoutURL creates the PayMob checkout iframe URL
func (s *Service) GetCheckoutURL(paymentKey string) string {
	iframeID := s.cfg.PayMobIframeID
	if iframeID == "" {
		iframeID = s.cfg.PayMobIntegrationID
	}
	return fmt.Sprintf("%s/api/acceptance/iframes/%s?payment_token=%s",
		s.baseURL, iframeID, paymentKey)
}

// VerifyWebhookSignature verifies the HMAC signature from PayMob webhook
func (s *Service) VerifyWebhookSignature(signature string, payload []byte) bool {
	if s.cfg.PayMobHMACSecret == "" {
		return true
	}

	mac := hmac.New(sha512.New, []byte(s.cfg.PayMobHMACSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// authenticate gets an auth token from PayMob
func (s *Service) authenticate(ctx context.Context) (string, error) {
	payload := map[string]string{"api_key": s.cfg.PayMobAPIKey}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/auth/tokens", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp authResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.Token == "" {
		return "", fmt.Errorf("no token in auth response: %s", string(body))
	}

	return authResp.Token, nil
}

// createOrder registers an order with PayMob
func (s *Service) createOrder(ctx context.Context, authToken, merchantOrderID string, amountCents int, currency string) (int, error) {
	merchantID, _ := strconv.Atoi(s.cfg.PayMobMerchantID)

	payload := map[string]interface{}{
		"auth_token":        authToken,
		"delivery_needed":   false,
		"merchant_id":       merchantID,
		"amount_cents":      amountCents,
		"currency":          currency,
		"merchant_order_id": merchantOrderID,
		"items":             []interface{}{},
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/ecommerce/orders", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("order creation failed: %s", string(body))
	}

	var orderResp orderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return 0, err
	}

	return orderResp.ID, nil
}

// getPaymentKey generates a payment key for the iframe
func (s *Service) getPaymentKey(ctx context.Context, authToken string, orderID int, amountCents int, currency string, req domain.PaymentRequest) (string, error) {
	integrationID, _ := strconv.Atoi(s.cfg.PayMobIntegrationID)

	payload := map[string]interface{}{
		"auth_token":      authToken,
		"amount_cents":    amountCents,
		"expiration":      3600,
		"order_id":        orderID,
		"billing_data": map[string]string{
			"first_name":   req.Name,
			"last_name":    "Demo",
			"email":        req.Email,
			"phone_number": req.Phone,
			"country":      "EG",
			"city":         "Cairo",
			"street":       "Demo Street",
			"building":     "1",
			"floor":        "1",
			"apartment":    "1",
		},
		"currency":              currency,
		"integration_id":        integrationID,
		"lock_order_when_paid":  true,
		"redirect_url":          s.cfg.ServerURL + "/success",
	}

	jsonData, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/acceptance/payment_keys", bytes.NewReader(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("payment key failed with status %d: %s", resp.StatusCode, string(body))
	}

	var keyResp paymentKeyResponse
	if err := json.Unmarshal(body, &keyResp); err != nil {
		return "", fmt.Errorf("failed to parse payment key response: %w", err)
	}

	if keyResp.Token == "" {
		return "", fmt.Errorf("no token in payment key response: %s", string(body))
	}

	return keyResp.Token, nil
}

// createDemoPayment creates a mock payment for demo mode
func (s *Service) createDemoPayment(orderID string, amount int, currency string, req domain.PaymentRequest) *domain.Payment {
	mockPaymentKey := "demo_" + uuid.New().String()[:8]

	return &domain.Payment{
		ID:               uuid.New().String(),
		OrderID:          orderID,
		Amount:           amount,
		Currency:         currency,
		Status:           domain.PaymentStatusPending,
		CheckoutURL:      fmt.Sprintf("%s/pay/simulate?order_id=%s", s.cfg.ServerURL, orderID),
		PayMobOrderID:    0,
		PayMobPaymentKey: mockPaymentKey,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
