package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-subscription-bot/database"
	"telegram-subscription-bot/models"
)

type ModerationHandler struct {
	bot      *tgbotapi.BotAPI
	db       *database.DB
	userRepo *models.UserRepository
}

func NewModerationHandler(bot *tgbotapi.BotAPI, db *database.DB) *ModerationHandler {
	return &ModerationHandler{
		bot:      bot,
		db:       db,
		userRepo: models.NewUserRepository(db.DB),
	}
}

// Структура для хранения настроек модерации
type ModerationSettings struct {
	AutoBanEnabled       bool `json:"auto_ban_enabled"`
	WarningThreshold     int  `json:"warning_threshold"`
	TempBanDuration      int  `json:"temp_ban_duration"`
	PermanentBanThreshold int  `json:"permanent_ban_threshold"`
}

// Структура для нарушений
type Violation struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	ChatID         int64     `json:"chat_id"`
	ViolationType  string    `json:"violation_type"`
	ViolationReason string   `json:"violation_reason"`
	MessageText    string    `json:"message_text"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
}

// Обработка сообщений для модерации
func (h *ModerationHandler) ProcessMessage(message *tgbotapi.Message) {
	if message.Chat.Type == "private" {
		return
	}

	// Проверяем запрещенные слова
	if h.containsForbiddenWords(message.Text) {
		h.handleViolation(message, "forbidden_words", "Использование запрещенных слов")
	}

	// Проверяем спам (слишком много сообщений за короткое время)
	if h.isSpam(message) {
		h.handleViolation(message, "spam", "Спам сообщения")
	}
}

// Проверка на запрещенные слова
func (h *ModerationHandler) containsForbiddenWords(text string) bool {
	if text == "" {
		return false
	}

	text = strings.ToLower(text)
	
	// Получаем список запрещенных слов из базы
	query := `SELECT word FROM forbidden_words WHERE is_active = TRUE`
	rows, err := h.db.DB.Query(query)
	if err != nil {
		log.Printf("Error fetching forbidden words: %v", err)
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			continue
		}
		
		if strings.Contains(text, strings.ToLower(word)) {
			return true
		}
	}

	return false
}

// Проверка на спам
func (h *ModerationHandler) isSpam(message *tgbotapi.Message) bool {
	// Проверяем количество сообщений от пользователя за последние 30 секунд
	query := `
		SELECT COUNT(*) 
		FROM user_violations 
		WHERE user_id = (SELECT id FROM users WHERE telegram_id = $1) 
		AND chat_id = $2 
		AND created_at > NOW() - INTERVAL '30 seconds'
	`
	
	var count int
	err := h.db.DB.QueryRow(query, message.From.ID, message.Chat.ID).Scan(&count)
	if err != nil {
		log.Printf("Error checking spam: %v", err)
		return false
	}

	return count > 5 // Более 5 сообщений за 30 секунд считается спамом
}

// Обработка нарушения
func (h *ModerationHandler) handleViolation(message *tgbotapi.Message, violationType, reason string) {
	// Получаем или создаем пользователя
	user, err := h.userRepo.GetByTelegramID(message.From.ID)
	if err != nil {
		// Создаем нового пользователя
		user = &models.User{
			TelegramID:   message.From.ID,
			Username:     message.From.UserName,
			FirstName:    message.From.FirstName,
			LastName:     message.From.LastName,
			LanguageCode: message.From.LanguageCode,
		}
		if err := h.userRepo.CreateOrUpdate(user); err != nil {
			log.Printf("Error creating user: %v", err)
			return
		}
	}

	// Получаем настройки модерации
	settings := h.getModerationSettings(message.Chat.ID)
	
	// Считаем количество нарушений
	violationCount := h.getViolationCount(user.ID, message.Chat.ID)
	
	var action string
	var duration *time.Time
	
	if violationCount >= settings.PermanentBanThreshold {
		action = "permanent_ban"
		h.banUser(message.Chat.ID, message.From.ID, 0) // Постоянный бан
	} else if violationCount >= settings.WarningThreshold {
		action = "temp_ban"
		banUntil := time.Now().Add(time.Duration(settings.TempBanDuration) * time.Hour)
		duration = &banUntil
		h.banUser(message.Chat.ID, message.From.ID, int(banUntil.Unix()))
	} else {
		action = "warning"
		h.warnUser(message.Chat.ID, message.From.ID, reason)
	}

	// Сохраняем нарушение в базу
	h.saveViolation(user.ID, message.Chat.ID, action, reason, message.Text, duration)
	
	// Удаляем сообщение нарушителя
	h.deleteMessage(message.Chat.ID, message.MessageID)
	
	// Отправляем уведомление в группу
	h.sendModerationNotification(message.Chat.ID, user, action, reason, violationCount+1)
}

// Получение настроек модерации
func (h *ModerationHandler) getModerationSettings(chatID int64) ModerationSettings {
	return ModerationSettings{
		AutoBanEnabled:       true,
		WarningThreshold:     3,
		TempBanDuration:      24,
		PermanentBanThreshold: 5,
	}
}

// Подсчет нарушений пользователя
func (h *ModerationHandler) getViolationCount(userID int, chatID int64) int {
	query := `
		SELECT COUNT(*) 
		FROM user_violations 
		WHERE user_id = $1 AND chat_id = $2 AND is_active = TRUE
	`
	
	var count int
	err := h.db.DB.QueryRow(query, userID, chatID).Scan(&count)
	if err != nil {
		log.Printf("Error getting violation count: %v", err)
		return 0
	}
	
	return count
}

// Бан пользователя
func (h *ModerationHandler) banUser(chatID int64, userID int64, untilDate int) {
	banConfig := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate: int64(untilDate),
	}
	
	if _, err := h.bot.Request(banConfig); err != nil {
		log.Printf("Error banning user: %v", err)
	}
}

// Предупреждение пользователя
func (h *ModerationHandler) warnUser(chatID int64, userID int64, reason string) {
	// Предупреждение отправляется через уведомление в группу
	log.Printf("Warning user %d in chat %d: %s", userID, chatID, reason)
}

// Удаление сообщения
func (h *ModerationHandler) deleteMessage(chatID int64, messageID int) {
	deleteConfig := tgbotapi.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	}
	
	if _, err := h.bot.Request(deleteConfig); err != nil {
		log.Printf("Error deleting message: %v", err)
	}
}

// Сохранение нарушения в базу
func (h *ModerationHandler) saveViolation(userID int, chatID int64, violationType, reason, messageText string, expiresAt *time.Time) {
	query := `
		INSERT INTO user_violations (user_id, chat_id, violation_type, violation_reason, message_text, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	
	_, err := h.db.DB.Exec(query, userID, chatID, violationType, reason, messageText, expiresAt)
	if err != nil {
		log.Printf("Error saving violation: %v", err)
	}
}

