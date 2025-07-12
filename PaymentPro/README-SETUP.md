# Complete Setup Guide for Telegram Bot Payment System

## 🚀 Quick Start (Local Development)

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

### 1. Clone and Setup
```bash
git clone <repository-url>
cd telegram-bot
cp .env.example .env
```

### 2. Configure Environment
Edit `.env` file with your credentials:
```bash
# Bot Configuration
BOT_TOKEN=your_bot_token_here
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123

# Database
DATABASE_URL=postgres://username:password@localhost:5432/telegram_bot

# Payment Providers (optional for testing)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
YOOMONEY_SECRET_KEY=...
TELEGRAM_PAYMENT_PROVIDER_TOKEN=...
```

### 3. Run Locally
```bash
go mod tidy
go run main.go
```

Access: http://localhost:5000

## 🌐 VPS Production Deployment

### Automated Deployment Script
```bash
# Download and run the deployment script
curl -fsSL https://raw.githubusercontent.com/your-repo/deploy-vps.sh -o deploy-vps.sh
chmod +x deploy-vps.sh

# Deploy with domain and bot token
sudo ./deploy-vps.sh yourdomain.com your_bot_token admin_password
```

### Manual VPS Setup

#### 1. Server Requirements
- Ubuntu 20.04+ or Debian 11+
- 2GB RAM minimum
- 20GB disk space
- Domain name pointed to server IP

#### 2. Initial Server Setup
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y curl wget git nginx postgresql postgresql-contrib certbot python3-certbot-nginx ufw fail2ban

# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Database Setup
```bash
# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql
CREATE DATABASE telegram_bot;
CREATE USER botuser WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE telegram_bot TO botuser;
\q
```

#### 4. Application Setup
```bash
# Create application directory
sudo mkdir -p /opt/telegram-bot
cd /opt/telegram-bot

# Clone your repository
git clone <your-repo-url> .

# Configure environment
sudo cp .env.example .env
sudo nano .env  # Edit with your configuration

# Build application
go mod tidy
go build -o telegram-bot main.go
```

#### 5. Nginx Configuration
```bash
# Create nginx site
sudo nano /etc/nginx/sites-available/telegram-bot
```

Add this configuration:
```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    
    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
    
    location /webhook/ {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/telegram-bot /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

#### 6. SSL Certificate
```bash
# Get SSL certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

#### 7. Systemd Service
```bash
# Create service file
sudo nano /etc/systemd/system/telegram-bot.service
```

Add this configuration:
```ini
[Unit]
Description=Telegram Bot Service
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/telegram-bot
ExecStart=/opt/telegram-bot/telegram-bot
Restart=always
RestartSec=10
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=/opt/telegram-bot/.env

[Install]
WantedBy=multi-user.target
```

```bash
# Start service
sudo systemctl daemon-reload
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

#### 8. Firewall Setup
```bash
# Configure UFW
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw enable
```

## 💳 Payment Provider Setup

### Stripe (Credit Cards)
1. Create account at https://stripe.com
2. Get API keys from Dashboard → API keys
3. Add webhooks endpoint: `https://yourdomain.com/webhook/stripe`
4. Add keys to `.env`:
```bash
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLISHABLE_KEY=pk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

### YooMoney (Russian Payments)
1. Register at https://yoomoney.ru
2. Create application in developer console
3. Add webhook: `https://yourdomain.com/webhook/yoomoney`
4. Add keys to `.env`:
```bash
YOOMONEY_SECRET_KEY=...
YOOMONEY_WEBHOOK_SECRET=...
```

### PayPal
1. Create account at https://developer.paypal.com
2. Create app and get client ID/secret
3. Add webhook: `https://yourdomain.com/webhook/paypal`
4. Add keys to `.env`:
```bash
PAYPAL_CLIENT_ID=...
PAYPAL_CLIENT_SECRET=...
PAYPAL_WEBHOOK_ID=...
PAYPAL_MODE=live  # or sandbox
```

### Telegram Payments
1. Contact @BotFather on Telegram
2. Use `/mybots` → Select bot → Payments
3. Choose payment provider and get token
4. Add to `.env`:
```bash
TELEGRAM_PAYMENT_PROVIDER_TOKEN=...
```

### Cryptocurrency
1. Choose processor (BTCPay, CoinGate, etc.)
2. Create account and get API credentials
3. Add webhook: `https://yourdomain.com/webhook/crypto`
4. Add keys to `.env`:
```bash
CRYPTO_PROCESSOR_API_KEY=...
CRYPTO_PROCESSOR_URL=...
CRYPTO_WALLET_ADDRESS=...
```

## 🤖 Bot Configuration

### 1. Create Bot
1. Message @BotFather on Telegram
2. Use `/newbot` command
3. Follow instructions to create bot
4. Save the bot token

### 2. Set Bot Commands
```bash
# Set bot commands via BotFather
/setcommands
```

Add these commands:
```
start - Start bot
plans - View subscription plans
subscribe - Subscribe to a plan
myplan - View current subscription
cancel - Cancel subscription
support - Get support
help - Show help
```

### 3. Set Bot Webhook
```bash
# Set webhook URL
curl -X POST "https://api.telegram.org/bot<BOT_TOKEN>/setWebhook" \
  -d "url=https://yourdomain.com/webhook/telegram" \
  -d "secret_token=your_webhook_secret"
```

### 4. Bot Menu Configuration
```bash
# Set bot menu via BotFather
/setmenu
```

