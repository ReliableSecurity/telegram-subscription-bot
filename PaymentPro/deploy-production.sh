#!/bin/bash

# Telegram Bot Production Deployment Script
# Usage: ./deploy-production.sh
# Requires: TELEGRAM_BOT_TOKEN and DOMAIN_NAME environment variables

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Check required environment variables
if [[ -z "$TELEGRAM_BOT_TOKEN" ]]; then
    error "TELEGRAM_BOT_TOKEN environment variable is required"
fi

if [[ -z "$DOMAIN_NAME" ]]; then
    error "DOMAIN_NAME environment variable is required"
fi

log "Starting Telegram Bot Production Deployment"
log "Domain: $DOMAIN_NAME"
log "Bot Token: ${TELEGRAM_BOT_TOKEN:0:10}..."

# Detect OS
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
else
    error "Cannot detect OS. This script supports Ubuntu 20.04+ and Debian 11+"
fi

log "Detected OS: $OS $VER"

# Verify supported OS
case $OS in
    "Ubuntu")
        if [[ $(echo "$VER >= 20.04" | bc -l) -ne 1 ]]; then
            error "Ubuntu 20.04 or higher required"
        fi
        ;;
    "Debian GNU/Linux")
        if [[ $(echo "$VER >= 11" | bc -l) -ne 1 ]]; then
            error "Debian 11 or higher required"
        fi
        ;;
    *)
        error "Unsupported OS. This script supports Ubuntu 20.04+ and Debian 11+"
        ;;
esac

# Generate secure passwords if not provided
DB_PASSWORD=${DB_PASSWORD:-$(openssl rand -base64 32)}
JWT_SECRET=${JWT_SECRET:-$(openssl rand -hex 32)}
ENCRYPT_KEY=${ENCRYPT_KEY:-$(openssl rand -hex 32)}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-$(openssl rand -base64 16)}

log "Generated secure passwords and keys"

# Update system
log "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install required packages
log "Installing required packages..."
sudo apt install -y \
    curl \
    wget \
    git \
    nginx \
    postgresql \
    postgresql-contrib \
    certbot \
    python3-certbot-nginx \
    ufw \
    htop \
    unzip \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release \
    bc

# Install Go
log "Installing Go..."
GO_VERSION="1.21.5"
cd /tmp
wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"

# Add Go to PATH
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi
export PATH=$PATH:/usr/local/go/bin

# Verify Go installation
go version || error "Go installation failed"
log "Go installed successfully"

# Setup PostgreSQL
log "Configuring PostgreSQL..."
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE telegram_bot;
CREATE USER telegram_user WITH PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE telegram_bot TO telegram_user;
ALTER USER telegram_user CREATEDB;
\q
EOF

log "PostgreSQL configured successfully"

# Clone repository
log "Cloning repository..."
cd ~
if [[ -d "telegram-subscription-bot" ]]; then
    rm -rf telegram-subscription-bot
fi

git clone https://github.com/yourusername/telegram-subscription-bot.git
cd telegram-subscription-bot

# Build application
log "Building application..."
go mod tidy
go build -o telegram-bot main.go

# Create .env file
log "Creating configuration file..."
cat > .env << EOF
# Bot Configuration
TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN
BOT_USERNAME=@$(echo $TELEGRAM_BOT_TOKEN | cut -d: -f1)

# Database Configuration
DATABASE_URL=postgresql://telegram_user:$DB_PASSWORD@localhost/telegram_bot?sslmode=disable
PGHOST=localhost
PGPORT=5432
PGUSER=telegram_user
PGPASSWORD=$DB_PASSWORD
PGDATABASE=telegram_bot

# Web Server Configuration
WEB_PORT=5000
WEB_HOST=0.0.0.0
DOMAIN_NAME=$DOMAIN_NAME

# Security Configuration
ADMIN_USERNAME=admin
ADMIN_PASSWORD=$ADMIN_PASSWORD
JWT_SECRET=$JWT_SECRET
ENCRYPT_KEY=$ENCRYPT_KEY

# Application Settings
ENVIRONMENT=production
DEBUG=false
LOG_LEVEL=info

# Features
FEATURE_AI_RECOMMENDATIONS=true
FEATURE_MULTI_LANGUAGE=true
FEATURE_USER_DASHBOARD=true
FEATURE_ADVANCED_ANALYTICS=true

# Backup Configuration
BACKUP_ENABLED=true
BACKUP_SCHEDULE=0 2 * * *
BACKUP_RETENTION_DAYS=7
BACKUP_PATH=$HOME/backups

