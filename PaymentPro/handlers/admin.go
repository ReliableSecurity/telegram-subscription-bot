package handlers

import (
        "fmt"
        "strconv"
        "strings"
        "time"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
        "telegram-subscription-bot/database"
        "telegram-subscription-bot/models"
        "telegram-subscription-bot/services"
)

type AdminHandler struct {
        bot                 *tgbotapi.BotAPI
        db                  *database.DB
        subscriptionService *services.SubscriptionService
        paymentService      *services.PaymentService
        userRepo            *models.UserRepository
        paymentRepo         *models.PaymentRepository
        planRepo            *models.SubscriptionRepository
        adminUserIDs        []int64
}

func NewAdminHandler(bot *tgbotapi.BotAPI, db *database.DB, subscriptionService *services.SubscriptionService, paymentService *services.PaymentService) *AdminHandler {
        return &AdminHandler{
                bot:                 bot,
                db:                  db,
                subscriptionService: subscriptionService,
                paymentService:      paymentService,
                userRepo:            models.NewUserRepository(db.DB),
                paymentRepo:         models.NewPaymentRepository(db.DB),
                planRepo:            models.NewSubscriptionRepository(db.DB),
                adminUserIDs:        []int64{}, // Should be loaded from config
        }
}

func (h *AdminHandler) HandleAdminCommand(update tgbotapi.Update) {
        if update.Message == nil || !update.Message.IsCommand() {
                return
        }

        // Check if user is admin
        if !h.isAdmin(update.Message.From.ID) {
                return
        }

        command := update.Message.Command()
        args := strings.Fields(update.Message.CommandArguments())

        switch command {
        case "admin_stats":
                h.handleStats(update)
        case "admin_users":
                h.handleUsers(update, args)
        case "admin_payments":
                h.handlePayments(update, args)
        case "admin_plans":
                h.handlePlans(update, args)
        case "admin_grant":
                h.handleGrant(update, args)
        case "admin_revoke":
                h.handleRevoke(update, args)
        }
}

