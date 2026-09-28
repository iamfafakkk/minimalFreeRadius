# FreeRADIUS API Installation Guide

## Prerequisites

### System Requirements

- **Operating System:** Linux (Ubuntu 20.04+/Debian 11+)
- **Go:** Version 1.18 atau lebih baru
- **C toolchain:** gcc + build-essential (CGO, driver SQLite)
- **MySQL:** Version 5.7+ atau MariaDB 10.3+
- **systemd:** wajib (service dijalankan sebagai unit systemd)
- **Memory:** Minimum 512MB RAM
- **Storage:** Minimum 1GB free space

### Required Software

1. **Go + build tools**
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install -y golang-go build-essential
   ```

2. **MySQL/MariaDB**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install -y mysql-server
   ```

3. **Git**
   ```bash
   sudo apt-get install -y git
   ```

> Untuk instalasi RADIUS + MySQL + schema lengkap, jalankan `install.sh` di
> root repo (lihat [README](../../README.md)). Panduan ini fokus pada panel API.

## Database Setup

Database `radius`, user `radius`, dan schema FreeRADIUS dibuat oleh `install.sh`
(root repo) dengan schema resmi MySQL dari
`/etc/freeradius/3.0/mods-config/sql/main/mysql/schema.sql` — mencakup
`nas`, `radcheck`, `radreply`, `radacct`, dan `radpostauth`.

```bash
# Di root repo
sudo ./install.sh
```

`setup.sh` hanya **memverifikasi** schema, tidak membuat/menghapus tabel.

## Application Installation

### 1. Clone or Download Source Code

```bash
# Jika menggunakan Git
git clone <repository-url> freeradius-api
cd freeradius-api

# Atau extract dari archive
tar -xzf freeradius-api.tar.gz
cd freeradius-api
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

```bash
# Copy file environment
cp .env.example .env

# Edit konfigurasi
nano .env
```

**Konfigurasi .env:**
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=radius
DB_USER=radius
DB_PASSWORD=radiuspass123!

# Server Configuration
PORT=3000
NODE_ENV=production

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

### 4. Build & Install Service (systemd)

```bash
# Build + install unit freeradius-api.service + start + health check
sudo ./setup.sh
```

Opsi lain:

```bash
CGO_ENABLED=1 go build -o freeradius-api .   # build saja
sudo ./setup.sh --no-start                    # install service tanpa start
sudo ./setup.sh --systemd-only                # build + install service saja
sudo ./setup.sh --remove-systemd              # hapus service
```

### 5. Build Frontend (opsional)

Backend menyajikan build statis dari `../freeradius-web/build`:

```bash
cd ../freeradius-web && npm install && npm run build
```

## Reverse Proxy Setup (Nginx)

### 1. Install Nginx

```bash
# Ubuntu/Debian
sudo apt-get install nginx

