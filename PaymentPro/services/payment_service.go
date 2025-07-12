package services

import (
        "crypto/rand"
        "encoding/hex"
        "fmt"
        "time"

        "telegram-subscription-bot/config"
        "telegram-subscription-bot/database"
        "telegram-subscription-bot/models"
        "telegram-subscription-bot/utils"
)

type PaymentService struct {
        db          *database.DB
        config      *config.Config
        paymentRepo *models.PaymentRepository
        planRepo    *models.SubscriptionRepository
        cryptoUtils *utils.CryptoUtils
}

func NewPaymentService(db *database.DB, config *config.Config) *PaymentService {
        return &PaymentService{
                db:          db,
                config:      config,
                paymentRepo: models.NewPaymentRepository(db.DB),
                planRepo:    models.NewSubscriptionRepository(db.DB),
                cryptoUtils: utils.NewCryptoUtils(),
        }
}

func (s *PaymentService) GetProviderToken() string {
        // Return appropriate provider token based on configuration
        if s.config.StripeToken != "" {
                return s.config.StripeToken
        }
        if s.config.YooMoneyToken != "" {
                return s.config.YooMoneyToken
        }
        if s.config.PayPalToken != "" {
                return s.config.PayPalToken
        }
        return ""
}

func (s *PaymentService) CreateCardPayment(userID int, planID int) (*models.Payment, error) {
        plan, err := s.planRepo.GetByID(planID)
        if err != nil {
                return nil, err
        }

        payment := &models.Payment{
                UserID:          int64(userID),
                PlanID:          int64(planID),
                Amount:          plan.PriceCents,
                Currency:        plan.Currency,
                PaymentMethod:   "card",
                PaymentProvider: "telegram",
                Status:          "pending",
                Description:     s.generateInvoicePayload(userID, planID),
                CreatedAt:       time.Now(),
                UpdatedAt:       time.Now(),
        }

        err = s.paymentRepo.Create(payment)
        if err != nil {
                return nil, err
        }

        return payment, nil
}

func (s *PaymentService) CreateCryptoPayment(userID int, planID int, cryptoCurrency string) (*models.Payment, error) {
        plan, err := s.planRepo.GetByID(planID)
        if err != nil {
                return nil, err
        }

        // Get crypto rate and calculate amount
        cryptoAmount, err := s.cryptoUtils.ConvertToCrypto(float64(plan.PriceCents)/100, plan.Currency, cryptoCurrency)
        if err != nil {
                return nil, err
        }

        // Get crypto address
        cryptoAddress, err := s.getCryptoAddress(cryptoCurrency)
        if err != nil {
                return nil, err
        }

        payment := &models.Payment{
                UserID:          int64(userID),
                PlanID:          int64(planID),
                Amount:          plan.PriceCents,
                Currency:        plan.Currency,
                PaymentMethod:   "crypto",
                PaymentProvider: cryptoCurrency,
                Status:          "pending",
                Description:     fmt.Sprintf("Crypto payment: %s to %s", fmt.Sprintf("%.8f", cryptoAmount), cryptoAddress),
                CreatedAt:       time.Now(),
                UpdatedAt:       time.Now(),
        }

        err = s.paymentRepo.Create(payment)
        if err != nil {
                return nil, err
        }

        return payment, nil
}

func (s *PaymentService) VerifyPayment(paymentID int64) error {
        payment, err := s.paymentRepo.GetByID(paymentID)
        if err != nil {
                return err
        }

        if payment.PaymentMethod == "crypto" {
                return s.verifyCryptoPayment(payment)
        }

        return nil
}

func (s *PaymentService) verifyCryptoPayment(payment *models.Payment) error {
        // Simplified crypto payment verification
        // In production, integrate with blockchain APIs
        payment.Status = "completed"
        payment.CompletedAt = time.Now()
        return s.paymentRepo.Update(payment)
}

func (s *PaymentService) ProcessPendingCryptoPayments() error {
        // Simplified implementation
        // In production, query for pending crypto payments and verify them
        return nil
}

func (s *PaymentService) generateInvoicePayload(userID int, planID int) string {
        // Generate random bytes for uniqueness
        randomBytes := make([]byte, 8)
        rand.Read(randomBytes)
        
        return fmt.Sprintf("payment_%d_%d_%d_%s", userID, planID, time.Now().Unix(), hex.EncodeToString(randomBytes))
}

func (s *PaymentService) getCryptoAddress(currency string) (string, error) {
        switch currency {
        case "BTC":
                if s.config.BTCAddress != "" {
                        return s.config.BTCAddress, nil
                }
                return "", fmt.Errorf("BTC address not configured")
        case "ETH":
                if s.config.ETHAddress != "" {
                        return s.config.ETHAddress, nil
                }
                return "", fmt.Errorf("ETH address not configured")
        case "USDT":
                if s.config.USDTAddress != "" {
                        return s.config.USDTAddress, nil
                }
                return "", fmt.Errorf("USDT address not configured")
        default:
                return "", fmt.Errorf("unsupported cryptocurrency: %s", currency)
        }
}

func (s *PaymentService) GetPaymentStats() (map[string]interface{}, error) {
        stats := make(map[string]interface{})
        
        // Total payments
        var totalPayments int
        var totalRevenue float64
        
        err := s.db.QueryRow("SELECT COUNT(*), COALESCE(SUM(amount_cents), 0) FROM payments WHERE status = 'completed'").Scan(&totalPayments, &totalRevenue)
        if err != nil {
                return nil, err
        }
        
        stats["total_payments"] = totalPayments
        stats["total_revenue"] = totalRevenue / 100
        
        // Payment methods breakdown
        rows, err := s.db.Query("SELECT payment_method, COUNT(*), SUM(amount_cents) FROM payments WHERE status = 'completed' GROUP BY payment_method")
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        methodStats := make(map[string]map[string]float64)
        for rows.Next() {
                var method string
                var count int
                var revenue float64
                
                rows.Scan(&method, &count, &revenue)
                methodStats[method] = map[string]float64{
                        "count":   float64(count),
                        "revenue": revenue / 100,
                }
        }
        
        stats["payment_methods"] = methodStats
        
        return stats, nil
}
