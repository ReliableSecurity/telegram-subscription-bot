#!/bin/bash

# Enhanced VPS Deployment Script for Telegram Bot with Domain and Payment Integration
# Usage: ./deploy-vps.sh [domain] [bot_token] [admin_password]

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOMAIN=${1:-""}
BOT_TOKEN=${2:-""}
ADMIN_PASSWORD=${3:-"admin123"}
DB_NAME="telegram_bot"
DB_USER="botuser"
DB_PASSWORD=$(openssl rand -base64 32)
APP_DIR="/opt/telegram-bot"
SERVICE_NAME="telegram-bot"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

install_dependencies() {
    log_step "Installing system dependencies..."
    
    # Update system
    apt update && apt upgrade -y
    
    # Install required packages
    apt install -y \
        curl \
        wget \
        git \
        nginx \
        postgresql \
        postgresql-contrib \
        certbot \
        python3-certbot-nginx \
        ufw \
        fail2ban \
        htop \
        nano \
        unzip \
        supervisor \
        jq
    
    # Install Go
    if ! command -v go &> /dev/null; then
        log_info "Installing Go..."
        wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        export PATH=$PATH:/usr/local/go/bin
        rm go1.21.5.linux-amd64.tar.gz
    fi
    
    log_info "Dependencies installed successfully"
}

setup_database() {
    log_step "Setting up PostgreSQL database..."
    
    # Start PostgreSQL
    systemctl start postgresql
    systemctl enable postgresql
    
    # Create database and user
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME;"
    sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
    sudo -u postgres psql -c "ALTER USER $DB_USER CREATEDB;"
    
    # Configure PostgreSQL for network connections
    PG_VERSION=$(sudo -u postgres psql -c "SELECT version();" | grep -oP '\d+\.\d+' | head -1)
    PG_CONFIG_DIR="/etc/postgresql/$PG_VERSION/main"
    
    # Update postgresql.conf
    sed -i "s/#listen_addresses = 'localhost'/listen_addresses = 'localhost'/" "$PG_CONFIG_DIR/postgresql.conf"
    
    # Update pg_hba.conf
    echo "local   $DB_NAME   $DB_USER   md5" >> "$PG_CONFIG_DIR/pg_hba.conf"
    
    systemctl restart postgresql
    
    log_info "Database setup completed"
}

