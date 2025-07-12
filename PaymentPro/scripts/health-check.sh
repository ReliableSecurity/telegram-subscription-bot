#!/bin/bash

# Health check script for Telegram Bot
# Monitors all system components and reports status

set -e

# Configuration
BOT_SERVICE="telegram-bot"
WEB_PORT=${WEB_PORT:-5000}
DB_NAME=${PGDATABASE:-"telegram_bot"}
DOMAIN_NAME=${DOMAIN_NAME:-"localhost"}
LOG_FILE="/tmp/health-check.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Status tracking
OVERALL_STATUS=0
ISSUES=()

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}✓${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1" | tee -a "$LOG_FILE"
    ISSUES+=("WARNING: $1")
}

error() {
    echo -e "${RED}✗${NC} $1" | tee -a "$LOG_FILE"
    ISSUES+=("ERROR: $1")
    OVERALL_STATUS=1
}

# Check if service is running
check_service() {
    local service_name=$1
    if systemctl is-active --quiet "$service_name"; then
        success "$service_name service is running"
        return 0
    else
        error "$service_name service is not running"
        return 1
    fi
}

# Check if port is listening
check_port() {
    local port=$1
    local service_name=$2
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        success "$service_name is listening on port $port"
        return 0
    elif ss -tuln 2>/dev/null | grep -q ":$port "; then
        success "$service_name is listening on port $port"
        return 0
    else
        error "$service_name is not listening on port $port"
        return 1
    fi
}

# Check HTTP endpoint
check_http() {
    local url=$1
    local service_name=$2
    local expected_code=${3:-200}
    
    if command -v curl &> /dev/null; then
        local response_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" --max-time 10)
        if [[ "$response_code" == "$expected_code" ]]; then
            success "$service_name HTTP endpoint is responding ($response_code)"
            return 0
        else
            error "$service_name HTTP endpoint returned $response_code (expected $expected_code)"
            return 1
        fi
    else
        warning "curl not available, skipping HTTP check for $service_name"
        return 1
    fi
}

# Check database connection
check_database() {
    log "Checking database connection..."
    
    if command -v psql &> /dev/null; then
        if [[ -n "$DATABASE_URL" ]]; then
            if psql "$DATABASE_URL" -c "SELECT 1;" &> /dev/null; then
                success "Database connection successful"
                
                # Check database size
                local db_size=$(psql "$DATABASE_URL" -t -c "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));" 2>/dev/null | xargs)
                if [[ -n "$db_size" ]]; then
                    log "Database size: $db_size"
                fi
                
                return 0
            else
                error "Database connection failed"
                return 1
            fi
        else
            warning "DATABASE_URL not set, skipping database check"
            return 1
        fi
    else
        warning "psql not available, skipping database check"
        return 1
    fi
}

# Check disk space
check_disk_space() {
    log "Checking disk space..."
    
    local usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    local available=$(df -h / | awk 'NR==2 {print $4}')
    
    if [[ $usage -lt 80 ]]; then
        success "Disk usage: ${usage}% (${available} available)"
    elif [[ $usage -lt 90 ]]; then
        warning "Disk usage: ${usage}% (${available} available) - Consider cleanup"
    else
        error "Disk usage: ${usage}% (${available} available) - Critical!"
    fi
}

# Check memory usage
check_memory() {
    log "Checking memory usage..."
    
    if command -v free &> /dev/null; then
        local mem_info=$(free -h | grep "Mem:")
        local total=$(echo $mem_info | awk '{print $2}')
        local used=$(echo $mem_info | awk '{print $3}')
        local available=$(echo $mem_info | awk '{print $7}')
        local usage_percent=$(free | grep "Mem:" | awk '{printf "%.0f", $3/$2 * 100.0}')
        
        if [[ $usage_percent -lt 80 ]]; then
            success "Memory usage: ${usage_percent}% (${used}/${total}, ${available} available)"
        elif [[ $usage_percent -lt 90 ]]; then
            warning "Memory usage: ${usage_percent}% (${used}/${total}, ${available} available)"
        else
            error "Memory usage: ${usage_percent}% (${used}/${total}, ${available} available) - Critical!"
        fi
    else
        warning "free command not available, skipping memory check"
    fi
}

# Check SSL certificate
check_ssl() {
    local domain=$1
    
    if [[ "$domain" == "localhost" ]]; then
        log "Skipping SSL check for localhost"
        return 0
    fi
    
    log "Checking SSL certificate for $domain..."
    
    if command -v openssl &> /dev/null; then
        local cert_info=$(echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null | openssl x509 -noout -dates 2>/dev/null)
        if [[ $? -eq 0 ]]; then
            local expiry_date=$(echo "$cert_info" | grep "notAfter" | cut -d= -f2)
            local expiry_epoch=$(date -d "$expiry_date" +%s 2>/dev/null)
            local current_epoch=$(date +%s)
            local days_until_expiry=$(( (expiry_epoch - current_epoch) / 86400 ))
            
            if [[ $days_until_expiry -gt 30 ]]; then
                success "SSL certificate valid (expires in $days_until_expiry days)"
            elif [[ $days_until_expiry -gt 7 ]]; then
                warning "SSL certificate expires in $days_until_expiry days - consider renewal"
            else
                error "SSL certificate expires in $days_until_expiry days - urgent renewal needed!"
            fi
        else
            error "Failed to check SSL certificate for $domain"
        fi
    else
        warning "openssl not available, skipping SSL check"
    fi
}

