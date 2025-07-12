package handlers

import (
        "bytes"
        "crypto/hmac"
        "crypto/sha256"
        "encoding/hex"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "os"
        "strconv"
        "time"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
        "telegram-subscription-bot/models"
)

type PaymentHandler struct {
        bot        *tgbotapi.BotAPI
        userRepo   *models.UserRepository
        paymentRepo *models.PaymentRepository
}

func NewPaymentHandler(bot *tgbotapi.BotAPI, userRepo *models.UserRepository, paymentRepo *models.PaymentRepository) *PaymentHandler {
        return &PaymentHandler{
                bot:         bot,
                userRepo:    userRepo,
                paymentRepo: paymentRepo,
        }
}

// Stripe Payment Integration
func (h *PaymentHandler) CreateStripePayment(userID int64, amount int, currency string, description string) (*models.Payment, error) {
        stripeKey := os.Getenv("STRIPE_SECRET_KEY")
        if stripeKey == "" {
                return nil, fmt.Errorf("stripe secret key not configured")
        }

        // Create payment record
        payment := &models.Payment{
                UserID:          userID,
                Amount:          amount,
                Currency:        currency,
                PaymentMethod:   "credit_card",
                PaymentProvider: "stripe",
                Status:          "pending",
                Description:     description,
                CreatedAt:       time.Now(),
        }

        // Save to database
        err := h.paymentRepo.Create(payment)
        if err != nil {
                return nil, err
        }

        // Create Stripe payment intent
        paymentIntent, err := h.createStripePaymentIntent(amount, currency, payment.ID)
        if err != nil {
                payment.Status = "failed"
                h.paymentRepo.Update(payment)
                return nil, err
        }

        payment.TransactionID = paymentIntent.ID
        payment.Status = "processing"
        h.paymentRepo.Update(payment)

        return payment, nil
}

func (h *PaymentHandler) createStripePaymentIntent(amount int, currency string, paymentID int64) (*StripePaymentIntent, error) {
        url := "https://api.stripe.com/v1/payment_intents"
        
        data := fmt.Sprintf("amount=%d&currency=%s&metadata[payment_id]=%d", amount, currency, paymentID)
        
        req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
        if err != nil {
                return nil, err
        }

        req.Header.Set("Authorization", "Bearer "+os.Getenv("STRIPE_SECRET_KEY"))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        var paymentIntent StripePaymentIntent
        if err := json.NewDecoder(resp.Body).Decode(&paymentIntent); err != nil {
                return nil, err
        }

        return &paymentIntent, nil
}

// YooMoney Payment Integration
func (h *PaymentHandler) CreateYooMoneyPayment(userID int64, amount int, description string) (*models.Payment, error) {
        yooMoneyKey := os.Getenv("YOOMONEY_SECRET_KEY")
        if yooMoneyKey == "" {
                return nil, fmt.Errorf("yoomoney secret key not configured")
        }

        payment := &models.Payment{
                UserID:          userID,
                Amount:          amount,
                Currency:        "RUB",
                PaymentMethod:   "yoomoney",
                PaymentProvider: "yoomoney",
                Status:          "pending",
                Description:     description,
                CreatedAt:       time.Now(),
        }

        err := h.paymentRepo.Create(payment)
        if err != nil {
                return nil, err
        }

        // Create YooMoney payment
        yooPayment, err := h.createYooMoneyPayment(amount, description, payment.ID)
        if err != nil {
                payment.Status = "failed"
                h.paymentRepo.Update(payment)
                return nil, err
        }

        payment.TransactionID = yooPayment.ID
        payment.Status = "processing"
        h.paymentRepo.Update(payment)

        return payment, nil
}

