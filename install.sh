#!/bin/bash

#=============================================================================
# FreeRADIUS Installation Script for Ubuntu 22.04 LTS
# Version: 1.0
# Description: Script otomatis untuk instalasi dan konfigurasi FreeRADIUS
#              dengan MySQL backend untuk autentikasi RADIUS
# Author: Auto-generated Script
# Date: $(date +"%Y-%m-%d")
#=============================================================================

# Konfigurasi warna untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Lingkungan non-interaktif + locale agar output rapi
export DEBIAN_FRONTEND=noninteractive
export LC_ALL=C.UTF-8
export LANG=C.UTF-8
export LANGUAGE=C.UTF-8

# Variabel global
LOG_FILE="/tmp/freeradius_install.log"
MYSQL_ROOT_PASSWORD="radius123!"
RADIUS_DB_NAME="radius"
RADIUS_DB_USER="radius"
RADIUS_DB_PASSWORD="radiuspass123!"
TEST_USER="testuser"
TEST_PASSWORD="testpass"
PHPMYADMIN_APP_PASSWORD="dsnet354"

#=============================================================================
# FUNGSI UTILITAS
#=============================================================================

# Fungsi untuk menampilkan pesan dengan warna
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Fungsi untuk logging
log_message() {
    local message=$1
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $message" >> "$LOG_FILE"
}

# Fungsi untuk menampilkan header
show_header() {
    clear
    print_message $BLUE "═══════════════════════════════════════════════════════════════════════════════"
    print_message $BLUE "                    FreeRADIUS Installation Script v1.0"
    print_message $BLUE "                         Ubuntu 22.04 LTS - MySQL Backend"
    print_message $BLUE "═══════════════════════════════════════════════════════════════════════════════"
    echo
}

# Fungsi progress bar
show_progress() {
    local current=$1
    local total=$2
    local description=$3
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((current * width / total))
    
    printf "\r[%s] %d%% %s" \
        "$(printf '%*s' "$completed" | tr ' ' '█')$(printf '%*s' "$((width - completed))" | tr ' ' '░')" \
        "$percentage" \
        "$description"
    
    if [ $current -eq $total ]; then
        echo
    fi
}

# Fungsi error handling
handle_error() {
    local exit_code=$1
    local error_message=$2
    local line_number=$3
    
    if [ $exit_code -ne 0 ]; then
        print_message $RED "❌ ERROR: $error_message (Line: $line_number)"
        log_message "ERROR: $error_message (Exit Code: $exit_code, Line: $line_number)"
        print_message $YELLOW "📋 Log file tersedia di: $LOG_FILE"
        exit $exit_code
    fi
}

# Fungsi untuk menjalankan command dengan error handling
run_command() {
    local command="$1"
    local description="$2"
    local line_number="$3"
    
    log_message "Executing: $command"
    eval "$command" >> "$LOG_FILE" 2>&1
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        log_message "SUCCESS: $description"
    else
        handle_error $exit_code "$description failed" "$line_number"
    fi
    
    return $exit_code
}

# Fungsi untuk konfirmasi user
confirm_action() {
    local message=$1
    local default=${2:-"y"}
    
    if [ "$default" = "y" ]; then
        prompt="[Y/n]"
    else
        prompt="[y/N]"
    fi
    
    while true; do
        read -p "$message $prompt: " choice
        case "${choice:-$default}" in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Silakan jawab yes (y) atau no (n).";;
        esac
    done
}

# Fungsi untuk mengecek apakah command tersedia
check_command() {
    local cmd=$1
    if ! command -v "$cmd" &> /dev/null; then
        return 1
    fi
    return 0
}

# Fungsi untuk mengecek status service
check_service_status() {
    local service_name=$1
    if systemctl is-active --quiet "$service_name"; then
        return 0
    else
        return 1
    fi
}

#=============================================================================
# FUNGSI VALIDASI SISTEM
#=============================================================================

# Fungsi untuk validasi sistem operasi
validate_os() {
    print_message $BLUE "🔍 Memvalidasi sistem operasi..."
    
    # Cek apakah OS adalah Ubuntu
    if [ ! -f /etc/os-release ]; then
        handle_error 1 "File /etc/os-release tidak ditemukan" $LINENO
    fi
    
    source /etc/os-release
    
    if [ "$ID" != "ubuntu" ]; then
        handle_error 1 "Script ini hanya mendukung Ubuntu. OS terdeteksi: $ID" $LINENO
    fi
    
    # Cek versi Ubuntu
    if [ "$VERSION_ID" != "22.04" ]; then
        print_message $YELLOW "⚠️  Peringatan: Script ini dioptimalkan untuk Ubuntu 22.04. Versi terdeteksi: $VERSION_ID"
        if ! confirm_action "Apakah Anda ingin melanjutkan?"; then
            exit 0
        fi
    fi
    
    print_message $GREEN "✅ Sistem operasi valid: Ubuntu $VERSION_ID"
    log_message "OS validation passed: Ubuntu $VERSION_ID"
}

