#!/bin/bash
# FreeRADIUS API Setup Script (Go version)
set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
print_status()  { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error()   { echo -e "${RED}[ERROR]${NC} $1"; }
command_exists() { command -v "$1" >/dev/null 2>&1; }

check_requirements() {
    print_status "Checking system requirements..."
    if command_exists go; then
        print_success "Go $(go version | awk '{print $3}') is installed"
    else
        print_error "Go is not installed. Install Go 1.18+ from https://go.dev/dl/"
        exit 1
    fi
    if command_exists mysql; then
        print_success "MySQL client is installed"
    else
        print_warning "MySQL client not found. Please ensure MySQL/MariaDB is installed."
    fi
}

install_dependencies() {
    print_status "Downloading Go modules..."
    go mod download
    print_success "Dependencies ready"
}

build_binary() {
    print_status "Building freeradius-api binary..."
    go build -o freeradius-api .
    print_success "Binary built at ./freeradius-api"
}

setup_environment() {
    print_status "Setting up environment configuration..."
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_success "Environment file created from template"
            if command_exists openssl; then
                JWT_SECRET=$(openssl rand -base64 48 | tr -d '\n')
                # escape & and / for sed
                ESCAPED=$(printf '%s' "$JWT_SECRET" | sed -e 's/[\/&]/\\&/g')
                sed -i "s/your-super-secret-jwt-key-change-this-in-production/$ESCAPED/g" .env
                print_success "JWT secret generated automatically"
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

create_directories() {
    mkdir -p logs
    print_success "Directories created"
}

init_database() {
    print_status "Initializing database..."
    if [ -f "database/init.sql" ]; then
        if [ -f ".env" ]; then
            set -a; . ./.env 2>/dev/null || export $(grep -v '^#' .env | xargs); set +a
        fi
        DB_HOST=${DB_HOST:-localhost}; DB_PORT=${DB_PORT:-3306}
        DB_USER=${DB_USER:-radius}; DB_PASSWORD=${DB_PASSWORD:-radiuspass123!}; DB_NAME=${DB_NAME:-radius}
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < database/init.sql
        print_success "Database initialized successfully"
    else
        print_warning "Database initialization script not found"
    fi
}

create_systemd_service() {
    print_status "Creating systemd service file..."
    CURRENT_DIR=$(pwd)
    SERVICE_CONTENT="[Unit]
Description=FreeRADIUS API Service (Go)
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=$CURRENT_DIR
EnvironmentFile=$CURRENT_DIR/.env
ExecStart=$CURRENT_DIR/freeradius-api
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=freeradius-api

[Install]
WantedBy=multi-user.target"
    echo "$SERVICE_CONTENT" | sudo tee /etc/systemd/system/freeradius-api.service > /dev/null
    sudo systemctl daemon-reload
    sudo systemctl enable freeradius-api.service
    print_success "Service enabled to start on boot"
}

show_usage() {
    echo "FreeRADIUS API Setup Script (Go)"
    echo "Usage: $0 [OPTIONS]"
    echo "  --check-only      Only check system requirements"
    echo "  --install-only    Only download modules"
    echo "  --build-only      Only build the binary"
    echo "  --db-only         Only initialize database"
    echo "  --systemd-only    Only create systemd service (builds first)"
    echo "  --no-start        Don't start the application"
    echo "  --help            Show this help"
}

main() {
    CHECK_ONLY=false; INSTALL_ONLY=false; BUILD_ONLY=false; DB_ONLY=false
    SYSTEMD_ONLY=false; NO_START=false
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check-only) CHECK_ONLY=true; shift ;;
            --install-only) INSTALL_ONLY=true; shift ;;
            --build-only) BUILD_ONLY=true; shift ;;
            --db-only) DB_ONLY=true; shift ;;
            --systemd-only) SYSTEMD_ONLY=true; shift ;;
            --no-start) NO_START=true; shift ;;
            --help) show_usage; exit 0 ;;
            *) print_error "Unknown option: $1"; show_usage; exit 1 ;;
        esac
    done

    if [ "$CHECK_ONLY" = true ]; then check_requirements; exit 0; fi
    if [ "$INSTALL_ONLY" = true ]; then check_requirements; install_dependencies; exit 0; fi
    if [ "$BUILD_ONLY" = true ]; then install_dependencies; build_binary; exit 0; fi
    if [ "$DB_ONLY" = true ]; then init_database; exit 0; fi
    if [ "$SYSTEMD_ONLY" = true ]; then
        install_dependencies; build_binary; create_systemd_service
        sudo systemctl start freeradius-api.service
        print_success "Systemd service started"
        exit 0
    fi

    check_requirements
    install_dependencies
    setup_environment
    create_directories
    build_binary

    if [ "$NO_START" = false ]; then
        if command_exists systemctl; then
            read -p "Run the API as a systemd service? (Y/n): " -n 1 -r; echo
            if [[ -z "$REPLY" ]] || [[ ! $REPLY =~ ^[Nn]$ ]]; then
                create_systemd_service
                sudo systemctl start freeradius-api.service
                print_success "Application started as systemd service"
            else
                print_status "Starting directly (Ctrl+C to stop)..."
                ./freeradius-api
            fi
        else
            ./freeradius-api
        fi
    else
        print_success "Setup completed. Run with: ./freeradius-api"
    fi
}

main "$@"