func (h *PaymentHandler) createYooMoneyPayment(amount int, description string, paymentID int64) (*YooMoneyPayment, error) {
        url := "https://api.yoomoney.ru/v3/payments"
        
        payload := map[string]interface{}{
                "amount": map[string]interface{}{
                        "value":    fmt.Sprintf("%.2f", float64(amount)/100),
                        "currency": "RUB",
                },
                "confirmation": map[string]interface{}{
                        "type":       "redirect",
                        "return_url": fmt.Sprintf("%s/payment/success", os.Getenv("DOMAIN")),
                },
                "description": description,
                "metadata": map[string]interface{}{
                        "payment_id": paymentID,
                },
        }

        jsonPayload, err := json.Marshal(payload)
        if err != nil {
                return nil, err
        }

        req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
        if err != nil {
                return nil, err
        }

        req.Header.Set("Authorization", "Bearer "+os.Getenv("YOOMONEY_SECRET_KEY"))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Idempotence-Key", fmt.Sprintf("payment_%d_%d", paymentID, time.Now().Unix()))

        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        var payment YooMoneyPayment
        if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
                return nil, err
        }

        return &payment, nil
}

// Telegram Native Payment Integration
func (h *PaymentHandler) SendTelegramInvoice(chatID int64, planID int, plan *models.SubscriptionPlan) error {
        providerToken := os.Getenv("TELEGRAM_PAYMENT_PROVIDER_TOKEN")
        if providerToken == "" {
                return fmt.Errorf("telegram payment provider token not configured")
        }

        invoice := tgbotapi.InvoiceConfig{
                BaseChat: tgbotapi.BaseChat{
                        ChatID: chatID,
                },
                Title:       fmt.Sprintf("Подписка %s", plan.Name),
                Description: plan.Description,
                Payload:     fmt.Sprintf("plan_%d_%d", planID, time.Now().Unix()),
                Currency:    plan.Currency,
                Prices: []tgbotapi.LabeledPrice{
                        {
                                Label:  plan.Name,
                                Amount: int(plan.PriceCents),
                        },
                },
                ProviderToken: providerToken,
                StartParameter: fmt.Sprintf("pay_%d", planID),
                PhotoURL:      "",
                PhotoSize:     0,
                PhotoWidth:    0,
                PhotoHeight:   0,
                NeedName:      false,
                NeedPhoneNumber: false,
                NeedEmail:     false,
                NeedShippingAddress: false,
                SendPhoneNumberToProvider: false,
                SendEmailToProvider: false,
                IsFlexible:    false,
        }

        _, err := h.bot.Send(invoice)
        return err
}

// Cryptocurrency Payment Integration
func (h *PaymentHandler) CreateCryptoPayment(userID int64, amount int, currency string, description string) (*models.Payment, error) {
        cryptoProcessor := os.Getenv("CRYPTO_PROCESSOR_API_KEY")
        if cryptoProcessor == "" {
                return nil, fmt.Errorf("crypto processor not configured")
        }

        payment := &models.Payment{
                UserID:          userID,
                Amount:          amount,
                Currency:        currency,
                PaymentMethod:   "cryptocurrency",
                PaymentProvider: "crypto",
                Status:          "pending",
                Description:     description,
                CreatedAt:       time.Now(),
        }

        err := h.paymentRepo.Create(payment)
        if err != nil {
                return nil, err
        }

        // Create crypto payment via BTCPay or similar
        cryptoPayment, err := h.createCryptoInvoice(amount, currency, description, payment.ID)
        if err != nil {
                payment.Status = "failed"
                h.paymentRepo.Update(payment)
                return nil, err
        }

        payment.TransactionID = cryptoPayment.ID
        payment.Status = "processing"
        h.paymentRepo.Update(payment)

        return payment, nil
}

