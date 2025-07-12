# Telegram Bot VPS Deployment Guide

## Quick Setup (One Command)

### 1. Copy project files to VPS
```bash
# On your local machine, create deployment package
tar -czf telegram-bot.tar.gz --exclude='.git' --exclude='node_modules' .

# Copy to VPS
scp telegram-bot.tar.gz user@your-server-ip:/tmp/

# On VPS, extract files
ssh user@your-server-ip
cd /tmp
tar -xzf telegram-bot.tar.gz
mv telegram-bot-* telegram-bot
```

### 2. Run deployment script
```bash
# Make script executable and run
chmod +x deploy.sh
./deploy.sh
```

The script will automatically:
- Install all dependencies (Go, PostgreSQL, Nginx)
- Set up database with all tables
- Build and configure the application
- Create systemd service
- Configure Nginx reverse proxy
- Set up firewall rules
- Create management scripts

## Manual Configuration

### 1. Set Telegram Bot Token
```bash
cd /opt/telegram-bot
nano .env
```

Add your bot token:
```
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

### 2. Configure Domain (Optional)
```bash
sudo nano /etc/nginx/sites-available/telegram-bot
```

Replace `server_name _;` with your domain:
```
server_name yourdomain.com;
```

### 3. Set up SSL Certificate
```bash
sudo certbot --nginx -d yourdomain.com
```

### 4. Start the bot
```bash
./restart.sh
```

## Management Commands

### Service Management
```bash
# Check status
sudo systemctl status telegram-bot

# Start service
sudo systemctl start telegram-bot

# Stop service
sudo systemctl stop telegram-bot

# Restart service
sudo systemctl restart telegram-bot

# View logs
sudo journalctl -u telegram-bot -f
```

### Quick Management Scripts
```bash
# Start bot
./start.sh

# Stop bot
./stop.sh

# Restart bot
./restart.sh

# Check status and logs
./status.sh

# Update from git
./update.sh
```

## Database Access

### Connect to database
```bash
PGPASSWORD=telegrambot123 psql -h localhost -U telegrambot -d telegrambot
```

### Common database commands
```sql
-- Check users
SELECT * FROM users;

-- Check groups
SELECT * FROM user_groups;

-- Check subscriptions
SELECT * FROM group_subscriptions;

-- Check violations
SELECT * FROM user_violations;
```

## Web Interface

### Access URLs
- Main interface: `http://your-server-ip/`
- Admin dashboard: `http://your-server-ip/dashboard`
- User dashboard: `http://your-server-ip/user-dashboard`
- Group management: `http://your-server-ip/group-selector`

### Default Login
- **Admin**: username: `admin`, password: `admin123`
- **User**: Create account via Telegram bot

## Security Configuration

### 1. Change default passwords
```bash
# Update .env file
nano .env

# Change ADMIN_PASSWORD and database password
# Then restart service
./restart.sh
```

### 2. Configure firewall
```bash
# Check firewall status
sudo ufw status

# Allow specific IPs only (optional)
sudo ufw allow from YOUR_IP_ADDRESS to any port 22
sudo ufw deny 22
```

### 3. Regular backups
```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/telegram-bot"
mkdir -p $BACKUP_DIR

# Backup database
PGPASSWORD=telegrambot123 pg_dump -h localhost -U telegrambot telegrambot > $BACKUP_DIR/db_$DATE.sql

# Backup application files
tar -czf $BACKUP_DIR/app_$DATE.tar.gz /opt/telegram-bot

echo "Backup completed: $BACKUP_DIR"
EOF

chmod +x backup.sh
```

## Troubleshooting

### Check service status
```bash
./status.sh
```

### Common issues and solutions

#### 1. Service won't start
```bash
# Check logs
sudo journalctl -u telegram-bot -n 100

# Check configuration
cd /opt/telegram-bot
./telegram-bot --help
```

#### 2. Database connection error
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Test connection
PGPASSWORD=telegrambot123 psql -h localhost -U telegrambot -d telegrambot -c "SELECT 1;"
```

#### 3. Web interface not accessible
```bash
# Check Nginx status
sudo systemctl status nginx

# Check configuration
sudo nginx -t

# Check if port is open
sudo netstat -tlnp | grep :80
```

#### 4. Bot not responding
```bash
# Check if bot token is set
cat .env | grep TELEGRAM_BOT_TOKEN

# Check network connectivity
curl -s https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe
```

### View detailed logs
```bash
# Application logs
sudo journalctl -u telegram-bot -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*-main.log
```

## Performance Optimization

### 1. Database optimization
```sql
-- Connect to database
PGPASSWORD=telegrambot123 psql -h localhost -U telegrambot -d telegrambot

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_user_groups_user_id ON user_groups(user_id);
CREATE INDEX IF NOT EXISTS idx_user_violations_user_id ON user_violations(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
```

### 2. System optimization
```bash
# Increase file limits
echo "telegram-bot soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "telegram-bot hard nofile 65536" | sudo tee -a /etc/security/limits.conf
```

## Updates and Maintenance

### Update application
```bash
# Use update script
./update.sh

# Or manually
git pull origin main
go mod tidy
go build -o telegram-bot .
sudo systemctl restart telegram-bot
```

### Regular maintenance
```bash
# Clean old logs
sudo journalctl --vacuum-time=7d

# Update system packages
sudo apt update && sudo apt upgrade -y

# Restart services if needed
sudo systemctl restart telegram-bot
```

## Monitoring

### Set up monitoring script
```bash
cat > monitor.sh << 'EOF'
#!/bin/bash
# Check if service is running
if ! systemctl is-active --quiet telegram-bot; then
    echo "Telegram Bot service is down! Restarting..."
    sudo systemctl restart telegram-bot
fi

# Check disk space
DISK_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 85 ]; then
    echo "Warning: Disk usage is ${DISK_USAGE}%"
fi

# Check memory usage
MEM_USAGE=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')
if (( $(echo "$MEM_USAGE > 90" | bc -l) )); then
    echo "Warning: Memory usage is ${MEM_USAGE}%"
fi
EOF

chmod +x monitor.sh

# Add to crontab (run every 5 minutes)
(crontab -l 2>/dev/null; echo "*/5 * * * * /opt/telegram-bot/monitor.sh") | crontab -
```

## Support

For issues and support:
1. Check the logs using `./status.sh`
2. Review this documentation
3. Check the GitHub repository for updates
4. Contact the development team

---

**Note**: This deployment script is tested on Ubuntu 20.04/22.04 and Debian 11/12. For other distributions, you may need to adjust package names and paths.