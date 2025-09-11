# FreeRADIUS API Documentation

## Overview

FreeRADIUS API adalah REST API untuk mengelola FreeRADIUS server dengan fitur CRUD untuk NAS (Network Access Server) dan manajemen user.

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

API ini mendukung dua metode autentikasi:

### 1. JWT Token (Recommended)

```bash
Authorization: Bearer <your-jwt-token>
```

### 2. API Key

```bash
X-API-Key: <your-api-key>
```

## Response Format

Semua response menggunakan format JSON dengan struktur berikut:

```json
{
  "success": true|false,
  "message": "Description of the result",
  "data": {}, // Response data (optional)
  "errors": [] // Validation errors (optional)
}
```

## Authentication Endpoints

### Login

Mendapatkan JWT token untuk autentikasi.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin123!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "username": "admin",
      "role": "admin"
    },
    "expires_in": "24h"
  }
}
```

### Verify Token

Memverifikasi validitas JWT token.

**Endpoint:** `GET /auth/verify`

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "Token is valid",
  "data": {
    "user": {
      "username": "admin",
      "role": "admin",
      "iat": 1640995200
    },
    "valid": true
  }
}
```

### API Info

Mendapatkan informasi tentang API.

**Endpoint:** `GET /auth/info`

**Response:**
```json
{
  "success": true,
  "message": "API information retrieved successfully",
  "data": {
    "name": "FreeRADIUS API",
    "version": "1.0.0",
    "description": "REST API for FreeRADIUS management",
    "endpoints": {
      "authentication": {
        "login": "POST /api/v1/auth/login",
        "verify": "GET /api/v1/auth/verify"
      },
      "nas": {
        "list": "GET /api/v1/nas",
        "get": "GET /api/v1/nas/:id",
        "create": "POST /api/v1/nas",
        "update": "PUT /api/v1/nas/:id",
        "delete": "DELETE /api/v1/nas/:id"
      },
      "users": {
        "list": "GET /api/v1/users",
        "get": "GET /api/v1/users/:username",
        "getById": "GET /api/v1/users/id/:id",
        "create": "POST /api/v1/users",
        "update": "PUT /api/v1/users/:username",
        "updateById": "PUT /api/v1/users/id/:id",
        "delete": "DELETE /api/v1/users/:username"
      }
    }
  }
}
```

### Health Check

Memeriksa status kesehatan API.

