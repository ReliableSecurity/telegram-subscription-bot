package models

import (
        "database/sql"
        "encoding/json"
        "time"
)

type SubscriptionPlan struct {
        ID           int                    `json:"id"`
        Name         string                 `json:"name"`
        Description  string                 `json:"description"`
        PriceCents   int                    `json:"price_cents"`
        DurationDays int                    `json:"duration_days"`
        MaxGroups    int                    `json:"max_groups"`
        Features     map[string]interface{} `json:"features"`
        Currency     string                 `json:"currency"`
        IsActive     bool                   `json:"is_active"`
        CreatedAt    time.Time              `json:"created_at"`
}

type SubscriptionRepository struct {
        db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
        return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) GetAll() ([]SubscriptionPlan, error) {
        query := `
                SELECT id, name, description, price_cents, duration_days, max_groups, features, currency, is_active, created_at
                FROM subscription_plans WHERE is_active = true ORDER BY price_cents
        `
        rows, err := r.db.Query(query)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        var plans []SubscriptionPlan
        for rows.Next() {
                var plan SubscriptionPlan
                var featuresJSON []byte
                
                err := rows.Scan(
                        &plan.ID, &plan.Name, &plan.Description, &plan.PriceCents, &plan.DurationDays, &plan.MaxGroups,
                        &featuresJSON, &plan.Currency, &plan.IsActive, &plan.CreatedAt,
                )
                if err != nil {
                        return nil, err
                }
                
                if len(featuresJSON) > 0 {
                        json.Unmarshal(featuresJSON, &plan.Features)
                }
                
                plans = append(plans, plan)
        }
        return plans, nil
}

func (r *SubscriptionRepository) GetByID(id int) (*SubscriptionPlan, error) {
        plan := &SubscriptionPlan{}
        var featuresJSON []byte
        
        query := `
                SELECT id, name, description, price_cents, duration_days, max_groups, features, currency, is_active, created_at
                FROM subscription_plans WHERE id = $1
        `
        err := r.db.QueryRow(query, id).Scan(
                &plan.ID, &plan.Name, &plan.Description, &plan.PriceCents, &plan.DurationDays, &plan.MaxGroups,
                &featuresJSON, &plan.Currency, &plan.IsActive, &plan.CreatedAt,
        )
        if err != nil {
                return nil, err
        }
        
        if len(featuresJSON) > 0 {
                json.Unmarshal(featuresJSON, &plan.Features)
        }
        
        return plan, nil
}

func (r *SubscriptionRepository) Create(plan *SubscriptionPlan) error {
        featuresJSON, _ := json.Marshal(plan.Features)
        
        query := `
                INSERT INTO subscription_plans (name, price_cents, duration_days, max_groups, features, currency, is_active)
                VALUES ($1, $2, $3, $4, $5, $6, $7)
                RETURNING id, created_at
        `
        return r.db.QueryRow(query, plan.Name, plan.PriceCents, plan.DurationDays, plan.MaxGroups, featuresJSON, plan.Currency, plan.IsActive).Scan(&plan.ID, &plan.CreatedAt)
}

func (r *SubscriptionRepository) Update(plan *SubscriptionPlan) error {
        featuresJSON, _ := json.Marshal(plan.Features)
        
        query := `
                UPDATE subscription_plans 
                SET name = $1, price_cents = $2, duration_days = $3, max_groups = $4, features = $5, currency = $6, is_active = $7
                WHERE id = $8
        `
        _, err := r.db.Exec(query, plan.Name, plan.PriceCents, plan.DurationDays, plan.MaxGroups, featuresJSON, plan.Currency, plan.IsActive, plan.ID)
        return err
}

func (r *SubscriptionRepository) Delete(id int) error {
        query := `UPDATE subscription_plans SET is_active = false WHERE id = $1`
        _, err := r.db.Exec(query, id)
        return err
}
