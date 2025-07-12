#!/bin/bash

# Database setup script for Telegram Bot
# Creates database, user, and initializes schema

set -e

# Default values
DB_NAME="telegram_bot"
DB_USER="telegram_user"
DB_HOST="localhost"
DB_PORT="5432"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Check if PostgreSQL is running
check_postgresql() {
    if ! sudo systemctl is-active --quiet postgresql; then
        log "Starting PostgreSQL..."
        sudo systemctl start postgresql
    fi
    
    if ! sudo systemctl is-enabled --quiet postgresql; then
        log "Enabling PostgreSQL to start on boot..."
        sudo systemctl enable postgresql
    fi
    
    log "PostgreSQL is running"
}

# Generate secure password
generate_password() {
    if command -v openssl &> /dev/null; then
        openssl rand -base64 32
    else
        # Fallback to /dev/urandom
        tr -dc A-Za-z0-9 </dev/urandom | head -c 32
    fi
}

# Create database and user
setup_database() {
    log "Setting up database..."
    
    # Generate password if not provided
    if [[ -z "$DB_PASSWORD" ]]; then
        DB_PASSWORD=$(generate_password)
        log "Generated secure password for database user"
    fi
    
    # Create database and user
    sudo -u postgres psql << EOF
-- Create database if not exists
SELECT 'CREATE DATABASE $DB_NAME' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec

-- Create user if not exists
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
        CREATE ROLE $DB_USER LOGIN PASSWORD '$DB_PASSWORD';
    END IF;
END
\$\$;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
ALTER USER $DB_USER CREATEDB;

-- Connect to the database and grant schema privileges
\c $DB_NAME
GRANT ALL ON SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO $DB_USER;

\q
EOF
    
    log "Database and user created successfully"
}

# Initialize database schema
initialize_schema() {
    log "Initializing database schema..."
    
    # Check if schema file exists
    if [[ -f "database/schema.sql" ]]; then
        log "Found schema file, applying..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f database/schema.sql
    else
        log "No schema file found, creating basic schema..."
        create_basic_schema
    fi
    
    log "Database schema initialized"
}

# Create basic schema if file doesn't exist
create_basic_schema() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME << 'EOF'
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    is_premium BOOLEAN DEFAULT false,
    web_username VARCHAR(255),
    web_password VARCHAR(255),
    is_web_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Subscription plans table
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

-- User subscriptions table
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    plan_id INTEGER REFERENCES subscription_plans(id),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    auto_renew BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INTEGER REFERENCES user_subscriptions(id),
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_provider VARCHAR(50),
    provider_payment_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- User groups table
CREATE TABLE IF NOT EXISTS user_groups (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT NOT NULL,
    chat_type VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, chat_id)
);

-- Violations table
CREATE TABLE IF NOT EXISTS violations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT NOT NULL,
    violation_type VARCHAR(100) NOT NULL,
    description TEXT,
    severity INTEGER DEFAULT 1,
    is_resolved BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Analytics table
CREATE TABLE IF NOT EXISTS analytics (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI recommendations table
CREATE TABLE IF NOT EXISTS ai_recommendations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT,
    recommendation_type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'pending',
    confidence_score DECIMAL(3,2) DEFAULT 0.5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default subscription plans
INSERT INTO subscription_plans (name, price_cents, duration_days, max_groups, features) VALUES
('Basic', 19900, 30, 3, 'Basic moderation, Up to 3 groups, Email support'),
('Pro', 49900, 90, 10, 'Advanced moderation, Up to 10 groups, Analytics dashboard, Priority support'),
('Premium', 149900, 365, 50, 'All features, Up to 50 groups, AI recommendations, 24/7 support, Custom integrations')
ON CONFLICT DO NOTHING;

-- Create indexes for better performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_groups_user_id ON user_groups(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_groups_chat_id ON user_groups(chat_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_violations_user_id ON violations(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_violations_chat_id ON violations(chat_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_analytics_chat_id ON analytics(chat_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_analytics_created_at ON analytics(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ai_recommendations_user_id ON ai_recommendations(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ai_recommendations_status ON ai_recommendations(status);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ai_recommendations_updated_at ON ai_recommendations;
CREATE TRIGGER update_ai_recommendations_updated_at
    BEFORE UPDATE ON ai_recommendations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
EOF
}

# Test database connection
test_connection() {
    log "Testing database connection..."
    
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;" &> /dev/null; then
        log "Database connection successful"
    else
        error "Database connection failed"
    fi
}

# Save connection info
save_connection_info() {
    log "Saving database connection information..."
    
    cat > database-credentials.txt << EOF
Database Setup Complete
========================

Database Name: $DB_NAME
Database User: $DB_USER
Database Password: $DB_PASSWORD
Database Host: $DB_HOST
Database Port: $DB_PORT

Connection URL: postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME

Add this to your .env file:
DATABASE_URL=postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME
PGHOST=$DB_HOST
PGPORT=$DB_PORT
PGUSER=$DB_USER
PGPASSWORD=$DB_PASSWORD
PGDATABASE=$DB_NAME

IMPORTANT: Keep this file secure and delete after copying credentials!
EOF
    
    chmod 600 database-credentials.txt
    log "Database credentials saved to database-credentials.txt"
}

# Display usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n, --name NAME       Database name (default: telegram_bot)"
    echo "  -u, --user USER       Database user (default: telegram_user)"
    echo "  -p, --password PASS   Database password (generated if not provided)"
    echo "  -h, --host HOST       Database host (default: localhost)"
    echo "  --port PORT           Database port (default: 5432)"
    echo "  --help                Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 --name mybot --user botuser --password mypassword"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            DB_NAME="$2"
            shift 2
            ;;
        -u|--user)
            DB_USER="$2"
            shift 2
            ;;
        -p|--password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        -h|--host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Main execution
main() {
    log "Starting database setup for Telegram Bot..."
    log "Database: $DB_NAME, User: $DB_USER, Host: $DB_HOST:$DB_PORT"
    
    check_postgresql
    setup_database
    initialize_schema
    test_connection
    save_connection_info
    
    log "Database setup completed successfully!"
    log "Next steps:"
    log "1. Copy database credentials from database-credentials.txt to your .env file"
    log "2. Delete database-credentials.txt for security"
    log "3. Build and run your Telegram bot application"
}

# Run main function
main