# Check log files for errors
check_logs() {
    log "Checking recent logs for errors..."
    
    if command -v journalctl &> /dev/null; then
        local error_count=$(journalctl -u "$BOT_SERVICE" --since "1 hour ago" --no-pager | grep -i "error\|fatal\|panic" | wc -l)
        local warning_count=$(journalctl -u "$BOT_SERVICE" --since "1 hour ago" --no-pager | grep -i "warning\|warn" | wc -l)
        
        if [[ $error_count -eq 0 ]]; then
            success "No errors in recent logs"
        else
            warning "$error_count errors found in recent logs"
        fi
        
        if [[ $warning_count -gt 0 ]]; then
            log "$warning_count warnings found in recent logs"
        fi
    else
        warning "journalctl not available, skipping log check"
    fi
}

# Check bot API connectivity
check_bot_api() {
    log "Checking Telegram Bot API connectivity..."
    
    if [[ -n "$TELEGRAM_BOT_TOKEN" ]]; then
        local api_url="https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"
        if check_http "$api_url" "Telegram Bot API"; then
            # Get bot info
            local bot_info=$(curl -s "$api_url" --max-time 10)
            local bot_username=$(echo "$bot_info" | grep -o '"username":"[^"]*"' | cut -d'"' -f4)
            if [[ -n "$bot_username" ]]; then
                success "Bot API working - @$bot_username"
            fi
        fi
    else
        warning "TELEGRAM_BOT_TOKEN not set, skipping bot API check"
    fi
}

# Generate summary report
generate_summary() {
    echo ""
    echo "=========================================="
    if [[ $OVERALL_STATUS -eq 0 ]]; then
        echo -e "${GREEN}✓ SYSTEM HEALTH: ALL SYSTEMS OPERATIONAL${NC}"
    else
        echo -e "${RED}✗ SYSTEM HEALTH: ISSUES DETECTED${NC}"
    fi
    echo "=========================================="
    echo "Check completed at: $(date)"
    
    if [[ ${#ISSUES[@]} -gt 0 ]]; then
        echo ""
        echo "Issues found:"
        for issue in "${ISSUES[@]}"; do
            echo "  $issue"
        done
    fi
    
    echo ""
    echo "System Information:"
    echo "  Hostname: $(hostname)"
    echo "  Uptime: $(uptime -p 2>/dev/null || uptime)"
    echo "  Load: $(cat /proc/loadavg 2>/dev/null | cut -d' ' -f1-3 || echo 'N/A')"
    
    if command -v systemctl &> /dev/null; then
        echo ""
        echo "Service Status:"
        systemctl is-active "$BOT_SERVICE" &>/dev/null && echo "  ✓ $BOT_SERVICE: running" || echo "  ✗ $BOT_SERVICE: stopped"
        systemctl is-active postgresql &>/dev/null && echo "  ✓ postgresql: running" || echo "  ✗ postgresql: stopped"
        systemctl is-active nginx &>/dev/null && echo "  ✓ nginx: running" || echo "  ✗ nginx: stopped"
    fi
    
    echo ""
}

# Main health check function
main() {
    echo "Starting system health check..." > "$LOG_FILE"
    
    log "=========================================="
    log "Telegram Bot System Health Check"
    log "=========================================="
    
    # Load environment variables if available
    if [[ -f ".env" ]]; then
        set -a
        source .env
        set +a
        log "Loaded environment variables from .env"
    fi
    
    # Core service checks
    log "Checking core services..."
    check_service "$BOT_SERVICE"
    check_service "postgresql"
    
    # Port checks
    log "Checking network connectivity..."
    check_port "$WEB_PORT" "Web server"
    check_port "5432" "PostgreSQL"
    
    # HTTP endpoint checks
    log "Checking HTTP endpoints..."
    check_http "http://localhost:$WEB_PORT/health" "Local web server"
    
    if [[ "$DOMAIN_NAME" != "localhost" ]]; then
        check_http "https://$DOMAIN_NAME/health" "External web server"
    fi
    
    # Database connectivity
    check_database
    
    # System resource checks
    log "Checking system resources..."
    check_disk_space
    check_memory
    
    # SSL certificate check
    if [[ "$DOMAIN_NAME" != "localhost" ]]; then
        check_ssl "$DOMAIN_NAME"
    fi
    
    # Log analysis
    check_logs
    
    # Bot API check
    check_bot_api
    
    # Generate summary
    generate_summary
    
    # Save detailed log
    log "Detailed log saved to: $LOG_FILE"
    
    exit $OVERALL_STATUS
}

# Show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --quiet       Only show errors and warnings"
    echo "  --json        Output results in JSON format"
    echo "  --help        Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  WEB_PORT      Web server port (default: 5000)"
    echo "  DOMAIN_NAME   Domain name for SSL check (default: localhost)"
    echo "  DATABASE_URL  PostgreSQL connection string"
    echo ""
}

# Parse command line arguments
QUIET=false
JSON_OUTPUT=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --quiet)
            QUIET=true
            shift
            ;;
        --json)
            JSON_OUTPUT=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Run main function
if [[ "$JSON_OUTPUT" == "true" ]]; then
    # TODO: Implement JSON output format
    main 2>&1 | grep -E "(✓|⚠|✗)" | sed 's/.*\[.*\] //' > "$LOG_FILE.json"
else
    main
fi