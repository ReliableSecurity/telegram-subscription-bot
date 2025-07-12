#!/bin/bash

# Telegram Bot Backup Script
# This script creates automated backups of the database and configuration

set -e

# Configuration
BACKUP_DIR="/backups"
RETENTION_DAYS=7
DATE=$(date +%Y%m%d_%H%M%S)
LOG_FILE="$BACKUP_DIR/backup.log"

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting backup process..."

# Database backup
if [[ -n "$PGPASSWORD" && -n "$POSTGRES_USER" && -n "$POSTGRES_DB" ]]; then
    BACKUP_FILE="$BACKUP_DIR/telegram_bot_backup_$DATE.sql"
    
    log "Creating database backup..."
    pg_dump -h postgres -U "$POSTGRES_USER" -d "$POSTGRES_DB" > "$BACKUP_FILE"
    
    # Compress backup
    gzip "$BACKUP_FILE"
    log "Database backup created: ${BACKUP_FILE}.gz"
    
    # Calculate backup size
    BACKUP_SIZE=$(du -h "${BACKUP_FILE}.gz" | cut -f1)
    log "Backup size: $BACKUP_SIZE"
else
    log "ERROR: Database credentials not found in environment"
    exit 1
fi

# Clean old backups
log "Cleaning old backups (keeping last $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "telegram_bot_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true

# List current backups
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "telegram_bot_backup_*.sql.gz" | wc -l)
log "Current backup count: $BACKUP_COUNT"

# Disk usage check
DISK_USAGE=$(df -h "$BACKUP_DIR" | awk 'NR==2 {print $5}' | sed 's/%//')
if [[ $DISK_USAGE -gt 80 ]]; then
    log "WARNING: Disk usage is at ${DISK_USAGE}%"
fi

log "Backup process completed successfully"

# Exit with success
exit 0