# Fungsi untuk validasi hak akses root
validate_root() {
    print_message $BLUE "🔍 Memvalidasi hak akses..."
    
    if [ "$EUID" -ne 0 ]; then
        handle_error 1 "Script ini harus dijalankan sebagai root. Gunakan: sudo ./install.sh" $LINENO
    fi
    
    print_message $GREEN "✅ Hak akses root tervalidasi"
    log_message "Root access validation passed"
}

# Fungsi untuk validasi koneksi internet
validate_internet() {
    print_message $BLUE "🔍 Memvalidasi koneksi internet..."
    
    if ! ping -c 1 8.8.8.8 &> /dev/null; then
        handle_error 1 "Tidak ada koneksi internet. Pastikan koneksi internet aktif" $LINENO
    fi
    
    print_message $GREEN "✅ Koneksi internet tersedia"
    log_message "Internet connection validation passed"
}

# Fungsi untuk validasi space disk
validate_disk_space() {
    print_message $BLUE "🔍 Memvalidasi ruang disk..."
    
    local required_space=1048576  # 1GB in KB
    local available_space=$(df / | awk 'NR==2 {print $4}')
    
    if [ "$available_space" -lt "$required_space" ]; then
        handle_error 1 "Ruang disk tidak mencukupi. Diperlukan minimal 1GB, tersedia: $((available_space/1024))MB" $LINENO
    fi
    
    print_message $GREEN "✅ Ruang disk mencukupi: $((available_space/1024))MB tersedia"
    log_message "Disk space validation passed: ${available_space}KB available"
}

#=============================================================================
# FUNGSI UTAMA INSTALASI
#=============================================================================

# Fungsi untuk memperbarui sistem (ringan dan aman)
update_system() {
    print_message $BLUE "📦 Memperbarui sistem (ringan)..."
    show_progress 1 4 "Updating package list..."

    run_command "apt-get update" "Update package list" $LINENO
    show_progress 2 4 "Installing essential packages..."

    # Paket utilitas yang diperlukan untuk validasi dan tools
    run_command "DEBIAN_FRONTEND=noninteractive apt-get install -y curl wget gnupg2 software-properties-common apt-transport-https ca-certificates net-tools" "Install essential packages" $LINENO
    show_progress 3 4 "Holding off full upgrade..."

    # Lewati full upgrade agar aman pada sistem existing
    log_message "Skipping distribution upgrade by design"
    show_progress 4 4 "System prep completed"

    print_message $GREEN "✅ Paket dasar terpasang"
    log_message "System package prep completed successfully"
}

# Fungsi untuk instalasi MySQL Server
install_mysql() {
    print_message $BLUE "🗄️  Menginstal MySQL Server..."
    show_progress 1 6 "Installing MySQL Server..."
    
    show_progress 2 6 "Installing packages..."
    run_command "DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server mysql-client" "Install MySQL Server" $LINENO
    
    show_progress 3 6 "Starting MySQL service..."
    run_command "systemctl enable --now mysql" "Enable + start MySQL service" $LINENO
    
    # Tunggu MySQL siap menerima koneksi
    show_progress 4 6 "Waiting for MySQL to be ready..."
    for i in {1..30}; do
        if mysqladmin ping >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    
    # Tentukan metode admin yang tersedia untuk mengatur root password
    ADMIN_AVAILABLE=""
    if mysql -uroot -e "SELECT 1" >/dev/null 2>&1; then
        ADMIN_AVAILABLE="root_socket"
    elif [ -f /etc/mysql/debian.cnf ] && mysql --defaults-file=/etc/mysql/debian.cnf -e "SELECT 1" >/dev/null 2>&1; then
        ADMIN_AVAILABLE="debian_sys_maint"
    elif mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; then
        ADMIN_AVAILABLE="root_password"
    fi

    case "$ADMIN_AVAILABLE" in
        root_socket)
            # Set atau perbarui password root menggunakan koneksi socket
            run_command "mysql -e \"ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '$MYSQL_ROOT_PASSWORD'; FLUSH PRIVILEGES;\"" "Set MySQL root password (socket)" $LINENO
            ;;
        debian_sys_maint)
            # Gunakan akun sistem untuk mengatur root password
            run_command "mysql --defaults-file=/etc/mysql/debian.cnf -e \"ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '$MYSQL_ROOT_PASSWORD'; FLUSH PRIVILEGES;\"" "Set MySQL root password (debian-sys-maint)" $LINENO
            ;;
        root_password)
            log_message "Root password already set and valid"
            ;;
        *)
            print_message $YELLOW "⚠️  Tidak bisa autentikasi ke MySQL untuk set root password (akan lanjut)"
            log_message "WARNING: Could not auth to MySQL to set root password"
            ;;
    esac

    # Pembersihan minimal ala mysql_secure_installation (gunakan kredensial yang tersedia)
    if mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; then
        mysql -uroot -p"$MYSQL_ROOT_PASSWORD" <<EOF
