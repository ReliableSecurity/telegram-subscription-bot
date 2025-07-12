# GitHub Setup Guide

Follow these steps to upload your Telegram Bot to GitHub and deploy it to a VPS with domain binding.

## 📋 Prerequisites

- GitHub account
- Git installed on your local machine
- VPS with Ubuntu 20.04+ or Debian 11+
- Domain name pointed to your VPS
- Telegram Bot Token from [@BotFather](https://t.me/BotFather)

## 🎯 Step 1: Create GitHub Repository

### Option A: Using GitHub Web Interface

1. Go to [GitHub](https://github.com) and sign in
2. Click the "+" icon → "New repository"
3. Repository details:
   - **Name**: `telegram-subscription-bot`
   - **Description**: `Comprehensive Telegram bot with payment system and web dashboard`
   - **Visibility**: Choose Public or Private
   - **Initialize**: ✅ Add a README file
   - **Add .gitignore**: Choose "Go"
   - **Choose a license**: MIT License

4. Click "Create repository"

### Option B: Using GitHub CLI (if installed)

```bash
gh repo create telegram-subscription-bot --public --description "Telegram bot with payment system"
```

## 🎯 Step 2: Upload Your Code

### From Your Local Development Machine:

```bash
# Navigate to your project directory
cd /path/to/your/telegram-subscription-bot

# Initialize git repository (if not already done)
git init

# Add GitHub repository as origin
git remote add origin https://github.com/yourusername/telegram-subscription-bot.git

# Add all files to git
git add .

# Create initial commit
git commit -m "Initial commit: Complete Telegram bot with payment system

Features:
- Multi-group management with individual billing
- Payment processing (Stripe, YooMoney, PayPal, Crypto)
- AI-powered recommendations and moderation
- Web dashboard with analytics
- Automated VPS deployment scripts
- SSL certificate automation
- Database management and backups
- Docker containerization support"

# Push to GitHub
git push -u origin main
```

### From Replit (Current Environment):

```bash
# Configure git (replace with your information)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Add GitHub repository as origin
git remote add origin https://github.com/yourusername/telegram-subscription-bot.git

# Add all files
git add .

# Commit changes
git commit -m "Complete Telegram subscription bot with VPS deployment automation"

# Push to GitHub
git push -u origin main
```

**Note**: You may need to authenticate with GitHub. Use a [Personal Access Token](https://github.com/settings/tokens) if prompted for password.

## 🎯 Step 3: Repository Setup on GitHub

### 3.1 Update Repository Description

1. Go to your repository on GitHub
2. Click the ⚙️ gear icon next to "About"
3. Add description: `Comprehensive Telegram bot platform for group management with subscription handling, payment processing, AI recommendations, and web dashboard`
4. Add topics: `telegram-bot`, `golang`, `payment-processing`, `subscription-management`, `ai-recommendations`, `web-dashboard`, `vps-deployment`
5. Add website URL (will be your domain after deployment)

### 3.2 Enable Repository Features

- ✅ Issues (for bug reports and feature requests)
- ✅ Projects (for task management)
- ✅ Wiki (for additional documentation)
- ✅ Discussions (for community support)

### 3.3 Set Up Branch Protection (Optional but Recommended)

1. Go to Settings → Branches
2. Add rule for `main` branch:
   - ✅ Require pull request reviews before merging
   - ✅ Require status checks to pass before merging
   - ✅ Include administrators

## 🎯 Step 4: VPS Deployment from GitHub

### 4.1 Get Your Bot Token

1. Message [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` and follow instructions
3. Save your bot token (format: `123456789:ABCdef...`)
4. Send `/setcommands` to set bot commands:
   ```
   start - Start the bot and get welcome message
   plans - View available subscription plans
   subscribe - Subscribe to a plan with payment
   myplan - Check current subscription status
   cancel - Cancel active subscription
   help - Get help and support information
   ```

### 4.2 Prepare Your VPS

**Requirements:**
- Ubuntu 20.04+ or Debian 11+
- At least 1GB RAM
- 10GB+ disk space
- Root or sudo access

**Domain Setup:**
Point your domain to your VPS IP:
```
A record: yourdomain.com → your.vps.ip.address
A record: www.yourdomain.com → your.vps.ip.address
```

### 4.3 One-Command Deployment

**Option A: Direct from GitHub (Recommended)**

```bash
# Set your configuration
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export DOMAIN_NAME="yourdomain.com"
export GITHUB_REPO="yourusername/telegram-subscription-bot"

# Download and run deployment script
curl -sSL "https://raw.githubusercontent.com/$GITHUB_REPO/main/deploy-production.sh" | bash
```

**Option B: Clone and Deploy**

```bash
# Clone your repository
git clone https://github.com/yourusername/telegram-subscription-bot.git
cd telegram-subscription-bot

# Set environment variables
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export DOMAIN_NAME="yourdomain.com"

# Run deployment script
./deploy-production.sh
```

### 4.4 What the Deployment Script Does

The deployment script automatically:

1. **System Setup**:
   - Updates Ubuntu/Debian packages
   - Installs Go, PostgreSQL, Nginx, SSL tools
   - Configures firewall (UFW)

2. **Application Setup**:
   - Clones your GitHub repository
   - Builds the Go application
   - Creates secure database with random passwords
   - Sets up environment configuration

3. **Web Server Configuration**:
   - Configures Nginx as reverse proxy
   - Sets up SSL certificates with Let's Encrypt
   - Enables automatic certificate renewal

4. **System Integration**:
   - Creates systemd service for auto-start
   - Sets up automated backups
   - Creates management scripts

5. **Security Configuration**:
   - Firewall rules (SSH + HTTPS only)
   - Secure headers and SSL settings
   - Database access restrictions

## 🎯 Step 5: Post-Deployment Configuration

### 5.1 Access Your Dashboard

After deployment completes, you'll see output like:
```
🌐 Website URL: https://yourdomain.com
🔐 Admin Login: admin / [generated_password]
```

1. Visit `https://yourdomain.com`
2. Login with admin credentials
3. Change admin password immediately

### 5.2 Configure Payment Providers

In the web dashboard:

1. Go to Settings → Payment Configuration
2. Add your payment provider credentials:
   - **Stripe**: Get keys from [stripe.com](https://stripe.com)
   - **YooMoney**: Get keys from [yoomoney.ru](https://yoomoney.ru)
   - **PayPal**: Get keys from [developer.paypal.com](https://developer.paypal.com)

### 5.3 Test Your Bot

1. Find your bot on Telegram: `@your_bot_username`
2. Send `/start` command
3. Test commands: `/plans`, `/subscribe`, `/help`
4. Verify web dashboard functionality

## 🎯 Step 6: Management and Maintenance

### 6.1 Management Commands

```bash
# Check system status
~/status-bot.sh

# Start/stop/restart bot
~/start-bot.sh
~/stop-bot.sh
~/restart-bot.sh

# Update from GitHub
~/update-bot.sh

# Create backup
~/backup-bot.sh

# Health check
~/health-check.sh
```

### 6.2 View Logs

```bash
# Bot application logs
sudo journalctl -u telegram-bot -f

# Nginx access logs
sudo tail -f /var/log/nginx/access.log

# System resource usage
htop
```

### 6.3 Regular Updates

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Update bot from GitHub
cd ~/telegram-subscription-bot
git pull
go build -o telegram-bot main.go
sudo systemctl restart telegram-bot
```

## 🎯 Step 7: Monitoring and Scaling

### 7.1 Built-in Monitoring

- **Health Check**: `~/health-check.sh`
- **Web Dashboard**: Analytics and user metrics
- **Log Analysis**: Automated error detection

### 7.2 Optional External Monitoring

Consider setting up:
- **Uptime Robot**: Monitor website availability
- **New Relic**: Application performance monitoring
- **Grafana**: Custom dashboards and alerts

### 7.3 Scaling Considerations

For high traffic:
- **Database**: Consider PostgreSQL clustering
- **Load Balancing**: Multiple server instances
- **CDN**: Cloudflare for static assets
- **Caching**: Redis for session management

## 🆘 Troubleshooting

### Common Issues and Solutions

**Bot Not Responding:**
```bash
# Check bot service
sudo systemctl status telegram-bot
sudo journalctl -u telegram-bot -n 50

# Verify bot token
curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"
```

**Website Not Accessible:**
```bash
# Check Nginx status
sudo systemctl status nginx
sudo nginx -t

# Check SSL certificate
sudo certbot certificates

# Test local connection
curl -I http://localhost:5000
```

**Database Connection Issues:**
```bash
# Check PostgreSQL
sudo systemctl status postgresql

# Test database connection
psql "postgresql://telegram_user:password@localhost/telegram_bot" -c "SELECT 1;"
```

**SSL Certificate Problems:**
```bash
# Renew SSL certificate
sudo certbot renew --dry-run

# Check domain DNS
nslookup yourdomain.com
```

## 📞 Support and Community

### Getting Help

1. **Documentation**: Check README.md and DEPLOYMENT-GUIDE.md
2. **Issues**: Create issue on GitHub with details
3. **Logs**: Always include relevant log excerpts
4. **Environment**: Specify OS, Go version, error messages

### Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push branch: `git push origin feature/amazing-feature`
5. Open Pull Request

### Repository Structure

```
telegram-subscription-bot/
├── README.md                 # Main documentation
├── DEPLOYMENT-GUIDE.md       # Detailed deployment instructions
├── GITHUB-SETUP.md          # This file
├── docker-compose.yml       # Docker containerization
├── deploy-production.sh     # VPS deployment automation
├── main.go                  # Application entry point
├── web/                     # Web dashboard
├── handlers/                # Telegram bot handlers
├── database/                # Database schemas
├── scripts/                 # Automation scripts
├── nginx/                   # Nginx configuration
└── locales/                 # Internationalization
```

---

**Congratulations!** 🎉 Your Telegram bot is now on GitHub and ready for professional deployment with automated VPS setup, SSL certificates, and domain binding.

**Next Steps:**
1. Star your repository to bookmark it
2. Share with the community
3. Consider adding more features
4. Set up monitoring and alerts
5. Create documentation for your users