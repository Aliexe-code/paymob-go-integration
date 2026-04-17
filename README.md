# PayMob Go Integration

A Go backend payment integration with PayMob, featuring a clean API that can be used with any frontend (React, Next.js, etc.) or with the built-in HTMX demo UI.

## Features

- **API-First Design** - Clean JSON API, frontend-agnostic (use with React, Next.js, Vue, or any frontend)
- **Optional HTMX Demo UI** - Built-in web frontend for quick demos and testing
- **Real PayMob Integration** - Production-ready payment processing with authentication, order creation, and payment key generation
- **Webhook Support** - HMAC signature verification for secure webhook callbacks
- **SQLite Persistence** - Fast, lightweight database with WAL mode for concurrent access
- **Dashboard** - Admin panel with payment statistics and recent transactions

## Tech Stack

- **Backend**: Go 1.21+ with Fiber framework
- **Frontend**: HTMX + Tailwind CSS (optional, demo only)
- **Database**: SQLite with WAL mode
- **Payment Gateway**: PayMob

## Quick Start

### Prerequisites

- Go 1.21+
- PayMob account with API credentials (or use demo mode)

### Run with Web UI (Demo)

```bash
git clone https://github.com/Aliexe-code/paymob-go-integration.git
cd paymob-go-integration
cp .env.example .env
# Edit .env with your credentials
make run-web
```

Open http://localhost:3000 in your browser.

### Run API Only (No Web UI)

```bash
make run
```

The API runs on port 3000. Web routes (`/`, `/dashboard`, `/success`, `/failure`) return 404 when built without the web tag.

### Demo Mode

Set `DEMO_MODE=true` in `.env` to skip real PayMob API calls and test locally with simulated payments.

## Project Structure

```
.
├── api/                          # Go backend (main module)
│   ├── cmd/server/
│   │   ├── main.go               # API-only entry point
│   │   └── main_web.go           # Entry point with web UI (build tag: web)
│   ├── internal/
│   │   ├── config/               # Environment configuration
│   │   ├── domain/               # Models, interfaces, errors
│   │   ├── modules/
│   │   │   ├── payment/          # Payment processing (handler.go / handler_web.go)
│   │   │   ├── dashboard/        # Admin dashboard
│   │   │   └── webhook/          # PayMob webhook handling
│   │   └── views/                # HTML templates (embedded)
│   ├── pkg/utils/                # Utility functions
│   └── tests/                    # Unit and integration tests
├── web/                          # Standalone HTMX frontend (optional)
│   ├── cmd/server/               # Web server that proxies to API
│   └── templates/                # HTML templates (source of truth)
└── PROGRESS.md                   # Development progress
```

## Build Tags

The project uses Go build tags to separate the API from the web frontend:

| Command | Description |
|---------|-------------|
| `make build` / `make run` | API-only binary (JSON responses) |
| `make build-web` / `make run-web` | API + HTMX web frontend |
| `make dev` | Run API-only in dev mode |
| `make dev-web` | Run with web frontend in dev mode |

## API Endpoints

### Payments

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/payments` | POST | Create new payment (form-urlencoded) |
| `/api/payments/status` | GET | Check payment status by order_id |
| `/api/payments/paymob-status` | GET | Query PayMob API for transaction status |

### Dashboard

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/dashboard` | GET | Dashboard data (JSON) |
| `/api/dashboard/html` | GET | Dashboard HTML fragment (HTMX) |

### Webhook

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/webhook` | POST | PayMob webhook callback |

### Simulation (Demo Mode)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/simulate/:order_id` | POST | Simulate successful payment |
| `/api/simulate-failure/:order_id` | POST | Simulate failed payment |

### Utility

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |

## Web Routes (with `-tags web`)

| Route | Description |
|-------|-------------|
| `/` | Payment form page |
| `/success` | Payment success callback page |
| `/failure` | Payment failure callback page |
| `/dashboard` | Admin dashboard |
| `/pay/simulate` | Demo payment simulation page |

## Configuration

| Variable | Description | Required |
|----------|-------------|----------|
| `PAYMOB_API_KEY` | Your PayMob API key | Yes |
| `PAYMOB_MERCHANT_ID` | Your merchant ID | Yes |
| `PAYMOB_INTEGRATION_ID` | Integration ID | Yes |
| `PAYMOB_IFRAME_ID` | Iframe ID for card payments | Yes |
| `PAYMOB_BASE_URL` | PayMob API base URL | No (defaults to production) |
| `PAYMOB_HMAC_SECRET` | Webhook signature secret | No |
| `SERVER_URL` | Your server URL (for callbacks) | Yes |
| `SERVER_PORT` | Server port | No (default: 3000) |
| `DEMO_MODE` | Enable demo mode (`true`/`false`) | No (default: false) |
| `TEMPLATES_DIR` | Load templates from filesystem | No (uses embedded) |

## Using as API-Only (Your Own Frontend)

Build the API-only binary and integrate with your frontend:

```bash
make run
```

Your frontend (React, Next.js, etc.) makes requests to `/api/*` endpoints. The payment flow:

1. POST to `/api/payments` with amount, name, email, phone
2. Response contains `checkout_url` - redirect user to PayMob checkout
3. User completes payment on PayMob
4. PayMob redirects to `/success` or `/failure` (callback URLs)
5. PayMob sends webhook to `/api/webhook` (server-to-server)
6. Check payment status via GET `/api/payments/status?order_id=xxx`

## Testing

```bash
make test              # Run all tests
make test-coverage     # Run with coverage report
make bench             # Run benchmarks
```

## Standalone Web Frontend

The `web/` directory contains a standalone HTMX frontend that can run independently:

```bash
# Start API (port 3000)
make run

# In another terminal, start web frontend (port 3001)
cd web && API_URL=http://localhost:3000 go run ./cmd/server/
```

Or use runtime templates (no rebuild needed):

```bash
cd api && TEMPLATES_DIR=../web/templates go run -tags web ./cmd/server/
```

## License

MIT License
