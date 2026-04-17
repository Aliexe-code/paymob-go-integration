package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"paymob-demo/internal/domain"
	"paymob-demo/pkg/utils"
	"path/filepath"
	"sync"
)

//go:embed templates/*.html templates/partials/*.html
var templateFS embed.FS

// Renderer handles template rendering
type Renderer struct {
	templates *template.Template
	mu        sync.RWMutex
}

// NewRenderer creates a new template renderer.
// If TEMPLATES_DIR env var is set, templates are loaded from that directory at runtime.
// Otherwise, templates are loaded from the embedded filesystem.
func NewRenderer() (*Renderer, error) {
	if dir := os.Getenv("TEMPLATES_DIR"); dir != "" {
		return LoadTemplatesFromDir(dir)
	}

	funcs := templateFuncs()

	// Parse all templates at once to handle dependencies
	tmpl, err := template.New("").Funcs(funcs).ParseFS(templateFS,
		"templates/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Renderer{templates: tmpl}, nil
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatAmount": utils.FormatAmount,
		"statusClass":  utils.StatusClass,
		"statusText":   utils.StatusText,
	}
}

// Render renders a template with the given data
func (r *Renderer) Render(name string, data interface{}) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var buf bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderPaymentPage renders the payment form page
func (r *Renderer) RenderPaymentPage(data domain.PaymentPageData) (string, error) {
	return r.Render("payment", data)
}

// RenderPaymentResult renders the payment result fragment
func (r *Renderer) RenderPaymentResult(data domain.PaymentResultData) (string, error) {
	return r.Render("payment_result", data)
}

// RenderDashboard renders the full dashboard page
func (r *Renderer) RenderDashboard(data domain.DashboardPageData) (string, error) {
	return r.Render("dashboard", data)
}

// RenderDashboardHTML renders the dashboard HTML fragment for HTMX
func (r *Renderer) RenderDashboardHTML(data domain.DashboardPageData) (string, error) {
	return r.Render("dashboard_html", data)
}

// RenderSuccessPage renders the payment success page
func (r *Renderer) RenderSuccessPage(data domain.ResultPageData) (string, error) {
	return r.Render("success", data)
}

// RenderFailurePage renders the payment failure page
func (r *Renderer) RenderFailurePage(data domain.ResultPageData) (string, error) {
	return r.Render("failure", data)
}

// RenderSimulatePage renders the demo simulate payment page
func (r *Renderer) RenderSimulatePage(data domain.SimulatePageData) (string, error) {
	return r.Render("simulate", data)
}

// RenderPaymentRow renders a single payment row for HTMX updates
func (r *Renderer) RenderPaymentRow(data domain.PaymentTableRow) (string, error) {
	return r.Render("payment_row", data)
}

// LoadTemplatesFromDir loads templates from a directory (for development/runtime)
func LoadTemplatesFromDir(dir string) (*Renderer, error) {
	funcs := templateFuncs()

	// Find all HTML files in the directory and subdirectories
	var templatePaths []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			templatePaths = append(templatePaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk templates directory: %w", err)
	}

	if len(templatePaths) == 0 {
		return nil, fmt.Errorf("no template files found in %s", dir)
	}

	tmpl, err := template.New("").Funcs(funcs).ParseFiles(templatePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Renderer{templates: tmpl}, nil
}

// GetTemplateFS returns the embedded filesystem for external use
func GetTemplateFS() fs.FS {
	return templateFS
}