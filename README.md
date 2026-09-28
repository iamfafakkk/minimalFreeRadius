# minimalFreeRadius

> Instalasi minimal FreeRADIUS dengan REST API untuk manajemen NAS dan user authentication

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-00ADD8.svg)](https://go.dev/)
[![MySQL](https://img.shields.io/badge/mysql-%3E%3D5.7-blue.svg)](https://www.mysql.com/)

## 📖 Deskripsi

minimalFreeRadius adalah solusi lengkap untuk instalasi dan manajemen FreeRADIUS server dengan backend MySQL. Proyek ini menyediakan:

- **Script instalasi otomatis** untuk Ubuntu 22.04 LTS
- **REST API** untuk manajemen NAS (Network Access Server) dan user
- **Konfigurasi keamanan** dengan UFW firewall
- **Database schema** yang sudah dikonfigurasi
- **User testing** untuk validasi instalasi

## 🎯 Fitur Utama

### FreeRADIUS Server
- ✅ Instalasi otomatis FreeRADIUS 3.0
- ✅ Integrasi dengan MySQL backend
- ✅ Konfigurasi SQL module
- ✅ User authentication dengan database
- ✅ Logging dan monitoring

### REST API
- 🔐 JWT Authentication
- 🔑 API Key Authentication
- 📊 CRUD operations untuk NAS
- 👤 CRUD operations untuk User
- 🛡️ Rate limiting dan CORS
- 📝 Swagger documentation
- ❤️ Health check endpoint

### Keamanan
- 🔥 UFW firewall configuration
- 🔒 Secure MySQL setup
- 🛡️ Input validation
- 📋 Comprehensive logging

## 📋 Prasyarat Sistem

### Sistem Operasi
- **Ubuntu 22.04 LTS** (direkomendasikan)
- **Ubuntu 20.04 LTS** (kompatibel)
- **Debian 11+** (kompatibel)

### Hardware Requirements
- **RAM**: Minimum 1GB, direkomendasikan 2GB+
- **Storage**: Minimum 2GB free space
- **Network**: Koneksi internet untuk download packages

### Software Requirements
- **Root access** atau sudo privileges
- **MySQL 5.7+** atau **MariaDB 10.3+**
- **Go 1.18+** + **gcc/build-essential** (REST API, CGO untuk driver SQLite)
- **systemd** (service API dijalankan sebagai unit systemd)

## 🚀 Panduan Instalasi Lengkap

### Langkah 1: Clone Repository

```bash
# Clone repository
git clone https://github.com/iamfafakkk/minimalFreeRadius.git .

# Berikan permission execute pada script
chmod +x install.sh
```

### Langkah 2: Jalankan Script Instalasi

```bash
# Jalankan sebagai root
sudo ./install.sh
```

Script akan melakukan:
1. ✅ Validasi sistem operasi dan requirements
2. 📦 Update sistem dan install dependencies
3. 🗄️ Install dan konfigurasi MySQL Server
4. 📡 Install dan konfigurasi FreeRADIUS
5. 🔗 Setup integrasi MySQL dengan FreeRADIUS
6. 👤 Membuat user testing
7. 🔥 Konfigurasi UFW firewall
8. ✅ Validasi instalasi

### Langkah 3: Setup REST API (Opsional)

```bash
# Masuk ke direktori API
cd freeradius-api

# Berikan permission execute pada script setup
chmod +x setup.sh

# Jalankan script setup untuk instalasi lengkap
sudo ./setup.sh
```

Script `setup.sh` akan melakukan:
1. ✅ Cek system requirements (Go, gcc, MySQL)
2. 📦 Download Go modules
3. ⚙️ Setup environment file (.env, JWT secret di-generate)
4. 📁 Buat direktori yang diperlukan
5. 🏗️ Build binary Go (`CGO_ENABLED=1 go build`)
6. 🗄️ Verifikasi schema database (dibuat `install.sh`)
7. 🖥️ Build frontend web (bila `npm` tersedia)
8. 🔥 Buka port API di UFW (bila UFW aktif)
9. 🚀 Install + start systemd service `freeradius-api`
10. ✅ Health check endpoint

**Opsi Setup yang Tersedia:**
```bash
# Setup lengkap (direkomendasikan)
sudo ./setup.sh

# Hanya cek requirements
./setup.sh --check-only

# Hanya download modules
./setup.sh --install-only

# Hanya build binary
./setup.sh --build-only

# Hanya verifikasi schema database
./setup.sh --db-only

# Hanya install systemd service (build dulu)
sudo ./setup.sh --systemd-only

# Install tanpa auto-start
sudo ./setup.sh --no-start

# Hapus systemd service
sudo ./setup.sh --remove-systemd

# Lihat bantuan
./setup.sh --help
```

### Langkah 4: Konfigurasi Environment API

Setelah menjalankan `setup.sh`, edit file `.env` jika diperlukan:

```bash
# Edit konfigurasi environment
nano .env
```

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=radius
DB_USER=radius
DB_PASSWORD=radiuspass123!

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRES_IN=24h

# API Configuration
PORT=3000
API_KEY=your-api-key-here

# Rate Limiting
RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=100
```

### Langkah 5: Mengelola Service API

Setelah instalasi dengan `setup.sh`, API akan dikelola sebagai systemd service:

```bash
# Cek status service
sudo systemctl status freeradius-api

# Start service
sudo systemctl start freeradius-api

# Stop service
sudo systemctl stop freeradius-api

# Restart service
sudo systemctl restart freeradius-api

# Enable service untuk auto-start saat boot
sudo systemctl enable freeradius-api

# Disable auto-start
sudo systemctl disable freeradius-api

# Lihat log service
journalctl -u freeradius-api -f
```

## ⚙️ Konfigurasi Awal

### 1. Konfigurasi Database

Setelah instalasi, database `radius` akan dibuat dengan tabel:
- `radcheck` - User authentication
- `radreply` - User attributes
- `nas` - Network Access Servers
- `radacct` - Accounting records
- `radpostauth` - Post-authentication logs

### 2. Konfigurasi FreeRADIUS Clients

Edit `/etc/freeradius/3.0/clients.conf`:

```bash
client localhost {
    ipaddr = 127.0.0.1
    secret = testing123
    require_message_authenticator = no
    nas_type = other
}

client private-network-1 {
    ipaddr = 192.168.1.0/24
    secret = your-strong-secret-here
    require_message_authenticator = no
    nas_type = other
}
```

### 3. Konfigurasi UFW Firewall

Firewall sudah dikonfigurasi otomatis untuk:
- Port `1812/udp` - RADIUS Authentication
- Port `1813/udp` - RADIUS Accounting
- Port `3000/tcp` - REST API (jika digunakan)

```bash
# Cek status firewall
sudo ufw status

# Tambah rule custom jika diperlukan
sudo ufw allow from 192.168.1.0/24 to any port 1812 proto udp
```

## 💡 Contoh Penggunaan Dasar

### 1. Testing Autentikasi RADIUS

```bash
# Test dengan radtest
radtest testuser testpass localhost 1812 testing123

# Test dengan radclient
echo "User-Name = 'testuser', User-Password = 'testpass'" | \
  radclient localhost:1812 auth testing123
```

### 2. Manajemen User via Database

```sql
-- Tambah user baru
INSERT INTO radcheck (username, attribute, op, value) 
VALUES ('newuser', 'Cleartext-Password', ':=', 'newpassword');

-- Lihat semua user
SELECT * FROM radcheck WHERE attribute = 'Cleartext-Password';

-- Hapus user
DELETE FROM radcheck WHERE username = 'olduser';
```

### 3. Menggunakan REST API

```bash
# Health check
curl http://localhost:3000/api/health

# Login untuk mendapatkan JWT token
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Tambah NAS baru
curl -X POST http://localhost:3000/api/nas \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nasname": "192.168.1.1",
    "shortname": "Router1",
    "secret": "strongsecret123",
    "description": "Main Router"
  }'

# Lihat semua NAS
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/nas
```

### 4. Monitoring dan Logging

```bash
# Monitor log FreeRADIUS
sudo tail -f /var/log/freeradius/radius.log

# Debug mode FreeRADIUS
sudo freeradius -X

# Cek status service
sudo systemctl status freeradius
sudo systemctl status mysql
```

## 🔧 Informasi Troubleshooting

### Masalah Umum dan Solusi

#### 1. FreeRADIUS tidak start

```bash
# Cek konfigurasi
sudo freeradius -CX

# Cek log error
sudo journalctl -u freeradius -f

# Restart service
sudo systemctl restart freeradius
```

#### 2. Database connection error

```bash
# Test koneksi database
mysql -u radius -p radius

# Cek konfigurasi SQL module
sudo nano /etc/freeradius/3.0/mods-available/sql

# Restart setelah perubahan
sudo systemctl restart freeradius
```

#### 3. Authentication failed

```bash
# Cek user di database
mysql -u radius -p radius -e "SELECT * FROM radcheck;"

# Test dengan debug mode
sudo freeradius -X

# Cek client configuration
sudo nano /etc/freeradius/3.0/clients.conf
```

#### 4. Port tidak listening

```bash
# Cek port yang digunakan
sudo netstat -ulnp | grep 1812
sudo netstat -ulnp | grep 1813

# Cek firewall
sudo ufw status

# Restart networking
sudo systemctl restart networking
```

#### 5. REST API tidak bisa diakses

```bash
# Cek status service API
sudo systemctl status freeradius-api

# Cek log service API
journalctl -u freeradius-api -f

# Restart service API
sudo systemctl restart freeradius-api

# Cek environment variables
cat freeradius-api/.env

# Test health check API
curl http://localhost:3000/api/v1/auth/health

# Cek port yang digunakan API
sudo netstat -tlnp | grep 3000
```

### Log Files Penting

- **FreeRADIUS**: `/var/log/freeradius/radius.log`
- **MySQL**: `/var/log/mysql/error.log`
- **Installation**: `/tmp/freeradius_install.log`
- **System**: `journalctl -u freeradius`

### Perintah Diagnostik

```bash
# Cek semua service
sudo systemctl status freeradius mysql

# Cek konfigurasi FreeRADIUS
sudo freeradius -CX

# Cek koneksi database
mysql -u radius -p radius -e "SHOW TABLES;"

# Cek firewall rules
sudo ufw status numbered

# Cek listening ports
sudo ss -tulnp | grep -E '1812|1813|3000'

# Cek status service API
sudo systemctl status freeradius-api

# Cek log service API
journalctl -u freeradius-api --no-pager -n 50
```

### Reset Konfigurasi

Jika terjadi masalah serius:

```bash
# Backup konfigurasi
sudo cp -r /etc/freeradius/3.0 /etc/freeradius/3.0.backup

# Reset ke default
sudo apt-get purge --auto-remove freeradius freeradius-mysql
sudo rm -rf /etc/freeradius

# Install ulang
sudo ./install.sh
```

## 📚 Dokumentasi Tambahan

- **API Documentation**: `http://localhost:3000/api-docs` (Swagger UI)
- **FreeRADIUS Wiki**: [https://wiki.freeradius.org/](https://wiki.freeradius.org/)
- **MySQL Documentation**: [https://dev.mysql.com/doc/](https://dev.mysql.com/doc/)

## 🤝 Kontribusi

Kontribusi sangat diterima! Silakan:

1. Fork repository ini
2. Buat feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit perubahan (`git commit -m 'Add some AmazingFeature'`)
4. Push ke branch (`git push origin feature/AmazingFeature`)
5. Buat Pull Request

## 📄 Lisensi

Proyek ini dilisensikan di bawah MIT License - lihat file [LICENSE](LICENSE) untuk detail.

## 👨‍💻 Author

- **GitHub**: [@iamfafakkk](https://github.com/iamfafakkk)
- **Repository**: [minimalFreeRadius](https://github.com/iamfafakkk/minimalFreeRadius)

## 🙏 Acknowledgments

- [FreeRADIUS Project](https://freeradius.org/) untuk server RADIUS yang powerful
- [MySQL](https://www.mysql.com/) untuk database backend
- [Go](https://go.dev/) untuk REST API
- [chi](https://github.com/go-chi/chi) untuk routing HTTP
- [SvelteKit](https://svelte.dev/) + [shadcn-svelte](https://shadcn-svelte.com/) untuk panel web

---

**⚠️ Catatan Keamanan**: Pastikan untuk mengganti semua password default sebelum menggunakan di production environment!

**📞 Support**: Jika mengalami masalah, silakan buat issue di [GitHub Issues](https://github.com/iamfafakkk/minimalFreeRadius/issues)