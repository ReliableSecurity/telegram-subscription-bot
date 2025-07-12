package locales

import "strings"

var messages = map[string]map[string]string{
        "en": {
                "welcome":                     "🎉 Welcome to the Subscription Bot!\n\nI help you manage your subscriptions and access premium features. Use /help to see available commands.",
                "help":                        "🔧 Available Commands:\n\n/start - Welcome message\n/help - Show this help\n/plans - View subscription plans\n/myplan - Check your current plan\n/subscribe <plan_id> - Subscribe to a plan\n/cancel - Cancel subscription\n/history - View payment history\n/crypto <plan_id> <currency> - Pay with crypto\n/setup - Bot setup instructions\n/addbot - How to add bot to group/channel",
                "available_plans":             "💎 Available Subscription Plans:",
                "current_plan":                "Current Plan",
                "expires_at":                  "Expires At",
                "plan_expired":                "Your plan has expired. Please upgrade to continue using premium features.",
                "max_groups":                  "Max Groups",
                "duration":                    "Duration",
                "days":                        "days",
                "free":                        "Free",
                "subscribe":                   "Subscribe",
                "payment_options":             "💳 Payment Options for",
                "price":                       "Price",
                "pay_with_card":               "Pay with Card",
                "invalid_plan_id":             "Invalid plan ID. Please use a valid plan ID.",
                "plan_not_found":              "Plan not found. Please check the plan ID.",
                "free_plan_no_payment":        "Free plan doesn't require payment. It's already active!",
                "subscription_cancelled":      "✅ Your subscription has been cancelled. You are now on the free plan.",
                "error_occurred":              "❌ An error occurred. Please try again later.",
                "payment_history":             "💰 Payment History:",
                "no_payment_history":          "No payment history found.",
                "completed":                   "Completed",
                "pending":                     "Pending",
                "failed":                      "Failed",
                "crypto_usage":                "Usage: /crypto <plan_id> <currency>\nExample: /crypto 2 BTC",
                "crypto_payment_instructions": "💰 Crypto Payment Instructions for",
                "address":                     "Address",
                "amount":                      "Amount",
                "crypto_payment_note":         "⚠️ Please send the exact amount to the address above. Payment will be confirmed automatically within 10 minutes.",
                "payment_successful":          "✅ Payment Successful! Your subscription has been activated.",
                "plan":                        "Plan",
                "payment_not_found":           "Payment not found. Please contact support.",
                "unknown_command":             "Unknown command. Use /help to see available commands.",
                "subscription_expired":        "⚠️ Your subscription has expired. You are now on the free plan.",
                "upgrade_prompt":              "💎 Upgrade to a premium plan to unlock advanced features:\n\n/plans - View available plans",
                "subscription_expiring_3_days": "⏰ Your %s subscription expires in 3 days!",
                "subscription_expiring_1_day":  "⚠️ Your %s subscription expires tomorrow!",
                "renew_prompt":                "🔄 Renew your subscription to continue enjoying premium features:\n\n/plans - View renewal options",
                "payment_failed_reminder":     "❌ Hi %s! Your recent payment failed.",
                "retry_payment_prompt":        "💳 Please try again or contact support if you need help:\n\n/plans - Try payment again",
                "welcome_message":             "🎉 Welcome to our premium subscription service!",
                "getting_started":             "🚀 Getting Started:\n\n1. Use /plans to view available plans\n2. Choose a plan that suits your needs\n3. Complete payment to unlock premium features\n4. Enjoy advanced functionality!",
                "payment_confirmation":        "✅ Payment confirmed for %s plan: $%.2f %s",
                "enjoy_features":              "🎉 Enjoy your premium features! Use /help to see what you can do.",
        },
        "ru": {
                "welcome":                     "🎉 Добро пожаловать в бота подписок!\n\nЯ помогаю управлять подписками и получать доступ к премиум функциям. Используйте /help для просмотра доступных команд.",
                "help":                        "🔧 Доступные команды:\n\n/start - Приветственное сообщение\n/help - Показать эту справку\n/plans - Посмотреть планы подписок\n/myplan - Проверить текущий план\n/subscribe <plan_id> - Подписаться на план\n/cancel - Отменить подписку\n/history - Посмотреть историю платежей\n/crypto <plan_id> <currency> - Оплатить криптой\n/setup - Инструкция по настройке бота\n/addbot - Как добавить бота в группу/канал",
                "available_plans":             "💎 Доступные планы подписок:",
                "current_plan":                "Текущий план",
                "expires_at":                  "Истекает",
                "plan_expired":                "Ваш план истек. Пожалуйста, обновите план для продолжения использования премиум функций.",
                "max_groups":                  "Макс. групп",
                "duration":                    "Длительность",
                "days":                        "дней",
                "free":                        "Бесплатно",
                "subscribe":                   "Подписаться",
                "payment_options":             "💳 Способы оплаты для",
                "price":                       "Цена",
                "pay_with_card":               "Оплатить картой",
                "invalid_plan_id":             "Неверный ID плана. Пожалуйста, используйте действительный ID плана.",
                "plan_not_found":              "План не найден. Пожалуйста, проверьте ID плана.",
                "free_plan_no_payment":        "Бесплатный план не требует оплаты. Он уже активен!",
                "subscription_cancelled":      "✅ Ваша подписка отменена. Теперь у вас бесплатный план.",
                "error_occurred":              "❌ Произошла ошибка. Пожалуйста, попробуйте позже.",
                "payment_history":             "💰 История платежей:",
                "no_payment_history":          "История платежей не найдена.",
                "completed":                   "Завершен",
                "pending":                     "Ожидает",
                "failed":                      "Неудачно",
                "crypto_usage":                "Использование: /crypto <plan_id> <currency>\nПример: /crypto 2 BTC",
                "crypto_payment_instructions": "💰 Инструкции по оплате криптой для",
                "address":                     "Адрес",
                "amount":                      "Сумма",
                "crypto_payment_note":         "⚠️ Пожалуйста, отправьте точную сумму на указанный адрес. Платеж будет подтвержден автоматически в течение 10 минут.",
                "payment_successful":          "✅ Платеж успешен! Ваша подписка активирована.",
                "plan":                        "План",
                "payment_not_found":           "Платеж не найден. Пожалуйста, обратитесь в поддержку.",
                "unknown_command":             "Неизвестная команда. Используйте /help для просмотра доступных команд.",
                "subscription_expired":        "⚠️ Ваша подписка истекла. Теперь у вас бесплатный план.",
                "upgrade_prompt":              "💎 Обновитесь до премиум плана для разблокировки дополнительных функций:\n\n/plans - Посмотреть доступные планы",
                "subscription_expiring_3_days": "⏰ Ваша подписка %s истекает через 3 дня!",
                "subscription_expiring_1_day":  "⚠️ Ваша подписка %s истекает завтра!",
                "renew_prompt":                "🔄 Продлите подписку для продолжения использования премиум функций:\n\n/plans - Посмотреть варианты продления",
                "payment_failed_reminder":     "❌ Привет %s! Ваш последний платеж не прошел.",
                "retry_payment_prompt":        "💳 Пожалуйста, попробуйте еще раз или обратитесь в поддержку:\n\n/plans - Попробовать оплату снова",
                "welcome_message":             "🎉 Добро пожаловать в наш премиум сервис подписок!",
                "getting_started":             "🚀 Начало работы:\n\n1. Используйте /plans для просмотра доступных планов\n2. Выберите подходящий план\n3. Завершите оплату для разблокировки премиум функций\n4. Наслаждайтесь расширенным функционалом!",
                "payment_confirmation":        "✅ Платеж подтвержден для плана %s: $%.2f %s",
                "enjoy_features":              "🎉 Наслаждайтесь вашими премиум функциями! Используйте /help чтобы узнать что вы можете делать.",
        },
}

func GetMessage(languageCode, key string) string {
        // Default to English if language not supported
        if languageCode == "" {
                languageCode = "en"
        }
        
        // Normalize language code (e.g., "en-US" -> "en")
        if strings.Contains(languageCode, "-") {
                languageCode = strings.Split(languageCode, "-")[0]
        }
        
        if lang, exists := messages[languageCode]; exists {
                if message, exists := lang[key]; exists {
                        return message
                }
        }
        
        // Fallback to English
        if enLang, exists := messages["en"]; exists {
                if message, exists := enLang[key]; exists {
                        return message
                }
        }
        
        // Last resort fallback
        return key
}

func GetSupportedLanguages() []string {
        var languages []string
        for lang := range messages {
                languages = append(languages, lang)
        }
        return languages
}

func AddLanguage(languageCode string, translations map[string]string) {
        messages[languageCode] = translations
}

func UpdateTranslation(languageCode, key, value string) {
        if _, exists := messages[languageCode]; !exists {
                messages[languageCode] = make(map[string]string)
        }
        messages[languageCode][key] = value
}
