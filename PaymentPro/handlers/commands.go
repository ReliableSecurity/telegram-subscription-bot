package handlers

import (
        "fmt"
        "math/rand"
        "os"
        "strconv"
        "strings"
        "time"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
        "telegram-subscription-bot/database"
        "telegram-subscription-bot/locales"
        "telegram-subscription-bot/models"
        "telegram-subscription-bot/services"
)

type CommandHandler struct {
        bot                 *tgbotapi.BotAPI
        db                  *database.DB
        subscriptionService *services.SubscriptionService
        paymentService      *services.PaymentService
        userRepo            *models.UserRepository
        planRepo            *models.SubscriptionRepository
}

func NewCommandHandler(bot *tgbotapi.BotAPI, db *database.DB, subscriptionService *services.SubscriptionService, paymentService *services.PaymentService) *CommandHandler {
        return &CommandHandler{
                bot:                 bot,
                db:                  db,
                subscriptionService: subscriptionService,
                paymentService:      paymentService,
                userRepo:            models.NewUserRepository(db.DB),
                planRepo:            models.NewSubscriptionRepository(db.DB),
        }
}

func (h *CommandHandler) Handle(update tgbotapi.Update) {
        if update.Message == nil {
                return
        }

        user := h.ensureUser(update.Message.From)
        if user == nil {
                return
        }

        command := update.Message.Command()
        args := strings.Fields(update.Message.CommandArguments())

        switch command {
        case "start":
                h.handleStart(update, user)
        case "help":
                h.handleHelp(update, user)
        case "plans":
                h.handlePlans(update, user)
        case "myplan":
                h.handleMyPlan(update, user)
        case "subscribe":
                h.handleSubscribe(update, user, args)
        case "cancel":
                h.handleCancel(update, user)
        case "history":
                h.handleHistory(update, user)
        case "crypto":
                h.handleCrypto(update, user, args)
        case "setup":
                h.handleSetup(update, user)
        case "addbot":
                h.handleAddBot(update, user)
        case "id":
                h.handleID(update, user)
        case "violations":
                h.handleViolations(update, user)
        default:
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "unknown_command"))
        }
}

func (h *CommandHandler) HandleCallback(update tgbotapi.Update) {
        if update.CallbackQuery == nil {
                return
        }

        user := h.ensureUser(update.CallbackQuery.From)
        if user == nil {
                return
        }

        data := update.CallbackQuery.Data
        parts := strings.Split(data, ":")

        switch parts[0] {
        case "show_plans":
                h.handlePlansCallback(update, user)
        case "my_plan":
                h.handleMyPlanCallback(update, user)
        case "payment_history":
                h.handleHistoryCallback(update, user)
        case "my_account":
                h.handleMyAccountCallback(update, user)
        case "setup_guide":
                h.handleSetupCallback(update, user)
        case "manage_groups":
                h.handleManageGroupsCallback(update, user)
        case "help_menu":
                h.handleHelpCallback(update, user)
        case "subscribe":
                if len(parts) > 1 {
                        planID, _ := strconv.Atoi(parts[1])
                        h.handleSubscribeCallback(update, user, planID)
                }
        case "pay_card":
                if len(parts) > 1 {
                        planID, _ := strconv.Atoi(parts[1])
                        h.handleCardPayment(update, user, planID)
                }
        case "pay_crypto":
                if len(parts) > 1 {
                        planID, _ := strconv.Atoi(parts[1])
                        h.handleCryptoPayment(update, user, planID)
                }
        case "back_to_menu":
                h.handleBackToMenu(update, user)
        case "change_password":
                h.handleChangePasswordCallback(update, user)
        case "change_username":
                h.handleChangeUsernameCallback(update, user)
        }

        // Answer callback query
        callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
        h.bot.Send(callback)
}