// Отправка уведомления о модерации
func (h *ModerationHandler) sendModerationNotification(chatID int64, user *models.User, action, reason string, violationCount int) {
	var message string
	var emoji string
	
	switch action {
	case "warning":
		emoji = "⚠️"
		message = fmt.Sprintf("%s **Предупреждение**\n\n👤 Пользователь: %s\n🔢 Нарушение: %d\n📝 Причина: %s\n\n💡 При достижении 3 предупреждений будет временный бан", 
			emoji, h.getUserMention(user), violationCount, reason)
	case "temp_ban":
		emoji = "🚫"
		message = fmt.Sprintf("%s **Временный бан**\n\n👤 Пользователь: %s\n🔢 Нарушение: %d\n📝 Причина: %s\n⏰ Длительность: 24 часа", 
			emoji, h.getUserMention(user), violationCount, reason)
	case "permanent_ban":
		emoji = "🔒"
		message = fmt.Sprintf("%s **Постоянный бан**\n\n👤 Пользователь: %s\n🔢 Нарушение: %d\n📝 Причина: %s\n⛔ Пользователь заблокирован навсегда", 
			emoji, h.getUserMention(user), violationCount, reason)
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending moderation notification: %v", err)
	}
}

// Получение упоминания пользователя
func (h *ModerationHandler) getUserMention(user *models.User) string {
	if user.Username != "" {
		return fmt.Sprintf("@%s", user.Username)
	}
	return fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.TelegramID)
}