func (h *AdminHandler) handleStats(update tgbotapi.Update) {
        // Get basic statistics
        var totalUsers, activeSubscriptions, totalPayments int
        var totalRevenue float64

        // Query total users
        h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
        
        // Query active subscriptions
        h.db.QueryRow("SELECT COUNT(*) FROM users WHERE current_plan_id > 1 AND (plan_expires_at IS NULL OR plan_expires_at > CURRENT_TIMESTAMP)").Scan(&activeSubscriptions)
        
        // Query total payments
        h.db.QueryRow("SELECT COUNT(*) FROM payments WHERE status = 'completed'").Scan(&totalPayments)
        
        // Query total revenue
        h.db.QueryRow("SELECT COALESCE(SUM(amount_cents), 0) FROM payments WHERE status = 'completed'").Scan(&totalRevenue)
        totalRevenue /= 100 // Convert from cents

        // Get today's stats
        var todayUsers, todayPayments int
        var todayRevenue float64
        
        h.db.QueryRow("SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE").Scan(&todayUsers)
        h.db.QueryRow("SELECT COUNT(*) FROM payments WHERE status = 'completed' AND DATE(completed_at) = CURRENT_DATE").Scan(&todayPayments)
        h.db.QueryRow("SELECT COALESCE(SUM(amount_cents), 0) FROM payments WHERE status = 'completed' AND DATE(completed_at) = CURRENT_DATE").Scan(&todayRevenue)
        todayRevenue /= 100

        message := "📊 **Admin Statistics**\n\n"
        message += "**Overall:**\n"
        message += fmt.Sprintf("👥 Total Users: %d\n", totalUsers)
        message += fmt.Sprintf("💎 Active Subscriptions: %d\n", activeSubscriptions)
        message += fmt.Sprintf("💰 Total Payments: %d\n", totalPayments)
        message += fmt.Sprintf("💵 Total Revenue: $%.2f\n\n", totalRevenue)
        
        message += "**Today:**\n"
        message += fmt.Sprintf("👥 New Users: %d\n", todayUsers)
        message += fmt.Sprintf("💰 Payments: %d\n", todayPayments)
        message += fmt.Sprintf("💵 Revenue: $%.2f\n", todayRevenue)

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *AdminHandler) handleUsers(update tgbotapi.Update, args []string) {
        var message string
        
        if len(args) == 0 {
                // Show recent users
                rows, err := h.db.Query(`
                        SELECT telegram_id, username, first_name, current_plan_id, plan_expires_at, created_at
                        FROM users 
                        ORDER BY created_at DESC 
                        LIMIT 10
                `)
                if err != nil {
                        h.sendMessage(update.Message.Chat.ID, "Error fetching users")
                        return
                }
                defer rows.Close()
                
                message = "👥 **Recent Users:**\n\n"
                
                for rows.Next() {
                        var telegramID int64
                        var username, firstName string
                        var currentPlanID int
                        var planExpiresAt *time.Time
                        var createdAt time.Time
                        
                        rows.Scan(&telegramID, &username, &firstName, &currentPlanID, &planExpiresAt, &createdAt)
                        
                        name := firstName
                        if username != "" {
                                name = fmt.Sprintf("%s (@%s)", firstName, username)
                        }
                        
                        planStatus := "Free"
                        if currentPlanID > 1 {
                                if planExpiresAt != nil && planExpiresAt.After(time.Now()) {
                                        planStatus = "Premium"
                                } else {
                                        planStatus = "Expired"
                                }
                        }
                        
                        message += fmt.Sprintf("🔹 %s\n", name)
                        message += fmt.Sprintf("   ID: %d\n", telegramID)
                        message += fmt.Sprintf("   Plan: %s\n", planStatus)
                        message += fmt.Sprintf("   Joined: %s\n\n", createdAt.Format("2006-01-02"))
                }
        } else {
                // Show specific user
                userID, err := strconv.ParseInt(args[0], 10, 64)
                if err != nil {
                        h.sendMessage(update.Message.Chat.ID, "Invalid user ID")
                        return
                }
                
                user, err := h.userRepo.GetByTelegramID(userID)
                if err != nil {
                        h.sendMessage(update.Message.Chat.ID, "User not found")
                        return
                }
                
                plan, _ := h.planRepo.GetByID(user.CurrentPlanID)
                payments, _ := h.paymentRepo.GetByUserID(int64(user.ID))
                
                message = fmt.Sprintf("👤 **User Details:**\n\n")
                message += fmt.Sprintf("**Name:** %s", user.FirstName)
                if user.Username != "" {
                        message += fmt.Sprintf(" (@%s)", user.Username)
                }
                message += "\n"
                message += fmt.Sprintf("**ID:** %d\n", user.TelegramID)
                message += fmt.Sprintf("**Plan:** %s\n", plan.Name)
                if user.PlanExpiresAt != nil {
                        message += fmt.Sprintf("**Expires:** %s\n", user.PlanExpiresAt.Format("2006-01-02 15:04:05"))
                }
                message += fmt.Sprintf("**Joined:** %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
                message += fmt.Sprintf("**Total Payments:** %d\n", len(payments))
                
                totalSpent := 0
                for _, payment := range payments {
                        if payment.Status == "completed" {
                                totalSpent += payment.Amount
                        }
                }
                message += fmt.Sprintf("**Total Spent:** $%.2f\n", float64(totalSpent)/100)
        }
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *AdminHandler) handlePayments(update tgbotapi.Update, args []string) {
        var message string
        
        if len(args) == 0 {
                // Show recent payments
                rows, err := h.db.Query(`
                        SELECT p.id, p.amount_cents, p.currency, p.payment_method, p.status, p.created_at, u.first_name, u.username
                        FROM payments p
                        JOIN users u ON p.user_id = u.id
                        ORDER BY p.created_at DESC 
                        LIMIT 10
                `)
                if err != nil {
                        h.sendMessage(update.Message.Chat.ID, "Error fetching payments")
                        return
                }
                defer rows.Close()
                
                message = "💰 **Recent Payments:**\n\n"
                
                for rows.Next() {
                        var paymentID, amountCents int
                        var currency, paymentMethod, status, firstName, username string
                        var createdAt time.Time
                        
                        rows.Scan(&paymentID, &amountCents, &currency, &paymentMethod, &status, &createdAt, &firstName, &username)
                        
                        name := firstName
                        if username != "" {
                                name = fmt.Sprintf("%s (@%s)", firstName, username)
                        }
                        
                        statusEmoji := "⏳"
                        if status == "completed" {
                                statusEmoji = "✅"
                        } else if status == "failed" {
                                statusEmoji = "❌"
                        }
                        
                        message += fmt.Sprintf("%s $%.2f %s via %s\n", statusEmoji, float64(amountCents)/100, currency, paymentMethod)
                        message += fmt.Sprintf("   By: %s\n", name)
                        message += fmt.Sprintf("   Date: %s\n\n", createdAt.Format("2006-01-02 15:04"))
                }
        }
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *AdminHandler) handlePlans(update tgbotapi.Update, args []string) {
        plans, err := h.planRepo.GetAll()
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "Error fetching plans")
                return
        }
        
        message := "💎 **Subscription Plans:**\n\n"
        
        for _, plan := range plans {
                message += fmt.Sprintf("**%s** (ID: %d)\n", plan.Name, plan.ID)
                message += fmt.Sprintf("Price: $%.2f %s\n", float64(plan.PriceCents)/100, plan.Currency)
                message += fmt.Sprintf("Duration: %d days\n", plan.DurationDays)
                message += fmt.Sprintf("Max Groups: %d\n", plan.MaxGroups)
                message += fmt.Sprintf("Active: %t\n\n", plan.IsActive)
        }
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *AdminHandler) handleGrant(update tgbotapi.Update, args []string) {
        if len(args) < 2 {
                h.sendMessage(update.Message.Chat.ID, "Usage: /admin_grant <user_id> <plan_id> [days]")
                return
        }
        
        userID, err := strconv.ParseInt(args[0], 10, 64)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "Invalid user ID")
                return
        }
        
        planID, err := strconv.Atoi(args[1])
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "Invalid plan ID")
                return
        }
        
        days := 30
        if len(args) > 2 {
                days, err = strconv.Atoi(args[2])
                if err != nil {
                        h.sendMessage(update.Message.Chat.ID, "Invalid days")
                        return
                }
        }
        
        user, err := h.userRepo.GetByTelegramID(userID)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "User not found")
                return
        }
        
        expiresAt := time.Now().AddDate(0, 0, days)
        err = h.userRepo.UpdateSubscription(user.ID, planID, &expiresAt)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "Error updating subscription")
                return
        }
        
        h.sendMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Granted plan %d to user %d for %d days", planID, userID, days))
}

func (h *AdminHandler) handleRevoke(update tgbotapi.Update, args []string) {
        if len(args) < 1 {
                h.sendMessage(update.Message.Chat.ID, "Usage: /admin_revoke <user_id>")
                return
        }
        
        userID, err := strconv.ParseInt(args[0], 10, 64)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "Invalid user ID")
                return
        }
        
        user, err := h.userRepo.GetByTelegramID(userID)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "User not found")
                return
        }
        
        err = h.userRepo.UpdateSubscription(user.ID, 1, nil) // Reset to free plan
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "Error revoking subscription")
                return
        }
        
        h.sendMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Revoked subscription for user %d", userID))
}

func (h *AdminHandler) isAdmin(userID int64) bool {
        for _, adminID := range h.adminUserIDs {
                if adminID == userID {
                        return true
                }
        }
        return false
}

func (h *AdminHandler) sendMessage(chatID int64, text string) {
        msg := tgbotapi.NewMessage(chatID, text)
        h.bot.Send(msg)
}
