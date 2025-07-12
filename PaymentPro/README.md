# Telegram Subscription Bot

Comprehensive Telegram bot platform for advanced group management with subscription handling, payment processing, AI-powered moderation, and web dashboard administration.

## 🚀 Features

### Core Functionality
- **Multi-Group Management**: Manage multiple Telegram groups/channels from single account
- **Subscription System**: Flexible subscription plans with individual billing per group
- **Payment Processing**: Support for Stripe, YooMoney, PayPal, and cryptocurrency payments
- **AI Recommendations**: Intelligent bot behavior optimization and moderation suggestions
- **Web Dashboard**: Professional admin interface with analytics and user management
- **Advanced Moderation**: Automated spam detection, user management, and violation tracking

### Technical Features
- **Dual Authentication**: Separate admin and user access systems
- **Real-time Analytics**: Revenue tracking, user growth, and engagement metrics
- **Database Integration**: PostgreSQL with automated schema management
- **Responsive Design**: Mobile-friendly web interface
- **SSL/TLS Security**: Automated certificate management
- **Backup System**: Automated daily backups with retention policy

## 📋 Prerequisites

### For Local Development
- Go 1.21+
- PostgreSQL 12+
- Git

### For VPS Production Deployment
- Ubuntu 20.04+ or Debian 11+ VPS
- Domain name pointed to your VPS
- Telegram Bot Token from [@BotFather](https://t.me/BotFather)
- SSL certificate (automated via Let's Encrypt)

## 🛠 Quick Start (Local Development)

### 1. Clone Repository
```bash
git clone https://github.com/yourusername/telegram-subscription-bot.git
cd telegram-subscription-bot
```

### 2. Install Dependencies
```bash
# Install Go dependencies
go mod tidy

# Install PostgreSQL (Ubuntu/Debian)
sudo apt update
sudo apt install postgresql postgresql-contrib
```

### 3. Setup Database
```bash
# Create database and user
sudo -u postgres createdb telegram_bot
sudo -u postgres createuser -P telegram_user

# Grant privileges
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE telegram_bot TO telegram_user;"

# Create schema
sudo -u postgres psql telegram_bot < database/schema.sql
```

### 4. Configure Environment
```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

Required environment variables:
```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_URL=postgresql://telegram_user:password@localhost/telegram_bot
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
WEB_PORT=5000
```

### 5. Build and Run
```bash
# Build application
go build -o telegram-bot main.go

# Run bot
./telegram-bot
```

Access dashboard at: http://localhost:5000

## 🌐 Production Deployment (VPS + Domain)

### Prerequisites
- VPS with Ubuntu 20.04+
- Domain name (example.com) pointed to VPS IP
- Telegram Bot Token

### One-Command Deployment

```bash
# Set required variables
export TELEGRAM_BOT_TOKEN="your_bot_token"
export DOMAIN_NAME="yourdomain.com"

# Download and run deployment script
curl -sSL https://raw.githubusercontent.com/yourusername/telegram-subscription-bot/main/deploy-production.sh | bash
```

### Manual Deployment Steps

#### 1. Prepare VPS
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y curl wget git nginx postgresql postgresql-contrib certbot python3-certbot-nginx ufw
```

#### 2. Install Go
```bash
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Clone and Build
```bash
git clone https://github.com/yourusername/telegram-subscription-bot.git
cd telegram-subscription-bot
go mod tidy
go build -o telegram-bot main.go
```

#### 4. Setup Database
```bash
sudo -u postgres createdb telegram_bot
sudo -u postgres createuser -P telegram_user
sudo -u postgres psql telegram_bot < database/schema.sql
```

#### 5. Configure Environment
```bash
# Create production .env file
cat > .env << EOF
TELEGRAM_BOT_TOKEN=your_bot_token
DATABASE_URL=postgresql://telegram_user:password@localhost/telegram_bot
WEB_PORT=5000
WEB_HOST=0.0.0.0
DOMAIN_NAME=yourdomain.com
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
ENVIRONMENT=production
EOF
```

#### 6. Setup System Service
```bash
# Create systemd service
sudo tee /etc/systemd/system/telegram-bot.service > /dev/null << EOF
[Unit]
Description=Telegram Subscription Bot
After=network.target postgresql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME/telegram-subscription-bot
ExecStart=$HOME/telegram-subscription-bot/telegram-bot
Restart=always
RestartSec=10
EnvironmentFile=$HOME/telegram-subscription-bot/.env

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

#### 7. Configure Nginx
```bash
# Create nginx configuration
sudo tee /etc/nginx/sites-available/yourdomain.com > /dev/null << EOF
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;
    
    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location /static/ {
        alias $HOME/telegram-subscription-bot/web/static/;
        expires 30d;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/yourdomain.com /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl restart nginx
```

#### 8. Setup SSL Certificate
```bash
# Configure firewall
sudo ufw enable
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'

# Get SSL certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Setup auto-renewal
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

## 🎮 Bot Commands

### User Commands
- `/start` - Start bot and get welcome message
- `/plans` - View available subscription plans
- `/subscribe` - Subscribe to a plan
- `/myplan` - Check current subscription status
- `/cancel` - Cancel active subscription
- `/help` - Get help information

### Admin Commands (Web Dashboard)
- Dashboard: `https://yourdomain.com/dashboard`
- Login: admin / admin123 (change after first login)
- AI Recommendations: `https://yourdomain.com/ai-recommendations`
- Analytics: `https://yourdomain.com/analytics`
- User Management: `https://yourdomain.com/users`

## 🔧 Configuration

### Environment Variables
```env
# Bot Settings
TELEGRAM_BOT_TOKEN=             # Bot token from @BotFather
BOT_USERNAME=                   # Bot username (optional)

# Database
DATABASE_URL=                   # PostgreSQL connection string
PGHOST=localhost
PGPORT=5432
PGUSER=telegram_user
PGPASSWORD=your_password
PGDATABASE=telegram_bot

# Web Server
WEB_PORT=5000
WEB_HOST=0.0.0.0
DOMAIN_NAME=yourdomain.com

# Authentication
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
JWT_SECRET=                     # Generate with: openssl rand -hex 32

# Payment Providers (optional)
STRIPE_SECRET_KEY=              # Stripe secret key
YOOMONEY_SECRET_KEY=            # YooMoney secret key
PAYPAL_SECRET_KEY=              # PayPal secret key

# Security
ENCRYPT_KEY=                    # Generate with: openssl rand -hex 32
ENVIRONMENT=production
DEBUG=false
```

### Payment Provider Setup

#### Stripe
1. Create account at [stripe.com](https://stripe.com)
2. Get API keys from Dashboard → Developers → API keys
3. Add secret key to `.env`

#### YooMoney
1. Register at [yoomoney.ru](https://yoomoney.ru)
2. Create application and get secret key
3. Configure webhooks for payment notifications

#### PayPal
1. Create developer account at [developer.paypal.com](https://developer.paypal.com)
2. Create application and get credentials
3. Configure webhooks for payment events

## 📊 Management Commands

After deployment, use these commands to manage your bot:

```bash
# Service management
~/start-bot.sh          # Start bot service
~/stop-bot.sh           # Stop bot service
~/restart-bot.sh        # Restart bot service
~/status-bot.sh         # Check all services status

# Maintenance
~/update-bot.sh         # Update bot from git
~/backup-bot.sh         # Create backup

# Logs
sudo journalctl -u telegram-bot -f    # Follow live logs
sudo journalctl -u telegram-bot -n 100 # Last 100 log entries
```

## 🔒 Security

### Best Practices
- Change default admin password immediately
- Use strong passwords and JWT secrets
- Enable firewall (UFW configured automatically)
- Keep system updated
- Monitor logs regularly
- Setup backup strategy

### Security Features
- HTTPS with automatic SSL renewal
- Secure headers (CSP, HSTS, X-Frame-Options)
- Input validation and sanitization
- Rate limiting on API endpoints
- Encrypted sensitive data storage

## 📈 Monitoring

### Built-in Monitoring
- System health dashboard
- Real-time performance metrics
- Error logging and tracking
- Automated backup system
- SSL certificate monitoring

### External Monitoring (Optional)
Consider setting up:
- Uptime monitoring (UptimeRobot, Pingdom)
- Log aggregation (ELK stack, Grafana)
- Performance monitoring (New Relic, DataDog)
- Alerting system (PagerDuty, Slack)

## 🛠 Troubleshooting

### Common Issues

#### Bot Not Starting
```bash
# Check service status
sudo systemctl status telegram-bot

# Check logs
sudo journalctl -u telegram-bot -n 50

# Check configuration
cat .env | grep -v PASSWORD
```

#### Database Connection Issues
```bash
# Test database connection
psql $DATABASE_URL -c "SELECT 1;"

# Check PostgreSQL status
sudo systemctl status postgresql

# Reset database password
sudo -u postgres psql -c "ALTER USER telegram_user PASSWORD 'new_password';"
```

#### Web Dashboard Not Accessible
```bash
# Check nginx status
sudo systemctl status nginx

# Test nginx configuration
sudo nginx -t

# Check firewall
sudo ufw status

# Test local connectivity
curl -I http://localhost:5000
```

#### SSL Certificate Issues
```bash
# Check certificate status
sudo certbot certificates

# Renew certificate manually
sudo certbot renew

# Test SSL configuration
curl -I https://yourdomain.com
```

### Log Locations
- Bot logs: `sudo journalctl -u telegram-bot`
- Nginx logs: `/var/log/nginx/`
- PostgreSQL logs: `/var/log/postgresql/`
- System logs: `/var/log/syslog`

## 🔄 Updates and Maintenance

### Regular Updates
```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Update bot code
cd ~/telegram-subscription-bot
git pull
go build -o telegram-bot main.go
sudo systemctl restart telegram-bot

# Or use the update script
~/update-bot.sh
```

### Backup Strategy
- Automated daily database backups
- Configuration file backups
- 7-day retention policy
- Manual backup: `~/backup-bot.sh`

### Database Maintenance
```bash
# Vacuum database (weekly)
psql $DATABASE_URL -c "VACUUM ANALYZE;"

# Check database size
psql $DATABASE_URL -c "SELECT pg_size_pretty(pg_database_size('telegram_bot'));"
```

## 📞 Support

### Getting Help
1. Check this README for common solutions
2. Review logs for error messages
3. Check GitHub Issues for known problems
4. Create new issue with detailed information

### Reporting Issues
Include:
- Operating system and version
- Go version
- Error messages from logs
- Steps to reproduce
- Expected vs actual behavior

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 Changelog

### Version 1.0.0 (Current)
- Initial release
- Multi-group management
- Payment processing
- AI recommendations
- Web dashboard
- Production deployment automation

---

**Made with ❤️ for the Telegram community**