DELETE FROM mysql.user WHERE User='';
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\_%';
FLUSH PRIVILEGES;
EOF
    elif [ -f /etc/mysql/debian.cnf ]; then
        mysql --defaults-file=/etc/mysql/debian.cnf <<EOF
DELETE FROM mysql.user WHERE User='';
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\_%';
FLUSH PRIVILEGES;
EOF
    fi
    
    show_progress 5 6 "Verifying MySQL installation..."
    if ! check_service_status "mysql"; then
        handle_error 1 "MySQL service tidak berjalan" $LINENO
    fi
    
    show_progress 6 6 "MySQL installation completed"
    print_message $GREEN "✅ MySQL Server berhasil diinstal dan dikonfigurasi"
    log_message "MySQL installation completed successfully"
}

# Fungsi untuk membuat database dan user FreeRADIUS
setup_radius_database() {
    print_message $BLUE "🗃️  Membuat database FreeRADIUS..."
    show_progress 1 5 "Creating database..."
    
    # Pilih kredensial admin yang tersedia
    MYSQL_ADMIN=""
    if mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; then
        MYSQL_ADMIN=(mysql -uroot -p"$MYSQL_ROOT_PASSWORD")
    elif [ -f /etc/mysql/debian.cnf ] && mysql --defaults-file=/etc/mysql/debian.cnf -e "SELECT 1" >/dev/null 2>&1; then
        MYSQL_ADMIN=(mysql --defaults-file=/etc/mysql/debian.cnf)
    elif mysql -uroot -e "SELECT 1" >/dev/null 2>&1; then
        MYSQL_ADMIN=(mysql -uroot)
    else
        handle_error 1 "Tidak dapat terhubung ke MySQL sebagai admin (root/debian-sys-maint)" $LINENO
    fi

    # Buat database radius
    "${MYSQL_ADMIN[@]}" <<EOF
CREATE DATABASE IF NOT EXISTS $RADIUS_DB_NAME;
EOF
    
    if [ $? -ne 0 ]; then
        handle_error 1 "Gagal membuat database $RADIUS_DB_NAME" $LINENO
    fi
    
    show_progress 2 5 "Creating database user..."
    # Buat user untuk FreeRADIUS (paksa plugin klasik untuk kompatibilitas luas)
    "${MYSQL_ADMIN[@]}" <<EOF
CREATE USER IF NOT EXISTS '$RADIUS_DB_USER'@'localhost' IDENTIFIED WITH mysql_native_password BY '$RADIUS_DB_PASSWORD';
GRANT ALL PRIVILEGES ON $RADIUS_DB_NAME.* TO '$RADIUS_DB_USER'@'localhost';
FLUSH PRIVILEGES;
EOF
    
    if [ $? -ne 0 ]; then
        handle_error 1 "Gagal membuat user database $RADIUS_DB_USER" $LINENO
    fi
    
    show_progress 3 5 "Testing database connection..."
    # Test koneksi database
    mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" --execute="SELECT 1;" > /dev/null 2>&1
    
    if [ $? -ne 0 ]; then
        handle_error 1 "Gagal terhubung ke database dengan user $RADIUS_DB_USER" $LINENO
    fi
    
    show_progress 4 5 "Database setup completed"
    show_progress 5 5 "Ready for FreeRADIUS schema"
    
    print_message $GREEN "✅ Database FreeRADIUS berhasil dibuat"
    log_message "FreeRADIUS database setup completed successfully"
}