func (h *CommandHandler) handleStart(update tgbotapi.Update, user *models.User) {
        message := locales.GetMessage(user.LanguageCode, "welcome")
        
        // Create main menu keyboard
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💎 Планы подписок", "show_plans"),
                        tgbotapi.NewInlineKeyboardButtonData("📊 Мой план", "my_plan"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💰 История платежей", "payment_history"),
                        tgbotapi.NewInlineKeyboardButtonData("🔗 Мой аккаунт", "my_account"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔧 Управление группами", "manage_groups"),
                        tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleHelp(update tgbotapi.Update, user *models.User) {
        message := locales.GetMessage(user.LanguageCode, "help")
        h.sendMessage(update.Message.Chat.ID, message)
}

func (h *CommandHandler) handlePlans(update tgbotapi.Update, user *models.User) {
        plans, err := h.planRepo.GetAll()
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }

        message := locales.GetMessage(user.LanguageCode, "available_plans") + "\n\n"
        
        var keyboard [][]tgbotapi.InlineKeyboardButton
        
        for _, plan := range plans {
                price := fmt.Sprintf("%.2f USD", float64(plan.PriceCents)/100)
                if plan.PriceCents == 0 {
                        price = "Бесплатно"
                }
                
                // Display simple plan info
                planText := fmt.Sprintf("💎 %s - %s", plan.Name, price)
                if plan.Name == "Basic" {
                        planText += " (1 месяц)"
                } else if plan.Name == "Pro" {
                        planText += " (3 месяца)"
                } else if plan.Name == "Premium" {
                        planText += " (1 год)"
                }
                
                message += planText + "\n"
                
                subscribeBtn := tgbotapi.NewInlineKeyboardButtonData(
                        fmt.Sprintf("💎 %s - %s", plan.Name, price),
                        fmt.Sprintf("subscribe:%d", plan.ID),
                )
                keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{subscribeBtn})
        }
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        if len(keyboard) > 0 {
                msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
        }
        h.bot.Send(msg)
}

func (h *CommandHandler) handleMyPlan(update tgbotapi.Update, user *models.User) {
        plan, err := h.planRepo.GetByID(user.CurrentPlanID)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }

        message := fmt.Sprintf("%s: %s\n", locales.GetMessage(user.LanguageCode, "current_plan"), plan.Name)
        
        if user.PlanExpiresAt != nil {
                message += fmt.Sprintf("%s: %s\n", locales.GetMessage(user.LanguageCode, "expires_at"), user.PlanExpiresAt.Format("2006-01-02 15:04:05"))
                
                if user.PlanExpiresAt.Before(time.Now()) {
                        message += "\n⚠️ " + locales.GetMessage(user.LanguageCode, "plan_expired")
                }
        }
        
        message += fmt.Sprintf("\n%s: %d", locales.GetMessage(user.LanguageCode, "max_groups"), plan.MaxGroups)
        
        h.sendMessage(update.Message.Chat.ID, message)
}

func (h *CommandHandler) handleSubscribe(update tgbotapi.Update, user *models.User, args []string) {
        if len(args) == 0 {
                h.handlePlans(update, user)
                return
        }
        
        planID, err := strconv.Atoi(args[0])
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "invalid_plan_id"))
                return
        }
        
        h.handleSubscribePlan(update, user, planID)
}

