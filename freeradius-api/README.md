# FreeRADIUS REST API

REST API untuk mengelola FreeRADIUS server dengan fitur CRUD untuk NAS (Network Access Server) dan manajemen user.

## 🚀 Fitur Utama

- **CRUD Operations untuk NAS** - Kelola Network Access Server
- **CRUD Operations untuk User** - Kelola user authentication (radcheck & radreply)
- **JWT Authentication** - Keamanan berbasis token
- **API Key Authentication** - Alternatif autentikasi
- **Input Validation** - Validasi data menggunakan Joi
- **Rate Limiting** - Perlindungan dari abuse
- **CORS Support** - Cross-origin resource sharing
- **Comprehensive Logging** - Log sistem yang lengkap
- **Health Check** - Monitoring kesehatan API
- **Documentation** - Dokumentasi API yang lengkap

## 📋 Persyaratan Sistem

- **Node.js** 16.x atau lebih baru
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
npm install
```

### 3. Setup Database

```sql
-- Buat database dan user
CREATE DATABASE radius;
CREATE USER 'radius'@'localhost' IDENTIFIED BY 'radiuspass123!';
GRANT ALL PRIVILEGES ON radius.* TO 'radius'@'localhost';
FLUSH PRIVILEGES;
```

```sql
-- Buat tabel yang diperlukan
USE radius;

-- Tabel NAS
CREATE TABLE nas (
  id int(10) NOT NULL AUTO_INCREMENT,
  nasname varchar(128) NOT NULL,
  shortname varchar(32),
  type varchar(30) DEFAULT 'other',
  ports int(5),
  secret varchar(60) DEFAULT 'secret' NOT NULL,
  server varchar(64),
  community varchar(50),
  description varchar(200) DEFAULT 'RADIUS Client',
  PRIMARY KEY (id),
  KEY nasname (nasname)
);

-- Tabel radcheck
CREATE TABLE radcheck (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  username varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '==',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY username (username(32))
);

-- Tabel radreply
CREATE TABLE radreply (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  username varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '=',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY username (username(32))
);
```

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

### 5. Jalankan Aplikasi

```bash
# Development mode
npm run dev

# Production mode
npm start
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
├── src/
│   ├── config/
│   │   └── database.js          # Konfigurasi database
│   ├── controllers/
│   │   ├── authController.js    # Controller autentikasi
│   │   ├── nasController.js     # Controller NAS
│   │   └── userController.js    # Controller user
│   ├── middleware/
│   │   ├── auth.js             # Middleware autentikasi
│   │   └── validation.js       # Middleware validasi
│   ├── models/
│   │   ├── NasModel.js         # Model NAS
│   │   └── UserModel.js        # Model user
│   ├── routes/
│   │   ├── authRoutes.js       # Routes autentikasi
│   │   ├── nasRoutes.js        # Routes NAS
│   │   └── userRoutes.js       # Routes user
│   └── utils/
├── docs/
│   ├── API_DOCUMENTATION.md    # Dokumentasi API
│   └── INSTALLATION_GUIDE.md   # Panduan instalasi
├── logs/                       # Directory log
├── .env                        # Environment variables
├── .env.example               # Contoh environment
├── package.json               # Dependencies
├── server.js                  # Entry point
└── README.md                  # File ini
```

## 🔒 Keamanan

- **JWT Authentication** - Token berbasis keamanan
- **API Key Support** - Alternatif autentikasi
- **Rate Limiting** - 100 requests per 15 menit per IP
- **Input Validation** - Validasi semua input menggunakan Joi
- **CORS Protection** - Konfigurasi CORS yang aman
- **Security Headers** - Helmet.js untuk security headers
- **Password Hashing** - Bcrypt untuk hash password (jika diperlukan)

## 🚀 Deployment

### Menggunakan PM2

```bash
# Install PM2
npm install -g pm2

# Start aplikasi
pm2 start server.js --name freeradius-api

# Monitor
pm2 monit

# Logs
pm2 logs freeradius-api
```

### Menggunakan Docker

```bash
# Build image
docker build -t freeradius-api .

# Run container
docker run -d \
  --name freeradius-api \
  -p 3000:3000 \
  -e DB_HOST=your-db-host \
  -e DB_USER=radius \
  -e DB_PASSWORD=radiuspass123! \
  freeradius-api
```

### Menggunakan Docker Compose

```bash
# Start semua services
docker-compose up -d

# Check logs
docker-compose logs -f
```

### Menggunakan Nginx Reverse Proxy dengan Cloudflare SSL

```bash
# Jalankan setup script dengan opsi nginx-only
./setup.sh --nginx-only

# Ikuti petunjuk untuk mengkonfigurasi domain dan mode SSL Cloudflare
```

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
    "timestamp": "2024-01-01T12:00:00.000Z",
    "uptime": 3600,
    "database": "connected",
    "memory_usage": {...},
    "node_version": "v18.17.0"
  }
}
```

### Log Files

- **Application Logs:** `logs/app.log`
- **Error Logs:** `logs/error.log`
- **Access Logs:** `logs/access.log`

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

## 🐛 Troubleshooting

### Database Connection Issues

```bash
# Test database connection
node -e "const db = require('./src/config/database'); db.testConnection().then(() => console.log('OK')).catch(console.error);"
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