func (h *PaymentHandler) createCryptoInvoice(amount int, currency string, description string, paymentID int64) (*CryptoPayment, error) {
        // This is a placeholder for BTCPay Server integration
        // Replace with actual implementation based on your crypto payment processor
        
        url := fmt.Sprintf("%s/api/v1/invoices", os.Getenv("CRYPTO_PROCESSOR_URL"))
        
        payload := map[string]interface{}{
                "amount":          fmt.Sprintf("%.2f", float64(amount)/100),
                "currency":        currency,
                "orderId":         fmt.Sprintf("order_%d", paymentID),
                "itemDesc":        description,
                "redirectURL":     fmt.Sprintf("%s/payment/success", os.Getenv("DOMAIN")),
                "notificationURL": fmt.Sprintf("%s/webhook/crypto", os.Getenv("DOMAIN")),
                "metadata": map[string]interface{}{
                        "payment_id": paymentID,
                },
        }

        jsonPayload, err := json.Marshal(payload)
        if err != nil {
                return nil, err
        }

        req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
        if err != nil {
                return nil, err
        }

        req.Header.Set("Authorization", "token "+os.Getenv("CRYPTO_PROCESSOR_API_KEY"))
        req.Header.Set("Content-Type", "application/json")

        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        var payment CryptoPayment
        if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
                return nil, err
        }

        return &payment, nil
}

// Webhook handlers
func (h *PaymentHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
        payload, err := io.ReadAll(r.Body)
        if err != nil {
                http.Error(w, "Error reading request body", http.StatusBadRequest)
                return
        }

        // Verify webhook signature
        signature := r.Header.Get("Stripe-Signature")
        if !h.verifyStripeSignature(payload, signature) {
                http.Error(w, "Invalid signature", http.StatusUnauthorized)
                return
        }

        var event StripeEvent
        if err := json.Unmarshal(payload, &event); err != nil {
                http.Error(w, "Error parsing JSON", http.StatusBadRequest)
                return
        }

        switch event.Type {
        case "payment_intent.succeeded":
                h.handleStripePaymentSuccess(event.Data.Object)
        case "payment_intent.payment_failed":
                h.handleStripePaymentFailed(event.Data.Object)
        }

        w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) HandleYooMoneyWebhook(w http.ResponseWriter, r *http.Request) {
        var notification YooMoneyNotification
        if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
                http.Error(w, "Invalid JSON", http.StatusBadRequest)
                return
        }

        // Verify notification
        if !h.verifyYooMoneySignature(notification, r.Header.Get("X-Signature")) {
                http.Error(w, "Invalid signature", http.StatusUnauthorized)
                return
        }

        // Process payment
        paymentID, _ := strconv.ParseInt(notification.Object.Metadata["payment_id"], 10, 64)
        payment, err := h.paymentRepo.GetByID(paymentID)
        if err != nil {
                http.Error(w, "Payment not found", http.StatusNotFound)
                return
        }

        if notification.Event == "payment.succeeded" {
                h.processSuccessfulPayment(payment)
        } else if notification.Event == "payment.canceled" {
                payment.Status = "cancelled"
                h.paymentRepo.Update(payment)
        }

        w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) HandleTelegramPayment(update tgbotapi.Update) {
        if update.PreCheckoutQuery != nil {
                // Answer pre-checkout query
                preCheckoutConfig := tgbotapi.PreCheckoutConfig{
                        PreCheckoutQueryID: update.PreCheckoutQuery.ID,
                        OK:                 true,
                }
                h.bot.Send(preCheckoutConfig)
        }

        if update.Message != nil && update.Message.SuccessfulPayment != nil {
                // Payment successful
                payment := update.Message.SuccessfulPayment
                chatID := update.Message.Chat.ID

                // Extract plan ID from payload
                var planID int
                fmt.Sscanf(payment.InvoicePayload, "plan_%d_", &planID)

                // Process subscription
                err := h.activateSubscription(chatID, int64(planID), payment.TotalAmount)
                if err != nil {
                        fmt.Printf("Failed to activate subscription: %v\n", err)
                        return
                }

                // Send confirmation
                msg := tgbotapi.NewMessage(chatID, "✅ Оплата успешна! Ваша подписка активирована.")
                h.bot.Send(msg)
        }
}