func (h *CommandHandler) handleSubscribePlan(update tgbotapi.Update, user *models.User, planID int) {
        plan, err := h.planRepo.GetByID(planID)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "plan_not_found"))
                return
        }

        if plan.PriceCents == 0 {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "free_plan_no_payment"))
                return
        }

        // Create payment options
        message := fmt.Sprintf("%s %s\n\n", locales.GetMessage(user.LanguageCode, "payment_options"), plan.Name)
        message += fmt.Sprintf("%s: %.2f %s\n", locales.GetMessage(user.LanguageCode, "price"), float64(plan.PriceCents)/100, plan.Currency)
        
        var keyboard [][]tgbotapi.InlineKeyboardButton
        
        // Card payment button
        cardBtn := tgbotapi.NewInlineKeyboardButtonData(
                "💳 "+locales.GetMessage(user.LanguageCode, "pay_with_card"),
                fmt.Sprintf("subscribe:%d", planID),
        )
        keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{cardBtn})
        
        // Crypto payment buttons
        cryptoRow := []tgbotapi.InlineKeyboardButton{
                tgbotapi.NewInlineKeyboardButtonData("₿ Bitcoin", fmt.Sprintf("crypto_pay:%d:BTC", planID)),
                tgbotapi.NewInlineKeyboardButtonData("Ξ Ethereum", fmt.Sprintf("crypto_pay:%d:ETH", planID)),
                tgbotapi.NewInlineKeyboardButtonData("₮ USDT", fmt.Sprintf("crypto_pay:%d:USDT", planID)),
        }
        keyboard = append(keyboard, cryptoRow)
        
        var chatID int64
        if update.Message != nil {
                chatID = update.Message.Chat.ID
        } else if update.CallbackQuery != nil {
                chatID = update.CallbackQuery.Message.Chat.ID
        }
        
        msg := tgbotapi.NewMessage(chatID, message)
        msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
        h.bot.Send(msg)
}

func (h *CommandHandler) handleCancel(update tgbotapi.Update, user *models.User) {
        // Reset user to free plan
        err := h.userRepo.UpdateSubscription(user.ID, 1, nil)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }
        
        h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "subscription_cancelled"))
}

func (h *CommandHandler) handleHistory(update tgbotapi.Update, user *models.User) {
        paymentRepo := models.NewPaymentRepository(h.db.DB)
        payments, err := paymentRepo.GetByUserID(int64(user.ID))
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }
        
        if len(payments) == 0 {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "no_payment_history"))
                return
        }
        
        message := locales.GetMessage(user.LanguageCode, "payment_history") + "\n\n"
        
        for _, payment := range payments {
                status := payment.Status
                if status == "completed" {
                        status = "✅ " + locales.GetMessage(user.LanguageCode, "completed")
                } else if status == "pending" {
                        status = "⏳ " + locales.GetMessage(user.LanguageCode, "pending")
                } else if status == "failed" {
                        status = "❌ " + locales.GetMessage(user.LanguageCode, "failed")
                }
                
                message += fmt.Sprintf("💰 %.2f %s - %s\n", float64(payment.Amount)/100, payment.Currency, status)
                message += fmt.Sprintf("📅 %s\n", payment.CreatedAt.Format("2006-01-02 15:04:05"))
                message += fmt.Sprintf("💳 %s\n\n", payment.PaymentMethod)
        }
        
        h.sendMessage(update.Message.Chat.ID, message)
}

func (h *CommandHandler) handleCrypto(update tgbotapi.Update, user *models.User, args []string) {
        if len(args) < 2 {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "crypto_usage"))
                return
        }
        
        planID, err := strconv.Atoi(args[0])
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "invalid_plan_id"))
                return
        }
        
        currency := strings.ToUpper(args[1])
        h.handleCryptoPay(update, user, planID, currency)
}

func (h *CommandHandler) handleCryptoPay(update tgbotapi.Update, user *models.User, planID int, currency string) {
        _, err := h.planRepo.GetByID(planID)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "plan_not_found"))
                return
        }

        _, err = h.paymentService.CreateCryptoPayment(user.ID, planID, currency)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }

        message := fmt.Sprintf("%s %s\n\n", locales.GetMessage(user.LanguageCode, "crypto_payment_instructions"), currency)
        message += fmt.Sprintf("%s: `%s`\n", locales.GetMessage(user.LanguageCode, "address"), "Address from description")
        message += fmt.Sprintf("%s: `%s`\n", locales.GetMessage(user.LanguageCode, "amount"), "Amount from description")
        message += fmt.Sprintf("\n%s", locales.GetMessage(user.LanguageCode, "crypto_payment_note"))
        
        var chatID int64
        if update.Message != nil {
                chatID = update.Message.Chat.ID
        } else if update.CallbackQuery != nil {
                chatID = update.CallbackQuery.Message.Chat.ID
        }
        
        msg := tgbotapi.NewMessage(chatID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *CommandHandler) ensureUser(from *tgbotapi.User) *models.User {
        user, err := h.userRepo.GetByTelegramID(from.ID)
        if err != nil {
                // Create new user
                user = &models.User{
                        TelegramID:    from.ID,
                        Username:      from.UserName,
                        FirstName:     from.FirstName,
                        LastName:      from.LastName,
                        LanguageCode:  from.LanguageCode,
                        CurrentPlanID: 1, // Basic plan
                }
                
                if err := h.userRepo.CreateOrUpdate(user); err != nil {
                        return nil
                }
        }
        
        return user
}

