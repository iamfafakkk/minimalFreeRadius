# Cloudflare SSL Configuration for FreeRADIUS API

This guide explains how to configure Cloudflare SSL with the FreeRADIUS API and Nginx reverse proxy.

## Overview

Cloudflare offers three SSL modes for securing traffic between visitors and your origin server:

1. **Flexible SSL**: Traffic is encrypted between the visitor and Cloudflare, but not between Cloudflare and your origin server.
2. **Full SSL**: Traffic is encrypted between the visitor and Cloudflare, and between Cloudflare and your origin server.
3. **Full SSL (Strict)**: Traffic is encrypted between the visitor and Cloudflare, and between Cloudflare and your origin server with strict certificate validation.

## Prerequisites

- A domain name pointed to Cloudflare nameservers
- Cloudflare account with your domain added
- FreeRADIUS API installed and running
- Nginx configured as a reverse proxy (using the provided setup script)

## Configuration Steps

### 1. Run the Setup Script with Nginx Configuration

```bash
# Navigate to the FreeRADIUS API directory
cd /path/to/freeradius-api

# Run the setup script with nginx-only option
./setup.sh --nginx-only
```

During the setup, you'll be prompted to:
1. Enter your domain name
2. Select your preferred Cloudflare SSL mode

### 2. Configure Cloudflare SSL/TLS Settings

Log in to your Cloudflare dashboard and navigate to your domain:

1. Go to **SSL/TLS > Overview**
2. Select the appropriate encryption mode based on your setup:
   - **Flexible**: If you chose Flexible SSL during setup
   - **Full**: If you chose Full SSL during setup
   - **Full (Strict)**: If you chose Full SSL (Strict) during setup

### 3. Configure DNS Records

In your Cloudflare dashboard:

1. Go to **DNS**
2. Add an A record pointing your domain to your server's IP address
3. Ensure the cloud icon is orange (proxied through Cloudflare)

## SSL Mode Details

### Flexible SSL (Not Recommended for Production)

- No SSL certificate required on your origin server
- Traffic between Cloudflare and your server is unencrypted
- Easiest to set up but least secure

### Full SSL

- Requires an SSL certificate on your origin server
- Self-signed certificates are acceptable
- Traffic is encrypted between all parties
- Good balance of security and ease of setup

### Full SSL (Strict) - Recommended

- Requires a valid SSL certificate on your origin server
- Certificate must be signed by a public CA or Cloudflare Origin CA
- Most secure option
- Recommended for production environments

## Certificate Management

### For Full SSL Mode

During setup, a self-signed certificate is automatically generated. For production use, replace this with a valid certificate:

```bash
# Replace with your valid certificate files
sudo cp your_domain.crt /etc/nginx/ssl/server.crt
sudo cp your_domain.key /etc/nginx/ssl/server.key
sudo nginx -t && sudo nginx -s reload
```

### For Full SSL (Strict) Mode

You have two options for obtaining a valid certificate:

#### Option 1: Cloudflare Origin CA

1. In Cloudflare dashboard, go to **SSL/TLS > Origin Server**
2. Click "Create Certificate"
3. Generate a certificate for your domain
4. Copy the certificate and private key to your server
5. Replace the self-signed certificate:

```bash
# Save the certificate and key to files
sudo nano /etc/nginx/ssl/server.crt  # Paste certificate
sudo nano /etc/nginx/ssl/server.key  # Paste private key

# Set proper permissions
sudo chmod 600 /etc/nginx/ssl/server.key
sudo chmod 644 /etc/nginx/ssl/server.crt

# Test and reload Nginx
sudo nginx -t && sudo nginx -s reload
```

#### Option 2: Let's Encrypt

```bash
# Install Certbot
sudo apt-get install certbot python3-certbot-nginx  # Ubuntu/Debian
# or
sudo yum install certbot python3-certbot-nginx      # CentOS/RHEL

# Obtain certificate
sudo certbot --nginx -d your-domain.com

# Set up auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

## Testing Your Configuration

After configuration, verify that everything is working correctly:

1. Visit your domain in a browser - you should see a secure connection
2. Check that the FreeRADIUS API endpoints are accessible
3. Verify that Cloudflare headers are being passed correctly

You can test the API health endpoint:
```bash
curl -I https://your-domain.com/health
```

## Troubleshooting

### Common Issues

1. **Mixed Content Errors**: Ensure all resources are loaded over HTTPS
2. **Certificate Errors**: Verify your certificate matches your domain
3. **502 Bad Gateway**: Check that the FreeRADIUS API is running and accessible

### Checking Nginx Logs

```bash
# Check Nginx error logs
sudo tail -f /var/log/nginx/error.log

# Check Nginx access logs
sudo tail -f /var/log/nginx/access.log
```

### Verifying Cloudflare Headers

The Nginx configuration passes Cloudflare headers to the application. You can verify these are working by checking the response headers:

```bash
curl -I https://your-domain.com/health
```

Look for headers like:
- `CF-RAY`
- `CF-IPCountry`
- `CF-Visitor`

## Security Considerations

1. **Always use Full SSL (Strict)** for production environments
2. **Keep your certificates up to date** with auto-renewal
3. **Regularly update Nginx** to the latest version
4. **Use strong cipher suites** (already configured in the provided nginx.conf)
5. **Enable HSTS** (already configured in the provided nginx.conf)

## Additional Resources

- [Cloudflare SSL Documentation](https://developers.cloudflare.com/ssl/)
- [Cloudflare Origin CA Documentation](https://developers.cloudflare.com/ssl/origin-configuration/origin-ca/)
- [Nginx Security Configuration](https://nginx.org/en/docs/)