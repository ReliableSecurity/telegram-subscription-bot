package main

import (
        "log"
        "os"
        "os/signal"
        "strings"
        "syscall"

        "github.com/gin-gonic/gin"
        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
        "github.com/joho/godotenv"
        "telegram-subscription-bot/config"
        "telegram-subscription-bot/database"
        "telegram-subscription-bot/handlers"
        "telegram-subscription-bot/models"
        "telegram-subscription-bot/services"
        "telegram-subscription-bot/utils"
        "telegram-subscription-bot/web"
)

func main() {
        // Load .env file
        if err := godotenv.Load(); err != nil {
                log.Println("Warning: .env file not found")
        }
        
        // Initialize logger
        logger := utils.NewLogger()
        
        // Load configuration
        cfg, err := config.Load()
        if err != nil {
                log.Fatal("Failed to load configuration:", err)
        }

        // Initialize database
        db, err := database.Connect(cfg.DatabaseURL)
        if err != nil {
                log.Fatal("Failed to connect to database:", err)
        }
        defer db.Close()

        // Run migrations
        if err := database.Migrate(db); err != nil {
                log.Fatal("Failed to run migrations:", err)
        }

        // Initialize Telegram bot
        bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
        if err != nil {
                log.Fatal("Failed to create bot:", err)
        }

        bot.Debug = cfg.Debug
        logger.Info("Authorized on account %s", bot.Self.UserName)

        // Initialize services
        paymentService := services.NewPaymentService(db, cfg)
        subscriptionService := services.NewSubscriptionService(db)
        notificationService := services.NewNotificationService(bot, db)

        // Initialize repositories
        userRepo := models.NewUserRepository(db.DB)
        paymentRepo := models.NewPaymentRepository(db.DB)
        
        // Initialize handlers
        commandHandler := handlers.NewCommandHandler(bot, db, subscriptionService, paymentService)
        paymentHandler := handlers.NewPaymentHandler(bot, userRepo, paymentRepo)
        adminHandler := handlers.NewAdminHandler(bot, db, subscriptionService, paymentService)
        moderationHandler := handlers.NewModerationHandler(bot, db)

        // Start notification service
        go notificationService.Start()

        // Start web dashboard
        go startWebDashboard(db, cfg)

        // Start bot polling
        u := tgbotapi.NewUpdate(0)
        u.Timeout = 60

        updates := bot.GetUpdatesChan(u)

        // Handle graceful shutdown
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)

        go func() {
                for update := range updates {
                        go handleUpdate(update, commandHandler, paymentHandler, adminHandler, moderationHandler, logger)
                }
        }()

        <-c
        logger.Info("Bot stopped")
}

func handleUpdate(update tgbotapi.Update, commandHandler *handlers.CommandHandler, paymentHandler *handlers.PaymentHandler, adminHandler *handlers.AdminHandler, moderationHandler *handlers.ModerationHandler, logger *utils.Logger) {
        defer func() {
                if r := recover(); r != nil {
                        logger.Error("Panic in update handler: %v", r)
                }
        }()

        if update.Message != nil && update.Message.IsCommand() {
                // Проверяем команды модерации
                switch update.Message.Command() {
                case "violations":
                        moderationHandler.HandleViolationsCommand(update.Message)
                case "unban":
                        args := strings.Split(update.Message.CommandArguments(), " ")
                        moderationHandler.HandleUnbanCommand(update.Message, args)
                default:
                        commandHandler.Handle(update)
                }
        } else if update.PreCheckoutQuery != nil {
                paymentHandler.HandleTelegramPayment(update)
        } else if update.Message != nil && update.Message.SuccessfulPayment != nil {
                paymentHandler.HandleTelegramPayment(update)
        } else if update.CallbackQuery != nil {
                commandHandler.HandleCallback(update)
        } else if update.Message != nil && update.Message.Text != "" {
                // Проверяем сообщения на нарушения
                moderationHandler.ProcessMessage(update.Message)
        }
}

func startWebDashboard(db *database.DB, cfg *config.Config) {
        if !cfg.WebDashboard {
                return
        }

        gin.SetMode(gin.ReleaseMode)
        r := gin.New()
        r.Use(gin.Recovery())

        dashboard := web.NewDashboard(db)
        dashboard.SetupRoutes(r)

        if err := r.Run(":5000"); err != nil {
                log.Printf("Failed to start web dashboard: %v", err)
        }
}
