package config

import "os"

// Config holds application configuration
type Config struct {
	// PayMob API Configuration
	PayMobAPIKey       string
	PayMobMerchantID   string
	PayMobBaseURL      string
	PayMobIntegrationID string
	PayMobIframeID     string

	// Server Configuration
	ServerPort string
	ServerURL  string

	// Webhook Security
	PayMobHMACSecret string

	// Demo Mode
	DemoMode bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		PayMobAPIKey:        getEnv("PAYMOB_API_KEY", ""),
		PayMobMerchantID:    getEnv("PAYMOB_MERCHANT_ID", ""),
		PayMobBaseURL:       getEnv("PAYMOB_BASE_URL", "https://accept.paymobsolutions.com"),
		PayMobIntegrationID: getEnv("PAYMOB_INTEGRATION_ID", ""),
		PayMobIframeID:      getEnv("PAYMOB_IFRAME_ID", ""),
		ServerPort:          getEnv("PORT", "3000"),
		ServerURL:           getEnv("SERVER_URL", "http://localhost:3000"),
		PayMobHMACSecret:    getEnv("PAYMOB_HMAC_SECRET", ""),
		DemoMode:            getEnv("DEMO_MODE", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