// Получение списка нарушений для группы
func (h *ModerationHandler) GetViolations(chatID int64) ([]Violation, error) {
	query := `
		SELECT v.id, v.user_id, v.chat_id, v.violation_type, v.violation_reason, 
		       v.message_text, v.created_at, v.expires_at, v.is_active,
		       u.username, u.first_name
		FROM user_violations v
		JOIN users u ON v.user_id = u.id
		WHERE v.chat_id = $1 AND v.is_active = TRUE
		ORDER BY v.created_at DESC
		LIMIT 50
	`
	
	rows, err := h.db.DB.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var violations []Violation
	for rows.Next() {
		var v Violation
		err := rows.Scan(&v.ID, &v.UserID, &v.ChatID, &v.ViolationType, &v.ViolationReason,
			&v.MessageText, &v.CreatedAt, &v.ExpiresAt, &v.IsActive, &v.Username, &v.FirstName)
		if err != nil {
			log.Printf("Error scanning violation: %v", err)
			continue
		}
		violations = append(violations, v)
	}
	
	return violations, nil
}

// Команда для отображения статистики нарушений
func (h *ModerationHandler) HandleViolationsCommand(message *tgbotapi.Message) {
	if message.Chat.Type == "private" {
		return
	}

	violations, err := h.GetViolations(message.Chat.ID)
	if err != nil {
		log.Printf("Error getting violations: %v", err)
		return
	}

	if len(violations) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "📊 В этой группе нет активных нарушений")
		h.bot.Send(msg)
		return
	}

	// Группируем по типам нарушений
	warnings := 0
	tempBans := 0
	permanentBans := 0
	
	var violationList strings.Builder
	violationList.WriteString("📊 **Статистика нарушений**\n\n")
	
	for _, v := range violations {
		switch v.ViolationType {
		case "warning":
			warnings++
		case "temp_ban":
			tempBans++
		case "permanent_ban":
			permanentBans++
		}
		
		username := v.Username
		if username == "" {
			username = v.FirstName
		}
		
		violationList.WriteString(fmt.Sprintf("• %s - %s (%s)\n", 
			username, v.ViolationReason, v.CreatedAt.Format("02.01.2006")))
	}
	
	summary := fmt.Sprintf("⚠️ Предупреждения: %d\n🚫 Временные баны: %d\n🔒 Постоянные баны: %d\n\n", 
		warnings, tempBans, permanentBans)
	
	finalMessage := summary + violationList.String()
	
	// Ограничиваем длину сообщения
	if len(finalMessage) > 4000 {
		finalMessage = finalMessage[:4000] + "..."
	}
	
	msg := tgbotapi.NewMessage(message.Chat.ID, finalMessage)
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

// Команда для разблокировки пользователя
func (h *ModerationHandler) HandleUnbanCommand(message *tgbotapi.Message, args []string) {
	if message.Chat.Type == "private" {
		return
	}

	if len(args) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Укажите ID пользователя для разблокировки")
		h.bot.Send(msg)
		return
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат ID пользователя")
		h.bot.Send(msg)
		return
	}

	// Разблокируем пользователя
	unbanConfig := tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: message.Chat.ID,
			UserID: userID,
		},
		OnlyIfBanned: true,
	}

	if _, err := h.bot.Request(unbanConfig); err != nil {
		log.Printf("Error unbanning user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при разблокировке пользователя")
		h.bot.Send(msg)
		return
	}

	// Деактивируем нарушения в базе
	query := `
		UPDATE user_violations 
		SET is_active = FALSE 
		WHERE user_id = (SELECT id FROM users WHERE telegram_id = $1) 
		AND chat_id = $2
	`
	
	_, err = h.db.DB.Exec(query, userID, message.Chat.ID)
	if err != nil {
		log.Printf("Error deactivating violations: %v", err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Пользователь разблокирован")
	h.bot.Send(msg)
}