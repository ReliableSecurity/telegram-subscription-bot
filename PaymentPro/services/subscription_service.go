package services

import (
        "time"

        "telegram-subscription-bot/database"
        "telegram-subscription-bot/models"
)

type SubscriptionService struct {
        db       *database.DB
        userRepo *models.UserRepository
        planRepo *models.SubscriptionRepository
}

func NewSubscriptionService(db *database.DB) *SubscriptionService {
        return &SubscriptionService{
                db:       db,
                userRepo: models.NewUserRepository(db.DB),
                planRepo: models.NewSubscriptionRepository(db.DB),
        }
}

func (s *SubscriptionService) ActivateSubscription(userID int, planID int) error {
        plan, err := s.planRepo.GetByID(planID)
        if err != nil {
                return err
        }

        var expiresAt *time.Time
        if plan.DurationDays > 0 {
                expiry := time.Now().AddDate(0, 0, plan.DurationDays)
                expiresAt = &expiry
        }

        return s.userRepo.UpdateSubscription(userID, planID, expiresAt)
}

func (s *SubscriptionService) ExtendSubscription(userID int, days int) error {
        user, err := s.userRepo.GetByTelegramID(int64(userID))
        if err != nil {
                return err
        }

        var newExpiresAt time.Time
        if user.PlanExpiresAt != nil && user.PlanExpiresAt.After(time.Now()) {
                // Extend from current expiry date
                newExpiresAt = user.PlanExpiresAt.AddDate(0, 0, days)
        } else {
                // Extend from now
                newExpiresAt = time.Now().AddDate(0, 0, days)
        }

        return s.userRepo.UpdateSubscription(user.ID, user.CurrentPlanID, &newExpiresAt)
}

func (s *SubscriptionService) CheckSubscriptionStatus(userID int64) (*models.User, bool, error) {
        user, err := s.userRepo.GetByTelegramID(userID)
        if err != nil {
                return nil, false, err
        }

        isActive := true
        if user.PlanExpiresAt != nil && user.PlanExpiresAt.Before(time.Now()) {
                isActive = false
        }

        return user, isActive, nil
}

func (s *SubscriptionService) GetUserPlan(userID int64) (*models.SubscriptionPlan, error) {
        user, err := s.userRepo.GetByTelegramID(userID)
        if err != nil {
                return nil, err
        }

        return s.planRepo.GetByID(user.CurrentPlanID)
}

func (s *SubscriptionService) CanUserAccessFeature(userID int64, feature string) (bool, error) {
        user, isActive, err := s.CheckSubscriptionStatus(userID)
        if err != nil {
                return false, err
        }

        if !isActive {
                // User subscription is expired, check if they can access free features
                freePlan, err := s.planRepo.GetByID(1) // Free plan
                if err != nil {
                        return false, err
                }

                if freePlan.Features != nil {
                        if featureEnabled, exists := freePlan.Features[feature]; exists {
                                return featureEnabled.(bool), nil
                        }
                }
                return false, nil
        }

        plan, err := s.planRepo.GetByID(user.CurrentPlanID)
        if err != nil {
                return false, err
        }

        if plan.Features != nil {
                if featureEnabled, exists := plan.Features[feature]; exists {
                        return featureEnabled.(bool), nil
                }
        }

        return false, nil
}

func (s *SubscriptionService) GetExpiredUsers() ([]models.User, error) {
        return s.userRepo.GetExpiredSubscriptions()
}

func (s *SubscriptionService) GetUsersExpiringSoon(days int) ([]models.User, error) {
        return s.userRepo.GetExpiringSoon(days)
}

func (s *SubscriptionService) ProcessExpiredSubscriptions() error {
        expiredUsers, err := s.GetExpiredUsers()
        if err != nil {
                return err
        }

        for _, user := range expiredUsers {
                // Reset to free plan
                err = s.userRepo.UpdateSubscription(user.ID, 1, nil)
                if err != nil {
                        continue // Continue with other users
                }
        }

        return nil
}

func (s *SubscriptionService) GetSubscriptionStats() (map[string]interface{}, error) {
        stats := make(map[string]interface{})

        // Total users
        var totalUsers int
        s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
        stats["total_users"] = totalUsers

        // Active subscriptions by plan
        rows, err := s.db.Query(`
                SELECT sp.name, COUNT(u.id) as count
                FROM subscription_plans sp
                LEFT JOIN users u ON sp.id = u.current_plan_id 
                    AND (u.plan_expires_at IS NULL OR u.plan_expires_at > CURRENT_TIMESTAMP)
                GROUP BY sp.id, sp.name
                ORDER BY sp.id
        `)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        planStats := make(map[string]int)
        for rows.Next() {
                var planName string
                var count int
                rows.Scan(&planName, &count)
                planStats[planName] = count
        }
        stats["active_subscriptions"] = planStats

        // Expiring soon (next 7 days)
        var expiringSoon int
        s.db.QueryRow(`
                SELECT COUNT(*) 
                FROM users 
                WHERE plan_expires_at BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + INTERVAL '7 days'
                AND current_plan_id > 1
        `).Scan(&expiringSoon)
        stats["expiring_soon"] = expiringSoon

        return stats, nil
}

func (s *SubscriptionService) CreatePlan(name string, priceCents int, durationDays int, maxGroups int, features map[string]interface{}) error {
        plan := &models.SubscriptionPlan{
                Name:         name,
                PriceCents:   priceCents,
                DurationDays: durationDays,
                MaxGroups:    maxGroups,
                Features:     features,
                Currency:     "USD",
                IsActive:     true,
        }

        return s.planRepo.Create(plan)
}

func (s *SubscriptionService) UpdatePlan(planID int, updates map[string]interface{}) error {
        plan, err := s.planRepo.GetByID(planID)
        if err != nil {
                return err
        }

        // Update fields based on the updates map
        if name, exists := updates["name"]; exists {
                plan.Name = name.(string)
        }
        if priceCents, exists := updates["price_cents"]; exists {
                plan.PriceCents = priceCents.(int)
        }
        if durationDays, exists := updates["duration_days"]; exists {
                plan.DurationDays = durationDays.(int)
        }
        if maxGroups, exists := updates["max_groups"]; exists {
                plan.MaxGroups = maxGroups.(int)
        }
        if features, exists := updates["features"]; exists {
                plan.Features = features.(map[string]interface{})
        }
        if currency, exists := updates["currency"]; exists {
                plan.Currency = currency.(string)
        }
        if isActive, exists := updates["is_active"]; exists {
                plan.IsActive = isActive.(bool)
        }

        return s.planRepo.Update(plan)
}

func (s *SubscriptionService) DeletePlan(planID int) error {
        return s.planRepo.Delete(planID)
}
