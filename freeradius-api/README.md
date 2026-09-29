# FreeRADIUS REST API (Go)

REST API untuk mengelola FreeRADIUS server dengan fitur CRUD untuk NAS (Network Access Server) dan manajemen user. Port Go dari implementasi Node.js/Express sebelumnya — kontrak endpoint dan bentuk respons JSON dipertahankan.

## 🚀 Fitur Utama

- **CRUD Operations untuk NAS** - Kelola Network Access Server
- **CRUD Operations untuk User** - Kelola user authentication (radcheck & radreply)
- **JWT Authentication** - Keamanan berbasis token
- **API Key Authentication** - Alternatif autentikasi
- **Input Validation** - Validasi manual setara aturan Joi sebelumnya
- **Rate Limiting** - Perlindungan dari abuse (fixed-window per-IP)
- **CORS Support** - Cross-origin resource sharing
- **Comprehensive Logging** - Log sistem yang lengkap
- **Health Check** - Monitoring kesehatan API
- **Documentation** - Dokumentasi API yang lengkap (`swagger.json`, `/docs`)

## 📋 Persyaratan Sistem

- **Go** 1.18 atau lebih baru
- **MySQL** 5.7+ atau MariaDB 10.3+
- **Linux** (Ubuntu 18.04+, CentOS 7+, Debian 9+)
- **Memory** Minimum 512MB RAM
- **Storage** Minimum 1GB free space

## 🛠️ Instalasi Cepat

### 1. Clone Repository

```bash
git clone <repository-url> freeradius-api
cd freeradius-api
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build & Run

```bash
cp .env.example .env   # sesuaikan kredensial DB
go build -o freeradius-api .
./freeradius-api
```

Atau setup otomatis (env, build, verifikasi DB, systemd):

```bash
sudo ./setup.sh
```

### 3. Setup Database

Database `radius`, user `radius`, dan schema FreeRADIUS (schema resmi
`/etc/freeradius/3.0/mods-config/sql/main/mysql/schema.sql`, termasuk
`radcheck`, `radreply`, `nas`, `radacct`, `radpostauth`) dibuat oleh
`install.sh` di root repo. Jalankan itu lebih dulu; `setup.sh` hanya
memverifikasi schema.

### 4. Konfigurasi Environment

```bash
# Copy dan edit file environment
cp .env.example .env
nano .env
```

**Contoh konfigurasi .env:**
```env
DB_HOST=localhost
DB_PORT=3306
DB_NAME=radius
DB_USER=radius
DB_PASSWORD=radiuspass123!

PORT=3000
NODE_ENV=development

JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRES_IN=24h

ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123!
```

### 5. Jalankan Aplikasi (systemd)

```bash
# Build + install service + start + health check
sudo ./setup.sh
```

Service dikelola systemd (wajib). Perintah manual:

```bash
CGO_ENABLED=1 go build -o freeradius-api .   # build saja
sudo ./setup.sh --no-start                    # install service tanpa start
```

### 6. Build Frontend (opsional, panel web)

Backend menyajikan hasil `npm run build` dari `../freeradius-web/build`; kalau
belum ada, panel jalan API-only. `setup.sh` membangunnya otomatis bila npm ada.

```bash
cd ../freeradius-web && npm install && npm run build
```

## 📚 Dokumentasi

- **[API Documentation](docs/API_DOCUMENTATION.md)** - Dokumentasi lengkap endpoint API
- **[Installation Guide](docs/INSTALLATION_GUIDE.md)** - Panduan instalasi dan deployment
- **[Cloudflare SSL Configuration](docs/CLOUDFLARE_SSL_CONFIGURATION.md)** - Panduan konfigurasi SSL dengan Cloudflare

## 🔗 Endpoint Utama

### Authentication
- `POST /api/v1/auth/login` - Login dan dapatkan JWT token
- `GET /api/v1/auth/verify` - Verifikasi token
- `GET /api/v1/auth/health` - Health check

### NAS Management
- `GET /api/v1/nas` - Daftar semua NAS
- `GET /api/v1/nas/:id` - Detail NAS
- `POST /api/v1/nas` - Buat NAS baru
- `PUT /api/v1/nas/:id` - Update NAS
- `DELETE /api/v1/nas/:id` - Hapus NAS

### User Management
- `GET /api/v1/users` - Daftar semua user
- `GET /api/v1/users/:username` - Detail user
- `POST /api/v1/users` - Buat user baru
- `PUT /api/v1/users/:username` - Update user
- `DELETE /api/v1/users/:username` - Hapus user

## 🧪 Testing API

### 1. Health Check

```bash
curl -X GET http://localhost:3000/api/v1/auth/health
```

### 2. Login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123!"}'
```

### 3. Buat NAS

```bash
# Gunakan token dari login
TOKEN="your-jwt-token-here"

curl -X POST http://localhost:3000/api/v1/nas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "router1",
    "ip": "192.168.1.1",
    "secret": "secret123",
    "type": "cisco",
    "description": "Main router"
  }'
```

### 4. Buat User

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "password": "testpass123",
    "profile": "PPP"
  }'
