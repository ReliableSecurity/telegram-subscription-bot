# Payment Integration Guide

## Overview
This document explains how to integrate various payment providers with the Telegram Bot system.

## Supported Payment Providers

### 1. Stripe (Credit Cards)
**Setup:**
1. Create account at https://stripe.com
2. Get API keys from Dashboard → API keys
3. Add to `.env` file:
```bash
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLISHABLE_KEY=pk_live_...
```

**Implementation:**
```go
// In handlers/payment.go
func handleStripePayment(amount int, currency string) error {
    stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
    
    params := &stripe.PaymentIntentParams{
        Amount:   stripe.Int64(int64(amount)),
        Currency: stripe.String(currency),
    }
    
    _, err := paymentintent.New(params)
    return err
}
```

### 2. YooMoney (Russian payments)
**Setup:**
1. Register at https://yoomoney.ru
2. Get API keys from developer console
3. Add to `.env` file:
```bash
YOOMONEY_SECRET_KEY=...
YOOMONEY_SHOP_ID=...
```

**Implementation:**
```go
func handleYooMoneyPayment(amount int, description string) error {
    client := &http.Client{}
    payload := map[string]interface{}{
        "amount": map[string]interface{}{
            "value":    fmt.Sprintf("%.2f", float64(amount)/100),
            "currency": "RUB",
        },
        "confirmation": map[string]interface{}{
            "type":       "redirect",
            "return_url": "https://yourdomain.com/payment/success",
        },
        "description": description,
    }
    
    jsonPayload, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", "https://api.yoomoney.ru/v3/payments", bytes.NewBuffer(jsonPayload))
    req.Header.Set("Authorization", "Bearer "+os.Getenv("YOOMONEY_SECRET_KEY"))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    return err
}
```

### 3. PayPal
**Setup:**
1. Create account at https://developer.paypal.com
2. Create app and get client ID/secret
3. Add to `.env` file:
```bash
PAYPAL_CLIENT_ID=...
PAYPAL_CLIENT_SECRET=...
PAYPAL_MODE=sandbox # or live
```

**Implementation:**
```go
func handlePayPalPayment(amount int, currency string) error {
    client, err := paypal.NewClient(
        os.Getenv("PAYPAL_CLIENT_ID"),
        os.Getenv("PAYPAL_CLIENT_SECRET"),
        paypal.APIBaseSandBox, // or paypal.APIBaseLive
    )
    
    if err != nil {
        return err
    }
    
    order := paypal.Order{
        Intent: "CAPTURE",
        PurchaseUnits: []paypal.PurchaseUnit{
            {
                Amount: &paypal.PurchaseUnitAmount{
                    Currency: currency,
                    Value:    fmt.Sprintf("%.2f", float64(amount)/100),
                },
            },
        },
    }
    
    _, err = client.CreateOrder(context.Background(), order)
    return err
}
```

### 4. Cryptocurrency
**Setup:**
1. Choose a crypto payment processor (BTCPay, CoinGate, etc.)
2. Or use direct wallet addresses
3. Add to `.env` file:
```bash
CRYPTO_WALLET_ADDRESS=1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
CRYPTO_PROCESSOR_API_KEY=...
```

**Implementation:**
```go
func handleCryptoPayment(amount int, currency string) (string, error) {
    // For BTCPay Server
    client := &http.Client{}
    payload := map[string]interface{}{
        "amount":      fmt.Sprintf("%.2f", float64(amount)/100),
        "currency":    currency,
        "orderId":     generateOrderID(),
        "redirectURL": "https://yourdomain.com/payment/success",
        "notificationURL": "https://yourdomain.com/webhook/crypto",
    }
    
    jsonPayload, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", "https://btcpay.server/api/v1/invoices", bytes.NewBuffer(jsonPayload))
    req.Header.Set("Authorization", "token "+os.Getenv("CRYPTO_PROCESSOR_API_KEY"))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result["url"].(string), nil
}
```

## Telegram Native Payments

### Setup
1. Contact @BotFather on Telegram
2. Use `/mybots` → Select bot → Payments
3. Choose payment provider and get provider token
4. Add to `.env` file:
```bash
TELEGRAM_PAYMENT_PROVIDER_TOKEN=...
```

