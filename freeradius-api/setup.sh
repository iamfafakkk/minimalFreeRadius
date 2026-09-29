#!/bin/bash
# FreeRADIUS API Setup Script (Go) — systemd-based.
#
# Prasyarat: jalankan install.sh (FreeRADIUS + MySQL + schema) lebih dulu,
# lalu ./setup.sh di direktori freeradius-api.
set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
print_status()  { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error()   { echo -e "${RED}[ERROR]${NC} $1"; }
command_exists() { command -v "$1" >/dev/null 2>&1; }

SERVICE_NAME="freeradius-api"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
ENV_FILE="$(pwd)/.env"

# systemd, bukan opsi. Script berhenti kalau tidak ada systemctl.
require_root() {
    if [ "$(id -u)" -ne 0 ]; then
        print_error "Jalankan sebagai root: sudo ./setup.sh"
        exit 1
    fi
}

require_systemd() {
    if ! command_exists systemctl; then
        print_error "systemd (systemctl) tidak ditemukan. Script ini hanya mendukung systemd."
        exit 1
    fi
}

check_requirements() {
    print_status "Checking system requirements..."
    if command_exists go; then
        print_success "Go $(go version | awk '{print $3}') is installed"
    else
        print_error "Go is not installed. Install Go 1.18+ from https://go.dev/dl/"
        exit 1
    fi
    # CGO diperlukan: driver SQLite (mattn/go-sqlite3) butuh gcc + libc headers.
    if command_exists gcc; then
        print_success "C compiler (gcc) is installed"
    else
        print_error "gcc is required (CGO for the SQLite driver). Install: apt-get install -y build-essential"
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
    CGO_ENABLED=1 go build -o freeradius-api .
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

# Skema RADIUS dibuat oleh install.sh (schema resmi FreeRADIUS MySQL).
# setup.sh TIDAK membuat/menghapus tabel — hanya memverifikasi.
verify_database() {
    print_status "Verifying FreeRADIUS database schema (created by install.sh)..."
    if [ -f ".env" ]; then
        set -a; . ./.env 2>/dev/null || export $(grep -v '^#' .env | xargs); set +a
    fi
    DB_HOST=${DB_HOST:-localhost}; DB_PORT=${DB_PORT:-3306}
    DB_USER=${DB_USER:-radius}; DB_PASSWORD=${DB_PASSWORD:-radiuspass123!}; DB_NAME=${DB_NAME:-radius}

    if ! command_exists mysql; then
        print_warning "mysql client not found, skipping schema check"
        return 0
    fi
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" \
        -e "SELECT 1 FROM radcheck LIMIT 1;" >/dev/null 2>&1; then
        print_success "FreeRADIUS schema present (radcheck accessible)"
    else
        print_warning "Tidak bisa memverifikasi tabel radcheck. Pastikan install.sh sudah dijalankan."
    fi
}

# Frontend di-serve backend dari direktori build (adapter-static). Cari
# ../freeradius-web dan build (butuh npm) bila belum ada.
build_frontend() {
    print_status "Preparing web panel (static build)..."
    local web_dir
    web_dir="$(cd "$(dirname "$(pwd)")" 2>/dev/null && pwd)/freeradius-web"
    if [ ! -d "$web_dir" ]; then
        print_warning "freeradius-web/ tidak ditemukan — panel akan jalan API-only."
        return 0
    fi
    if [ -f "$web_dir/build/index.html" ]; then
        print_success "Web build sudah ada ($web_dir/build)"
        return 0
    fi
    if ! command_exists npm; then
        print_warning "npm tidak ditemukan — build frontend dilewati. Jalankan 'npm run build' di freeradius-web."
        return 0
    fi
    print_status "Building frontend (npm install && npm run build)..."
    ( cd "$web_dir" && npm install --no-audit --no-fund && npm run build )
    print_success "Frontend built at $web_dir/build"
}

# UFW aktif (dari install.sh) default-deny, jadi port API perlu dibuka.
configure_firewall() {
    if command_exists ufw && ufw status 2>/dev/null | grep -q "Status: active"; then
        local port
        port=$(grep -E '^PORT=' .env 2>/dev/null | cut -d= -f2 | tr -d ' ')
        port=${port:-3000}
        ufw allow "${port}/tcp" comment 'FreeRADIUS API' >/dev/null 2>&1 || true
        print_success "UFW: port ${port}/tcp diizinkan"
    else
        print_status "UFW tidak aktif, melewati konfigurasi firewall"
    fi
}

# Service selalu root: tombol "Restart FreeRADIUS" dan "Clear Log File" di panel
# menjalankan `systemctl restart freeradius` dan menulis ke /var/log/freeradius,
# yang keduanya butuh root. Menurunkan ke freerad akan mematikan kedua fitur itu.
# ponytail: root saja; pakai sudoers NOPASSWD + izin log bila kelak butuh
# service non-root.
service_user() {
    echo "root"
}

create_systemd_service() {
    print_status "Creating systemd service file..."
    local current_dir service_user
    current_dir=$(pwd)
    service_user=$(service_user)
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=FreeRADIUS API Service (Go)
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=${service_user}
WorkingDirectory=${current_dir}
EnvironmentFile=${ENV_FILE}
ExecStart=${current_dir}/freeradius-api
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${SERVICE_NAME}

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME.service"
    print_success "Service enabled to start on boot (User=${service_user})"
}

start_service() {
    print_status "Restarting ${SERVICE_NAME} service..."
    systemctl restart "$SERVICE_NAME.service"
    sleep 1
    if systemctl is-active --quiet "$SERVICE_NAME.service"; then
        print_success "Service active"
    else
        print_error "Service gagal start. Cek: journalctl -u ${SERVICE_NAME} -n 50"
        exit 1
    fi
}

remove_systemd_service() {
    print_status "Removing systemd service..."
    systemctl stop "$SERVICE_NAME.service" 2>/dev/null || true
    systemctl disable "$SERVICE_NAME.service" 2>/dev/null || true
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
    systemctl reset-failed "$SERVICE_NAME.service" 2>/dev/null || true
    print_success "Service removed"
}

health_check() {
    local port
    port=$(grep -E '^PORT=' .env 2>/dev/null | cut -d= -f2 | tr -d ' ')
    port=${port:-3000}
    if curl -fsS "http://localhost:${port}/api/v1/auth/health" >/dev/null 2>&1; then
        print_success "Health check OK (http://localhost:${port}/api/v1/auth/health)"
    else
        print_warning "Health check gagal; cek: journalctl -u ${SERVICE_NAME} -n 50"
    fi
}

show_usage() {
    echo "FreeRADIUS API Setup Script (Go, systemd)"
    echo "Usage: $0 [OPTIONS]"
    echo "  --check-only      Only check system requirements"
    echo "  --install-only    Only download modules"
    echo "  --build-only      Only build the binary"
    echo "  --db-only         Only verify the FreeRADIUS database schema"
    echo "  --systemd-only    Only install the systemd service (builds first)"
    echo "  --remove-systemd  Stop, disable and remove the systemd service"
    echo "  --no-start        Install the service but don't start it"
    echo "  --help            Show this help"
}

main() {
    CHECK_ONLY=false; INSTALL_ONLY=false; BUILD_ONLY=false; DB_ONLY=false
    SYSTEMD_ONLY=false; REMOVE_SYSTEMD=false; NO_START=false
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check-only) CHECK_ONLY=true; shift ;;
            --install-only) INSTALL_ONLY=true; shift ;;
            --build-only) BUILD_ONLY=true; shift ;;
            --db-only) DB_ONLY=true; shift ;;
            --systemd-only) SYSTEMD_ONLY=true; shift ;;
            --remove-systemd) REMOVE_SYSTEMD=true; shift ;;
            --no-start) NO_START=true; shift ;;
            --help) show_usage; exit 0 ;;
            *) print_error "Unknown option: $1"; show_usage; exit 1 ;;
        esac
    done

    if [ "$CHECK_ONLY" = true ]; then check_requirements; exit 0; fi
    if [ "$INSTALL_ONLY" = true ]; then check_requirements; install_dependencies; exit 0; fi
    if [ "$BUILD_ONLY" = true ]; then install_dependencies; build_binary; exit 0; fi
    if [ "$DB_ONLY" = true ]; then verify_database; exit 0; fi
    if [ "$REMOVE_SYSTEMD" = true ]; then require_root; require_systemd; remove_systemd_service; exit 0; fi
    if [ "$SYSTEMD_ONLY" = true ]; then
        require_root; require_systemd
        install_dependencies; build_binary
        create_systemd_service
        systemctl start "$SERVICE_NAME.service"
        print_success "Systemd service started"
        exit 0
    fi

    require_root
    require_systemd
    check_requirements
    install_dependencies
    setup_environment
    create_directories
    build_binary
    verify_database
    build_frontend
    configure_firewall
    create_systemd_service

    if [ "$NO_START" = true ]; then
        print_success "Setup completed. Service installed but not started."
    else
        start_service
        health_check
    fi
}

main "$@"