func (h *CommandHandler) handleID(update tgbotapi.Update, user *models.User) {
        chatID := update.Message.Chat.ID
        chatType := update.Message.Chat.Type
        
        var message string
        if chatType == "private" {
                message = fmt.Sprintf("🆔 **ID этого чата:** `%d`\n\n", chatID)
                message += "💡 Для получения ID группы или канала используйте эту команду в самой группе/канале"
        } else {
                groupName := update.Message.Chat.Title
                if groupName == "" {
                        groupName = "Неизвестная группа"
                }
                
                message = fmt.Sprintf("🆔 **ID группы/канала:** `%d`\n", chatID)
                message += fmt.Sprintf("📝 **Название:** %s\n", groupName)
                message += fmt.Sprintf("🔧 **Тип:** %s\n\n", chatType)
                message += "💡 Скопируйте ID и используйте его в веб-интерфейсе для добавления группы"
        }
        
        msg := tgbotapi.NewMessage(chatID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *CommandHandler) ensureWebAccount(user *models.User) {
        if user.WebUsername != "" {
                return
        }
        
        // Генерируем данные для веб-аккаунта
        username := h.generateRandomUsername()
        password := h.generateRandomPassword()
        
        err := h.userRepo.UpdateWebCredentials(user.ID, username, password)
        if err != nil {
                return
        }
        
        user.WebUsername = username
        user.WebPassword = password
        user.IsWebActive = true
}

func (h *CommandHandler) generateUserToken(user *models.User) string {
        return fmt.Sprintf("user_%d_%s", user.ID, h.generateRandomPassword()[:32])
}

func (h *CommandHandler) sendMessage(chatID int64, text string) {
        msg := tgbotapi.NewMessage(chatID, text)
        h.bot.Send(msg)
}

func (h *CommandHandler) handleViolations(update tgbotapi.Update, user *models.User) {
        // Эта команда работает только в группах
        if update.Message.Chat.Type == "private" {
                h.sendMessage(update.Message.Chat.ID, "❌ Эта команда работает только в группах")
                return
        }

        // Получаем список нарушений для этой группы
        query := `
                SELECT v.id, v.user_id, v.chat_id, v.violation_type, v.violation_reason, 
                       v.message_text, v.created_at, v.expires_at, v.is_active,
                       u.username, u.first_name
                FROM user_violations v
                JOIN users u ON v.user_id = u.id
                WHERE v.chat_id = $1 AND v.is_active = TRUE
                ORDER BY v.created_at DESC
                LIMIT 20
        `
        
        rows, err := h.db.DB.Query(query, update.Message.Chat.ID)
        if err != nil {
                h.sendMessage(update.Message.Chat.ID, "❌ Ошибка получения данных о нарушениях")
                return
        }
        defer rows.Close()
        
        var violations []string
        warnings := 0
        tempBans := 0
        permanentBans := 0
        
        for rows.Next() {
                var id, userID int
                var chatID int64
                var violationType, violationReason, messageText, username, firstName string
                var createdAt time.Time
                var expiresAt *time.Time
                var isActive bool
                
                err := rows.Scan(&id, &userID, &chatID, &violationType, &violationReason,
                        &messageText, &createdAt, &expiresAt, &isActive, &username, &firstName)
                if err != nil {
                        continue
                }
                
                // Подсчитываем статистику
                switch violationType {
                case "warning":
                        warnings++
                case "temp_ban":
                        tempBans++
                case "permanent_ban":
                        permanentBans++
                }
                
                // Формируем имя пользователя
                displayName := firstName
                if username != "" {
                        displayName = "@" + username
                }
                
                // Формируем описание нарушения
                var icon string
                switch violationType {
                case "warning":
                        icon = "⚠️"
                case "temp_ban":
                        icon = "🚫"
                case "permanent_ban":
                        icon = "🔒"
                }
                
                violationText := fmt.Sprintf("%s %s - %s (%s)", 
                        icon, displayName, violationReason, createdAt.Format("02.01.2006"))
                violations = append(violations, violationText)
        }
        
        // Формируем сообщение
        var message strings.Builder
        message.WriteString("📊 **Статистика нарушений в группе**\n\n")
        message.WriteString(fmt.Sprintf("⚠️ Предупреждения: %d\n", warnings))
        message.WriteString(fmt.Sprintf("🚫 Временные баны: %d\n", tempBans))
        message.WriteString(fmt.Sprintf("🔒 Постоянные баны: %d\n\n", permanentBans))
        
        if len(violations) == 0 {
                message.WriteString("✅ В этой группе нет активных нарушений")
        } else {
                message.WriteString("**Последние нарушения:**\n")
                for _, violation := range violations {
                        message.WriteString(violation + "\n")
                }
                
                if len(violations) == 20 {
                        message.WriteString("\n... и другие (показаны последние 20)")
                }
        }
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message.String())
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *CommandHandler) handleSetup(update tgbotapi.Update, user *models.User) {
        message := "🔧 **Инструкция по настройке бота**\n\n"
        message += "**Шаг 1: Создание бота**\n"
        message += "1. Найдите @BotFather в Telegram\n"
        message += "2. Отправьте команду /newbot\n"
        message += "3. Следуйте инструкциям для создания бота\n"
        message += "4. Получите токен бота\n\n"
        
        message += "**Шаг 2: Добавление в группы**\n"
        message += "1. Добавьте бота в группу или канал\n"
        message += "2. Дайте боту права администратора\n"
        message += "3. Настройте функции модерации\n\n"
        
        message += "**Шаг 3: Веб-панель**\n"
        message += "1. Откройте веб-панель управления\n"
        message += "2. Войдите с данными: admin / admin123\n"
        message += "3. Смените пароль в настройках\n\n"
        
        message += "**Полезные команды:**\n"
        message += "/addbot - Как добавить бота в группу\n"
        message += "/plans - Посмотреть тарифы\n"
        message += "/help - Список всех команд\n"
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *CommandHandler) handleAddBot(update tgbotapi.Update, user *models.User) {
        message := "🤖 **Как добавить бота в группу или канал**\n\n"
        message += "**Для групп:**\n"
        message += "1. Перейдите в группу\n"
        message += "2. Нажмите на название группы\n"
        message += "3. Выберите \"Добавить участника\"\n"
        message += "4. Найдите и добавьте этого бота\n"
        message += "5. Дайте боту права администратора\n\n"
        
        message += "**Для каналов:**\n"
        message += "1. Перейдите в канал\n"
        message += "2. Нажмите на название канала\n"
        message += "3. Выберите \"Администраторы\"\n"
        message += "4. Нажмите \"Добавить администратора\"\n"
        message += "5. Найдите и добавьте этого бота\n"
        message += "6. Настройте права администратора\n\n"
        
        message += "**Необходимые права:**\n"
        message += "✅ Удаление сообщений\n"
        message += "✅ Блокировка пользователей\n"
        message += "✅ Закрепление сообщений\n"
        message += "✅ Изменение информации\n\n"
        
        message += "**Функции бота:**\n"
        message += "• Модерация контента\n"
        message += "• Управление подписками\n"
        message += "• Антиспам защита\n"
        message += "• Аналитика группы\n"
        message += "• Кастомные команды\n\n"
        
        message += "**Тарифы:**\n"
        message += "• Бесплатный план: 1 группа\n"
        message += "• Premium: 5 групп ($5/месяц)\n"
        message += "• Pro: 20 групп ($10/месяц)\n\n"
        
        message += "Используйте /plans для просмотра тарифов и оплаты."
        
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

// Button-based callback handlers
func (h *CommandHandler) handlePlansCallback(update tgbotapi.Update, user *models.User) {
        plans, err := h.planRepo.GetAll()
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }

        message := "💎 Доступные планы:\n\n"
        var keyboard [][]tgbotapi.InlineKeyboardButton
        
        // Group plans by type to avoid duplicates
        seenPlans := make(map[string]bool)
        
        for _, plan := range plans {
                if plan.PriceCents == 0 {
                        continue
                }
                
                planKey := fmt.Sprintf("%s_%d_%d", plan.Name, plan.PriceCents, plan.DurationDays)
                if seenPlans[planKey] {
                        continue
                }
                seenPlans[planKey] = true
                
                price := fmt.Sprintf("%.2f %s", float64(plan.PriceCents)/100, plan.Currency)
                
                subscribeBtn := tgbotapi.NewInlineKeyboardButtonData(
                        fmt.Sprintf("💎 %s - %s", plan.Name, price),
                        fmt.Sprintf("subscribe:%d", plan.ID),
                )
                keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{subscribeBtn})
        }
        
        backBtn := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu")
        keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{backBtn})
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
        h.bot.Send(msg)
}

