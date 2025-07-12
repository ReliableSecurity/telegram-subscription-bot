package models

import (
	"database/sql"
	"time"
)

type Payment struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	PlanID          int64     `json:"plan_id" db:"plan_id"`
	Amount          int       `json:"amount" db:"amount"`
	Currency        string    `json:"currency" db:"currency"`
	PaymentMethod   string    `json:"payment_method" db:"payment_method"`
	PaymentProvider string    `json:"payment_provider" db:"payment_provider"`
	TransactionID   string    `json:"transaction_id" db:"transaction_id"`
	Status          string    `json:"status" db:"status"`
	Description     string    `json:"description" db:"description"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	CompletedAt     time.Time `json:"completed_at" db:"completed_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(payment *Payment) error {
	query := `
		INSERT INTO payments (user_id, plan_id, amount, currency, payment_method, payment_provider, transaction_id, status, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		payment.UserID,
		payment.PlanID,
		payment.Amount,
		payment.Currency,
		payment.PaymentMethod,
		payment.PaymentProvider,
		payment.TransactionID,
		payment.Status,
		payment.Description,
		payment.CreatedAt,
		payment.UpdatedAt,
	).Scan(&payment.ID)
	
	return err
}

func (r *PaymentRepository) GetByID(id int64) (*Payment, error) {
	query := `
		SELECT id, user_id, plan_id, amount, currency, payment_method, payment_provider, transaction_id, status, description, created_at, completed_at, updated_at
		FROM payments
		WHERE id = $1
	`
	
	payment := &Payment{}
	err := r.db.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.UserID,
		&payment.PlanID,
		&payment.Amount,
		&payment.Currency,
		&payment.PaymentMethod,
		&payment.PaymentProvider,
		&payment.TransactionID,
		&payment.Status,
		&payment.Description,
		&payment.CreatedAt,
		&payment.CompletedAt,
		&payment.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return payment, nil
}

func (r *PaymentRepository) GetByUserID(userID int64) ([]*Payment, error) {
	query := `
		SELECT id, user_id, plan_id, amount, currency, payment_method, payment_provider, transaction_id, status, description, created_at, completed_at, updated_at
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var payments []*Payment
	for rows.Next() {
		payment := &Payment{}
		err := rows.Scan(
			&payment.ID,
			&payment.UserID,
			&payment.PlanID,
			&payment.Amount,
			&payment.Currency,
			&payment.PaymentMethod,
			&payment.PaymentProvider,
			&payment.TransactionID,
			&payment.Status,
			&payment.Description,
			&payment.CreatedAt,
			&payment.CompletedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	
	return payments, nil
}

func (r *PaymentRepository) Update(payment *Payment) error {
	query := `
		UPDATE payments
		SET plan_id = $2, amount = $3, currency = $4, payment_method = $5, payment_provider = $6, transaction_id = $7, status = $8, description = $9, completed_at = $10, updated_at = $11
		WHERE id = $1
	`
	
	payment.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		payment.ID,
		payment.PlanID,
		payment.Amount,
		payment.Currency,
		payment.PaymentMethod,
		payment.PaymentProvider,
		payment.TransactionID,
		payment.Status,
		payment.Description,
		payment.CompletedAt,
		payment.UpdatedAt,
	)
	
	return err
}

func (r *PaymentRepository) GetPlanByID(planID int64) (*SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price_cents, duration_days, currency, is_active, max_groups, features, created_at
		FROM subscription_plans
		WHERE id = $1 AND is_active = true
	`
	
	plan := &SubscriptionPlan{}
	err := r.db.QueryRow(query, planID).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&plan.PriceCents,
		&plan.DurationDays,
		&plan.Currency,
		&plan.IsActive,
		&plan.MaxGroups,
		&plan.Features,
		&plan.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return plan, nil
}

func (r *PaymentRepository) GetAllPlans() ([]*SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price_cents, duration_days, currency, is_active, max_groups, features, created_at
		FROM subscription_plans
		WHERE is_active = true
		ORDER BY price_cents ASC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var plans []*SubscriptionPlan
	for rows.Next() {
		plan := &SubscriptionPlan{}
		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Description,
			&plan.PriceCents,
			&plan.DurationDays,
			&plan.Currency,
			&plan.IsActive,
			&plan.MaxGroups,
			&plan.Features,
			&plan.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	
	return plans, nil
}

func (r *PaymentRepository) GetPaymentStats() (*PaymentStats, error) {
	stats := &PaymentStats{}
	
	// Total revenue
	query := `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed'`
	err := r.db.QueryRow(query).Scan(&stats.TotalRevenue)
	if err != nil {
		return nil, err
	}
	
	// Today's revenue
	query = `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed' AND DATE(created_at) = CURRENT_DATE`
	err = r.db.QueryRow(query).Scan(&stats.TodayRevenue)
	if err != nil {
		return nil, err
	}
	
	// Total payments
	query = `SELECT COUNT(*) FROM payments WHERE status = 'completed'`
	err = r.db.QueryRow(query).Scan(&stats.TotalPayments)
	if err != nil {
		return nil, err
	}
	
	// Today's payments
	query = `SELECT COUNT(*) FROM payments WHERE status = 'completed' AND DATE(created_at) = CURRENT_DATE`
	err = r.db.QueryRow(query).Scan(&stats.TodayPayments)
	if err != nil {
		return nil, err
	}
	
	// Success rate (last 30 days)
	var total, successful int
	query = `SELECT COUNT(*) FROM payments WHERE created_at > NOW() - INTERVAL '30 days'`
	err = r.db.QueryRow(query).Scan(&total)
	if err != nil {
		return nil, err
	}
	
	query = `SELECT COUNT(*) FROM payments WHERE status = 'completed' AND created_at > NOW() - INTERVAL '30 days'`
	err = r.db.QueryRow(query).Scan(&successful)
	if err != nil {
		return nil, err
	}
	
	if total > 0 {
		stats.SuccessRate = float64(successful) / float64(total) * 100
	}
	
	return stats, nil
}

type PaymentStats struct {
	TotalRevenue  int     `json:"total_revenue"`
	TodayRevenue  int     `json:"today_revenue"`
	TotalPayments int     `json:"total_payments"`
	TodayPayments int     `json:"today_payments"`
	SuccessRate   float64 `json:"success_rate"`
}