# CentOS/RHEL
sudo yum install nginx
```

### 2. Configure Nginx

```bash
# Create Nginx configuration
sudo tee /etc/nginx/sites-available/freeradius-api > /dev/null << 'EOF'
server {
    listen 80;
    server_name your-domain.com;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "strict-origin-when-cross-origin";

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://127.0.0.1:3000/api/v1/auth/health;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/freeradius-api /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

### 3. SSL Certificate (Let's Encrypt)

```bash
# Install Certbot
sudo apt-get install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add line: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 4. Cloudflare SSL Configuration

For detailed instructions on configuring Cloudflare SSL with the FreeRADIUS API, please refer to the [Cloudflare SSL Configuration Guide](CLOUDFLARE_SSL_CONFIGURATION.md).

## Security Hardening

### 1. Firewall Configuration

```bash
# UFW (Ubuntu)
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable

# Firewalld (CentOS)
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 2. Database Security

```bash
# Run MySQL secure installation
sudo mysql_secure_installation

# Create dedicated database user with limited privileges
mysql -u root -p
CREATE USER 'radius_api'@'localhost' IDENTIFIED BY 'strong_password_here';
GRANT SELECT, INSERT, UPDATE, DELETE ON radius.nas TO 'radius_api'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON radius.radcheck TO 'radius_api'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON radius.radreply TO 'radius_api'@'localhost';
FLUSH PRIVILEGES;
```

### 3. Application Security

```bash
# Update .env with strong secrets
JWT_SECRET=$(openssl rand -base64 64)
echo "JWT_SECRET=$JWT_SECRET" >> .env

# Set proper file permissions
chmod 600 .env
chown root:root .env
```

## Monitoring and Logging

### 1. Log Rotation (journald)

Service menulis ke journald (`StandardOutput=journal`), bukan file aplikasi.
Batasi umur log di `/etc/systemd/journald.conf` (mis. `SystemMaxUse=200M`,
`MaxRetentionSec=2week`) lalu `sudo systemctl restart systemd-journald`.

Radius log punya logrotate sendiri dari paket FreeRADIUS.

### 2. Health Monitoring

```bash
# Create health check script
cat > /usr/local/bin/freeradius-api-health.sh << 'EOF'
#!/bin/bash
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/v1/auth/health)
if [ $response -eq 200 ]; then
    echo "API is healthy"
    exit 0
else
    echo "API is unhealthy (HTTP $response)"
    exit 1
fi
EOF

chmod +x /usr/local/bin/freeradius-api-health.sh

# Add to crontab for monitoring
echo "*/5 * * * * /usr/local/bin/freeradius-api-health.sh" | crontab -
```

## Testing Installation

### 1. API Health Check

```bash
curl -X GET http://localhost:3000/api/v1/auth/health
```

### 2. Authentication Test

```bash
# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123!"}'

# Use returned token for authenticated requests
TOKEN="your-jwt-token-here"
curl -X GET http://localhost:3000/api/v1/nas \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Database Connection Test

```bash
# Test koneksi database (kredensial dari .env)
mysql -h"${DB_HOST:-localhost}" -u"${DB_USER:-radius}" -p"${DB_PASSWORD:-radiuspass123!}" "${DB_NAME:-radius}" -e "SELECT COUNT(*) FROM radcheck;"
```

## Troubleshooting

### Common Issues

1. **Database Connection Error**
   ```bash
   # Check MySQL service
   sudo systemctl status mysql
   
   # Check database credentials
   mysql -u radius -p radius
   ```

2. **Port Already in Use**
   ```bash
   # Find process using port 3000
   sudo lsof -i :3000
   
   # Kill process if needed
   sudo kill -9 <PID>
   ```

3. **Permission Denied**
   ```bash
   # Fix file permissions (service berjalan sebagai root/freerad)
   sudo chown -R root:root /opt/freeradius-api
   sudo chmod -R 755 /opt/freeradius-api
   sudo chmod 600 /opt/freeradius-api/.env
   ```

4. **Memory Issues**
   ```bash
   # Check memory usage
   free -h
   
   # Increase swap if needed
   sudo fallocate -l 1G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   ```

### Log Locations

- **API Logs:** `journalctl -u freeradius-api`
- **Nginx Logs:** `/var/log/nginx/`
- **MySQL Logs:** `/var/log/mysql/`
- **FreeRADIUS Log:** `/var/log/freeradius/radius.log`
- **System Logs:** `/var/log/syslog`

### Performance Tuning

1. **Go Build**
   ```bash
   # Build binary teroptimasi
   CGO_ENABLED=1 go build -ldflags="-s -w" -o freeradius-api .
   ```

2. **MySQL Optimization**
   ```sql
   -- Add indexes for better performance
   CREATE INDEX idx_radcheck_username ON radcheck(username);
   CREATE INDEX idx_radreply_username ON radreply(username);
   CREATE INDEX idx_nas_nasname ON nas(nasname);
   ```

3. **Nginx Optimization**
   ```nginx
   # Add to nginx.conf
   worker_processes auto;
   worker_connections 1024;
   keepalive_timeout 65;
   gzip on;
   gzip_types text/plain application/json;
   ```

## Backup and Recovery

### Database Backup

```bash
# Create backup script
cat > /usr/local/bin/backup-radius-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/radius"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
mysqldump -u radius -p'radiuspass123!' radius > $BACKUP_DIR/radius_$DATE.sql
gzip $BACKUP_DIR/radius_$DATE.sql

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
EOF

chmod +x /usr/local/bin/backup-radius-db.sh

# Schedule daily backup
echo "0 2 * * * /usr/local/bin/backup-radius-db.sh" | crontab -
```

### Application Backup

```bash
# Backup application files
tar -czf /opt/backups/freeradius-api_$(date +%Y%m%d).tar.gz \
  --exclude=freeradius-api \
  --exclude=freeradius.db \
  --exclude=logs \
  /opt/freeradius-api
```

Setelah mengikuti panduan ini, FreeRADIUS API akan berjalan dengan aman dan optimal di lingkungan production Anda.