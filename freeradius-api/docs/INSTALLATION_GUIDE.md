# FreeRADIUS API Installation Guide

## Prerequisites

### System Requirements

- **Operating System:** Linux (Ubuntu 18.04+, CentOS 7+, Debian 9+)
- **Node.js:** Version 16.x atau lebih baru
- **MySQL:** Version 5.7+ atau MariaDB 10.3+
- **Memory:** Minimum 512MB RAM
- **Storage:** Minimum 1GB free space

### Required Software

1. **Node.js dan npm**
   ```bash
   # Ubuntu/Debian
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
   sudo apt-get install -y nodejs
   
   # CentOS/RHEL
   curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
   sudo yum install -y nodejs
   ```

2. **MySQL/MariaDB**
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install mysql-server
   
   # CentOS/RHEL
   sudo yum install mysql-server
   sudo systemctl start mysqld
   sudo systemctl enable mysqld
   ```

3. **Git**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install git
   
   # CentOS/RHEL
   sudo yum install git
   ```

## Database Setup

### 1. Create Database and User

```sql
-- Login ke MySQL sebagai root
mysql -u root -p

-- Buat database
CREATE DATABASE radius;

-- Buat user untuk FreeRADIUS
CREATE USER 'radius'@'localhost' IDENTIFIED BY 'radiuspass123!';
GRANT ALL PRIVILEGES ON radius.* TO 'radius'@'localhost';
FLUSH PRIVILEGES;

-- Keluar dari MySQL
EXIT;
```

### 2. Create FreeRADIUS Tables

```sql
-- Login dengan user radius
mysql -u radius -p radius

-- Buat tabel nas
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

-- Buat tabel radcheck
CREATE TABLE radcheck (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  username varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '==',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY username (username(32))
);

-- Buat tabel radreply
CREATE TABLE radreply (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  username varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '=',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY username (username(32))
);

-- Buat tabel radgroupcheck (optional)
CREATE TABLE radgroupcheck (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  groupname varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '==',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY groupname (groupname(32))
);

-- Buat tabel radgroupreply (optional)
CREATE TABLE radgroupreply (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  groupname varchar(64) NOT NULL DEFAULT '',
  attribute varchar(64) NOT NULL DEFAULT '',
  op char(2) NOT NULL DEFAULT '=',
  value varchar(253) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY groupname (groupname(32))
);

-- Buat tabel radusergroup (optional)
CREATE TABLE radusergroup (
  username varchar(64) NOT NULL DEFAULT '',
  groupname varchar(64) NOT NULL DEFAULT '',
  priority int(11) NOT NULL DEFAULT '1',
  KEY username (username(32))
);

-- Insert sample data
INSERT INTO nas (nasname, shortname, type, ports, secret, description) VALUES
('127.0.0.1', 'localhost', 'other', 1812, 'testing123', 'Local test server'),
('192.168.1.1', 'router1', 'cisco', 1812, 'secret123', 'Main router');

INSERT INTO radcheck (username, attribute, op, value) VALUES
('testuser', 'Cleartext-Password', ':=', 'testpass'),
('admin', 'Cleartext-Password', ':=', 'admin123!');

INSERT INTO radreply (username, attribute, op, value) VALUES
('testuser', 'Framed-Protocol', ':=', 'PPP'),
('admin', 'Framed-Protocol', ':=', 'PPP');

EXIT;
```

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
npm install
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
JWT_EXPIRES_IN=24h

# API Configuration
API_PREFIX=/api/v1

# Rate Limiting
RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=100

# CORS Configuration
CORS_ORIGIN=*

# Admin User
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123!
```

### 4. Test Database Connection

```bash
# Test koneksi database
node -e "const db = require('./src/config/database'); db.testConnection().then(() => console.log('Database connected')).catch(console.error);"
```

### 5. Start Application

```bash
# Development mode
npm run dev

# Production mode
npm start
```

## Production Deployment

### 1. Using PM2 (Recommended)

```bash
# Install PM2 globally
npm install -g pm2

# Create PM2 ecosystem file
cat > ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'freeradius-api',
    script: 'server.js',
    instances: 'max',
    exec_mode: 'cluster',
    env: {
      NODE_ENV: 'development'
    },
    env_production: {
      NODE_ENV: 'production',
      PORT: 3000
    },
    error_file: './logs/err.log',
    out_file: './logs/out.log',
    log_file: './logs/combined.log',
    time: true
  }]
};
EOF

# Create logs directory
mkdir -p logs

# Start with PM2
pm2 start ecosystem.config.js --env production

# Save PM2 configuration
pm2 save

# Setup PM2 startup
pm2 startup
```

### 2. Using Systemd Service

```bash
# Create systemd service file
sudo tee /etc/systemd/system/freeradius-api.service > /dev/null << 'EOF'
[Unit]
Description=FreeRADIUS API Server
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/freeradius-api
EnvironmentFile=/opt/freeradius-api/.env
ExecStart=/usr/bin/node server.js
Restart=on-failure
RestartSec=10
KillMode=process

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
sudo systemctl daemon-reload
sudo systemctl enable freeradius-api
sudo systemctl start freeradius-api

# Check status
sudo systemctl status freeradius-api
```

### 3. Using Docker

**Dockerfile:**
```dockerfile
FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Create non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nodejs -u 1001

# Change ownership
RUN chown -R nodejs:nodejs /app
USER nodejs

EXPOSE 3000

CMD ["node", "server.js"]
```

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  freeradius-api:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DB_HOST=mysql
      - DB_NAME=radius
      - DB_USER=radius
      - DB_PASSWORD=radiuspass123!
    depends_on:
      - mysql
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=rootpass123!
      - MYSQL_DATABASE=radius
      - MYSQL_USER=radius
      - MYSQL_PASSWORD=radiuspass123!
    volumes:
      - mysql_data:/var/lib/mysql
      - ./database/init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped

volumes:
  mysql_data:
```

**Deploy dengan Docker:**
```bash
# Build dan start
docker-compose up -d

# Check logs
docker-compose logs -f freeradius-api
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
chown www-data:www-data .env
```

## Monitoring and Logging

### 1. Log Rotation

```bash
# Create logrotate configuration
sudo tee /etc/logrotate.d/freeradius-api > /dev/null << 'EOF'
/opt/freeradius-api/logs/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
    postrotate
        systemctl reload freeradius-api
    endscript
}
EOF
```

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
# Test database connectivity
node -e "
const db = require('./src/config/database');
db.testConnection()
  .then(() => console.log('✓ Database connection successful'))
  .catch(err => console.error('✗ Database connection failed:', err.message));
"
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
   # Fix file permissions
   sudo chown -R www-data:www-data /opt/freeradius-api
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

- **Application Logs:** `/opt/freeradius-api/logs/`
- **PM2 Logs:** `~/.pm2/logs/`
- **Nginx Logs:** `/var/log/nginx/`
- **MySQL Logs:** `/var/log/mysql/`
- **System Logs:** `/var/log/syslog`

### Performance Tuning

1. **Node.js Optimization**
   ```bash
   # Set NODE_ENV to production
   export NODE_ENV=production
   
   # Increase memory limit if needed
   node --max-old-space-size=4096 server.js
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
  --exclude=node_modules \
  --exclude=logs \
  /opt/freeradius-api
```

Setelah mengikuti panduan ini, FreeRADIUS API akan berjalan dengan aman dan optimal di lingkungan production Anda.