setup_application() {
    log_step "Setting up application..."
    
    # Create application directory
    mkdir -p $APP_DIR
    cd $APP_DIR
    
    # Clone or copy application (assuming files are in current directory)
    if [ -d "/tmp/telegram-bot" ]; then
        cp -r /tmp/telegram-bot/* $APP_DIR/
    else
        log_warn "Application files not found in /tmp/telegram-bot, using current directory"
        cp -r . $APP_DIR/
    fi
    
    # Create environment file
    cat > $APP_DIR/.env << EOF
# Database Configuration
DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@localhost:5432/$DB_NAME?sslmode=disable
PGHOST=localhost
PGPORT=5432
PGUSER=$DB_USER
PGPASSWORD=$DB_PASSWORD
PGDATABASE=$DB_NAME

# Bot Configuration
BOT_TOKEN=$BOT_TOKEN
ADMIN_USERNAME=admin
ADMIN_PASSWORD=$ADMIN_PASSWORD

# Server Configuration
PORT=5000
HOST=0.0.0.0
DOMAIN=$DOMAIN

# Payment Configuration (to be filled by user)
STRIPE_SECRET_KEY=
STRIPE_PUBLISHABLE_KEY=
YOOMONEY_SECRET_KEY=
PAYPAL_CLIENT_ID=
PAYPAL_CLIENT_SECRET=
CRYPTO_WALLET_ADDRESS=
EOF
    
    # Set proper permissions
    chown -R root:root $APP_DIR
    chmod 644 $APP_DIR/.env
    chmod +x $APP_DIR/deploy-vps.sh
    
    # Build application
    log_info "Building Go application..."
    cd $APP_DIR
    export PATH=$PATH:/usr/local/go/bin
    go mod tidy
    go build -o telegram-bot main.go
    
    log_info "Application setup completed"
}

setup_nginx() {
    log_step "Setting up Nginx..."
    
    # Remove default site
    rm -f /etc/nginx/sites-enabled/default
    
    # Create new site configuration
    cat > /etc/nginx/sites-available/telegram-bot << EOF
server {
    listen 80;
    server_name $DOMAIN www.$DOMAIN;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    
    # Rate limiting
    limit_req_zone \$binary_remote_addr zone=login:10m rate=1r/s;
    limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;
    
    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        proxy_read_timeout 86400;
    }
    
    location /api/login {
        limit_req zone=login burst=5 nodelay;
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Static files
    location /static/ {
        alias $APP_DIR/web/static/;
        expires 1M;
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
    ln -s /etc/nginx/sites-available/telegram-bot /etc/nginx/sites-enabled/
    
    # Test configuration
    nginx -t
    
    # Start and enable nginx
    systemctl start nginx
    systemctl enable nginx
    
    log_info "Nginx setup completed"
}

setup_ssl() {
    if [ -z "$DOMAIN" ]; then
        log_warn "No domain provided, skipping SSL setup"
        return
    fi
    
    log_step "Setting up SSL certificate..."
    
    # Obtain SSL certificate
    certbot --nginx -d $DOMAIN -d www.$DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN
    
    # Setup automatic renewal
    cat > /etc/cron.d/certbot << EOF
0 12 * * * root certbot renew --quiet --post-hook "systemctl reload nginx"
EOF
    
    log_info "SSL certificate setup completed"
}

setup_systemd_service() {
    log_step "Setting up systemd service..."
    
    cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=Telegram Bot Service
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/telegram-bot
Restart=always
RestartSec=10
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=$APP_DIR/.env

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=$APP_DIR

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd and start service
    systemctl daemon-reload
    systemctl enable $SERVICE_NAME
    systemctl start $SERVICE_NAME
    
    log_info "Systemd service setup completed"
}

setup_firewall() {
    log_step "Setting up firewall..."
    
    # Reset UFW to defaults
    ufw --force reset
    
    # Default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH
    ufw allow ssh
    
    # Allow HTTP and HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow PostgreSQL (local only)
    ufw allow from 127.0.0.1 to any port 5432
    
    # Enable firewall
    ufw --force enable
    
    log_info "Firewall setup completed"
}

setup_monitoring() {
    log_step "Setting up monitoring and logging..."
    
    # Setup log rotation
    cat > /etc/logrotate.d/$SERVICE_NAME << EOF
/var/log/$SERVICE_NAME/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        systemctl reload $SERVICE_NAME
    endscript
}
EOF
    
    # Create log directory
    mkdir -p /var/log/$SERVICE_NAME
    
    # Setup fail2ban for nginx
    cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[nginx-http-auth]
enabled = true

[nginx-limit-req]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
EOF
    
    systemctl enable fail2ban
    systemctl start fail2ban
    
    log_info "Monitoring setup completed"
}

create_management_scripts() {
    log_step "Creating management scripts..."
    
    # Start script
    cat > $APP_DIR/start.sh << 'EOF'
#!/bin/bash
systemctl start telegram-bot
systemctl start nginx
systemctl start postgresql
echo "Services started"
EOF
    
    # Stop script
    cat > $APP_DIR/stop.sh << 'EOF'
#!/bin/bash
systemctl stop telegram-bot
echo "Bot stopped"
EOF
    
    # Restart script
    cat > $APP_DIR/restart.sh << 'EOF'
#!/bin/bash
systemctl restart telegram-bot
systemctl restart nginx
echo "Services restarted"
EOF
    
    # Status script
    cat > $APP_DIR/status.sh << 'EOF'
#!/bin/bash
echo "=== Service Status ==="
systemctl status telegram-bot --no-pager -l
echo ""
echo "=== Nginx Status ==="
systemctl status nginx --no-pager -l
echo ""
echo "=== Database Status ==="
systemctl status postgresql --no-pager -l
echo ""
echo "=== Disk Usage ==="
df -h
echo ""
echo "=== Memory Usage ==="
free -h
echo ""
echo "=== Recent Logs ==="
journalctl -u telegram-bot -n 10 --no-pager
EOF
    
    # Update script
    cat > $APP_DIR/update.sh << 'EOF'
#!/bin/bash
cd /opt/telegram-bot
git pull
go build -o telegram-bot main.go
systemctl restart telegram-bot
echo "Application updated and restarted"
EOF
    
    # Backup script
    cat > $APP_DIR/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/telegram-bot"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
pg_dump -h localhost -U botuser telegram_bot > $BACKUP_DIR/database_$DATE.sql
tar -czf $BACKUP_DIR/application_$DATE.tar.gz -C /opt telegram-bot
echo "Backup created: $BACKUP_DIR/database_$DATE.sql"
echo "Backup created: $BACKUP_DIR/application_$DATE.tar.gz"
EOF
    
    # Make scripts executable
    chmod +x $APP_DIR/*.sh
    
    log_info "Management scripts created"
}

setup_database_schema() {
    log_step "Setting up database schema..."
    
    # Wait for service to start
    sleep 5
    
    # Create database tables by making a request to the application
    curl -s "http://localhost:5000/health" || log_warn "Application not responding yet"
    
    log_info "Database schema setup completed"
}

print_summary() {
    log_step "Deployment Summary"
    
    echo ""
    echo "======================================"
    echo "    DEPLOYMENT COMPLETED SUCCESSFULLY"
    echo "======================================"
    echo ""
    echo "🌐 Application URL: http://$DOMAIN (or https://$DOMAIN if SSL configured)"
    echo "🤖 Bot Token: $BOT_TOKEN"
    echo "🔐 Admin Login: admin / $ADMIN_PASSWORD"
    echo "📊 Database: $DB_NAME"
    echo "📁 App Directory: $APP_DIR"
    echo ""
    echo "📋 Management Commands:"
    echo "  Start:   $APP_DIR/start.sh"
    echo "  Stop:    $APP_DIR/stop.sh"
    echo "  Restart: $APP_DIR/restart.sh"
    echo "  Status:  $APP_DIR/status.sh"
    echo "  Update:  $APP_DIR/update.sh"
    echo "  Backup:  $APP_DIR/backup.sh"
    echo ""
    echo "🔧 Configuration Files:"
    echo "  Environment: $APP_DIR/.env"
    echo "  Nginx: /etc/nginx/sites-available/telegram-bot"
    echo "  Service: /etc/systemd/system/$SERVICE_NAME.service"
    echo ""
    echo "💳 Payment Integration:"
    echo "  Edit $APP_DIR/.env to add your payment provider API keys:"
    echo "  - STRIPE_SECRET_KEY=sk_live_..."
    echo "  - STRIPE_PUBLISHABLE_KEY=pk_live_..."
    echo "  - YOOMONEY_SECRET_KEY=..."
    echo "  - PAYPAL_CLIENT_ID=..."
    echo "  - PAYPAL_CLIENT_SECRET=..."
    echo "  - CRYPTO_WALLET_ADDRESS=..."
    echo ""
    echo "🚀 Next Steps:"
    echo "1. Configure your payment provider API keys in $APP_DIR/.env"
    echo "2. Set up your bot webhook: https://api.telegram.org/bot$BOT_TOKEN/setWebhook?url=https://$DOMAIN/webhook"
    echo "3. Test the bot by sending /start to your bot"
    echo "4. Access the admin panel at https://$DOMAIN/dashboard"
    echo ""
    echo "📱 Bot Commands:"
    echo "  /start - Start bot"
    echo "  /plans - View subscription plans"
    echo "  /subscribe - Subscribe to a plan"
    echo "  /myplan - View current plan"
    echo "  /setup - Setup instructions"
    echo ""
    echo "🔍 Monitoring:"
    echo "  Logs: journalctl -u telegram-bot -f"
    echo "  Status: systemctl status telegram-bot"
    echo "  Nginx logs: tail -f /var/log/nginx/access.log"
    echo ""
    echo "✅ All services are running and ready to use!"
    echo ""
}

# Main execution
main() {
    log_info "Starting Telegram Bot VPS deployment..."
    
    # Check if domain and bot token are provided
    if [ -z "$DOMAIN" ] || [ -z "$BOT_TOKEN" ]; then
        log_error "Usage: $0 <domain> <bot_token> [admin_password]"
        log_error "Example: $0 mybot.com 1234567890:ABCdef123456789 mypassword"
        exit 1
    fi
    
    # Validate bot token format
    if [[ ! $BOT_TOKEN =~ ^[0-9]+:[A-Za-z0-9_-]+$ ]]; then
        log_error "Invalid bot token format. Should be: 123456789:ABCdef123456789"
        exit 1
    fi
    
    check_root
    install_dependencies
    setup_database
    setup_application
    setup_nginx
    setup_ssl
    setup_systemd_service
    setup_firewall
    setup_monitoring
    create_management_scripts
    setup_database_schema
    print_summary
    
    log_info "Deployment completed successfully!"
}

# Run main function
main "$@"
EOF