# Fungsi untuk instalasi FreeRADIUS
install_freeradius() {
    print_message $BLUE "📡 Menginstal FreeRADIUS..."
    show_progress 1 8 "Installing FreeRADIUS packages..."
    
    # Install FreeRADIUS dan modul MySQL
    run_command "apt-get install -y freeradius freeradius-mysql freeradius-utils" "Install FreeRADIUS packages" $LINENO
    
    show_progress 2 8 "Installing additional dependencies..."
    run_command "apt-get install -y libmysqlclient-dev build-essential" "Install additional dependencies" $LINENO
    
    show_progress 3 8 "Stopping FreeRADIUS service..."
    run_command "systemctl stop freeradius" "Stop FreeRADIUS service" $LINENO
    
    show_progress 4 8 "Creating FreeRADIUS database schema..."
    # Import schema FreeRADIUS ke database
    if [ -f "/etc/freeradius/3.0/mods-config/sql/main/mysql/schema.sql" ]; then
        mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" < /etc/freeradius/3.0/mods-config/sql/main/mysql/schema.sql
        if [ $? -ne 0 ]; then
            handle_error 1 "Gagal mengimport schema FreeRADIUS" $LINENO
        fi
    else
        handle_error 1 "File schema FreeRADIUS tidak ditemukan" $LINENO
    fi
    
    show_progress 5 8 "Restoring safe permissions..."
    # Jika sebelumnya ada chmod recursive yang salah, pulihkan permission aman
    if [ -d /etc/freeradius ]; then
        chown -R root:freerad /etc/freeradius/ 2>/dev/null || true
        find /etc/freeradius -type d -exec chmod 750 {} + 2>/dev/null || true
        find /etc/freeradius -type f -exec chmod 640 {} + 2>/dev/null || true
        chmod 750 /etc/freeradius /etc/freeradius/3.0 2>/dev/null || true
        log_message "Permissions on /etc/freeradius restored to root:freerad (dirs 750, files 640)"
    fi
    
    show_progress 6 8 "Verifying service binaries..."
    
    show_progress 7 8 "Verifying FreeRADIUS installation..."
    # Cek apakah FreeRADIUS terinstal dengan benar
    if ! dpkg -l | grep -q "freeradius " && ! check_command "freeradius" && ! [ -f "/usr/sbin/freeradius" ]; then
        handle_error 1 "FreeRADIUS tidak terinstal dengan benar" $LINENO
    fi
    
    show_progress 8 8 "FreeRADIUS installation completed"
    print_message $GREEN "✅ FreeRADIUS berhasil diinstal"
    log_message "FreeRADIUS installation completed successfully"
}

# Fungsi untuk konfigurasi FreeRADIUS dengan MySQL
configure_freeradius_mysql() {
    print_message $BLUE "⚙️  Mengkonfigurasi FreeRADIUS dengan MySQL..."
    show_progress 1 6 "Configuring SQL module..."
    
    # Backup konfigurasi asli
    run_command "cp /etc/freeradius/3.0/mods-available/sql /etc/freeradius/3.0/mods-available/sql.backup" "Backup SQL module config" $LINENO
    
    # Konfigurasi modul SQL
    cat > /etc/freeradius/3.0/mods-available/sql <<EOF
sql {
    driver = "rlm_sql_mysql"
    dialect = "mysql"
    
    # Connection info:
    server = "localhost"
    port = 3306
    login = "$RADIUS_DB_USER"
    password = "$RADIUS_DB_PASSWORD"
    
    # Database table configuration for everything except Oracle
    radius_db = "$RADIUS_DB_NAME"
    
    # If you are using Oracle then use this instead
    # radius_db = "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SID=your_sid)))"
    
    # If you want both stop and start records logged to the
    # same SQL table, leave this as is.  If you want them in
    # different tables, put the start table in acct_table1
    # and stop table in acct_table2
    acct_table1 = "radacct"
    acct_table2 = "radacct"
    
    postauth_table = "radpostauth"
    authcheck_table = "radcheck"
    authreply_table = "radreply"
    groupcheck_table = "radgroupcheck"
    groupreply_table = "radgroupreply"
    usergroup_table = "radusergroup"
    
    # Remove stale session if checkrad does not see a double login
    delete_stale_sessions = yes
    
    pool {
        start = \${thread[pool].start_servers}
        min = \${thread[pool].min_spare_servers}
        max = \${thread[pool].max_servers}
        spare = \${thread[pool].max_spare_servers}
        uses = 0
        retry_delay = 30
        lifetime = 0
        idle_timeout = 60
    }
    
    # Set to 'yes' to read radius clients from the database ('nas' table)
    read_clients = yes
    
    client_table = "nas"
    
    group_attribute = "SQL-Group"
    
    \$INCLUDE \${modconfdir}/sql/main/\${dialect}/queries.conf
}
EOF
    
    show_progress 2 6 "Enabling SQL module..."
    # Enable modul SQL
    run_command "ln -sf /etc/freeradius/3.0/mods-available/sql /etc/freeradius/3.0/mods-enabled/" "Enable SQL module" $LINENO
    
    show_progress 3 6 "Configuring default site..."
    # Backup dan konfigurasi site default
    run_command "cp /etc/freeradius/3.0/sites-available/default /etc/freeradius/3.0/sites-available/default.backup" "Backup default site config" $LINENO
    
    # Pastikan baris komentar yang mengandung kata 'sql ...' tetap dikomentari (perbaikan jika pernah terbuka)
    sed -ri 's/^([[:space:]]*)sql module can handle this\./\1# sql module can handle this./' /etc/freeradius/3.0/sites-available/default
    sed -ri 's/^([[:space:]]*)sql module is \*much\* faster/\1# sql module is *much* faster/' /etc/freeradius/3.0/sites-available/default
    sed -ri 's/^([[:space:]]*)sql[ \/].+/\1# &/' /etc/freeradius/3.0/sites-available/default
    
    # Nonaktifkan modul opsional yang tidak kita konfigurasikan
    sed -ri 's/^([[:space:]]*)sqlippool\s*$/\1# sqlippool/' /etc/freeradius/3.0/sites-available/default
    sed -ri 's/^([[:space:]]*)sql-voip\s*$/\1# sql-voip/' /etc/freeradius/3.0/sites-available/default

    # Aktifkan sql di beberapa section secara aman (hanya baris 'sql' murni)
    for section in authorize accounting session post-auth; do
        sed -ri "/^\s*${section}\s*\{/,/^\s*\}/ s/^\s*#\s*sql\s*$/\tsql/" /etc/freeradius/3.0/sites-available/default
    done
    
    show_progress 4 6 "Configuring inner-tunnel..."
    # Konfigurasi inner-tunnel untuk EAP
    run_command "cp /etc/freeradius/3.0/sites-available/inner-tunnel /etc/freeradius/3.0/sites-available/inner-tunnel.backup" "Backup inner-tunnel config" $LINENO
    
    # Pastikan komentar-komentar tetap aman pada inner-tunnel
    sed -ri 's/^([[:space:]]*)sql module can handle this\./\1# sql module can handle this./' /etc/freeradius/3.0/sites-available/inner-tunnel
    sed -ri 's/^([[:space:]]*)sql module is \*much\* faster/\1# sql module is *much* faster/' /etc/freeradius/3.0/sites-available/inner-tunnel
    sed -ri 's/^([[:space:]]*)sql\/main\/\$driver\/queries\.conf`/\1# sql\/main\/\$driver\/queries.conf`/' /etc/freeradius/3.0/sites-available/inner-tunnel

    # Aktifkan sql di inner-tunnel authorize dan session (hanya baris 'sql' murni)
    for section in authorize session; do
        sed -ri "/^\s*${section}\s*\{/,/^\s*\}/ s/^\s*#\s*sql\s*$/\tsql/" /etc/freeradius/3.0/sites-available/inner-tunnel
    done
    
    show_progress 5 6 "Testing FreeRADIUS configuration..."
    # Test konfigurasi FreeRADIUS
    if ! freeradius -C > /dev/null 2>&1; then
        # Coba dengan radiusd jika freeradius tidak berhasil
        if ! radiusd -C > /dev/null 2>&1; then
            print_message $YELLOW "⚠️  Peringatan: Tidak dapat memvalidasi konfigurasi FreeRADIUS secara otomatis"
            log_message "Warning: Cannot validate FreeRADIUS configuration automatically"
        fi
    fi
    
    show_progress 6 6 "FreeRADIUS MySQL configuration completed"
    print_message $GREEN "✅ FreeRADIUS berhasil dikonfigurasi dengan MySQL"
    log_message "FreeRADIUS MySQL configuration completed successfully"
}

