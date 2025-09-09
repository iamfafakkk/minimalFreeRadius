# 📖 Swagger API Documentation Guide

## Overview

FreeRADIUS API menggunakan Swagger/OpenAPI 3.0 untuk dokumentasi API yang interaktif dan komprehensif. Dokumentasi ini memungkinkan Anda untuk:

- 📋 Melihat semua endpoint yang tersedia
- 🔍 Memahami struktur request dan response
- 🧪 Menguji API secara langsung dari browser
- 📝 Melihat contoh kode dan data

## Akses Dokumentasi

### 1. Swagger UI (Standard)
```
URL: http://localhost:3000/api-docs
```
Interface Swagger UI standar dengan semua fitur lengkap.

### 2. Swagger UI (Custom)
```
URL: http://localhost:3000/swagger
```
Interface kustom dengan desain yang lebih menarik dan informasi tambahan.

### 3. Swagger JSON
```
URL: http://localhost:3000/swagger.json
```
File JSON mentah untuk integrasi dengan tools lain.

## Fitur Dokumentasi

### 🔐 Authentication Testing
1. Klik tombol **"Authorize"** di bagian atas
2. Masukkan JWT token dengan format: `Bearer your-jwt-token`
3. Token akan otomatis ditambahkan ke semua request

### 🧪 Try It Out
1. Pilih endpoint yang ingin ditest
2. Klik **"Try it out"**
3. Isi parameter yang diperlukan
4. Klik **"Execute"**
5. Lihat response langsung di browser

### 📋 Model Schemas
- Lihat struktur data lengkap untuk request dan response
- Contoh data untuk setiap field
- Validasi rules dan constraints

## Endpoint Categories

### 🔐 Authentication
- `POST /api/v1/auth/login` - Login dan dapatkan JWT token
- `GET /api/v1/auth/verify` - Verifikasi token validity

### 🖥️ NAS Management
- `GET /api/v1/nas` - List semua NAS devices
- `POST /api/v1/nas` - Tambah NAS device baru
- `GET /api/v1/nas/{id}` - Detail NAS device
- `PUT /api/v1/nas/{id}` - Update NAS device
- `DELETE /api/v1/nas/{id}` - Hapus NAS device

### 👥 User Management
- `GET /api/v1/users` - List semua users
- `POST /api/v1/users` - Tambah user baru
- `GET /api/v1/users/{username}` - Detail user
- `PUT /api/v1/users/{username}` - Update user
- `DELETE /api/v1/users/{username}` - Hapus user
- `GET /api/v1/users/{username}/attributes` - List user attributes
- `POST /api/v1/users/{username}/attributes` - Tambah user attribute

### 💚 System Health
- `GET /api/v1/health` - Health check status

## Quick Start Testing

### 1. Login untuk mendapatkan token
```bash
curl -X POST "http://localhost:3000/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 2. Gunakan token untuk request lain
```bash
curl -X GET "http://localhost:3000/api/v1/nas" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. Test melalui Swagger UI
1. Buka http://localhost:3000/api-docs
2. Klik **"Authorize"**
3. Masukkan: `Bearer YOUR_JWT_TOKEN`
4. Test endpoint apapun dengan klik **"Try it out"**

## Response Format

Semua response menggunakan format JSON konsisten:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

### Paginated Response
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "pages": 10
  }
}
```

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request berhasil |
| 201 | Created - Resource berhasil dibuat |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Token tidak valid |
| 403 | Forbidden - Tidak memiliki permission |
| 404 | Not Found - Resource tidak ditemukan |
| 409 | Conflict - Resource sudah ada |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |

## Security Features

### 🔒 JWT Authentication
- Token expires dalam 24 jam (default)
- Refresh token mechanism
- Role-based access control

### 🛡️ Rate Limiting
- 100 requests per 15 menit per IP
- Login endpoint: 5 attempts per 15 menit
- API endpoints: 1000 requests per jam

### 🔐 Input Validation
- Joi schema validation
- SQL injection protection
- XSS protection
- CORS configuration

## Integration Examples

### JavaScript/Node.js
```javascript
const axios = require('axios');

// Login
const login = async () => {
  const response = await axios.post('http://localhost:3000/api/v1/auth/login', {
    username: 'admin',
    password: 'admin123'
  });
  return response.data.token;
};

// Get NAS devices
const getNasDevices = async (token) => {
  const response = await axios.get('http://localhost:3000/api/v1/nas', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.data;
};
```

### Python
```python
import requests

# Login
response = requests.post('http://localhost:3000/api/v1/auth/login', json={
    'username': 'admin',
    'password': 'admin123'
})
token = response.json()['token']

# Get NAS devices
headers = {'Authorization': f'Bearer {token}'}
response = requests.get('http://localhost:3000/api/v1/nas', headers=headers)
nas_devices = response.json()
```

### PHP
```php
<?php
// Login
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, 'http://localhost:3000/api/v1/auth/login');
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
    'username' => 'admin',
    'password' => 'admin123'
]));
curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$data = json_decode($response, true);
$token = $data['token'];
curl_close($ch);

// Get NAS devices
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, 'http://localhost:3000/api/v1/nas');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Authorization: Bearer ' . $token
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$nas_devices = json_decode($response, true);
curl_close($ch);
?>
```

## Troubleshooting

### Common Issues

1. **401 Unauthorized**
   - Pastikan JWT token valid dan belum expired
   - Format: `Bearer your-jwt-token`

2. **429 Too Many Requests**
   - Tunggu beberapa menit sebelum mencoba lagi
   - Implementasi exponential backoff

3. **400 Bad Request**
   - Periksa format JSON request
   - Validasi semua required fields

4. **CORS Issues**
   - Pastikan origin domain sudah dikonfigurasi
   - Check CORS_ORIGIN environment variable

### Debug Mode
Set `NODE_ENV=development` untuk error details lengkap.

## Support

Jika mengalami masalah:
1. Periksa [Installation Guide](INSTALLATION_GUIDE.md)
2. Lihat [API Documentation](API_DOCUMENTATION.md)
3. Check server logs untuk error details
4. Test dengan Postman atau curl untuk isolasi masalah

---

**Happy API Testing! 🚀**