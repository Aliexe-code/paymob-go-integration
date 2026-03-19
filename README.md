# PayMob Go Integration

A production-ready Go payment integration with PayMob (Accept). Features modular monolith architecture, HTMX frontend, comprehensive test coverage, and real-time payment status tracking.

## Features

- **Modular Monolith Architecture** - Clean separation of concerns with domain-driven design
- **Real PayMob Integration** - Production-ready payment processing
- **HTMX Frontend** - Modern, responsive UI with minimal JavaScript
- **Comprehensive Testing** - 78%+ test coverage including mocked API tests
- **Real-time Status Tracking** - Live payment status updates
- **Docker Support** - Containerized deployment ready

## Tech Stack

- **Backend**: Go 1.21+ with Fiber framework
- **Frontend**: HTMX + Tailwind CSS
- **Database**: SQLite (with repository pattern for easy swapping)
- **Payment Gateway**: PayMob (Accept)
- **Testing**: Go test with httptest for API mocking

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PayMob account with API credentials

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Aliexe-code/paymob-go-integration.git
cd paymob-go-integration
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Edit `.env` with your PayMob credentials:
```env
PAYMOB_API_KEY=your_api_key_here
PAYMOB_MERCHANT_ID=your_merchant_id
PAYMOB_INTEGRATION_ID=your_integration_id
PAYMOB_IFRAME_ID=your_iframe_id
```

4. Run the application:
```bash
make run
```

5. Open http://localhost:3000 in your browser

### Docker Deployment

```bash
docker-compose up -d
```

## Project Structure

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/          # Domain models and interfaces
│   ├── modules/         # Business modules
│   │   ├── payment/     # Payment processing
│   │   ├── dashboard/   # Admin dashboard
│   │   └── webhook/     # PayMob webhooks
│   └── views/           # HTML templates
├── pkg/utils/           # Shared utilities
└── tests/               # Test suites
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Payment form |
| `/api/payments` | POST | Create new payment |
| `/api/payments/status` | GET | Check payment status |
| `/dashboard` | GET | Admin dashboard |
| `/api/dashboard/html` | GET | Dashboard data (HTMX) |
| `/webhook` | POST | PayMob webhook |
| `/success` | GET | Payment success callback |
| `/failure` | GET | Payment failure callback |

## Testing

Run all tests:
```bash
make test
```

Run with coverage:
```bash
make test-coverage
```

## Configuration

| Variable | Description | Required |
|----------|-------------|----------|
| `PAYMOB_API_KEY` | Your PayMob API key | Yes |
| `PAYMOB_MERCHANT_ID` | Your merchant ID | Yes |
| `PAYMOB_INTEGRATION_ID` | Integration ID | Yes |
| `PAYMOB_IFRAME_ID` | Iframe ID for card payments | Yes |
| `SERVER_URL` | Your server URL | Yes |
| `DEMO_MODE` | Enable demo mode (true/false) | No |
| `WEBHOOK_SECRET` | Webhook signature secret | No |

## Security Notes

- Never commit `.env` file to version control
- Use HTTPS in production
- Set strong webhook secret for production
- Rotate API keys regularly

## License

MIT License - see LICENSE file for details

## Contributing

Pull requests are welcome. For major changes, please open an issue first.

## Support

For PayMob API documentation, visit: https://docs.paymob.com
