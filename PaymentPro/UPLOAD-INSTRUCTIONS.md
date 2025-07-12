# 🚀 Инструкции по загрузке на GitHub

## Быстрая загрузка на GitHub

### Шаг 1: Создание репозитория на GitHub

1. Откройте [GitHub.com](https://github.com) и войдите в аккаунт
2. Нажмите "+" → "New repository"
3. Настройки репозитория:
   - **Название**: `telegram-subscription-bot`
   - **Описание**: `Comprehensive Telegram bot with payment system and VPS deployment`
   - **Видимость**: Public (рекомендуется) или Private
   - **Инициализация**: ❌ НЕ добавляйте README (у нас уже есть)
   - **Add .gitignore**: ❌ НЕ добавляйте (у нас есть свой)
4. Нажмите "Create repository"

### Шаг 2: Подключение к GitHub

Скопируйте и выполните команды (замените `yourusername` на ваш GitHub username):

```bash
# Добавить GitHub как удаленный репозиторий
git remote add origin https://github.com/yourusername/telegram-subscription-bot.git

# Загрузить все файлы на GitHub
git push -u origin main
```

**Примечание**: При запросе пароля используйте [Personal Access Token](https://github.com/settings/tokens) вместо пароля.

### Шаг 3: Настройка репозитория

После загрузки:

1. **Обновите описание**: Перейдите в настройки репозитория → About
   - Описание: `Telegram bot platform with subscription management, payment processing, AI recommendations, and automated VPS deployment`
   - Теги: `telegram-bot`, `golang`, `payment-processing`, `vps-deployment`, `ai-recommendations`

2. **Включите Issues и Wiki** для поддержки пользователей

3. **Добавьте веб-сайт**: URL вашего домена после развертывания

## Развертывание на VPS

### Получение токена бота

1. Напишите [@BotFather](https://t.me/BotFather) в Telegram
2. Отправьте `/newbot` и следуйте инструкциям
3. Сохраните токен бота (формат: `123456789:ABCdef...`)

### Односкомандное развертывание

```bash
# Установите переменные окружения
export TELEGRAM_BOT_TOKEN="ваш_токен_бота"
export DOMAIN_NAME="yourdomain.com"

# Скачайте и запустите скрипт развертывания
curl -sSL https://raw.githubusercontent.com/yourusername/telegram-subscription-bot/main/deploy-production.sh | bash
```

### Что делает скрипт автоматически:

✅ Устанавливает все зависимости (Go, PostgreSQL, Nginx)
✅ Настраивает базу данных с безопасными паролями
✅ Собирает и развертывает приложение
✅ Настраивает SSL сертификаты (Let's Encrypt)
✅ Конфигурирует firewall и безопасность
✅ Создает автоматические резервные копии
✅ Настраивает мониторинг системы

## После развертывания

### Доступ к панели управления

После завершения развертывания вы увидите:
```
🌐 Website URL: https://yourdomain.com
🔐 Admin Login: admin / [сгенерированный_пароль]
```

### Команды управления

```bash
# Проверка статуса системы
~/status-bot.sh

# Запуск/остановка/перезапуск бота
~/start-bot.sh
~/stop-bot.sh
~/restart-bot.sh

# Обновление с GitHub
~/update-bot.sh

# Создание резервной копии
~/backup-bot.sh

# Проверка здоровья системы
~/health-check.sh
```

### Настройка платежных провайдеров

В веб-панели управления:
1. Перейдите в Настройки → Платежи
2. Добавьте ключи ваших платежных провайдеров:
   - **Stripe**: [stripe.com](https://stripe.com)
   - **YooMoney**: [yoomoney.ru](https://yoomoney.ru)  
   - **PayPal**: [developer.paypal.com](https://developer.paypal.com)

## Тестирование

### Проверка бота
1. Найдите вашего бота в Telegram: `@ваш_бот_username`
2. Отправьте `/start`
3. Протестируйте команды: `/plans`, `/subscribe`, `/help`

### Проверка веб-сайта
1. Откройте `https://yourdomain.com`
2. Войдите с учетными данными администратора
3. Проверьте все разделы панели управления

## Поддержка

Если возникли проблемы:
1. Проверьте логи: `sudo journalctl -u telegram-bot -f`
2. Запустите проверку здоровья: `~/health-check.sh`
3. Обратитесь к [документации](README.md) или создайте Issue на GitHub

---

**Готово!** 🎉 Ваш Telegram бот теперь на GitHub и готов к профессиональному развертыванию!