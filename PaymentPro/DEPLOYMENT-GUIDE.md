# Telegram Bot Production Deployment Guide

This guide provides step-by-step instructions for deploying the Telegram Subscription Bot to a VPS with domain binding and SSL configuration.

## 🎯 Quick Start (Recommended)

### Prerequisites
- VPS with Ubuntu 20.04+ or Debian 11+
- Domain name pointed to your VPS IP address
- Telegram Bot Token from [@BotFather](https://t.me/BotFather)

### One-Command Deployment

```bash
# Set your configuration
export TELEGRAM_BOT_TOKEN="your_bot_token_from_botfather"
export DOMAIN_NAME="yourdomain.com"

# Download and run deployment script
curl -sSL https://raw.githubusercontent.com/yourusername/telegram-subscription-bot/main/deploy-production.sh | bash
```

That's it! The script will:
- ✅ Install all dependencies (Go, PostgreSQL, Nginx)
- ✅ Configure database with secure passwords
- ✅ Build and deploy the application
- ✅ Setup SSL certificates with Let's Encrypt
- ✅ Configure firewall and security
- ✅ Create management scripts
- ✅ Setup automated backups

## 📋 Detailed Manual Deployment

If you prefer manual deployment or need customization:

### Step 1: Prepare Your VPS

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install basic tools
sudo apt install -y curl wget git nginx postgresql postgresql-contrib \
                    certbot python3-certbot-nginx ufw htop
```

### Step 2: Get Bot Token

1. Start a chat with [@BotFather](https://t.me/BotFather)
2. Send `/newbot` command
3. Follow instructions to create your bot
4. Save the bot token (format: `123456789:ABCdef...`)
5. Send `/setcommands` and add these commands:
   ```
   start - Start the bot
   plans - View subscription plans
   subscribe - Subscribe to a plan
   myplan - Check current subscription
   cancel - Cancel subscription
   help - Get help
   ```

### Step 3: Domain Configuration

Point your domain to your VPS:
```
A record: yourdomain.com → your.vps.ip.address
A record: www.yourdomain.com → your.vps.ip.address
```

Wait for DNS propagation (usually 5-30 minutes).

### Step 4: Clone and Setup

```bash
# Clone repository
git clone https://github.com/yourusername/telegram-subscription-bot.git
cd telegram-subscription-bot

# Install Go
curl -sSL https://golang.org/dl/go1.21.5.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Build application
go mod tidy
go build -o telegram-bot main.go
```

### Step 5: Database Setup

```bash
# Start PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres createdb telegram_bot
sudo -u postgres createuser -P telegram_user  # Enter a secure password
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE telegram_bot TO telegram_user;"

# Initialize schema
PGPASSWORD=your_password psql -h localhost -U telegram_user -d telegram_bot -f database/schema.sql
```

### Step 6: Configuration

```bash
# Create .env file
cp .env.example .env
nano .env
```

Update the following variables:
```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DOMAIN_NAME=yourdomain.com
DATABASE_URL=postgresql://telegram_user:your_password@localhost/telegram_bot
ADMIN_PASSWORD=secure_admin_password
JWT_SECRET=generate_with_openssl_rand_hex_32
```

### Step 7: System Service

```bash
# Create systemd service
sudo tee /etc/systemd/system/telegram-bot.service > /dev/null << EOF
[Unit]
Description=Telegram Subscription Bot
After=network.target postgresql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$PWD
ExecStart=$PWD/telegram-bot
Restart=always
EnvironmentFile=$PWD/.env

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot
```

### Step 8: Nginx Configuration

```bash
# Create Nginx config
sudo tee /etc/nginx/sites-available/yourdomain.com > /dev/null << EOF
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;
    
    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location /static/ {
        alias $PWD/web/static/;
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

### Step 9: SSL Certificate

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

## 🔧 Configuration Options

### Payment Providers

Add to your `.env` file:

```env
# Stripe
STRIPE_SECRET_KEY=sk_live_your_stripe_key
STRIPE_PUBLIC_KEY=pk_live_your_public_key

# YooMoney (Russian market)
YOOMONEY_SECRET_KEY=your_yoomoney_key

# PayPal
PAYPAL_CLIENT_ID=your_paypal_client_id
PAYPAL_SECRET_KEY=your_paypal_secret
```

### Email Notifications

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

### Monitoring

```env
MONITORING_ENABLED=true
HEALTH_CHECK_INTERVAL=30
BACKUP_ENABLED=true
```

## 🎮 Testing Your Deployment

### 1. Test Bot Connection
```bash
# Check if bot responds
curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"
```

### 2. Test Website
```bash
# Test local connection
curl -I http://localhost:5000

# Test external connection
curl -I https://yourdomain.com
```

### 3. Test Database
```bash
# Connect to database
PGPASSWORD=your_password psql -h localhost -U telegram_user -d telegram_bot -c "SELECT COUNT(*) FROM users;"
```

### 4. Test Bot Commands
1. Find your bot on Telegram: `@your_bot_username`
2. Send `/start` command
3. Send `/plans` command
4. Test web dashboard: `https://yourdomain.com`

## 📊 Management

### Service Commands
```bash
# Check status
sudo systemctl status telegram-bot
sudo systemctl status nginx
sudo systemctl status postgresql

# View logs
sudo journalctl -u telegram-bot -f
sudo tail -f /var/log/nginx/access.log

# Restart services
sudo systemctl restart telegram-bot
sudo systemctl restart nginx
```

### Database Management
```bash
# Create backup
pg_dump -h localhost -U telegram_user telegram_bot > backup.sql

# Restore backup
psql -h localhost -U telegram_user -d telegram_bot < backup.sql

# Monitor database size
psql -h localhost -U telegram_user -d telegram_bot -c "SELECT pg_size_pretty(pg_database_size('telegram_bot'));"
```

### Updates
```bash
# Update application
cd ~/telegram-subscription-bot
git pull
go build -o telegram-bot main.go
sudo systemctl restart telegram-bot
```

## 🔒 Security Checklist

- ✅ Use strong passwords for database and admin account
- ✅ Enable firewall (UFW) with minimal required ports
- ✅ SSL certificate with auto-renewal
- ✅ Regular security updates
- ✅ Backup strategy implemented
- ✅ Monitor logs for suspicious activity
- ✅ Use environment variables for secrets
- ✅ Limit file permissions (600 for .env)

## 🆘 Troubleshooting

### Bot Not Starting
```bash
# Check logs
sudo journalctl -u telegram-bot -n 50

# Check configuration
cat .env | grep -v PASSWORD

# Test bot token
curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"
```

### Website Not Accessible
```bash
# Check nginx
sudo nginx -t
sudo systemctl status nginx

# Check firewall
sudo ufw status

# Check SSL
sudo certbot certificates
```

### Database Issues
```bash
# Check PostgreSQL
sudo systemctl status postgresql

# Test connection
PGPASSWORD=your_password psql -h localhost -U telegram_user -d telegram_bot -c "SELECT 1;"
```

### SSL Certificate Issues
```bash
# Check certificates
sudo certbot certificates

# Renew manually
sudo certbot renew --dry-run

# Check domain pointing
nslookup yourdomain.com
```

## 📈 Monitoring and Maintenance

### Daily Tasks
- Check bot status: `sudo systemctl status telegram-bot`
- Review error logs: `sudo journalctl -u telegram-bot --since yesterday`
- Monitor disk space: `df -h`

### Weekly Tasks
- Update system packages: `sudo apt update && sudo apt upgrade`
- Clean old logs: `sudo journalctl --vacuum-time=30d`
- Test SSL certificate: `curl -I https://yourdomain.com`

### Monthly Tasks
- Review and update bot code
- Check database performance
- Update dependencies
- Test backup/restore process

## 🎯 Performance Optimization

### Database Optimization
```sql
-- Run monthly
VACUUM ANALYZE;

-- Check database size
SELECT pg_size_pretty(pg_database_size('telegram_bot'));

-- Check table sizes
SELECT schemaname,tablename,pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables WHERE schemaname='public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Nginx Optimization
```nginx
# Add to nginx.conf for better performance
worker_processes auto;
worker_connections 1024;
keepalive_timeout 65;
gzip on;
gzip_comp_level 6;
```

### System Resources
```bash
# Monitor resource usage
htop
free -h
df -h
iotop
```

## 📞 Support

### Getting Help
1. Check this guide first
2. Review application logs
3. Check GitHub Issues
4. Create new issue with:
   - OS and version
   - Error messages
   - Steps to reproduce
   - Configuration (without secrets)

### Useful Commands
```bash
# Full system status
~/status-bot.sh

# Create backup
~/backup-bot.sh

# Update bot
~/update-bot.sh

# View live logs
sudo journalctl -u telegram-bot -f
```

---

**Congratulations!** 🎉 Your Telegram Subscription Bot is now deployed and ready to serve users with professional-grade security, monitoring, and scalability.