Add these menu items:
```
📊 Dashboard - https://yourdomain.com/user-dashboard
💳 Subscribe - /subscribe
📋 My Plan - /myplan
🆘 Support - /support
```

## 🔧 Configuration Options

### Environment Variables
```bash
# Server Configuration
PORT=5000
HOST=0.0.0.0
DOMAIN=yourdomain.com

# Database Configuration
DATABASE_URL=postgres://user:pass@localhost:5432/dbname
PGHOST=localhost
PGPORT=5432
PGUSER=botuser
PGPASSWORD=password
PGDATABASE=telegram_bot

# Bot Configuration
BOT_TOKEN=your_bot_token
WEBHOOK_SECRET=your_webhook_secret
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123

# Payment Configuration
STRIPE_SECRET_KEY=sk_...
STRIPE_PUBLISHABLE_KEY=pk_...
STRIPE_WEBHOOK_SECRET=whsec_...
YOOMONEY_SECRET_KEY=...
PAYPAL_CLIENT_ID=...
PAYPAL_CLIENT_SECRET=...
TELEGRAM_PAYMENT_PROVIDER_TOKEN=...
CRYPTO_PROCESSOR_API_KEY=...

# Security Configuration
JWT_SECRET=your_jwt_secret
WEBHOOK_SECRET=your_webhook_secret
SESSION_SECRET=your_session_secret

# Monitoring Configuration
SENTRY_DSN=your_sentry_dsn
LOG_LEVEL=info
METRICS_ENABLED=true
```

### Subscription Plans
Edit plans in admin dashboard or directly in database:
```sql
INSERT INTO subscription_plans (name, description, price_cents, duration_days, currency, max_groups) VALUES
('Basic', 'Basic plan with essential features', 1900, 30, 'USD', 1),
('Pro', 'Professional plan with advanced features', 4900, 90, 'USD', 5),
('Premium', 'Premium plan with all features', 14900, 365, 'USD', 999);
```

## 📊 Monitoring & Maintenance

### Health Checks
```bash
# Check application health
curl https://yourdomain.com/health

# Check service status
sudo systemctl status telegram-bot

# Check logs
sudo journalctl -u telegram-bot -f
```

### Database Maintenance
```bash
# Backup database
pg_dump -h localhost -U botuser telegram_bot > backup.sql

# Restore database
psql -h localhost -U botuser telegram_bot < backup.sql

# Vacuum database
sudo -u postgres psql -d telegram_bot -c "VACUUM ANALYZE;"
```

### Log Management
```bash
# View application logs
sudo journalctl -u telegram-bot -n 100

# View nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# View PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*-main.log
```

### Updates
```bash
# Update application
cd /opt/telegram-bot
git pull
go build -o telegram-bot main.go
sudo systemctl restart telegram-bot

# Update system packages
sudo apt update && sudo apt upgrade -y
```

## 🛡️ Security Best Practices

### Server Security
```bash
# Configure fail2ban
sudo nano /etc/fail2ban/jail.local

# Add these rules:
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[nginx-http-auth]
enabled = true

[nginx-limit-req]
enabled = true
```

### Database Security
```bash
# Secure PostgreSQL
sudo nano /etc/postgresql/*/main/postgresql.conf
# Set: listen_addresses = 'localhost'

sudo nano /etc/postgresql/*/main/pg_hba.conf
# Ensure only local connections are allowed
```

### Application Security
- Use strong passwords
- Enable HTTPS only
- Implement rate limiting
- Regular security updates
- Monitor for unusual activity

## 🔍 Troubleshooting

### Common Issues

#### Bot Not Responding
```bash
# Check bot token
curl "https://api.telegram.org/bot<BOT_TOKEN>/getMe"

# Check webhook
curl "https://api.telegram.org/bot<BOT_TOKEN>/getWebhookInfo"
```

#### Database Connection Issues
```bash
# Test database connection
psql -h localhost -U botuser -d telegram_bot

# Check PostgreSQL status
sudo systemctl status postgresql
```

#### Payment Issues
```bash
# Check webhook logs
sudo journalctl -u telegram-bot | grep webhook

# Test payment endpoints
curl -X POST "https://yourdomain.com/webhook/stripe" \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

#### SSL Certificate Issues
```bash
# Renew certificate
sudo certbot renew

# Test SSL
openssl s_client -connect yourdomain.com:443
```

### Performance Optimization
```bash
# Optimize PostgreSQL
sudo nano /etc/postgresql/*/main/postgresql.conf

# Add these settings:
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
```

## 📞 Support

For issues and questions:
1. Check logs first
2. Review configuration
3. Test with minimal setup
4. Contact support with detailed error messages

## 🔄 Backup Strategy

### Automated Backups
```bash
# Create backup script
sudo nano /opt/telegram-bot/backup.sh
```

```bash
#!/bin/bash
BACKUP_DIR="/opt/backups/telegram-bot"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Database backup
pg_dump -h localhost -U botuser telegram_bot > $BACKUP_DIR/database_$DATE.sql

# Application backup
tar -czf $BACKUP_DIR/application_$DATE.tar.gz -C /opt telegram-bot

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
```

```bash
# Make executable and add to cron
chmod +x /opt/telegram-bot/backup.sh
sudo crontab -e
# Add: 0 2 * * * /opt/telegram-bot/backup.sh
```

This completes the comprehensive setup guide. Follow these steps carefully for a successful deployment.