### Implementation
```go
func sendInvoice(chatID int64, title, description string, amount int) error {
    invoice := tgbotapi.InvoiceConfig{
        BaseChat: tgbotapi.BaseChat{
            ChatID: chatID,
        },
        Title:       title,
        Description: description,
        Payload:     fmt.Sprintf("payment_%d_%d", chatID, time.Now().Unix()),
        Currency:    "USD",
        Prices: []tgbotapi.LabeledPrice{
            {
                Label:  title,
                Amount: amount * 100, // Amount in cents
            },
        },
        ProviderToken: os.Getenv("TELEGRAM_PAYMENT_PROVIDER_TOKEN"),
    }
    
    _, err := bot.Send(invoice)
    return err
}
```

## Webhook Handling

### Payment Success Webhook
```go
func handlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
    var payment PaymentNotification
    if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Verify payment signature
    if !verifyPaymentSignature(payment, r.Header.Get("X-Signature")) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Process payment
    err := processPayment(payment)
    if err != nil {
        http.Error(w, "Payment processing failed", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}
```

### Telegram Payment Handler
```go
func handleTelegramPayment(update tgbotapi.Update) {
    if update.PreCheckoutQuery != nil {
        // Answer pre-checkout query
        preCheckoutConfig := tgbotapi.PreCheckoutConfig{
            PreCheckoutQueryID: update.PreCheckoutQuery.ID,
            OK:                 true,
        }
        bot.Send(preCheckoutConfig)
    }
    
    if update.Message.SuccessfulPayment != nil {
        // Payment successful
        payment := update.Message.SuccessfulPayment
        
        // Process subscription
        err := activateSubscription(
            update.Message.Chat.ID,
            payment.InvoicePayload,
            payment.TotalAmount,
        )
        
        if err != nil {
            log.Printf("Failed to activate subscription: %v", err)
            return
        }
        
        // Send confirmation
        msg := tgbotapi.NewMessage(
            update.Message.Chat.ID,
            "✅ Payment successful! Your subscription is now active.",
        )
        bot.Send(msg)
    }
}
```

## Database Schema for Payments

```sql
-- Payments table
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,
    payment_provider VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(255) UNIQUE,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payment provider tokens
CREATE TABLE payment_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    api_key_encrypted TEXT,
    webhook_secret TEXT,
    configuration JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Subscription activations
CREATE TABLE subscription_activations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    payment_id INTEGER REFERENCES payments(id),
    plan_id INTEGER REFERENCES subscription_plans(id),
    activated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);
```

## Security Best Practices

1. **Always verify webhook signatures**
2. **Use HTTPS for all payment endpoints**
3. **Store API keys encrypted**
4. **Implement rate limiting**
5. **Log all payment transactions**
6. **Use idempotency keys**
7. **Validate payment amounts**
8. **Implement fraud detection**

## Testing

### Test Cards (Stripe)
- Success: 4242424242424242
- Decline: 4000000000000002
- Insufficient funds: 4000000000009995

### Test Environment Setup
```bash
# Use test/sandbox APIs
STRIPE_SECRET_KEY=sk_test_...
PAYPAL_MODE=sandbox
YOOMONEY_MODE=test
```

## Monitoring & Analytics

### Payment Metrics
```go
func getPaymentMetrics() PaymentMetrics {
    metrics := PaymentMetrics{}
    
    // Total revenue
    db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed'").Scan(&metrics.TotalRevenue)
    
    // Today's revenue
    db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed' AND DATE(created_at) = CURRENT_DATE").Scan(&metrics.TodayRevenue)
    
    // Success rate
    var total, successful int
    db.QueryRow("SELECT COUNT(*) FROM payments WHERE created_at > NOW() - INTERVAL '30 days'").Scan(&total)
    db.QueryRow("SELECT COUNT(*) FROM payments WHERE status = 'completed' AND created_at > NOW() - INTERVAL '30 days'").Scan(&successful)
    
    if total > 0 {
        metrics.SuccessRate = float64(successful) / float64(total) * 100
    }
    
    return metrics
}
```

## Troubleshooting

### Common Issues
1. **Payment failed**: Check API keys and network connectivity
2. **Webhook not received**: Verify URL and SSL certificate
3. **Invalid signature**: Check webhook secret configuration
4. **Currency mismatch**: Ensure currency codes match
5. **Amount validation**: Check for proper amount formatting

### Debug Mode
```bash
# Enable debug logging
export PAYMENT_DEBUG=true
```

This comprehensive guide covers all aspects of payment integration for your Telegram bot system.