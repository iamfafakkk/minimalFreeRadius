#!/bin/bash

# FreeRADIUS API Setup Script
# This script automates the installation and configuration process

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Node.js and npm using nvm
install_nodejs_with_nvm() {
    print_status "Installing Node.js and npm using nvm..."
    
    # Check if nvm is already installed
    if [ -s "$HOME/.nvm/nvm.sh" ]; then
        print_status "NVM is already installed"
        # Load nvm
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
        [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
    else
        print_status "Installing NVM..."
        # Download and install nvm
        curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
        
        if [ $? -ne 0 ]; then
            print_error "Failed to install NVM"
            return 1
        fi
        
        # Load nvm
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
        [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
        
        print_success "NVM installed successfully"
    fi
    
    # Install latest LTS Node.js
    print_status "Installing latest LTS Node.js..."
    nvm install --lts
    nvm use --lts
    nvm alias default lts/*
    
    if [ $? -eq 0 ]; then
        print_success "Node.js $(node --version) installed successfully"
        print_success "npm $(npm --version) installed successfully"
        
        # Add nvm to shell profile for future sessions
        SHELL_PROFILE=""
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_PROFILE="$HOME/.bashrc"
        elif [ -f "$HOME/.zshrc" ]; then
            SHELL_PROFILE="$HOME/.zshrc"
        elif [ -f "$HOME/.profile" ]; then
            SHELL_PROFILE="$HOME/.profile"
        fi
        
        if [ -n "$SHELL_PROFILE" ]; then
            # Check if nvm is already in the profile
            if ! grep -q "NVM_DIR" "$SHELL_PROFILE"; then
                echo '' >> "$SHELL_PROFILE"
                echo '# NVM Configuration' >> "$SHELL_PROFILE"
                echo 'export NVM_DIR="$HOME/.nvm"' >> "$SHELL_PROFILE"
                echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"' >> "$SHELL_PROFILE"
                echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"' >> "$SHELL_PROFILE"
                print_success "NVM configuration added to $SHELL_PROFILE"
            fi
        fi
        
        return 0
    else
        print_error "Failed to install Node.js with NVM"
        return 1
    fi
}

# Function to check system requirements
check_requirements() {
    print_status "Checking system requirements..."
    
    # Check Node.js and npm
    NODE_NEEDS_INSTALL=false
    NPM_NEEDS_INSTALL=false
    
    if command_exists node; then
        NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
        if [ "$NODE_VERSION" -ge 16 ]; then
            print_success "Node.js $(node --version) is installed"
        else
            print_warning "Node.js version 16 or higher is required. Current: $(node --version)"
            NODE_NEEDS_INSTALL=true
        fi
    else
        print_warning "Node.js is not installed"
        NODE_NEEDS_INSTALL=true
    fi
    
    if command_exists npm; then
        print_success "npm $(npm --version) is installed"
    else
        print_warning "npm is not installed"
        NPM_NEEDS_INSTALL=true
    fi
    
    # Install Node.js and npm if needed
    if [ "$NODE_NEEDS_INSTALL" = true ] || [ "$NPM_NEEDS_INSTALL" = true ]; then
        print_status "Installing Node.js and npm automatically..."
        
        # Check if curl is available for nvm installation
        if ! command_exists curl; then
            print_error "curl is required to install Node.js via NVM. Please install curl first."
            print_status "On macOS: brew install curl"
            print_status "On Ubuntu/Debian: sudo apt-get install curl"
            print_status "On CentOS/RHEL: sudo yum install curl"
            exit 1
        fi
        
        install_nodejs_with_nvm
        if [ $? -ne 0 ]; then
            print_error "Failed to install Node.js and npm automatically"
            print_error "Please install Node.js 16 or higher manually from https://nodejs.org/"
            exit 1
        fi
        
        # Verify installation
        if command_exists node && command_exists npm; then
            NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
            if [ "$NODE_VERSION" -ge 16 ]; then
                print_success "Node.js $(node --version) and npm $(npm --version) installed successfully"
            else
                print_error "Installed Node.js version is still below 16. Please check the installation."
                exit 1
            fi
        else
            print_error "Node.js or npm installation verification failed"
            exit 1
        fi
    fi
    
    # Check MySQL
    if command_exists mysql; then
        print_success "MySQL client is installed"
    else
        print_warning "MySQL client not found. Please ensure MySQL/MariaDB is installed."
    fi
    
    # Check available memory (cross-platform)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        MEMORY_BYTES=$(sysctl -n hw.memsize 2>/dev/null || echo "0")
        MEMORY_MB=$((MEMORY_BYTES / 1024 / 1024))
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if [ -f "/proc/meminfo" ]; then
            MEMORY_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
            MEMORY_MB=$((MEMORY_KB / 1024))
        else
            MEMORY_MB=0
        fi
    else
        MEMORY_MB=0
    fi
    
    if [ "$MEMORY_MB" -gt 0 ]; then
        if [ "$MEMORY_MB" -ge 512 ]; then
            print_success "Memory: ${MEMORY_MB}MB (sufficient)"
        else
            print_warning "Memory: ${MEMORY_MB}MB (recommended: 512MB+)"
        fi
    else
        print_warning "Could not determine available memory"
    fi
    
    # Check disk space (cross-platform)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        DISK_SPACE=$(df -m . | awk 'NR==2 {print $4}')
    else
        # Linux and others
        DISK_SPACE=$(df -BM . | awk 'NR==2 {print $4}' | sed 's/M//')
    fi
    
    if [ "$DISK_SPACE" -ge 1024 ]; then
        print_success "Disk space: ${DISK_SPACE}MB (sufficient)"
    else
        print_warning "Disk space: ${DISK_SPACE}MB (recommended: 1GB+)"
    fi
}

# Function to install dependencies
install_dependencies() {
    print_status "Installing Node.js dependencies..."
    
    # Load nvm if it exists (in case Node.js was just installed)
    if [ -s "$HOME/.nvm/nvm.sh" ]; then
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
        [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
    fi
    
    if [ -f "package.json" ]; then
        npm install
        print_success "Dependencies installed successfully"
    else
        print_error "package.json not found. Are you in the correct directory?"
        exit 1
    fi
}

# Function to setup environment file
setup_environment() {
    print_status "Setting up environment configuration..."
    
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_success "Environment file created from template"
            
            # Generate JWT secret
            if command_exists openssl; then
                JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')
                sed -i "s/your-super-secret-jwt-key-change-this-in-production/$JWT_SECRET/g" .env
                print_success "JWT secret generated automatically"
            else
                print_warning "OpenSSL not found. Please manually set JWT_SECRET in .env file"
            fi
            
            print_warning "Please review and update the .env file with your database credentials"
        else
            print_error ".env.example not found"
            exit 1
        fi
    else
        print_success "Environment file already exists"
    fi
}

# Function to create directories
creat_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p logs
    mkdir -p database
    
    print_success "Directories created"
}

# Function to setup firewall
setup_firewall() {
    print_status "Configuring firewall..."
    
    # Check if ufw is available
    if command_exists ufw; then
        # Allow port 80 and 443 for Nginx
        print_status "Opening ports 80 and 443 for Nginx..."
        sudo ufw allow 80/tcp
        sudo ufw allow 443/tcp
        
        # Allow port 3000 for the API (in case direct access is needed)
        print_status "Opening port 3000 for FreeRADIUS API..."
        sudo ufw allow 3000/tcp
        
        if [ $? -eq 0 ]; then
            print_success "Ports 80, 443, and 3000 opened successfully"
        else
            print_warning "Failed to open ports. You may need to configure firewall manually"
        fi
    else
        print_warning "UFW not found. Please manually configure firewall to allow ports 80, 443, and 3000"
        print_status "Manual commands: sudo ufw allow 80/tcp; sudo ufw allow 443/tcp; sudo ufw allow 3000/tcp"
    fi
}

# Function to test database connection
test_database() {
    print_status "Testing database connection..."
    
    # Source environment variables
    if [ -f ".env" ]; then
        export $(grep -v '^#' .env | xargs)
    fi
    
    # Test connection using Node.js
    node -e "
        const mysql = require('mysql2/promise');
        async function testConnection() {
            try {
                const connection = await mysql.createConnection({
                    host: process.env.DB_HOST || 'localhost',
                    port: process.env.DB_PORT || 3306,
                    user: process.env.DB_USER || 'radius',
                    password: process.env.DB_PASSWORD || 'radiuspass123!',
                    database: process.env.DB_NAME || 'radius'
                });
                await connection.ping();
                console.log('Database connection successful');
                await connection.end();
                process.exit(0);
            } catch (error) {
                console.error('Database connection failed:', error.message);
                process.exit(1);
            }
        }
        testConnection();
    " 2>/dev/null
    
    if [ $? -eq 0 ]; then
        print_success "Database connection test passed"
        return 0
    else
        print_error "Database connection test failed"
        return 1
    fi
}

# Function to initialize database
init_database() {
    print_status "Initializing database..."
    
    if [ -f "database/init.sql" ]; then
        # Source environment variables
        if [ -f ".env" ]; then
            export $(grep -v '^#' .env | xargs)
        fi
        
        DB_HOST=${DB_HOST:-localhost}
        DB_PORT=${DB_PORT:-3306}
        DB_USER=${DB_USER:-radius}
        DB_PASSWORD=${DB_PASSWORD:-radiuspass123!}
        DB_NAME=${DB_NAME:-radius}
        
        print_status "Executing database initialization script..."
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < database/init.sql
        
        if [ $? -eq 0 ]; then
            print_success "Database initialized successfully"
        else
            print_error "Database initialization failed"
            exit 1
        fi
    else
        print_warning "Database initialization script not found"
    fi
}

# Function to remove systemd service
remove_systemd_service() {
    print_status "Removing systemd service..."
    
    # Stop the service if running
    if sudo systemctl is-active --quiet freeradius-api.service; then
        sudo systemctl stop freeradius-api.service
        print_status "Service stopped"
    fi
    
    # Disable the service
    if sudo systemctl is-enabled --quiet freeradius-api.service; then
        sudo systemctl disable freeradius-api.service
        print_status "Service disabled"
    fi
    
    # Remove service file
    if [ -f "/etc/systemd/system/freeradius-api.service" ]; then
        sudo rm /etc/systemd/system/freeradius-api.service
        sudo systemctl daemon-reload
        print_success "Systemd service removed successfully"
    else
        print_warning "Service file not found"
    fi
}

# Function to create systemd service file
create_systemd_service() {
    print_status "Creating systemd service file..."
    
    # Load nvm if it exists (in case Node.js was installed via nvm)
    if [ -s "$HOME/.nvm/nvm.sh" ]; then
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
        [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
    fi
    
    # Ensure dependencies are installed
    if [ ! -d "node_modules" ]; then
        print_status "Installing dependencies first..."
        npm install
        if [ $? -ne 0 ]; then
            print_error "Failed to install dependencies"
            return 1
        fi
    fi
    
    # Get current directory
    CURRENT_DIR=$(pwd)
    
    # Source environment variables
    if [ -f ".env" ]; then
        export $(grep -v '^#' .env | xargs)
    fi
    
    # Get Node.js path (prefer nvm version if available)
    if [ -s "$HOME/.nvm/nvm.sh" ] && command -v nvm >/dev/null 2>&1; then
        NODE_PATH=$(nvm which node 2>/dev/null || which node)
    else
        NODE_PATH=$(which node)
    fi
    
    # Verify Node.js path exists
    if [ ! -f "$NODE_PATH" ]; then
        print_error "Node.js executable not found at $NODE_PATH"
        return 1
    fi
    
    # Create service file content
    SERVICE_CONTENT="[Unit]
Description=FreeRADIUS API Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=$CURRENT_DIR
EnvironmentFile=$CURRENT_DIR/.env
ExecStart=$NODE_PATH $CURRENT_DIR/server.js
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=freeradius-api

[Install]
WantedBy=multi-user.target"
    
    # Write service file
    echo "$SERVICE_CONTENT" | sudo tee /etc/systemd/system/freeradius-api.service > /dev/null
    
    if [ $? -eq 0 ]; then
        print_success "Systemd service file created at /etc/systemd/system/freeradius-api.service"
        
        # Reload systemd daemon
        sudo systemctl daemon-reload
        
        # Enable service to start on boot
        sudo systemctl enable freeradius-api.service
        
        print_success "Service enabled to start on boot"
    else
        print_error "Failed to create systemd service file"
        return 1
    fi
}

# Function to start application
start_application() {
    print_status "Starting application..."
    
    # Check if systemctl is available and user wants to use it
    if command_exists systemctl; then
        read -p "Do you want to run the API as a systemd service? (Y/n): " -n 1 -r
        echo
        # If user just presses Enter, default to Y
        if [[ -z "$REPLY" ]] || [[ ! $REPLY =~ ^[Nn]$ ]]; then
            create_systemd_service
            if [ $? -eq 0 ]; then
                sudo systemctl start freeradius-api.service
                if [ $? -eq 0 ]; then
                    print_success "Application started as systemd service"
                    print_status "Use 'sudo systemctl status freeradius-api' to check status"
                    print_status "Use 'sudo systemctl stop freeradius-api' to stop the service"
                    print_status "Use 'sudo systemctl restart freeradius-api' to restart the service"
                    print_status "Use 'journalctl -u freeradius-api -f' to view logs"
                    return 0
                else
                    print_error "Failed to start systemd service"
                fi
            fi
        fi
    fi
    
    # Fallback to PM2 or development mode
    if command_exists pm2; then
        print_status "Starting with PM2..."
        pm2 start server.js --name freeradius-api
        print_success "Application started with PM2"
        print_status "Use 'pm2 logs freeradius-api' to view logs"
        print_status "Use 'pm2 stop freeradius-api' to stop the application"
    else
        print_status "PM2 not found. Starting in development mode..."
        print_status "Use Ctrl+C to stop the application"
        npm run dev
    fi
}

# Function to run health check
health_check() {
    print_status "Running health check..."
    
    # Wait a moment for the server to start
    sleep 3
    
    # Source environment variables
    if [ -f ".env" ]; then
        export $(grep -v '^#' .env | xargs)
    fi
    
    PORT=${PORT:-3000}
    
    if command_exists curl; then
        RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/api/v1/auth/health" || echo "000")
        
        if [ "$RESPONSE" = "200" ]; then
            print_success "Health check passed - API is running on port $PORT"
            print_success "API URL: http://localhost:$PORT/api/v1"
        else
            print_error "Health check failed - HTTP $RESPONSE"
        fi
    else
        print_warning "curl not found. Cannot perform health check"
        print_status "Please manually check: http://localhost:$PORT/api/v1/auth/health"
    fi
}

# Function to validate domain name
validate_domain() {
    local domain=$1
    
    # Check if it's localhost or IP address
    if [[ "$domain" == "localhost" ]] || [[ "$domain" == "127.0.0.1" ]]; then
        return 0
    fi
    
    # Simple domain validation regex
    # Allows alphanumeric characters, hyphens, and dots
    # Must have at least one dot and valid TLD
    if [[ "$domain" =~ ^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$ ]]; then
        return 0
    else
        return 1
    fi
}

# Function to get domain input from user
get_domain_input() {
    echo
    print_status "Nginx Domain Configuration"
    print_status "=========================="
    
    # Ask for domain name
    read -p "Enter your domain name (e.g., example.com) [localhost]: " -r DOMAIN_NAME
    echo
    
    # Use localhost as default if no input
    if [[ -z "$DOMAIN_NAME" ]]; then
        DOMAIN_NAME="localhost"
    fi
    
    # Validate domain format
    if ! validate_domain "$DOMAIN_NAME"; then
        print_warning "Invalid domain format. Using localhost as default."
        DOMAIN_NAME="localhost"
    fi
    
    print_success "Domain set to: $DOMAIN_NAME"
    export DOMAIN_NAME
}

# Function to get Cloudflare SSL mode
get_cloudflare_ssl_mode() {
    echo
    print_status "Cloudflare SSL Configuration"
    print_status "============================"
    echo "Cloudflare offers several SSL modes:"
    echo "1. Flexible SSL: Traffic is encrypted between visitor and Cloudflare, but not between Cloudflare and your server"
    echo "2. Full SSL: Traffic is encrypted between visitor and Cloudflare, and between Cloudflare and your server (self-signed certificate)"
    echo "3. Full SSL (Strict): Traffic is encrypted between visitor and Cloudflare, and between Cloudflare and your server (CA-signed or Cloudflare Origin CA certificate)"
    echo
    echo "For production environments, we recommend 'Full SSL (Strict)' for maximum security."
    echo
    read -p "Select SSL mode (1: Flexible, 2: Full, 3: Full Strict) [3]: " -r SSL_MODE
    echo
    
    # Use Full SSL (Strict) as default if no input
    if [[ -z "$SSL_MODE" ]]; then
        SSL_MODE=3
    fi
    
    # Validate input
    if [[ ! "$SSL_MODE" =~ ^[1-3]$ ]]; then
        print_warning "Invalid SSL mode. Using Full SSL (Strict) as default."
        SSL_MODE=3
    fi
    
    case $SSL_MODE in
        1)
            print_success "Cloudflare SSL mode set to: Flexible"
            ;;
        2)
            print_success "Cloudflare SSL mode set to: Full"
            ;;
        3)
            print_success "Cloudflare SSL mode set to: Full (Strict)"
            ;;
    esac
    
    export SSL_MODE
}

# Function to install and configure Nginx
install_nginx() {
    print_status "Installing and configuring Nginx..."
    
    # Detect OS and install Nginx accordingly
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists brew; then
            print_status "Installing Nginx via Homebrew..."
            brew install nginx
        else
            print_warning "Homebrew not found. Please install Homebrew first: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
            return 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux - detect distribution
        if command_exists apt-get; then
            # Debian/Ubuntu
            print_status "Installing Nginx via apt-get..."
            sudo apt-get update
            sudo apt-get install -y nginx
        elif command_exists yum; then
            # CentOS/RHEL
            print_status "Installing Nginx via yum..."
            sudo yum install -y nginx
        elif command_exists dnf; then
            # Fedora
            print_status "Installing Nginx via dnf..."
            sudo dnf install -y nginx
        else
            print_warning "Package manager not found. Please install Nginx manually."
            return 1
        fi
    else
        print_warning "Unsupported OS. Please install Nginx manually."
        return 1
    fi
    
    if [ $? -eq 0 ]; then
        print_success "Nginx installed successfully"
    else
        print_error "Failed to install Nginx"
        return 1
    fi
}

# Function to configure Nginx as reverse proxy
configure_nginx() {
    print_status "Configuring Nginx as reverse proxy..."
    
    # Backup existing nginx.conf if it exists
    if [ -f "/etc/nginx/nginx.conf" ]; then
        print_status "Backing up existing nginx.conf..."
        sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup.$(date +%Y%m%d_%H%M%S)
    fi
    
    # Copy our nginx configuration
    print_status "Copying Nginx configuration..."
    sudo cp nginx/nginx.conf /etc/nginx/nginx.conf
    
    # Replace placeholder with actual domain
    print_status "Updating domain configuration to ${DOMAIN_NAME}..."
    sudo sed -i "s/{{DOMAIN_NAME}}/${DOMAIN_NAME}/g" /etc/nginx/nginx.conf
    
    # Create SSL directory if it doesn't exist
    sudo mkdir -p /etc/nginx/ssl
    
    # For Cloudflare Flexible SSL, no certificate is needed on the origin server
    # For Full SSL or Full SSL (Strict), we need to generate or provide certificates
    if [ "$SSL_MODE" -eq 2 ] || [ "$SSL_MODE" -eq 3 ]; then
        # Create self-signed certificate for the domain (in production, user should replace with real certificate)
        if command_exists openssl; then
            print_status "Generating self-signed SSL certificate for ${DOMAIN_NAME}..."
            sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout /etc/nginx/ssl/server.key \
                -out /etc/nginx/ssl/server.crt \
                -subj "/C=US/ST=State/L=City/O=Organization/CN=${DOMAIN_NAME}"
            
            if [ $? -eq 0 ]; then
                print_success "Self-signed SSL certificate created for ${DOMAIN_NAME}"
                if [ "$SSL_MODE" -eq 3 ]; then
                    print_warning "For Full SSL (Strict), replace the self-signed certificate with a CA-signed or Cloudflare Origin CA certificate"
                fi
            else
                print_warning "Failed to create self-signed certificate"
            fi
        else
            print_warning "OpenSSL not found. Please manually configure SSL certificates"
        fi
    else
        print_status "Using Cloudflare Flexible SSL - no origin certificate required"
    fi
    
    # Test Nginx configuration
    print_status "Testing Nginx configuration..."
    sudo nginx -t
    
    if [ $? -eq 0 ]; then
        print_success "Nginx configuration test passed"
        return 0
    else
        print_error "Nginx configuration test failed"
        return 1
    fi
}

# Function to start Nginx service
start_nginx() {
    print_status "Starting Nginx service..."
    
    # Detect OS and start Nginx accordingly
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS - Homebrew installation
        if command_exists brew; then
            print_status "Starting Nginx via Homebrew services..."
            brew services start nginx
        else
            # Try to start directly
            sudo nginx
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux - use systemctl
        if command_exists systemctl; then
            print_status "Starting Nginx via systemctl..."
            sudo systemctl start nginx
            sudo systemctl enable nginx
        else
            # Try to start directly
            sudo nginx
        fi
    else
        # Try to start directly
        sudo nginx
    fi
    
    if [ $? -eq 0 ]; then
        print_success "Nginx started successfully"
        return 0
    else
        print_error "Failed to start Nginx"
        return 1
    fi
}

# Function to display usage information
show_usage() {
    echo "FreeRADIUS API Setup Script"
    echo ""
    echo "This script automatically sets up the FreeRADIUS API environment."
    echo "It will automatically install Node.js and npm using NVM if not available."
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --check-only      Only check system requirements"
    echo "  --install-only    Only install dependencies"
    echo "  --db-only         Only initialize database"
    echo "  --systemd-only    Only create and start systemd service"
    echo "  --nginx-only      Only install and configure Nginx"
    echo "  --remove-systemd  Remove systemd service"
    echo "  --no-start        Don't start the application"
    echo "  --help           Show this help message"
    echo ""
    echo "Features:"
    echo "  • Automatic Node.js and npm installation via NVM"
    echo "  • System requirements checking"
    echo "  • Environment configuration"
    echo "  • Database initialization"
    echo "  • Systemd service creation"
    echo "  • Nginx reverse proxy configuration"
    echo "  • Firewall configuration"
    echo ""
    echo "Requirements:"
    echo "  • curl (for Node.js installation)"
    echo "  • MySQL/MariaDB server"
    echo "  • 512MB+ RAM (recommended)"
    echo "  • 1GB+ disk space (recommended)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Full setup (recommended)"
    echo "  $0 --check-only      # Check requirements only"
    echo "  $0 --systemd-only    # Setup systemd service only"
    echo "  $0 --nginx-only      # Setup Nginx reverse proxy only"
    echo "  $0 --remove-systemd  # Remove systemd service"
    echo "  $0 --no-start        # Setup without starting"
}

# Function to start application in fallback mode (PM2 or development)
start_application_fallback() {
    # Fallback to PM2 or development mode
    if command_exists pm2; then
        print_status "Starting with PM2..."
        pm2 start server.js --name freeradius-api
        print_success "Application started with PM2"
        print_status "Use 'pm2 logs freeradius-api' to view logs"
        print_status "Use 'pm2 stop freeradius-api' to stop the application"
        return 0
    else
        print_status "PM2 not found. Starting in development mode..."
        print_status "Use Ctrl+C to stop the application"
        npm run dev
        return 0
    fi
}

# Main setup function
main() {
    echo "======================================"
    echo "    FreeRADIUS API Setup Script"
    echo "======================================"
    echo ""
    
    # Parse command line arguments
    CHECK_ONLY=false
    INSTALL_ONLY=false
    DB_ONLY=false
    SYSTEMD_ONLY=false
    NGINX_ONLY=false
    REMOVE_SYSTEMD=false
    NO_START=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check-only)
                CHECK_ONLY=true
                shift
                ;;
            --install-only)
                INSTALL_ONLY=true
                shift
                ;;
            --db-only)
                DB_ONLY=true
                shift
                ;;
            --systemd-only)
                SYSTEMD_ONLY=true
                shift
                ;;
            --nginx-only)
                NGINX_ONLY=true
                shift
                ;;
            --remove-systemd)
                REMOVE_SYSTEMD=true
                shift
                ;;
            --no-start)
                NO_START=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Execute based on options
    if [ "$CHECK_ONLY" = true ]; then
        check_requirements
        exit 0
    fi
    
    if [ "$INSTALL_ONLY" = true ]; then
        check_requirements
        install_dependencies
        exit 0
    fi
    
    if [ "$DB_ONLY" = true ]; then
        init_database
        exit 0
    fi
    
    if [ "$NGINX_ONLY" = true ]; then
        # Get domain input from user
        get_domain_input
        # Get Cloudflare SSL mode
        get_cloudflare_ssl_mode
        
        install_nginx
        if [ $? -eq 0 ]; then
            configure_nginx
            if [ $? -eq 0 ]; then
                start_nginx
                if [ $? -eq 0 ]; then
                    print_success "Nginx reverse proxy configured and started successfully"
                    print_status "Nginx is now serving as a reverse proxy for the FreeRADIUS API"
                    print_status "Access the API via: http://${DOMAIN_NAME} or https://${DOMAIN_NAME}"
                    
                    # Provide Cloudflare configuration guidance
                    echo
                    print_status "Cloudflare Configuration Guide:"
                    case $SSL_MODE in
                        1)
                            echo "1. In your Cloudflare dashboard, go to SSL/TLS > Overview"
                            echo "2. Set SSL/TLS encryption mode to 'Flexible'"
                            echo "3. No additional origin certificate configuration needed"
                            ;;
                        2)
                            echo "1. In your Cloudflare dashboard, go to SSL/TLS > Overview"
                            echo "2. Set SSL/TLS encryption mode to 'Full'"
                            echo "3. Your origin server is using a self-signed certificate"
                            echo "4. For production, replace with a valid certificate"
                            ;;
                        3)
                            echo "1. In your Cloudflare dashboard, go to SSL/TLS > Overview"
                            echo "2. Set SSL/TLS encryption mode to 'Full (Strict)'"
                            echo "3. Your origin server needs a valid certificate (CA-signed or Cloudflare Origin CA)"
                            echo "4. For production, obtain a certificate from Cloudflare Origin CA or a CA"
                            ;;
                    esac
                else
                    print_error "Failed to start Nginx"
                    exit 1
                fi
            else
                print_error "Failed to configure Nginx"
                exit 1
            fi
        else
            print_error "Failed to install Nginx"
            exit 1
        fi
        exit 0
    fi
    
    if [ "$SYSTEMD_ONLY" = true ]; then
        if command_exists systemctl; then
            create_systemd_service
            if [ $? -eq 0 ]; then
                sudo systemctl start freeradius-api.service
                if [ $? -eq 0 ]; then
                    print_success "Systemd service created and started successfully"
                    print_status "Use 'sudo systemctl status freeradius-api' to check status"
                    print_status "Use 'journalctl -u freeradius-api -f' to view logs"
                else
                    print_error "Failed to start systemd service"
                    exit 1
                fi
            else
                exit 1
            fi
        else
            print_error "systemctl not available on this system"
            exit 1
        fi
        exit 0
    fi
    
    if [ "$REMOVE_SYSTEMD" = true ]; then
        if command_exists systemctl; then
            remove_systemd_service
        else
            print_error "systemctl not available on this system"
            exit 1
        fi
        exit 0
    fi
    
    # Full setup
    check_requirements
    install_dependencies
    setup_environment
    create_directories
    setup_firewall
    
    # Test database connection
    if test_database; then
        read -p "Do you want to initialize the database with sample data? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            init_database
        fi
    else
        print_warning "Skipping database initialization due to connection failure"
        print_warning "Please check your database configuration in .env file"
    fi
    
    # Install and configure Nginx
    read -p "Do you want to install and configure Nginx as a reverse proxy? (Y/n): " -n 1 -r
    echo
    if [[ -z "$REPLY" ]] || [[ ! $REPLY =~ ^[Nn]$ ]]; then
        # Get domain input from user
        get_domain_input
        # Get Cloudflare SSL mode
        get_cloudflare_ssl_mode
        
        install_nginx
        if [ $? -eq 0 ]; then
            configure_nginx
            if [ $? -ne 0 ]; then
                print_error "Failed to configure Nginx"
            fi
        else
            print_error "Failed to install Nginx"
        fi
    fi

    if [ "$NO_START" = false ]; then
        # Start Nginx if it was installed
        if command_exists nginx && [[ (-z "$REPLY") || (! $REPLY =~ ^[Nn]$) ]]; then
            start_nginx
        fi
        
        # Check if systemctl is available and user wants systemd service
        if command_exists systemctl; then
            read -p "Do you want to run the API as a systemd service? (Y/n): " -n 1 -r
            echo
            # If user just presses Enter, default to Y
            if [[ -z "$REPLY" ]] || [[ ! $REPLY =~ ^[Nn]$ ]]; then
                create_systemd_service
                if [ $? -eq 0 ]; then
                    sudo systemctl start freeradius-api.service
                    if [ $? -eq 0 ]; then
                        print_success "Application started as systemd service"
                        print_status "Use 'sudo systemctl status freeradius-api' to check status"
                        print_status "Use 'sudo systemctl stop freeradius-api' to stop the service"
                        print_status "Use 'sudo systemctl restart freeradius-api' to restart the service"
                        print_status "Use 'journalctl -u freeradius-api -f' to view logs"
                    else
                        print_error "Failed to start systemd service"
                        # Fall back to PM2 or development mode
                        start_application_fallback
                    fi
                else
                    print_error "Failed to create systemd service"
                    # Fall back to PM2 or development mode
                    start_application_fallback
                fi
            else
                # User chose not to use systemd, fall back to PM2 or development mode
                start_application_fallback
            fi
        else
            # systemctl not available, fall back to PM2 or development mode
            start_application_fallback
        fi
        
        # Run health check
        health_check
    else
        print_success "Setup completed successfully!"
        print_status "To start the application manually, run: npm start"
    fi
    
    echo ""
    print_success "Setup completed!"
    echo ""
    echo "Next steps:"
    echo "1. Review the .env file and update database credentials if needed"
    echo "2. Check the API documentation in docs/API_DOCUMENTATION.md"
    echo "3. Test the API endpoints using the examples in the documentation"
    echo ""
    echo "Useful commands:"
    echo "  npm start                              # Start in production mode"
    echo "  npm run dev                            # Start in development mode"
    echo "  pm2 logs                               # View PM2 logs (if using PM2)"
    echo "  pm2 stop all                           # Stop all PM2 processes"
    echo "  sudo systemctl status freeradius-api   # Check systemd service status"
    echo "  sudo systemctl start freeradius-api    # Start systemd service"
    echo "  sudo systemctl stop freeradius-api     # Stop systemd service"
    echo "  sudo systemctl restart freeradius-api  # Restart systemd service"
    echo "  journalctl -u freeradius-api -f        # View systemd service logs"
    echo "  sudo systemctl status nginx            # Check Nginx service status"
    echo "  sudo systemctl start nginx             # Start Nginx service"
    echo "  sudo nginx -t                          # Test Nginx configuration"
    echo "  sudo nginx -s reload                   # Reload Nginx configuration"
    echo ""
}

# Run main function
main "$@"