// Helper functions
func (h *PaymentHandler) verifyStripeSignature(payload []byte, signature string) bool {
        webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
        if webhookSecret == "" {
                return false
        }

        mac := hmac.New(sha256.New, []byte(webhookSecret))
        mac.Write(payload)
        expected := hex.EncodeToString(mac.Sum(nil))

        return hmac.Equal([]byte(signature), []byte(expected))
}

func (h *PaymentHandler) verifyYooMoneySignature(notification YooMoneyNotification, signature string) bool {
        webhookSecret := os.Getenv("YOOMONEY_WEBHOOK_SECRET")
        if webhookSecret == "" {
                return false
        }

        data := fmt.Sprintf("%s&%s&%s", notification.Event, notification.Object.ID, webhookSecret)
        mac := hmac.New(sha256.New, []byte(webhookSecret))
        mac.Write([]byte(data))
        expected := hex.EncodeToString(mac.Sum(nil))

        return hmac.Equal([]byte(signature), []byte(expected))
}

func (h *PaymentHandler) processSuccessfulPayment(payment *models.Payment) error {
        payment.Status = "completed"
        payment.CompletedAt = time.Now()
        
        err := h.paymentRepo.Update(payment)
        if err != nil {
                return err
        }

        // Activate subscription
        return h.activateSubscription(payment.UserID, payment.PlanID, payment.Amount)
}

func (h *PaymentHandler) activateSubscription(userID int64, planID int64, amount int) error {
        // Get user
        user, err := h.userRepo.GetByTelegramID(userID)
        if err != nil {
                return err
        }

        // Get plan
        plan, err := h.paymentRepo.GetPlanByID(planID)
        if err != nil {
                return err
        }

        // Calculate expiration date
        expirationDate := time.Now().AddDate(0, 0, plan.DurationDays)

        // Update user subscription
        user.CurrentPlanID = int(planID)
        user.PlanExpiresAt = &expirationDate

        return h.userRepo.CreateOrUpdate(user)
}

func (h *PaymentHandler) handleStripePaymentSuccess(paymentIntent map[string]interface{}) {
        paymentID, _ := strconv.ParseInt(paymentIntent["metadata"].(map[string]interface{})["payment_id"].(string), 10, 64)
        
        payment, err := h.paymentRepo.GetByID(paymentID)
        if err != nil {
                fmt.Printf("Payment not found: %v\n", err)
                return
        }

        h.processSuccessfulPayment(payment)
}

func (h *PaymentHandler) handleStripePaymentFailed(paymentIntent map[string]interface{}) {
        paymentID, _ := strconv.ParseInt(paymentIntent["metadata"].(map[string]interface{})["payment_id"].(string), 10, 64)
        
        payment, err := h.paymentRepo.GetByID(paymentID)
        if err != nil {
                fmt.Printf("Payment not found: %v\n", err)
                return
        }

        payment.Status = "failed"
        h.paymentRepo.Update(payment)
}

// Data structures
type StripePaymentIntent struct {
        ID     string `json:"id"`
        Amount int    `json:"amount"`
        Currency string `json:"currency"`
        Status string `json:"status"`
        ClientSecret string `json:"client_secret"`
}

type StripeEvent struct {
        Type string `json:"type"`
        Data struct {
                Object map[string]interface{} `json:"object"`
        } `json:"data"`
}

type YooMoneyPayment struct {
        ID     string `json:"id"`
        Status string `json:"status"`
        Amount struct {
                Value    string `json:"value"`
                Currency string `json:"currency"`
        } `json:"amount"`
        Confirmation struct {
                Type            string `json:"type"`
                ConfirmationURL string `json:"confirmation_url"`
        } `json:"confirmation"`
}

type YooMoneyNotification struct {
        Event  string `json:"event"`
        Object struct {
                ID       string            `json:"id"`
                Status   string            `json:"status"`
                Metadata map[string]string `json:"metadata"`
        } `json:"object"`
}

type CryptoPayment struct {
        ID     string `json:"id"`
        Status string `json:"status"`
        URL    string `json:"url"`
        Amount string `json:"amount"`
        Currency string `json:"currency"`
}