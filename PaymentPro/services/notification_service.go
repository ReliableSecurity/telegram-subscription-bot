package services

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-subscription-bot/database"
	"telegram-subscription-bot/locales"
	"telegram-subscription-bot/models"
)

type NotificationService struct {
	bot                 *tgbotapi.BotAPI
	db                  *database.DB
	subscriptionService *SubscriptionService
	userRepo            *models.UserRepository
	ticker              *time.Ticker
	stopChan            chan bool
}

func NewNotificationService(bot *tgbotapi.BotAPI, db *database.DB) *NotificationService {
	return &NotificationService{
		bot:                 bot,
		db:                  db,
		subscriptionService: NewSubscriptionService(db),
		userRepo:            models.NewUserRepository(db.DB),
		ticker:              time.NewTicker(1 * time.Hour), // Check every hour
		stopChan:            make(chan bool),
	}
}

func (s *NotificationService) Start() {
	log.Println("Starting notification service...")
	
	for {
		select {
		case <-s.ticker.C:
			s.processNotifications()
		case <-s.stopChan:
			s.ticker.Stop()
			return
		}
	}
}

func (s *NotificationService) Stop() {
	s.stopChan <- true
}

func (s *NotificationService) processNotifications() {
	// Process expired subscriptions
	s.processExpiredSubscriptions()
	
	// Process expiring soon notifications
	s.processExpiringNotifications()
	
	// Process payment reminders
	s.processPaymentReminders()
}

func (s *NotificationService) processExpiredSubscriptions() {
	expiredUsers, err := s.subscriptionService.GetExpiredUsers()
	if err != nil {
		log.Printf("Error getting expired users: %v", err)
		return
	}

	for _, user := range expiredUsers {
		// Reset to free plan
		err = s.userRepo.UpdateSubscription(user.ID, 1, nil)
		if err != nil {
			log.Printf("Error resetting user %d to free plan: %v", user.ID, err)
			continue
		}

		// Send notification
		message := locales.GetMessage(user.LanguageCode, "subscription_expired")
		message += "\n\n" + locales.GetMessage(user.LanguageCode, "upgrade_prompt")
		
		s.sendNotification(user.TelegramID, message)
		s.logNotification(user.ID, "expired")
	}
}

func (s *NotificationService) processExpiringNotifications() {
	// Notify users 3 days before expiration
	expiring3Days, err := s.subscriptionService.GetUsersExpiringSoon(3)
	if err != nil {
		log.Printf("Error getting users expiring in 3 days: %v", err)
		return
	}

	for _, user := range expiring3Days {
		if s.wasNotificationSent(user.ID, "expiring_3_days") {
			continue
		}

		plan, err := s.subscriptionService.GetUserPlan(user.TelegramID)
		if err != nil {
			continue
		}

		message := fmt.Sprintf(locales.GetMessage(user.LanguageCode, "subscription_expiring_3_days"), plan.Name)
		message += "\n\n" + locales.GetMessage(user.LanguageCode, "renew_prompt")
		
		s.sendNotification(user.TelegramID, message)
		s.logNotification(user.ID, "expiring_3_days")
	}

	// Notify users 1 day before expiration
	expiring1Day, err := s.subscriptionService.GetUsersExpiringSoon(1)
	if err != nil {
		log.Printf("Error getting users expiring in 1 day: %v", err)
		return
	}

	for _, user := range expiring1Day {
		if s.wasNotificationSent(user.ID, "expiring_1_day") {
			continue
		}

		plan, err := s.subscriptionService.GetUserPlan(user.TelegramID)
		if err != nil {
			continue
		}

		message := fmt.Sprintf(locales.GetMessage(user.LanguageCode, "subscription_expiring_1_day"), plan.Name)
		message += "\n\n" + locales.GetMessage(user.LanguageCode, "renew_prompt")
		
		s.sendNotification(user.TelegramID, message)
		s.logNotification(user.ID, "expiring_1_day")
	}
}

func (s *NotificationService) processPaymentReminders() {
	// Get users with failed payments in the last 24 hours
	rows, err := s.db.Query(`
		SELECT DISTINCT u.id, u.telegram_id, u.language_code, u.first_name
		FROM users u
		JOIN payments p ON u.id = p.user_id
		WHERE p.status = 'failed' 
		AND p.created_at > CURRENT_TIMESTAMP - INTERVAL '24 hours'
		AND NOT EXISTS (
			SELECT 1 FROM payment_notifications pn 
			WHERE pn.user_id = u.id 
			AND pn.notification_type = 'payment_failed'
			AND pn.sent_at > CURRENT_TIMESTAMP - INTERVAL '24 hours'
		)
	`)
	if err != nil {
		log.Printf("Error getting users with failed payments: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		var telegramID int64
		var languageCode, firstName string
		
		rows.Scan(&userID, &telegramID, &languageCode, &firstName)
		
		message := fmt.Sprintf(locales.GetMessage(languageCode, "payment_failed_reminder"), firstName)
		message += "\n\n" + locales.GetMessage(languageCode, "retry_payment_prompt")
		
		s.sendNotification(telegramID, message)
		s.logNotification(userID, "payment_failed")
	}
}

func (s *NotificationService) sendNotification(telegramID int64, message string) {
	msg := tgbotapi.NewMessage(telegramID, message)
	_, err := s.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending notification to user %d: %v", telegramID, err)
	}
}

func (s *NotificationService) logNotification(userID int, notificationType string) {
	_, err := s.db.Exec(`
		INSERT INTO payment_notifications (user_id, notification_type, sent_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
	`, userID, notificationType)
	if err != nil {
		log.Printf("Error logging notification: %v", err)
	}
}

func (s *NotificationService) wasNotificationSent(userID int, notificationType string) bool {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM payment_notifications 
		WHERE user_id = $1 AND notification_type = $2 
		AND sent_at > CURRENT_TIMESTAMP - INTERVAL '24 hours'
	`, userID, notificationType).Scan(&count)
	
	return err == nil && count > 0
}

func (s *NotificationService) SendWelcomeMessage(userID int64, languageCode string) {
	message := locales.GetMessage(languageCode, "welcome_message")
	message += "\n\n" + locales.GetMessage(languageCode, "getting_started")
	
	s.sendNotification(userID, message)
}

func (s *NotificationService) SendPaymentConfirmation(userID int64, languageCode string, planName string, amount float64, currency string) {
	message := fmt.Sprintf(locales.GetMessage(languageCode, "payment_confirmation"), planName, amount, currency)
	message += "\n\n" + locales.GetMessage(languageCode, "enjoy_features")
	
	s.sendNotification(userID, message)
}

func (s *NotificationService) SendCustomNotification(userID int64, message string) {
	s.sendNotification(userID, message)
}

func (s *NotificationService) BroadcastMessage(message string, targetPlan int) error {
	query := `
		SELECT telegram_id, language_code 
		FROM users 
		WHERE current_plan_id = $1 
		AND (plan_expires_at IS NULL OR plan_expires_at > CURRENT_TIMESTAMP)
	`
	
	if targetPlan == 0 {
		// Broadcast to all users
		query = `SELECT telegram_id, language_code FROM users`
	}
	
	rows, err := s.db.Query(query, targetPlan)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var telegramID int64
		var languageCode string
		
		rows.Scan(&telegramID, &languageCode)
		s.sendNotification(telegramID, message)
		
		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}