**Endpoint:** `GET /auth/health`

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
    "memory_usage": {
      "rss": 50331648,
      "heapTotal": 20971520,
      "heapUsed": 15728640,
      "external": 1048576
    },
    "node_version": "v18.17.0"
  }
}
```

## NAS (Network Access Server) Endpoints

### Get All NAS

Mendapatkan daftar semua NAS.

**Endpoint:** `GET /nas`

**Query Parameters:**
- `search` (optional): Pencarian berdasarkan nama, IP, atau deskripsi
- `page` (optional): Nomor halaman (default: 1)
- `limit` (optional): Jumlah item per halaman (default: 10, max: 100)

**Example Request:**
```bash
GET /api/v1/nas?search=router&page=1&limit=10
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "NAS entries retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "router1",
      "ip": "192.168.1.1",
      "secret": "secret123",
      "type": "cisco",
      "ports": 1812,
      "community": "public",
      "description": "Main router"
    }
  ],
  "count": 1
}
```

### Get NAS by ID

Mendapatkan NAS berdasarkan ID.

**Endpoint:** `GET /nas/:id`

**Example Request:**
```bash
GET /api/v1/nas/1
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "NAS retrieved successfully",
  "data": {
    "id": 1,
    "name": "router1",
    "ip": "192.168.1.1",
    "secret": "secret123",
    "type": "cisco",
    "ports": 1812,
    "community": "public",
    "description": "Main router"
  }
}
```

### Create NAS

Membuat NAS baru.

**Endpoint:** `POST /nas`

**Request Body:**
```json
{
  "name": "router2",
  "ip": "192.168.1.2",
  "secret": "secret456",
  "type": "cisco",
  "ports": 1812,
  "community": "public",
  "description": "Secondary router"
}
```

**Response:**
```json
{
  "success": true,
  "message": "NAS created successfully",
  "data": {
    "id": 2,
    "name": "router2",
    "ip": "192.168.1.2",
    "secret": "secret456",
    "type": "cisco",
    "ports": 1812,
    "community": "public",
    "description": "Secondary router"
  }
}
```

### Update NAS

Memperbarui NAS yang ada.

**Endpoint:** `PUT /nas/:id`

**Request Body:**
```json
{
  "name": "router2_updated",
  "description": "Updated secondary router"
}
```

**Response:**
```json
{
  "success": true,
  "message": "NAS updated successfully",
  "data": {
    "id": 2,
    "name": "router2_updated",
    "ip": "192.168.1.2",
    "secret": "secret456",
    "type": "cisco",
    "ports": 1812,
    "community": "public",
    "description": "Updated secondary router"
  }
}
```

### Delete NAS

Menghapus NAS.

**Endpoint:** `DELETE /nas/:id`

**Example Request:**
```bash
DELETE /api/v1/nas/2
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "NAS deleted successfully"
}
```

### Get NAS Statistics

Mendapatkan statistik NAS.

**Endpoint:** `GET /nas/stats`

**Response:**
```json
{
  "success": true,
  "message": "NAS statistics retrieved successfully",
  "data": {
    "total_nas": 5
  }
}
```

## User Management Endpoints

### Get All Users

Mendapatkan daftar semua user.

**Endpoint:** `GET /users`

**Query Parameters:**
- `search` (optional): Pencarian berdasarkan username
- `page` (optional): Nomor halaman (default: 1)
- `limit` (optional): Jumlah item per halaman (default: 10, max: 100)

**Example Request:**
```bash
GET /api/v1/users?search=test&page=1&limit=10
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "user": "testuser1",
      "password": "password123",
      "profile": "PPP"
    }
  ],
  "count": 1
}
```

### Get User by Username

Mendapatkan user berdasarkan username.

**Endpoint:** `GET /users/:username`

**Example Request:**
```bash
GET /api/v1/users/testuser1
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "user": "testuser1",
    "password": "password123",
    "profile": "PPP"
  }
}
```

### Get User by ID

Mendapatkan user berdasarkan ID.

**Endpoint:** `GET /users/id/:id`

**Example Request:**
```bash
GET /api/v1/users/id/1
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": 1,
    "user": "testuser1",
    "password": "password123",
    "profile": "PPP"
  }
}
```

### Create User

Membuat user baru.

**Endpoint:** `POST /users`

**Request Body:**
```json
{
  "user": "newuser",
  "password": "newpassword123",
  "profile": "PPP"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": 2,
    "user": "newuser",
    "password": "newpassword123",
    "profile": "PPP"
  }
}
```

### Update User

Memperbarui user yang ada (legacy endpoint).

**Endpoint:** `PUT /users/:username`

**Request Body:**
```json
{
  "password": "updatedpassword123",
  "profile": "SLIP"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": 2,
    "user": "newuser",
    "password": "updatedpassword123",
    "profile": "SLIP"
  }
}
```

### Update User by ID

Memperbarui user yang ada berdasarkan ID (recommended).

**Endpoint:** `PUT /users/id/:id`

**Example Request:**
```bash
PUT /api/v1/users/id/2
Authorization: Bearer <jwt-token>
```

**Request Body:**
```json
{
  "password": "updatedpassword123",
  "profile": "SLIP"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": 2,
    "user": "newuser",
    "password": "updatedpassword123",
    "profile": "SLIP"
  }
}
```

### Delete User

Menghapus user.

**Endpoint:** `DELETE /users/:username`

**Example Request:**
```bash
DELETE /api/v1/users/newuser
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

### Get User Statistics

Mendapatkan statistik user.

**Endpoint:** `GET /users/stats`

**Response:**
```json
{
  "success": true,
  "message": "User statistics retrieved successfully",
  "data": {
    "total_users": 25
  }
}
```

### Get User Attributes

Mendapatkan semua atribut user dari tabel radcheck.

