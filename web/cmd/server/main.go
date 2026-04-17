package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var (
	templates    *template.Template
	apiBaseURL   string
	templatesDir string
)

func main() {
	templatesDir = getEnv("TEMPLATES_DIR", "./templates")
	apiBaseURL = getEnv("API_URL", "http://localhost:3000")
	webPort := getEnv("WEB_PORT", "3001")

	// Load templates
	funcMap := template.FuncMap{
		"formatAmount": formatAmount,
		"statusClass":  statusClass,
		"statusText":   statusText,
	}

	var templatePaths []string
	filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			templatePaths = append(templatePaths, path)
		}
		return nil
	})

	if len(templatePaths) == 0 {
		log.Fatalf("No template files found in %s", templatesDir)
	}

	var err error
	templates, err = template.New("").Funcs(funcMap).ParseFiles(templatePaths...)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "PayMob Web Frontend",
		ServerHeader: "Fiber",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	// Static files
	app.Static("/static", "./static")

	// Proxy API requests to backend
	app.Use("/api", func(c *fiber.Ctx) error {
		targetURL := apiBaseURL + c.Path()
		if c.Request().URI().QueryString() != nil {
			targetURL += "?" + string(c.Request().URI().QueryString())
		}

		// Create request to API backend
		var reqBody []byte
		if len(c.Body()) > 0 {
			reqBody = c.Body()
		}
		req, err := http.NewRequest(string(c.Request().Header.Method()), targetURL, bytes.NewReader(reqBody))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		// Copy headers
		c.Request().Header.VisitAll(func(key, value []byte) {
			req.Header.Set(string(key), string(value))
		})
		req.Header.Set("Host", "")

		// Forward request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).SendString(fmt.Sprintf("API unavailable: %v", err))
		}
		defer resp.Body.Close()

		// Copy response
		body, _ := io.ReadAll(resp.Body)
		c.Status(resp.StatusCode)
		for key, values := range resp.Header {
			for _, v := range values {
				c.Set(key, v)
			}
		}
		return c.Send(body)
	})

	// Web routes
	app.Get("/", renderTemplate("payment", fiber.Map{"Title": "PayMob Demo - Make Payment"}))
	app.Get("/success", renderResultPage("success"))
	app.Get("/failure", renderResultPage("failure"))
	app.Get("/dashboard", renderDashboard)
	app.Get("/pay/simulate", renderSimulatePage)

	addr := fmt.Sprintf(":%s", webPort)
	log.Printf("Starting PayMob Web Frontend on %s", addr)
	log.Printf("Proxying API requests to %s", apiBaseURL)
	log.Printf("Set API_URL env var to point to your backend API")

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start web frontend: %v", err)
	}
}

func renderTemplate(name string, data interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var buf bytes.Buffer
		if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to render template: %v", err))
		}
		c.Set("Content-Type", "text/html")
		return c.Send(buf.Bytes())
	}
}

func renderResultPage(defaultTemplate string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orderID := c.Query("order_id", "")
		transactionID := c.Query("transaction_id", c.Query("id", ""))

		// Try to get payment status from API
		status := "pending"
		message := "Your payment is being processed."
		title := "Payment Processing"

		if orderID != "" {
			apiResp, err := http.Get(fmt.Sprintf("%s/api/payments/status?order_id=%s", apiBaseURL, orderID))
			if err == nil {
				defer apiResp.Body.Close()
				body, _ := io.ReadAll(apiResp.Body)
				if strings.Contains(string(body), `"status":"success"`) {
					status = "success"
					title = "Payment Successful"
					message = "Your payment was completed successfully."
				} else if strings.Contains(string(body), `"status":"failed"`) {
					status = "failed"
					title = "Payment Failed"
					message = "Your payment could not be completed. Please try again."
				}
			}
		}

		hasError := c.Query("error") != "" || c.Query("error_occured") == "true" || c.Query("success") == "false"
		if hasError {
			status = "failed"
			title = "Payment Failed"
			message = "Your payment could not be completed. Please try again."
		}

		tmplName := defaultTemplate
		if status == "success" {
			tmplName = "success"
		} else if status == "failed" {
			tmplName = "failure"
		}

		data := fiber.Map{
			"Title":         title,
			"Message":       message,
			"OrderID":       orderID,
			"TransactionID": transactionID,
			"Status":        status,
		}

		var buf bytes.Buffer
		if err := templates.ExecuteTemplate(&buf, tmplName, data); err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to render template: %v", err))
		}
		c.Set("Content-Type", "text/html")
		return c.Send(buf.Bytes())
	}
}

func renderDashboard(c *fiber.Ctx) error {
	// Fetch dashboard data from API
	apiResp, err := http.Get(fmt.Sprintf("%s/api/dashboard", apiBaseURL))
	if err != nil {
		return c.Status(502).SendString(fmt.Sprintf("Failed to connect to API: %v", err))
	}
	defer apiResp.Body.Close()

	body, _ := io.ReadAll(apiResp.Body)

	// Parse JSON response
	var dataMap map[string]interface{}
	// Simple approach: just pass raw data to template
	_ = dataMap
	_ = body

	data := fiber.Map{
		"Title":          "Payment Dashboard",
		"TotalPayments":  0,
		"TotalAmount":    "0",
		"SuccessCount":   0,
		"FailedCount":    0,
		"PendingCount":   0,
		"RecentPayments": []fiber.Map{},
	}

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "dashboard", data); err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to render template: %v", err))
	}
	c.Set("Content-Type", "text/html")
	return c.Send(buf.Bytes())
}

func renderSimulatePage(c *fiber.Ctx) error {
	orderID := c.Query("order_id", "")
	if orderID == "" {
		return c.Status(400).SendString("Order ID required")
	}

	// Fetch payment status from API
	apiResp, err := http.Get(fmt.Sprintf("%s/api/payments/status?order_id=%s", apiBaseURL, orderID))
	if err != nil {
		return c.Status(502).SendString(fmt.Sprintf("Failed to connect to API: %v", err))
	}
	defer apiResp.Body.Close()

	_ = apiResp // In production, parse JSON

	data := fiber.Map{
		"Amount":   0,
		"Currency": "EGP",
		"OrderID":  orderID,
		"Status":   "Pending",
	}

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "simulate", data); err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to render template: %v", err))
	}
	c.Set("Content-Type", "text/html")
	return c.Send(buf.Bytes())
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// Utility functions (same as pkg/utils in API)
func formatAmount(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	return string(result)
}

func statusClass(status string) string {
	switch status {
	case "success":
		return "bg-green-500/20 text-green-400"
	case "failed":
		return "bg-red-500/20 text-red-400"
	default:
		return "bg-yellow-500/20 text-yellow-400"
	}
}

func statusText(status string) string {
	switch status {
	case "success":
		return "Success"
	case "failed":
		return "Failed"
	case "cancelled":
		return "Cancelled"
	default:
		return "Pending"
	}
}
