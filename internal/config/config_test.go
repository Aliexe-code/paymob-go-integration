package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	cfg := Load()
	assert.NotNil(t, cfg)
}

func TestLoad_WithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("PAYMOB_API_KEY", "test-api-key")
	os.Setenv("PAYMOB_MERCHANT_ID", "test-merchant")
	os.Setenv("PAYMOB_INTEGRATION_ID", "test-integration")
	os.Setenv("PAYMOB_IFRAME_ID", "test-iframe")
	os.Setenv("PAYMOB_HMAC_SECRET", "test-secret")
	os.Setenv("PORT", "8080")
	os.Setenv("SERVER_URL", "https://example.com")
	os.Setenv("DEMO_MODE", "true")

	defer func() {
		os.Unsetenv("PAYMOB_API_KEY")
		os.Unsetenv("PAYMOB_MERCHANT_ID")
		os.Unsetenv("PAYMOB_INTEGRATION_ID")
		os.Unsetenv("PAYMOB_IFRAME_ID")
		os.Unsetenv("PAYMOB_HMAC_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("SERVER_URL")
		os.Unsetenv("DEMO_MODE")
	}()

	cfg := Load()
	assert.Equal(t, "test-api-key", cfg.PayMobAPIKey)
	assert.Equal(t, "test-merchant", cfg.PayMobMerchantID)
	assert.Equal(t, "test-integration", cfg.PayMobIntegrationID)
	assert.Equal(t, "test-iframe", cfg.PayMobIframeID)
	assert.Equal(t, "test-secret", cfg.PayMobHMACSecret)
	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "https://example.com", cfg.ServerURL)
	assert.True(t, cfg.DemoMode)
}

func TestLoad_Defaults(t *testing.T) {
	// Clear all relevant env vars
	os.Unsetenv("PAYMOB_API_KEY")
	os.Unsetenv("PAYMOB_MERCHANT_ID")
	os.Unsetenv("PAYMOB_INTEGRATION_ID")
	os.Unsetenv("PAYMOB_IFRAME_ID")
	os.Unsetenv("PAYMOB_HMAC_SECRET")
	os.Unsetenv("PORT")
	os.Unsetenv("SERVER_URL")
	os.Unsetenv("DEMO_MODE")
	os.Unsetenv("PAYMOB_BASE_URL")

	cfg := Load()
	assert.Equal(t, "https://accept.paymobsolutions.com", cfg.PayMobBaseURL)
	assert.Equal(t, "3000", cfg.ServerPort)
	assert.Equal(t, "http://localhost:3000", cfg.ServerURL)
	assert.False(t, cfg.DemoMode)
}

func TestGetEnv(t *testing.T) {
	t.Run("WithValue", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test-value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnv("TEST_VAR", "default")
		assert.Equal(t, "test-value", result)
	})

	t.Run("WithDefault", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_VAR")
		result := getEnv("NONEXISTENT_VAR", "default-value")
		assert.Equal(t, "default-value", result)
	})

	t.Run("EmptyStringUsesDefault", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		result := getEnv("EMPTY_VAR", "default")
		assert.Equal(t, "default", result)
	})
}

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		PayMobAPIKey:        "key",
		PayMobMerchantID:    "merchant",
		PayMobBaseURL:       "url",
		PayMobIntegrationID: "integration",
		PayMobIframeID:      "iframe",
		ServerPort:          "port",
		ServerURL:           "server",
		PayMobHMACSecret:    "secret",
		DemoMode:            true,
	}

	assert.Equal(t, "key", cfg.PayMobAPIKey)
	assert.Equal(t, "merchant", cfg.PayMobMerchantID)
	assert.Equal(t, "url", cfg.PayMobBaseURL)
	assert.Equal(t, "integration", cfg.PayMobIntegrationID)
	assert.Equal(t, "iframe", cfg.PayMobIframeID)
	assert.Equal(t, "port", cfg.ServerPort)
	assert.Equal(t, "server", cfg.ServerURL)
	assert.Equal(t, "secret", cfg.PayMobHMACSecret)
	assert.True(t, cfg.DemoMode)
}