# Monitoring
MONITORING_ENABLED=true
HEALTH_CHECK_INTERVAL=30
EOF

# Set proper permissions for .env file
chmod 600 .env

# Create database schema
log "Initializing database schema..."
if [[ -f "database/schema.sql" ]]; then
    PGPASSWORD=$DB_PASSWORD psql -h localhost -U telegram_user -d telegram_bot -f database/schema.sql
else
    # Create basic schema if file doesn't exist
    PGPASSWORD=$DB_PASSWORD psql -h localhost -U telegram_user -d telegram_bot << 'EOF'
-- Create basic tables if they don't exist
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscription_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price_cents INTEGER NOT NULL,
    duration_days INTEGER NOT NULL,
    max_groups INTEGER DEFAULT 1,
    features TEXT,
    currency VARCHAR(3) DEFAULT 'USD',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default subscription plans
INSERT INTO subscription_plans (name, price_cents, duration_days, max_groups, features) VALUES
('Basic', 19900, 30, 3, 'Basic moderation, Up to 3 groups'),
('Pro', 49900, 90, 10, 'Advanced moderation, Up to 10 groups, Analytics'),
('Premium', 149900, 365, 50, 'All features, Up to 50 groups, Priority support')
ON CONFLICT DO NOTHING;
EOF
fi

log "Database schema initialized"

# Create systemd service
log "Creating systemd service..."
sudo tee /etc/systemd/system/telegram-bot.service > /dev/null << EOF
[Unit]
Description=Telegram Subscription Bot
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$HOME/telegram-subscription-bot
ExecStart=$HOME/telegram-subscription-bot/telegram-bot
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10
EnvironmentFile=$HOME/telegram-subscription-bot/.env
StandardOutput=journal
StandardError=journal
SyslogIdentifier=telegram-bot

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$HOME/telegram-subscription-bot $HOME/backups /tmp
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes

[Install]
WantedBy=multi-user.target
EOF

# Create backup directory
mkdir -p ~/backups

# Create management scripts
log "Creating management scripts..."

# Start script
cat > ~/start-bot.sh << 'EOF'
#!/bin/bash
sudo systemctl start telegram-bot
sudo systemctl status telegram-bot
EOF

# Stop script
cat > ~/stop-bot.sh << 'EOF'
#!/bin/bash
sudo systemctl stop telegram-bot
sudo systemctl status telegram-bot
EOF

# Restart script
cat > ~/restart-bot.sh << 'EOF'
#!/bin/bash
sudo systemctl restart telegram-bot
sudo systemctl status telegram-bot
EOF

# Status script
cat > ~/status-bot.sh << 'EOF'
#!/bin/bash
echo "=== Telegram Bot Status ==="
sudo systemctl status telegram-bot --no-pager
echo ""
echo "=== PostgreSQL Status ==="
sudo systemctl status postgresql --no-pager
echo ""
echo "=== Nginx Status ==="
sudo systemctl status nginx --no-pager
echo ""
echo "=== Disk Usage ==="
df -h /
echo ""
echo "=== Memory Usage ==="
free -h
echo ""
echo "=== Last 10 Bot Logs ==="
sudo journalctl -u telegram-bot -n 10 --no-pager
EOF

# Update script
cat > ~/update-bot.sh << 'EOF'
#!/bin/bash
set -e
cd ~/telegram-subscription-bot
echo "Stopping bot..."
sudo systemctl stop telegram-bot
echo "Pulling latest code..."
git pull
echo "Building application..."
go build -o telegram-bot main.go
echo "Starting bot..."
sudo systemctl start telegram-bot
echo "Update completed!"
sudo systemctl status telegram-bot --no-pager
EOF

# Backup script
cat > ~/backup-bot.sh << 'EOF'
#!/bin/bash
set -e
source ~/telegram-subscription-bot/.env
BACKUP_DIR="$HOME/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="telegram_bot_backup_$DATE.sql"

echo "Creating database backup..."
PGPASSWORD=$PGPASSWORD pg_dump -h localhost -U telegram_user -d telegram_bot > "$BACKUP_DIR/$BACKUP_FILE"

echo "Compressing backup..."
gzip "$BACKUP_DIR/$BACKUP_FILE"

echo "Cleaning old backups (keeping last 7 days)..."
find "$BACKUP_DIR" -name "telegram_bot_backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
ls -lah "$BACKUP_DIR/"
EOF

# Make scripts executable
# Create health check script
cat > ~/health-check.sh << 'EOF'
#!/bin/bash
cd ~/telegram-subscription-bot
./scripts/health-check.sh
EOF