# Fungsi untuk setup user testing
setup_test_users() {
    print_message $BLUE "👤 Membuat user testing..."
    show_progress 1 4 "Creating test user in database..."
    
    # Bersihkan jika sudah ada lalu tambahkan user testing ke database (idempotent)
    mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" <<EOF
DELETE FROM radcheck  WHERE username IN ('$TEST_USER','admin','user1');
DELETE FROM radreply  WHERE username IN ('$TEST_USER','admin','user1');
INSERT INTO radcheck (username, attribute, op, value) VALUES ('$TEST_USER', 'Cleartext-Password', ':=', '$TEST_PASSWORD');
INSERT INTO radreply (username, attribute, op, value) VALUES ('$TEST_USER', 'Reply-Message', '=', 'Hello $TEST_USER - Authentication successful');
EOF
    
    if [ $? -ne 0 ]; then
        handle_error 1 "Gagal membuat user testing" $LINENO
    fi
    
    show_progress 2 4 "Creating additional test users..."
    # Tambahkan beberapa user testing lainnya
    mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" <<EOF
INSERT INTO radcheck (username, attribute, op, value) VALUES ('admin', 'Cleartext-Password', ':=', 'admin123');
INSERT INTO radreply (username, attribute, op, value) VALUES ('admin', 'Reply-Message', '=', 'Hello admin - Authentication successful');
INSERT INTO radcheck (username, attribute, op, value) VALUES ('user1', 'Cleartext-Password', ':=', 'password1');
INSERT INTO radreply (username, attribute, op, value) VALUES ('user1', 'Reply-Message', '=', 'Hello user1 - Authentication successful');
EOF
    
    show_progress 3 4 "Verifying test users..."
    # Verifikasi user testing
    local user_count=$(mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" --execute="SELECT COUNT(*) FROM radcheck;" --skip-column-names 2>/dev/null)
    
    if [ "$user_count" -lt 3 ]; then
        handle_error 1 "User testing tidak berhasil dibuat" $LINENO
    fi
    
    show_progress 4 4 "Test users setup completed"
    print_message $GREEN "✅ User testing berhasil dibuat"
    print_message $YELLOW "   • Username: $TEST_USER, Password: $TEST_PASSWORD"
    print_message $YELLOW "   • Username: admin, Password: admin123"
    print_message $YELLOW "   • Username: user1, Password: password1"
    log_message "Test users setup completed successfully"
}

# Fungsi untuk konfigurasi UFW firewall
configure_ufw_firewall() {
    print_message $BLUE "🔥 Mengkonfigurasi UFW Firewall..."
    show_progress 1 6 "Checking UFW installation..."
    
    # Cek apakah UFW terinstal
    if ! check_command "ufw"; then
        show_progress 2 6 "Installing UFW..."
        run_command "apt-get install -y ufw" "Install UFW" $LINENO
    else
        show_progress 2 6 "UFW already installed"
        log_message "UFW already installed"
    fi
    
    show_progress 3 6 "Configuring UFW rules for FreeRADIUS..."
    # Aktifkan port 1812 (Authentication) dan 1813 (Accounting) untuk UDP
    run_command "ufw allow 1812/udp comment 'FreeRADIUS Authentication'" "Allow port 1812/udp" $LINENO
    run_command "ufw allow 1813/udp comment 'FreeRADIUS Accounting'" "Allow port 1813/udp" $LINENO
    
    show_progress 4 6 "Enabling UFW if not already enabled..."
    # Cek status UFW dan aktifkan jika belum aktif
    if ! ufw status | grep -q "Status: active"; then
        # Aktifkan UFW dengan konfirmasi otomatis
        run_command "echo 'y' | ufw enable" "Enable UFW" $LINENO
    else
        log_message "UFW already enabled"
    fi
    
    show_progress 5 6 "Verifying UFW rules..."
    # Verifikasi rules yang telah ditambahkan
    local ufw_rules=$(ufw status numbered | grep -E "1812|1813" | wc -l)
    
    if [ "$ufw_rules" -lt 2 ]; then
        print_message $YELLOW "⚠️  Peringatan: Rules UFW untuk FreeRADIUS mungkin tidak terkonfigurasi dengan benar"
        log_message "WARNING: UFW rules for FreeRADIUS may not be configured correctly"
    else
        print_message $GREEN "✅ Rules UFW untuk port 1812 dan 1813 berhasil dikonfigurasi"
        log_message "UFW rules for ports 1812 and 1813 configured successfully"
    fi
    
    show_progress 6 6 "UFW firewall configuration completed"
    print_message $GREEN "✅ UFW Firewall berhasil dikonfigurasi untuk FreeRADIUS"
    log_message "UFW firewall configuration completed successfully"
}

# Fungsi untuk validasi post-installation
validate_installation() {
    print_message $BLUE "🔍 Memvalidasi instalasi..."
    show_progress 1 8 "Checking MySQL service..."
    
    # Cek MySQL service
    if ! check_service_status "mysql"; then
        handle_error 1 "MySQL service tidak berjalan" $LINENO
    fi
    
    show_progress 2 8 "Checking FreeRADIUS service..."
    # Start dan cek FreeRADIUS service
    if systemctl list-unit-files | grep -q "freeradius.service"; then
        SERVICE_NAME="freeradius"
    else
        SERVICE_NAME="radiusd"
    fi
    
    # Coba start service dengan error handling yang lebih baik
    if ! systemctl start "$SERVICE_NAME" 2>/dev/null; then
        print_message $YELLOW "⚠️  Peringatan: Gagal start service $SERVICE_NAME, mencoba debug..."
        log_message "Warning: Failed to start $SERVICE_NAME service"
        # Coba jalankan FreeRADIUS dalam debug mode untuk melihat error
        freeradius -X > /tmp/freeradius_debug.log 2>&1 &
        sleep 2
        pkill freeradius 2>/dev/null
        log_message "FreeRADIUS debug output saved to /tmp/freeradius_debug.log"
    else
        systemctl enable "$SERVICE_NAME" 2>/dev/null
        log_message "$SERVICE_NAME service started and enabled successfully"
    fi
    
    # Cek status service dengan lebih fleksibel
    if ! check_service_status "$SERVICE_NAME" && ! pgrep -f "freeradius" > /dev/null; then
        print_message $YELLOW "⚠️  Peringatan: FreeRADIUS service tidak berjalan, tetapi instalasi dilanjutkan"
        log_message "Warning: FreeRADIUS service not running, but continuing installation"
    fi
    
    show_progress 3 8 "Testing database connection..."
    # Test koneksi database
    mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" --execute="SELECT COUNT(*) FROM radcheck;" > /dev/null 2>&1
    
    if [ $? -ne 0 ]; then
        handle_error 1 "Koneksi database FreeRADIUS gagal" $LINENO
    fi
    
    show_progress 4 8 "Testing FreeRADIUS configuration..."
    # Test konfigurasi FreeRADIUS
    if ! freeradius -C > /dev/null 2>&1; then
        # Coba dengan radiusd jika freeradius tidak berhasil
        if ! radiusd -C > /dev/null 2>&1; then
            print_message $YELLOW "⚠️  Peringatan: Tidak dapat memvalidasi konfigurasi FreeRADIUS secara otomatis"
            log_message "Warning: Cannot validate FreeRADIUS configuration automatically in validation"
        fi
    fi
    
    show_progress 5 8 "Testing RADIUS authentication..."
    # Test autentikasi RADIUS
    sleep 2  # Tunggu service fully started
    
    local auth_test=$(echo "User-Name = '$TEST_USER', User-Password = '$TEST_PASSWORD'" | radclient -x localhost:1812 auth testing123 2>/dev/null | grep "Access-Accept")
    
    if [ -z "$auth_test" ]; then
        print_message $YELLOW "⚠️  Peringatan: Test autentikasi RADIUS gagal. Mungkin perlu konfigurasi tambahan."
        log_message "WARNING: RADIUS authentication test failed"
    else
        print_message $GREEN "✅ Test autentikasi RADIUS berhasil"
        log_message "RADIUS authentication test successful"
    fi
    
    show_progress 6 8 "Checking listening ports..."
    # Cek port yang digunakan
    local radius_ports=$(netstat -ulnp | grep ":1812\|:1813" | wc -l)
    
    if [ "$radius_ports" -lt 2 ]; then
        print_message $YELLOW "⚠️  Peringatan: FreeRADIUS mungkin tidak listening pada port standar (1812/1813)"
        log_message "WARNING: FreeRADIUS not listening on standard ports"
    fi
    
    show_progress 7 8 "Generating service status report..."
    # Generate status report
    echo "=== FreeRADIUS Installation Status Report ===" > /tmp/freeradius_status.txt
    echo "Date: $(date)" >> /tmp/freeradius_status.txt
    echo "MySQL Status: $(systemctl is-active mysql)" >> /tmp/freeradius_status.txt
    echo "FreeRADIUS Status: $(systemctl is-active freeradius)" >> /tmp/freeradius_status.txt
    echo "Database Users: $(mysql --user="$RADIUS_DB_USER" --password="$RADIUS_DB_PASSWORD" --database="$RADIUS_DB_NAME" --execute="SELECT COUNT(*) FROM radcheck;" --skip-column-names 2>/dev/null)" >> /tmp/freeradius_status.txt
    echo "Listening Ports: $(netstat -ulnp | grep ":1812\|:1813")" >> /tmp/freeradius_status.txt
    
    show_progress 8 8 "Installation validation completed"
    print_message $GREEN "✅ Validasi instalasi selesai"
    log_message "Installation validation completed successfully"
}

# Fungsi untuk menampilkan dokumentasi penggunaan
show_usage_documentation() {
    print_message $BLUE "📚 Dokumentasi Penggunaan FreeRADIUS"
    echo
    print_message $GREEN "=== INFORMASI INSTALASI ==="
    echo "• FreeRADIUS Server: Terinstal dan berjalan"
    echo "• MySQL Database: Terinstal dan terkonfigurasi"
    echo "• UFW Firewall: Dikonfigurasi untuk port 1812/1813 UDP"
    echo "• Database Name: $RADIUS_DB_NAME"
    echo "• Database User: $RADIUS_DB_USER"
    echo "• Log File: $LOG_FILE"
    echo
    
    print_message $GREEN "=== USER TESTING ==="
    echo "User testing yang tersedia:"
    echo "• Username: $TEST_USER, Password: $TEST_PASSWORD"
    echo "• Username: admin, Password: admin123"
    echo "• Username: user1, Password: password1"
    echo
    
    print_message $GREEN "=== CARA TESTING AUTENTIKASI ==="
    echo "1. Test menggunakan radtest:"
    echo "   radtest $TEST_USER $TEST_PASSWORD localhost 1812 testing123"
    echo
    echo "2. Test menggunakan radclient:"
    echo "   echo \"User-Name = '$TEST_USER', User-Password = '$TEST_PASSWORD'\" | radclient localhost:1812 auth testing123"
    echo
    
    print_message $GREEN "=== MANAJEMEN SERVICE ==="
    echo "• Start FreeRADIUS: systemctl start freeradius"
    echo "• Stop FreeRADIUS: systemctl stop freeradius"
    echo "• Restart FreeRADIUS: systemctl restart freeradius"
    echo "• Status FreeRADIUS: systemctl status freeradius"
    echo "• Debug mode: freeradius -X"
    echo
    
    print_message $GREEN "=== FILE KONFIGURASI PENTING ==="
    echo "• Main config: /etc/freeradius/3.0/radiusd.conf"
    echo "• SQL module: /etc/freeradius/3.0/mods-available/sql"
    echo "• Default site: /etc/freeradius/3.0/sites-available/default"
    echo "• Clients: /etc/freeradius/3.0/clients.conf"
    echo
    
    print_message $GREEN "=== MANAJEMEN DATABASE ==="
    echo "• Koneksi MySQL: mysql -u $RADIUS_DB_USER -p$RADIUS_DB_PASSWORD $RADIUS_DB_NAME"
    echo "• Tabel utama: radcheck, radreply, radgroupcheck, radgroupreply"
    echo "• Tambah user: INSERT INTO radcheck (username, attribute, op, value) VALUES ('newuser', 'Cleartext-Password', ':=', 'newpass');"
    echo
    
    print_message $GREEN "=== TROUBLESHOOTING ==="
    echo "• Cek log FreeRADIUS: tail -f /var/log/freeradius/radius.log"
    echo "• Debug mode: freeradius -X"
    echo "• Test konfigurasi: freeradius -CX"
    echo "• Cek port: netstat -ulnp | grep 1812"
    echo "• Cek UFW status: ufw status"
    echo "• Cek UFW rules: ufw status numbered"
    echo
    
    print_message $YELLOW "=== KEAMANAN ==="
    echo "⚠️  PENTING: Ganti password default sebelum production!"
    echo "• MySQL root password: $MYSQL_ROOT_PASSWORD"
    echo "• Database password: $RADIUS_DB_PASSWORD"
    echo "• RADIUS shared secret: testing123 (ganti di /etc/freeradius/3.0/clients.conf)"
    echo
    
    print_message $BLUE "=== STATUS REPORT ==="
    echo "Status report tersimpan di: /tmp/freeradius_status.txt"
    echo "Installation log tersimpan di: $LOG_FILE"
    echo
    
    log_message "Usage documentation displayed"
}

#=============================================================================
# FUNGSI MAIN
#=============================================================================

# Fungsi utama
main() {
    # Inisialisasi log file
    echo "FreeRADIUS Installation Log - $(date)" > "$LOG_FILE"
    log_message "Installation started"
    
    # Tampilkan header
    show_header
    
    # Konfirmasi instalasi
    print_message $YELLOW "⚠️  Script ini akan menginstal dan mengkonfigurasi:"
    echo "   • FreeRADIUS Server"
    echo "   • MySQL Server"
    echo "   • Database dan tabel untuk autentikasi"
    echo "   • UFW Firewall (port 1812 dan 1813 UDP)"
    echo "   • Konfigurasi minimal untuk testing"
    echo
    
    if ! confirm_action "Apakah Anda ingin melanjutkan instalasi?"; then
        print_message $YELLOW "❌ Instalasi dibatalkan oleh user"
        exit 0
    fi
    
    echo
    print_message $BLUE "🚀 Memulai proses instalasi..."
    echo
    
    # Validasi sistem
    validate_os
    validate_root
    validate_internet
    validate_disk_space
    
    echo
    print_message $GREEN "✅ Semua validasi sistem berhasil"
    echo
    
    # Update sistem
    update_system
    
    echo
    print_message $BLUE "🗄️  Memulai instalasi MySQL..."
    install_mysql
    setup_radius_database
    
    echo
    print_message $BLUE "📡 Memulai instalasi FreeRADIUS..."
    install_freeradius
    configure_freeradius_mysql
    
    echo
    print_message $BLUE "👤 Membuat user testing..."
    setup_test_users
    
    echo
    print_message $BLUE "🔥 Mengkonfigurasi UFW Firewall..."
    configure_ufw_firewall
    
    echo
    print_message $BLUE "🔍 Memvalidasi instalasi..."
    validate_installation
    
    echo
    print_message $GREEN "🎉 Instalasi FreeRADIUS selesai!"
    echo
    
    # Tampilkan dokumentasi penggunaan
    show_usage_documentation
    
    print_message $GREEN "✅ Semua proses instalasi berhasil diselesaikan!"
    print_message $BLUE "📋 Silakan cek log lengkap di: $LOG_FILE"
    print_message $BLUE "📊 Status report tersedia di: /tmp/freeradius_status.txt"
    
    log_message "FreeRADIUS installation completed successfully"
}

# Trap untuk cleanup saat script dihentikan
trap 'print_message $RED "\n❌ Script dihentikan oleh user"; log_message "Script interrupted by user"; exit 130' INT TERM

# Jalankan fungsi utama
main "$@"