func (h *CommandHandler) handleMyPlanCallback(update tgbotapi.Update, user *models.User) {
        plan, err := h.planRepo.GetByID(user.CurrentPlanID)
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "error_occurred"))
                return
        }

        message := fmt.Sprintf("📊 Текущий план: %s\n", plan.Name)
        if user.PlanExpiresAt != nil {
                message += fmt.Sprintf("📅 Истекает: %s\n", user.PlanExpiresAt.Format("2006-01-02"))
        }
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💎 Улучшить план", "show_plans"),
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleHistoryCallback(update tgbotapi.Update, user *models.User) {
        // Get payments for user - simple implementation
        payments := []models.Payment{} // Empty for now
        err := error(nil)
        if err != nil || len(payments) == 0 {
                message := "💰 История платежей пуста"
                keyboard := tgbotapi.NewInlineKeyboardMarkup(
                        tgbotapi.NewInlineKeyboardRow(
                                tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                        ),
                )
                msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
                msg.ReplyMarkup = keyboard
                h.bot.Send(msg)
                return
        }

        message := "💰 История платежей:\n\n"
        message += "Пока нет платежей\n"
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleMyAccountCallback(update tgbotapi.Update, user *models.User) {
        var username, password string
        
        // Check if user already has web account
        if user.WebUsername != "" && user.IsWebActive {
                username = user.WebUsername
                password = user.WebPassword
        } else {
                // Create new web account with random credentials
                username = h.generateRandomUsername()
                password = h.generateRandomPassword()
                
                // Save to database
                err := h.userRepo.CreateWebAccount(user.ID, username, password)
                if err != nil {
                        h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка создания аккаунта")
                        return
                }
        }
        
        // Get website URL and generate token
        websiteURL := "http://localhost:5000"
        if domain := os.Getenv("REPLIT_DOMAIN"); domain != "" {
                websiteURL = fmt.Sprintf("https://%s", domain)
        }
        
        // Generate user token for URL
        userToken := h.generateUserToken(user)
        dashboardURL := fmt.Sprintf("%s/user-dashboard?token=%s", websiteURL, userToken)
        
        message := "🔗 **Ваш личный кабинет**\n\n"
        message += fmt.Sprintf("🌐 Прямая ссылка: `%s`\n\n", dashboardURL)
        message += "📝 Данные для входа:\n"
        message += fmt.Sprintf("👤 Логин: `%s`\n", username)
        message += fmt.Sprintf("🔑 Пароль: `%s`\n\n", password)
        message += "⚠️ **Важно!** Сохраните эти данные в безопасном месте.\n\n"
        message += "В личном кабинете:\n"
        message += "• Статистика использования\n"
        message += "• Управление подписками\n"
        message += "• История платежей\n"
        message += "• Настройки аккаунта\n"
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔑 Сменить пароль", "change_password"),
                        tgbotapi.NewInlineKeyboardButtonData("👤 Сменить логин", "change_username"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *CommandHandler) handleSetupCallback(update tgbotapi.Update, user *models.User) {
        message := "🔧 Настройка бота:\n\n"
        message += "1. Создайте бота в @BotFather\n"
        message += "2. Получите токен\n"
        message += "3. Добавьте в .env файл\n"
        message += "4. Добавьте в группы\n"
        message += "5. Настройте права\n\n"
        message += "Подробная инструкция: /setup"
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleManageGroupsCallback(update tgbotapi.Update, user *models.User) {
        // Проверяем, есть ли веб-аккаунт у пользователя
        if user.WebUsername == "" {
                h.ensureWebAccount(user)
        }
        
        message := "🔧 **Управление группами**\n\n"
        message += "🌐 Откройте веб-интерфейс для управления группами:\n"
        message += fmt.Sprintf("`http://localhost:5000/groups?token=%s`\n\n", h.generateUserToken(user))
        message += "В веб-интерфейсе вы можете:\n"
        message += "• Добавить новые группы и каналы\n"
        message += "• Просмотреть статистику групп\n"
        message += "• Настроить модерацию\n"
        message += "• Тестировать работу бота\n\n"
        message += "📝 **Как добавить группу:**\n"
        message += "1. Добавьте бота в группу: @jnhghyjuiokmuhbgbot\n"
        message += "2. Сделайте бота администратором\n"
        message += "3. Напишите в группе: `/id`\n"
        message += "4. Скопируйте ID и добавьте в веб-интерфейсе"
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleHelpCallback(update tgbotapi.Update, user *models.User) {
        message := "❓ Помощь:\n\n"
        message += "💎 Планы - тарифы подписок\n"
        message += "📊 Мой план - текущий статус\n"
        message += "💰 История - все платежи\n"
        message += "🔗 Аккаунт - личный кабинет\n"
        message += "🔧 Настройка - помощь по настройке\n"
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_to_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleSubscribeCallback(update tgbotapi.Update, user *models.User, planID int) {
        plan, err := h.planRepo.GetByID(planID)
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "plan_not_found"))
                return
        }

        message := fmt.Sprintf("💎 %s\n", plan.Name)
        message += fmt.Sprintf("💰 Цена: %.2f %s\n", float64(plan.PriceCents)/100, plan.Currency)
        message += fmt.Sprintf("⏰ На %d дней\n", plan.DurationDays)
        message += fmt.Sprintf("👥 До %d групп\n\n", plan.MaxGroups)
        message += "Выберите способ оплаты:"

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💳 Карта", fmt.Sprintf("pay_card:%d", planID)),
                        tgbotapi.NewInlineKeyboardButtonData("₿ Крипто", fmt.Sprintf("pay_crypto:%d", planID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "show_plans"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) handleCardPayment(update tgbotapi.Update, user *models.User, planID int) {
        plan, err := h.planRepo.GetByID(planID)
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "plan_not_found"))
                return
        }

        invoice := tgbotapi.NewInvoice(
                update.CallbackQuery.Message.Chat.ID,
                fmt.Sprintf("Подписка %s", plan.Name),
                fmt.Sprintf("Подписка на %d дней", plan.DurationDays),
                fmt.Sprintf("plan_%d", planID),
                "XTR",
                "XTR",
                "test-token",
                []tgbotapi.LabeledPrice{
                        {Label: plan.Name, Amount: plan.PriceCents},
                },
        )
        
        h.bot.Send(invoice)
}

func (h *CommandHandler) handleCryptoPayment(update tgbotapi.Update, user *models.User, planID int) {
        plan, err := h.planRepo.GetByID(planID)
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, locales.GetMessage(user.LanguageCode, "plan_not_found"))
                return
        }

        message := fmt.Sprintf("₿ Крипто оплата\n\n")
        message += fmt.Sprintf("💎 План: %s\n", plan.Name)
        message += fmt.Sprintf("💰 Сумма: %.2f USD\n\n", float64(plan.PriceCents)/100)
        message += "BTC: `1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa`\n"
        message += "ETH: `0x742d35Cc6634C0532925a3b8D41234567890`\n"
        message += "USDT: `TQKHhV5A1234567890abcdefghijklmnopqrst`\n\n"
        message += "Отправьте точную сумму. Активация автоматически."

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", fmt.Sprintf("subscribe:%d", planID)),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ParseMode = "Markdown"
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) sendCallbackMessage(chatID int64, text string) {
        msg := tgbotapi.NewMessage(chatID, text)
        h.bot.Send(msg)
}