chmod +x ~/start-bot.sh ~/stop-bot.sh ~/restart-bot.sh ~/status-bot.sh ~/update-bot.sh ~/backup-bot.sh ~/health-check.sh

# Configure firewall
log "Configuring firewall..."
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw --force enable

# Configure Nginx
log "Configuring Nginx..."
sudo tee /etc/nginx/sites-available/$DOMAIN_NAME > /dev/null << EOF
server {
    listen 80;
    server_name $DOMAIN_NAME www.$DOMAIN_NAME;
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name $DOMAIN_NAME www.$DOMAIN_NAME;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Main proxy
    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Static files
    location /static/ {
        alias $HOME/telegram-subscription-bot/web/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
    
    # Health check
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

# Enable site
sudo ln -sf /etc/nginx/sites-available/$DOMAIN_NAME /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t || error "Nginx configuration test failed"

# Start and enable services
log "Starting services..."
sudo systemctl daemon-reload
sudo systemctl enable telegram-bot
sudo systemctl enable nginx
sudo systemctl restart nginx

# Start the bot
sudo systemctl start telegram-bot

# Wait for bot to start
sleep 5

# Check if bot is running
if sudo systemctl is-active --quiet telegram-bot; then
    log "Bot service started successfully"
else
    error "Bot service failed to start. Check logs: sudo journalctl -u telegram-bot -f"
fi

# Setup SSL certificate
log "Setting up SSL certificate..."
if sudo certbot --nginx -d $DOMAIN_NAME -d www.$DOMAIN_NAME --non-interactive --agree-tos --email admin@$DOMAIN_NAME --redirect; then
    log "SSL certificate obtained successfully"
    
    # Setup auto-renewal
    (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | crontab -
    log "SSL auto-renewal configured"
else
    warn "SSL certificate setup failed. You can run it manually later: sudo certbot --nginx -d $DOMAIN_NAME"
fi

# Setup automated backups
log "Setting up automated backups..."
(crontab -l 2>/dev/null; echo "0 2 * * * $HOME/backup-bot.sh >> $HOME/backups/backup.log 2>&1") | crontab -

# Final verification
log "Running final verification..."
sleep 10

# Test local connectivity
if curl -sf http://localhost:5000/health > /dev/null; then
    log "Local health check passed"
else
    warn "Local health check failed"
fi

# Test external connectivity
if curl -sf https://$DOMAIN_NAME/health > /dev/null; then
    log "External health check passed"
else
    warn "External health check failed - may need DNS propagation time"
fi

# Display final information
echo ""
echo "=========================================="
echo -e "${GREEN}Deployment completed successfully!${NC}"
echo "=========================================="
echo ""
echo "🌐 Website URL: https://$DOMAIN_NAME"
echo "🔐 Admin Login: admin / $ADMIN_PASSWORD"
echo "🤖 Bot Token: ${TELEGRAM_BOT_TOKEN:0:10}..."
echo "💾 Database: telegram_bot (PostgreSQL)"
echo ""
echo "📋 Management Commands:"
echo "  Start bot:    ~/start-bot.sh"
echo "  Stop bot:     ~/stop-bot.sh"
echo "  Restart bot:  ~/restart-bot.sh"
echo "  Check status: ~/status-bot.sh"
echo "  Update bot:   ~/update-bot.sh"
echo "  Backup data:  ~/backup-bot.sh"
echo ""
echo "📊 Monitoring:"
echo "  Bot logs:     sudo journalctl -u telegram-bot -f"
echo "  Nginx logs:   sudo tail -f /var/log/nginx/access.log"
echo "  System:       htop"
echo ""
echo "🔒 Security:"
echo "  Firewall:     sudo ufw status"
echo "  SSL:          sudo certbot certificates"
echo ""
echo "⚠️  Important:"
echo "  1. Save your admin password: $ADMIN_PASSWORD"
echo "  2. Configure payment providers in web dashboard"
echo "  3. Test bot functionality: /start command in Telegram"
echo "  4. Setup monitoring and alerts"
echo "  5. Review and update .env file if needed"
echo ""
echo "📧 Support: Create issue on GitHub if you need help"
echo "=========================================="

# Save credentials to file
cat > ~/telegram-bot-credentials.txt << EOF
Telegram Bot Deployment Credentials
Generated: $(date)

Website: https://$DOMAIN_NAME
Admin Username: admin
Admin Password: $ADMIN_PASSWORD
Database Password: $DB_PASSWORD
JWT Secret: $JWT_SECRET

Important: Keep this file secure and delete after saving credentials elsewhere!
EOF

log "Credentials saved to ~/telegram-bot-credentials.txt"
log "Deployment completed successfully!"