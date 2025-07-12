#!/bin/bash

# Dependencies installation script for different operating systems
# Supports Ubuntu, Debian, CentOS, and macOS

set -e

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

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [[ -f /etc/os-release ]]; then
            . /etc/os-release
            OS=$ID
            VER=$VERSION_ID
        else
            error "Cannot detect Linux distribution"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        VER=$(sw_vers -productVersion)
    else
        error "Unsupported operating system: $OSTYPE"
    fi
    
    log "Detected OS: $OS $VER"
}

# Install Go
install_go() {
    log "Installing Go..."
    
    GO_VERSION="1.21.5"
    
    if command -v go &> /dev/null; then
        CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        if [[ "$CURRENT_GO_VERSION" == "$GO_VERSION" ]]; then
            log "Go $GO_VERSION is already installed"
            return
        fi
    fi
    
    case $OS in
        ubuntu|debian)
            # Remove old Go installation
            sudo rm -rf /usr/local/go
            
            # Download and install Go
            cd /tmp
            wget -q "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz"
            sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
            
            # Add to PATH
            if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
            fi
            export PATH=$PATH:/usr/local/go/bin
            ;;
        centos|rhel|fedora)
            sudo rm -rf /usr/local/go
            cd /tmp
            wget -q "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz"
            sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
            
            if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
            fi
            export PATH=$PATH:/usr/local/go/bin
            ;;
        macos)
            if command -v brew &> /dev/null; then
                brew install go
            else
                warn "Homebrew not found. Please install Go manually from https://golang.org/dl/"
            fi
            ;;
    esac
    
    # Verify installation
    if go version &> /dev/null; then
        log "Go installed successfully: $(go version)"
    else
        error "Go installation failed"
    fi
}

# Install PostgreSQL
install_postgresql() {
    log "Installing PostgreSQL..."
    
    case $OS in
        ubuntu|debian)
            sudo apt update
            sudo apt install -y postgresql postgresql-contrib
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            ;;
        centos|rhel)
            sudo yum install -y postgresql postgresql-server postgresql-contrib
            sudo postgresql-setup initdb
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            ;;
        fedora)
            sudo dnf install -y postgresql postgresql-server postgresql-contrib
            sudo postgresql-setup --initdb
            sudo systemctl start postgresql
            sudo systemctl enable postgresql
            ;;
        macos)
            if command -v brew &> /dev/null; then
                brew install postgresql
                brew services start postgresql
            else
                warn "Homebrew not found. Please install PostgreSQL manually"
            fi
            ;;
    esac
    
    log "PostgreSQL installed successfully"
}

# Install Nginx (for production)
install_nginx() {
    log "Installing Nginx..."
    
    case $OS in
        ubuntu|debian)
            sudo apt install -y nginx
            sudo systemctl enable nginx
            ;;
        centos|rhel|fedora)
            sudo yum install -y nginx || sudo dnf install -y nginx
            sudo systemctl enable nginx
            ;;
        macos)
            if command -v brew &> /dev/null; then
                brew install nginx
            else
                warn "Homebrew not found. Nginx installation skipped"
            fi
            ;;
    esac
    
    log "Nginx installed successfully"
}

# Install additional tools
install_tools() {
    log "Installing additional tools..."
    
    case $OS in
        ubuntu|debian)
            sudo apt install -y curl wget git htop unzip certbot python3-certbot-nginx ufw
            ;;
        centos|rhel)
            sudo yum install -y curl wget git htop unzip epel-release
            sudo yum install -y certbot python3-certbot-nginx
            ;;
        fedora)
            sudo dnf install -y curl wget git htop unzip certbot python3-certbot-nginx
            ;;
        macos)
            if command -v brew &> /dev/null; then
                brew install curl wget git htop
            fi
            ;;
    esac
    
    log "Additional tools installed successfully"
}

# Main installation process
main() {
    log "Starting dependency installation..."
    
    detect_os
    
    # Update package manager
    case $OS in
        ubuntu|debian)
            sudo apt update
            ;;
        centos|rhel)
            sudo yum update -y
            ;;
        fedora)
            sudo dnf update -y
            ;;
        macos)
            if ! command -v brew &> /dev/null; then
                log "Installing Homebrew..."
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            fi
            brew update
            ;;
    esac
    
    # Install dependencies
    install_go
    install_postgresql
    install_nginx
    install_tools
    
    log "All dependencies installed successfully!"
    log "You may need to restart your shell or run 'source ~/.bashrc' to update PATH"
    
    # Display versions
    echo ""
    echo "Installed versions:"
    go version 2>/dev/null || echo "Go: Not found in PATH"
    psql --version 2>/dev/null || echo "PostgreSQL: Installed but psql not in PATH"
    nginx -v 2>&1 | head -1 || echo "Nginx: Not installed or not in PATH"
}

# Run main function
main "$@"