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

# Function to check system requirements
check_requirements() {
    print_status "Checking system requirements..."
    
    # Check Node.js
    if command_exists node; then
        NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
        if [ "$NODE_VERSION" -ge 16 ]; then
            print_success "Node.js $(node --version) is installed"
        else
            print_error "Node.js version 16 or higher is required. Current: $(node --version)"
            exit 1
        fi
    else
        print_error "Node.js is not installed. Please install Node.js 16 or higher."
        exit 1
    fi
    
    # Check npm
    if command_exists npm; then
        print_success "npm $(npm --version) is installed"
    else
        print_error "npm is not installed. Please install npm."
        exit 1
    fi
    
    # Check MySQL
    if command_exists mysql; then
        print_success "MySQL client is installed"
    else
        print_warning "MySQL client not found. Please ensure MySQL/MariaDB is installed."
    fi
    
    # Check available memory
    MEMORY_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    MEMORY_MB=$((MEMORY_KB / 1024))
    if [ "$MEMORY_MB" -ge 512 ]; then
        print_success "Memory: ${MEMORY_MB}MB (sufficient)"
    else
        print_warning "Memory: ${MEMORY_MB}MB (recommended: 512MB+)"
    fi
    
    # Check disk space
    DISK_SPACE=$(df -BM . | awk 'NR==2 {print $4}' | sed 's/M//')
    if [ "$DISK_SPACE" -ge 1024 ]; then
        print_success "Disk space: ${DISK_SPACE}MB (sufficient)"
    else
        print_warning "Disk space: ${DISK_SPACE}MB (recommended: 1GB+)"
    fi
}

# Function to install dependencies
install_dependencies() {
    print_status "Installing Node.js dependencies..."
    
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
        # Allow port 3000 for the API
        print_status "Opening port 3000 for FreeRADIUS API..."
        sudo ufw allow 3000/tcp
        
        if [ $? -eq 0 ]; then
            print_success "Port 3000 opened successfully"
        else
            print_warning "Failed to open port 3000. You may need to configure firewall manually"
        fi
    else
        print_warning "UFW not found. Please manually configure firewall to allow port 3000"
        print_status "Manual command: sudo ufw allow 3000/tcp"
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
    
    # Get Node.js path
    NODE_PATH=$(which node)
    
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
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
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

# Function to display usage information
show_usage() {
    echo "FreeRADIUS API Setup Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --check-only      Only check system requirements"
    echo "  --install-only    Only install dependencies"
    echo "  --db-only         Only initialize database"
    echo "  --systemd-only    Only create and start systemd service"
    echo "  --remove-systemd  Remove systemd service"
    echo "  --no-start        Don't start the application"
    echo "  --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Full setup (recommended)"
    echo "  $0 --check-only      # Check requirements only"
    echo "  $0 --systemd-only    # Setup systemd service only"
    echo "  $0 --remove-systemd  # Remove systemd service"
    echo "  $0 --no-start        # Setup without starting"
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
    creat_directories
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
    
    if [ "$NO_START" = false ]; then
        start_application &
        APP_PID=$!
        
        # Run health check
        health_check
        
        # If not using PM2, wait for the application
        if ! command_exists pm2; then
            wait $APP_PID
        fi
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
    echo ""
}

# Run main function
main "$@"