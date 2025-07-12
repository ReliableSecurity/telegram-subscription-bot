package config

import (
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken string
	DatabaseURL      string
	Debug            bool
	WebDashboard     bool
	
	// Payment providers
	StripeToken       string
	YooMoneyToken     string
	PayPalToken       string
	CryptoPayToken    string
	
	// Subscription plans
	FreeGroupLimit    int
	PremiumPrice      int
	ProPrice          int
	
	// Admin settings
	AdminUserIDs      []int64
	
	// Crypto settings
	BTCAddress        string
	ETHAddress        string
	USDTAddress       string
}

func Load() (*Config, error) {
	cfg := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		Debug:            getBoolEnv("DEBUG", false),
		WebDashboard:     getBoolEnv("WEB_DASHBOARD", true),
		
		StripeToken:       os.Getenv("STRIPE_TOKEN"),
		YooMoneyToken:     os.Getenv("YOOMONEY_TOKEN"),
		PayPalToken:       os.Getenv("PAYPAL_TOKEN"),
		CryptoPayToken:    os.Getenv("CRYPTOPAY_TOKEN"),
		
		FreeGroupLimit:    getIntEnv("FREE_GROUP_LIMIT", 1),
		PremiumPrice:      getIntEnv("PREMIUM_PRICE", 500), // 5.00 USD in cents
		ProPrice:          getIntEnv("PRO_PRICE", 1000),    // 10.00 USD in cents
		
		BTCAddress:        os.Getenv("BTC_ADDRESS"),
		ETHAddress:        os.Getenv("ETH_ADDRESS"),
		USDTAddress:       os.Getenv("USDT_ADDRESS"),
	}
	
	// Parse admin user IDs
	if adminIDs := os.Getenv("ADMIN_USER_IDS"); adminIDs != "" {
		// Parse comma-separated admin IDs
		// Implementation would parse the string and convert to []int64
	}
	
	return cfg, nil
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
