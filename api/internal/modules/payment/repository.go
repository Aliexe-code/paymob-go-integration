package payment

import (
	"context"
	"database/sql"
	"paymob-demo/internal/domain"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db               *sql.DB
	stmtAdd          *sql.Stmt
	stmtGetByID      *sql.Stmt
	stmtGetByOrderID *sql.Stmt
	stmtUpdate       *sql.Stmt
	stmtRecent       *sql.Stmt
	cacheMu          sync.RWMutex
	dashboardCache   *domain.DashboardData
	cacheExpiry      time.Time
	cacheDuration    time.Duration
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, err
	}

	db.Exec(`PRAGMA synchronous=NORMAL`)
	db.Exec(`PRAGMA cache_size=10000`)
	db.Exec(`PRAGMA temp_store=MEMORY`)

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			order_id TEXT UNIQUE NOT NULL,
			amount INTEGER NOT NULL,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			checkout_url TEXT,
			paymob_order_id INTEGER,
			paymob_payment_key TEXT,
			transaction_id TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`); err != nil {
		return nil, err
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC)`)
	stmtAdd, _ := db.Prepare(`INSERT INTO payments (id, order_id, amount, currency, status, checkout_url, paymob_order_id, paymob_payment_key, transaction_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	stmtGetByID, _ := db.Prepare(`SELECT id, order_id, amount, currency, status, checkout_url, paymob_order_id, paymob_payment_key, transaction_id, created_at, updated_at FROM payments WHERE id = ?`)
	stmtGetByOrderID, _ := db.Prepare(`SELECT id, order_id, amount, currency, status, checkout_url, paymob_order_id, paymob_payment_key, transaction_id, created_at, updated_at FROM payments WHERE order_id = ?`)
	stmtUpdate, _ := db.Prepare(`UPDATE payments SET amount = ?, currency = ?, status = ?, checkout_url = ?, paymob_order_id = ?, paymob_payment_key = ?, transaction_id = ?, updated_at = ? WHERE id = ?`)
	stmtRecent, _ := db.Prepare(`SELECT id, order_id, amount, currency, status, checkout_url, paymob_order_id, paymob_payment_key, transaction_id, created_at, updated_at FROM payments ORDER BY created_at DESC LIMIT 10`)

	return &Repository{
		db:            db,
		stmtAdd:       stmtAdd,
		stmtGetByID:   stmtGetByID,
		stmtGetByOrderID: stmtGetByOrderID,
		stmtUpdate:    stmtUpdate,
		stmtRecent:    stmtRecent,
		cacheDuration: 500 * time.Millisecond,
	}, nil
}

// Add adds a new payment
func (r *Repository) Add(ctx context.Context, payment *domain.Payment) error {
	_, err := r.stmtAdd.ExecContext(ctx, payment.ID, payment.OrderID, payment.Amount, payment.Currency, payment.Status,
		payment.CheckoutURL, payment.PayMobOrderID, payment.PayMobPaymentKey, payment.TransactionID,
		payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return err
	}
	r.invalidateCache()
	return nil
}

// Get retrieves a payment by ID
func (r *Repository) Get(ctx context.Context, id string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	err := r.stmtGetByID.QueryRowContext(ctx, id).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Status,
		&payment.CheckoutURL, &payment.PayMobOrderID, &payment.PayMobPaymentKey, &payment.TransactionID,
		&payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	err := r.stmtGetByOrderID.QueryRowContext(ctx, orderID).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Status,
		&payment.CheckoutURL, &payment.PayMobOrderID, &payment.PayMobPaymentKey, &payment.TransactionID,
		&payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, order_id, amount, currency, status, checkout_url, paymob_order_id, paymob_payment_key, transaction_id, created_at, updated_at FROM payments ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		if err := rows.Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Status,
			&payment.CheckoutURL, &payment.PayMobOrderID, &payment.PayMobPaymentKey, &payment.TransactionID,
			&payment.CreatedAt, &payment.UpdatedAt); err != nil {
			continue
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

func (r *Repository) Update(ctx context.Context, payment *domain.Payment) error {
	payment.UpdatedAt = time.Now()
	_, err := r.stmtUpdate.ExecContext(ctx, payment.Amount, payment.Currency, payment.Status, payment.CheckoutURL,
		payment.PayMobOrderID, payment.PayMobPaymentKey, payment.TransactionID, payment.UpdatedAt, payment.ID)
	if err != nil {
		return err
	}
	r.invalidateCache()
	return nil
}

func (r *Repository) GetDashboardData(ctx context.Context) (*domain.DashboardData, error) {
	r.cacheMu.RLock()
	if r.dashboardCache != nil && time.Now().Before(r.cacheExpiry) {
		data := r.dashboardCache
		r.cacheMu.RUnlock()
		return data, nil
	}
	r.cacheMu.RUnlock()

	data := &domain.DashboardData{
		RecentPayments: make([]domain.Payment, 0),
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(amount), 0) as total_amount,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count
		FROM payments
	`).Scan(&data.TotalPayments, &data.TotalAmount, &data.SuccessCount, &data.FailedCount, &data.PendingCount)

	if err != nil {
		data.TotalPayments = 0
		data.TotalAmount = 0
	}

	rows, err := r.stmtRecent.QueryContext(ctx)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var payment domain.Payment
			if err := rows.Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Currency, &payment.Status,
				&payment.CheckoutURL, &payment.PayMobOrderID, &payment.PayMobPaymentKey, &payment.TransactionID,
				&payment.CreatedAt, &payment.UpdatedAt); err != nil {
				continue
			}
			data.RecentPayments = append(data.RecentPayments, payment)
		}
	}

	r.cacheMu.Lock()
	r.dashboardCache = data
	r.cacheExpiry = time.Now().Add(r.cacheDuration)
	r.cacheMu.Unlock()

	return data, nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	r.stmtAdd.Close()
	r.stmtGetByID.Close()
	r.stmtGetByOrderID.Close()
	r.stmtUpdate.Close()
	r.stmtRecent.Close()
	return r.db.Close()
}

func (r *Repository) invalidateCache() {
	r.cacheMu.Lock()
	r.cacheExpiry = time.Time{}
	r.cacheMu.Unlock()
}
