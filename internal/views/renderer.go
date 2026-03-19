package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"paymob-demo/internal/domain"
	"sync"
)

//go:embed templates/*.html templates/partials/*.html
var templateFS embed.FS

// Renderer handles template rendering
type Renderer struct {
	templates *template.Template
	mu        sync.RWMutex
}

// NewRenderer creates a new template renderer
func NewRenderer() (*Renderer, error) {
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
		"formatAmount": formatAmount,
		"statusClass":  statusClass,
		"statusText":   statusText,
	}
}

// formatAmount formats an integer with comma separators
func formatAmount(n int) string {
	s := fmt.Sprintf("%d", n)
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

// statusClass returns the CSS class for a status
func statusClass(status string) string {
	switch status {
	case string(domain.PaymentStatusSuccess):
		return "bg-green-500/20 text-green-400"
	case string(domain.PaymentStatusFailed):
		return "bg-red-500/20 text-red-400"
	default:
		return "bg-yellow-500/20 text-yellow-400"
	}
}

// statusText returns the display text for a status
func statusText(status string) string {
	switch status {
	case string(domain.PaymentStatusSuccess):
		return "Success"
	case string(domain.PaymentStatusFailed):
		return "Failed"
	case string(domain.PaymentStatusCancelled):
		return "Cancelled"
	default:
		return "Pending"
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

// LoadTemplatesFromDir loads templates from a directory (for development)
func LoadTemplatesFromDir(dir string) (*Renderer, error) {
	funcs := templateFuncs()

	tmpl, err := template.New("").Funcs(funcs).ParseGlob(dir + "/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Parse partials
	partials, err := template.New("").Funcs(funcs).ParseGlob(dir + "/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse partials: %w", err)
	}

	// Merge templates
	for _, t := range partials.Templates() {
		_, err := tmpl.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return nil, fmt.Errorf("failed to add partial %s: %w", t.Name(), err)
		}
	}

	return &Renderer{templates: tmpl}, nil
}

// GetTemplateFS returns the embedded filesystem for external use
func GetTemplateFS() fs.FS {
	return templateFS
}