**Endpoint:** `GET /users/:username/attributes`

**Response:**
```json
{
  "success": true,
  "message": "User attributes retrieved successfully",
  "data": {
    "id": 1,
    "username": "testuser1",
    "attributes": [
      {
        "attribute": "Cleartext-Password",
        "op": ":=",
        "value": "password123"
      }
    ]
  }
}
```

### Get User Reply Attributes

Mendapatkan semua atribut reply user dari tabel radreply.

**Endpoint:** `GET /users/:username/reply-attributes`

**Response:**
```json
{
  "success": true,
  "message": "User reply attributes retrieved successfully",
  "data": {
    "id": 1,
    "username": "testuser1",
    "reply_attributes": [
      {
        "attribute": "Framed-Protocol",
        "op": ":=",
        "value": "PPP"
      }
    ]
  }
}
```

### Add Custom Attribute

Menambahkan atribut kustom ke user.

**Endpoint:** `POST /users/:username/attributes`

**Request Body:**
```json
{
  "attribute": "Session-Timeout",
  "op": ":=",
  "value": "3600",
  "table": "radcheck"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Attribute added successfully",
  "data": {
    "id": 123,
    "username": "testuser1",
    "attribute": "Session-Timeout",
    "op": ":=",
    "value": "3600",
    "table": "radcheck"
  }
}
```

### Remove Custom Attribute

Menghapus atribut kustom dari user.

**Endpoint:** `DELETE /users/:username/attributes`

**Request Body:**
```json
{
  "attribute": "Session-Timeout",
  "table": "radcheck"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Attribute removed successfully"
}
```

## Error Responses

### Validation Error (400)

```json
{
  "success": false,
  "message": "Validation error",
  "errors": [
    {
      "field": "name",
      "message": "Name is required"
    },
    {
      "field": "ip",
      "message": "Must be a valid IP address"
    }
  ]
}
```

### Unauthorized (401)

```json
{
  "success": false,
  "message": "Access token required"
}
```

### Forbidden (403)

```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

### Not Found (404)

```json
{
  "success": false,
  "message": "NAS not found"
}
```

### Conflict (409)

```json
{
  "success": false,
  "message": "NAS with this name or IP already exists"
}
```

### Rate Limit (429)

```json
{
  "success": false,
  "message": "Too many requests from this IP, please try again later.",
  "retry_after": 900
}
```

### Internal Server Error (500)

```json
{
  "success": false,
  "message": "Internal server error"
}
```

## Data Models

### NAS Model

```json
{
  "id": "integer (auto-generated)",
  "name": "string (3-30 chars, alphanumeric)",
  "ip": "string (valid IPv4/IPv6)",
  "secret": "string (8-100 chars)",
  "type": "string (cisco|computone|livingston|juniper|max40xx|multitech|netserver|pathras|patton|portslave|tc|usrhiper|other)",
  "ports": "integer (1-65535, default: 1812)",
  "community": "string (max 50 chars, optional)",
  "description": "string (max 200 chars, optional)"
}
```

### User Model

```json
{
  "id": "integer (auto-generated)",
  "user": "string (3-64 chars, alphanumeric)",
  "password": "string (6-128 chars)",
  "profile": "string (PPP|SLIP|CSLIP|Shell-User|Telnet-User|Authenticate-Only|Promiscuous)"
}
```

## Rate Limiting

- **Window:** 15 menit
- **Max Requests:** 100 per IP
- **Headers:** `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset`

## Security Features

- **Helmet.js:** Security headers
- **CORS:** Cross-origin resource sharing
- **Rate Limiting:** Mencegah abuse
- **Input Validation:** Joi schema validation
- **JWT Authentication:** Secure token-based auth
- **API Key Authentication:** Alternative auth method

## Environment Variables

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_NAME=radius
DB_USER=radius
DB_PASSWORD=radiuspass123!

# Server
PORT=3000
NODE_ENV=development

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRES_IN=24h

# API
API_PREFIX=/api/v1

# Rate Limiting
RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=100

# CORS
CORS_ORIGIN=*

# Admin
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123!
```