func (h *CommandHandler) handleBackToMenu(update tgbotapi.Update, user *models.User) {
        message := locales.GetMessage(user.LanguageCode, "welcome")
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💎 Планы подписок", "show_plans"),
                        tgbotapi.NewInlineKeyboardButtonData("📊 Мой план", "my_plan"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("💰 История платежей", "payment_history"),
                        tgbotapi.NewInlineKeyboardButtonData("🔗 Мой аккаунт", "my_account"),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔧 Настройка", "setup_guide"),
                        tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help_menu"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        h.bot.Send(msg)
}

func (h *CommandHandler) generateRandomPassword() string {
        rand.Seed(time.Now().UnixNano())
        chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        password := make([]byte, 12)
        for i := range password {
                password[i] = chars[rand.Intn(len(chars))]
        }
        return string(password)
}

func getDomainURL() string {
        // Check if running on Replit
        if replit_url := os.Getenv("REPLIT_URL"); replit_url != "" {
                return replit_url
        }
        return "http://localhost:5000"
}

// Helper function for generating random usernames
func (h *CommandHandler) generateRandomUsername() string {
        // Generate truly unique username with timestamp
        timestamp := time.Now().UnixNano() / 1000000 // milliseconds
        randNum := rand.Intn(999) + 100
        
        prefixes := []string{"user", "member", "client", "guest", "player", "viewer", "person", "account", "profile", "visitor"}
        prefix := prefixes[rand.Intn(len(prefixes))]
        
        return fmt.Sprintf("%s_%d_%d", prefix, timestamp, randNum)
}

func (h *CommandHandler) handleChangePasswordCallback(update tgbotapi.Update, user *models.User) {
        newPassword := h.generateRandomPassword()
        
        err := h.userRepo.UpdateWebPassword(user.ID, newPassword)
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, "❌ Ошибка при смене пароля")
                return
        }
        
        message := "🔑 **Пароль успешно изменен!**\n\n"
        message += fmt.Sprintf("Новый пароль: `%s`\n\n", newPassword)
        message += "⚠️ **Важно!** Сохраните новый пароль в безопасном месте."
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "my_account"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}

func (h *CommandHandler) handleChangeUsernameCallback(update tgbotapi.Update, user *models.User) {
        newUsername := h.generateRandomUsername()
        
        err := h.userRepo.UpdateWebUsername(user.ID, newUsername)
        if err != nil {
                h.sendCallbackMessage(update.CallbackQuery.Message.Chat.ID, "❌ Ошибка при смене логина")
                return
        }
        
        message := "👤 **Логин успешно изменен!**\n\n"
        message += fmt.Sprintf("Новый логин: `%s`\n\n", newUsername)
        message += "⚠️ **Важно!** Используйте новый логин для входа на сайт."
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "my_account"),
                ),
        )
        
        msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, message)
        msg.ReplyMarkup = keyboard
        msg.ParseMode = "Markdown"
        h.bot.Send(msg)
}
