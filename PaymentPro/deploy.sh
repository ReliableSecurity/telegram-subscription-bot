#!/bin/bash

# Telegram Bot Deployment Script for VPS
# This script sets up the complete environment on Ubuntu/Debian VPS

set -e

echo "🚀 Starting Telegram Bot deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root. Please run as a regular user with sudo privileges."
   exit 1
fi

# Update system
print_status "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install required packages
print_status "Installing required packages..."
sudo apt install -y curl wget git build-essential postgresql postgresql-contrib nginx certbot python3-certbot-nginx ufw

# Install Go
print_status "Installing Go..."
if ! command -v go &> /dev/null; then
    GO_VERSION="1.21.5"
    wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    
    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
fi

# Verify Go installation
go version

# Create project directory
PROJECT_DIR="/opt/telegram-bot"
print_status "Creating project directory: $PROJECT_DIR"
sudo mkdir -p $PROJECT_DIR
sudo chown $USER:$USER $PROJECT_DIR

# Clone or copy project files
print_status "Setting up project files..."
cd $PROJECT_DIR

# Copy current project files if they exist
if [ -d "/tmp/telegram-bot" ]; then
    cp -r /tmp/telegram-bot/* ./
else
    print_warning "Project files not found in /tmp/telegram-bot. Please copy your project files to $PROJECT_DIR"
fi

# Set up PostgreSQL
print_status "Setting up PostgreSQL database..."
sudo -u postgres psql -c "CREATE USER telegrambot WITH PASSWORD 'telegrambot123';" || true
sudo -u postgres psql -c "CREATE DATABASE telegrambot OWNER telegrambot;" || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE telegrambot TO telegrambot;" || true

# Create database schema
print_status "Creating database schema..."
cat > schema.sql << 'EOF'
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    plan_name VARCHAR(100) DEFAULT 'Free',
    plan_expires_at TIMESTAMP,
    total_spent DECIMAL(10,2) DEFAULT 0.00,
    web_username VARCHAR(255),
    web_password VARCHAR(255),
    is_web_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Subscription plans
CREATE TABLE IF NOT EXISTS subscription_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    duration_days INTEGER NOT NULL,
    features TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    plan_id INTEGER REFERENCES subscription_plans(id),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'RUB',
    payment_method VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    payment_provider VARCHAR(50),
    provider_payment_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- User groups
CREATE TABLE IF NOT EXISTS user_groups (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    group_id BIGINT NOT NULL,
    group_title VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, group_id)
);

-- Group subscriptions (individual billing per group)
CREATE TABLE IF NOT EXISTS group_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES user_groups(id),
    plan_id INTEGER REFERENCES subscription_plans(id),
    status VARCHAR(20) DEFAULT 'active',
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Group settings
CREATE TABLE IF NOT EXISTS group_settings (
    id SERIAL PRIMARY KEY,
    group_id INTEGER REFERENCES user_groups(id) UNIQUE,
    moderation_enabled BOOLEAN DEFAULT TRUE,
    auto_ban_enabled BOOLEAN DEFAULT TRUE,
    warning_threshold INTEGER DEFAULT 3,
    temp_ban_duration INTEGER DEFAULT 24,
    permanent_ban_threshold INTEGER DEFAULT 5,
    welcome_message TEXT,
    goodbye_message TEXT,
    rules TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Group statistics
CREATE TABLE IF NOT EXISTS group_statistics (
    id SERIAL PRIMARY KEY,
    group_id INTEGER REFERENCES user_groups(id),
    date DATE DEFAULT CURRENT_DATE,
    messages_count INTEGER DEFAULT 0,
    new_members INTEGER DEFAULT 0,
    left_members INTEGER DEFAULT 0,
    warnings_issued INTEGER DEFAULT 0,
    bans_issued INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(group_id, date)
);

-- User violations
CREATE TABLE IF NOT EXISTS user_violations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    chat_id BIGINT NOT NULL,
    violation_type VARCHAR(50) NOT NULL,
    violation_reason TEXT,
    message_text TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    username VARCHAR(255),
    first_name VARCHAR(255),
    telegram_id BIGINT
);

-- Insert default subscription plans
INSERT INTO subscription_plans (name, description, price, duration_days, features) VALUES
('Базовый', 'Базовые функции модерации', 199.00, 30, ARRAY['Базовая модерация', 'До 100 участников', 'Статистика']),
('Стандартный', 'Расширенные возможности', 399.00, 30, ARRAY['Базовая модерация', 'До 500 участников', 'Статистика', 'Расширенная модерация', 'Кастомные правила']),
('Премиум', 'Все возможности', 799.00, 30, ARRAY['Базовая модерация', 'Безлимитные участники', 'Статистика', 'Расширенная модерация', 'Кастомные правила', 'Приоритетная поддержка', 'API доступ'])
ON CONFLICT DO NOTHING;

-- Insert test user
INSERT INTO users (telegram_id, username, first_name, web_username, web_password, is_web_active) VALUES
(1712174719, 'ReliableSecurity', 'RS', 'admin', 'admin123', true)
ON CONFLICT (telegram_id) DO UPDATE SET
web_username = EXCLUDED.web_username,
web_password = EXCLUDED.web_password,
is_web_active = EXCLUDED.is_web_active;
EOF

# Apply database schema
PGPASSWORD=telegrambot123 psql -h localhost -U telegrambot -d telegrambot -f schema.sql

print_status "Database setup completed!"

# Create environment file
print_status "Creating environment configuration..."
cat > .env << 'EOF'
# Database
DATABASE_URL=postgres://telegrambot:telegrambot123@localhost/telegrambot
PGHOST=localhost
PGPORT=5432
PGUSER=telegrambot
PGPASSWORD=telegrambot123
PGDATABASE=telegrambot

# Web server
PORT=5000
HOST=0.0.0.0

# Admin credentials
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
EOF

# Build the application
print_status "Building Go application..."
go mod tidy
go build -o telegram-bot .

# Create systemd service
print_status "Creating systemd service..."
sudo tee /etc/systemd/system/telegram-bot.service << EOF
[Unit]
Description=Telegram Bot Service
After=network.target postgresql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$PROJECT_DIR
ExecStart=$PROJECT_DIR/telegram-bot
Restart=always
RestartSec=10
Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
EnvironmentFile=$PROJECT_DIR/.env

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot

# Configure Nginx
print_status "Configuring Nginx..."
sudo tee /etc/nginx/sites-available/telegram-bot << 'EOF'
server {
    listen 80;
    server_name _;
    
    location / {
        proxy_pass http://127.0.0.1:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }
}
EOF

# Enable Nginx site
sudo ln -sf /etc/nginx/sites-available/telegram-bot /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl restart nginx

# Configure firewall
print_status "Configuring firewall..."
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable

# Create management scripts
print_status "Creating management scripts..."

# Start script
cat > start.sh << 'EOF'
#!/bin/bash
echo "Starting Telegram Bot..."
sudo systemctl start telegram-bot
sudo systemctl status telegram-bot
EOF

# Stop script
cat > stop.sh << 'EOF'
#!/bin/bash
echo "Stopping Telegram Bot..."
sudo systemctl stop telegram-bot
EOF

# Restart script
cat > restart.sh << 'EOF'
#!/bin/bash
echo "Restarting Telegram Bot..."
sudo systemctl restart telegram-bot
sudo systemctl status telegram-bot
EOF

# Status script
cat > status.sh << 'EOF'
#!/bin/bash
echo "=== Telegram Bot Status ==="
sudo systemctl status telegram-bot

echo ""
echo "=== Recent Logs ==="
sudo journalctl -u telegram-bot -n 50 --no-pager

echo ""
echo "=== Database Status ==="
sudo systemctl status postgresql

echo ""
echo "=== Nginx Status ==="
sudo systemctl status nginx
EOF

# Update script
cat > update.sh << 'EOF'
#!/bin/bash
echo "Updating Telegram Bot..."
git pull origin main
go mod tidy
go build -o telegram-bot .
sudo systemctl restart telegram-bot
echo "Update completed!"
EOF

# Make scripts executable
chmod +x start.sh stop.sh restart.sh status.sh update.sh

print_status "Deployment completed successfully!"
print_status "=== IMPORTANT INFORMATION ==="
print_status "Project directory: $PROJECT_DIR"
print_status "Database: telegrambot (user: telegrambot, password: telegrambot123)"
print_status "Web interface: http://your-server-ip"
print_status "Admin login: admin / admin123"
print_status ""
print_status "Available commands:"
print_status "  ./start.sh    - Start the bot"
print_status "  ./stop.sh     - Stop the bot"
print_status "  ./restart.sh  - Restart the bot"
print_status "  ./status.sh   - Check status and logs"
print_status "  ./update.sh   - Update from git"
print_status ""
print_status "Service management:"
print_status "  sudo systemctl status telegram-bot"
print_status "  sudo journalctl -u telegram-bot -f"
print_status ""
print_warning "NEXT STEPS:"
print_warning "1. Set your TELEGRAM_BOT_TOKEN in .env file"
print_warning "2. Configure your domain name in Nginx if needed"
print_warning "3. Set up SSL with: sudo certbot --nginx"
print_warning "4. Update database password for production use"

echo ""
echo "🎉 Deployment completed! Your Telegram Bot is ready to use."