```

## 📁 Struktur Proyek

```
freeradius-api/
├── internal/
│   ├── appdb/          # SQLite app DB (admin login, login history, audit)
│   ├── config/         # Konfigurasi env + resolusi WebDir
│   ├── database/       # Koneksi MySQL
│   ├── handlers/       # Handler HTTP (auth, nas, users, system, radius log)
│   ├── middleware/     # Auth, rate limit, security headers, activity log
│   ├── models/         # Tipe data NAS/user
│   ├── radiox/         # Tailer radius.log (SSE) + counters
│   └── router/         # Routing chi + penyajian SPA
├── database/           # mysql.cnf (client config)
├── docs/               # Dokumentasi + aset Swagger UI
├── nginx/nginx.conf    # Contoh reverse proxy (opsional)
├── main.go             # Entry point
├── setup.sh            # Setup env + build + systemd
├── swagger.json        # OpenAPI spec
├── .env.example        # Contoh environment
└── README.md           # File ini
```

## 🔒 Keamanan

- **JWT Authentication** - Token berbasis keamanan
- **API Key Support** - Alternatif autentikasi
- **Rate Limiting** - 1000 requests per 15 menit per IP (fixed-window)
- **Input Validation** - Validasi manual (setara aturan Joi sebelumnya)
- **CORS Protection** - Konfigurasi CORS yang aman
- **Security Headers** - CSP ketat, dibangun di middleware
- **Password Hashing** - bcrypt untuk akun admin panel

## 🚀 Deployment

Wajib systemd. `setup.sh` membuat unit `freeradius-api.service`, enable saat
boot, lalu start. Service berjalan sebagai **root** karena panel memakai
`systemctl restart freeradius` (tombol Restart) dan menulis ke
`/var/log/freeradius/radius.log` (tombol Clear Log File).

```bash
sudo ./setup.sh                  # build + install + start
sudo ./setup.sh --systemd-only   # build + install service saja
sudo ./setup.sh --no-start       # install tanpa start
sudo ./setup.sh --remove-systemd # stop + disable + hapus
```

```bash
sudo systemctl status freeradius-api
sudo systemctl restart freeradius-api
journalctl -u freeradius-api -f
```

Nginx/Cloudflare SSL dapat ditambahkan sebagai reverse proxy di depan service
(bind ke port API); konfigurasi contoh ada di `nginx/nginx.conf`. Tidak ada
opsi `--nginx-only` di `setup.sh`.

## 📊 Monitoring

### Health Check Endpoint

```bash
curl http://localhost:3000/api/v1/auth/health
```

**Response:**
```json
{
  "success": true,
  "message": "API is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "uptime": 3600,
    "database": "connected",
    "memory_usage": {...},
    "go_version": "go1.18.1"
  }
}
```

### Log Files

- **Service logs:** `journalctl -u freeradius-api`
- **Radius log (Live Logs page):** `/var/log/freeradius/radius.log`
- **Installation log:** `/tmp/freeradius_install.log`

## 🔧 Konfigurasi

### Environment Variables

| Variable | Description | Default |
|----------|-------------|----------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `DB_NAME` | Database name | `radius` |
| `DB_USER` | Database user | `radius` |
| `DB_PASSWORD` | Database password | - |
| `PORT` | Server port | `3000` |
| `NODE_ENV` | Environment | `development` |
| `JWT_SECRET` | JWT secret key | - |
| `JWT_EXPIRES_IN` | JWT expiration | `24h` |
| `API_PREFIX` | API prefix | `/api/v1` |
| `RATE_LIMIT_WINDOW_MS` | Rate limit window | `900000` |
| `RATE_LIMIT_MAX_REQUESTS` | Max requests per window | `100` |
| `CORS_ORIGIN` | CORS origin | `*` |
| `ADMIN_USERNAME` | Admin username | `admin` |
| `ADMIN_PASSWORD` | Admin password | - |
| `APP_DB_PATH` | SQLite app DB (login, audit) | `freeradius.db` |
| `RADIUS_LOG` | FreeRADIUS log yang di-tail | `/var/log/freeradius/radius.log` |
| `RADIUS_TEST_ADDR` | Target auth test | `127.0.0.1:1812` |
| `RADIUS_TEST_SECRET` | Secret client untuk auth test | `testing123` |
| `FREERADIUS_RELOAD_CMD` | Dijalankan setelah NAS berubah | `systemctl restart freeradius` |

## 🐛 Troubleshooting

### Database Connection Issues

```bash
# Test koneksi database (kredensial dari .env)
mysql -h"${DB_HOST:-localhost}" -u"${DB_USER:-radius}" -p"${DB_PASSWORD:-radiuspass123!}" "${DB_NAME:-radius}" -e "SELECT 1;"
```

### Port Already in Use

```bash
# Find process using port
sudo lsof -i :3000

# Kill process
sudo kill -9 <PID>
```

### Permission Issues

```bash
# Fix permissions
sudo chown -R $USER:$USER .
chmod 600 .env
```

## 📝 Contributing

1. Fork repository
2. Buat feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

## 📄 License

MIT License - lihat file [LICENSE](LICENSE) untuk detail.

## 🤝 Support

Jika Anda mengalami masalah atau memiliki pertanyaan:

1. Periksa [dokumentasi](docs/)
2. Lihat [troubleshooting guide](docs/INSTALLATION_GUIDE.md#troubleshooting)
3. Buat issue di repository

## 🔄 Changelog

### v1.0.0
- ✅ Initial release
- ✅ CRUD operations untuk NAS
- ✅ CRUD operations untuk User (radcheck/radreply)
- ✅ JWT Authentication
- ✅ API Key Authentication
- ✅ Input validation
- ✅ Rate limiting
- ✅ CORS support
- ✅ Health check endpoint
- ✅ Comprehensive documentation

---

**Dibuat dengan ❤️ untuk komunitas FreeRADIUS**