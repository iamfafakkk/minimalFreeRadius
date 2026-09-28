package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBMaxConns int

	Port      int
	APIPrefix string
	Env       string

	JWTSecret   string
	JWTExpiresH int // hours, parsed from JWT_EXPIRES_IN like "24h"

	RateWindowMs int
	RateMax      int

	CORSOrigin string
	APIKey     string

	AdminUsername string
	AdminPassword string

	// AppDBPath: file SQLite untuk user login + log aktivitas. Dibuat dan
	// di-seed otomatis saat server dijalankan (idempoten).
	AppDBPath string

	// RadiusLogPath: file log FreeRADIUS yang di-tail realtime oleh panel.
	RadiusLogPath string

	// WebDir: direktori hasil build frontend (adapter-static) yang di-serve
	// backend. Kosong = mode API-only. Bisa diisi via env WEB_DIR.
	WebDir string

	HTTPSEnabled    bool
	HTTPSPort       int
	RedirectToHTTPS bool
	SSLCertPath     string
	SSLKeyPath      string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvBool(key string) bool {
	return strings.ToLower(os.Getenv(key)) == "true"
}

// parse "24h" / "48h" / "30m" into hours (rounded up, min 1)
func parseExpiryHours(s string) int {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "h") {
		if n, err := strconv.Atoi(strings.TrimSuffix(s, "h")); err == nil && n > 0 {
			return n
		}
	}
	if strings.HasSuffix(s, "m") {
		if n, err := strconv.Atoi(strings.TrimSuffix(s, "m")); err == nil && n > 0 {
			h := n / 60
			if h < 1 {
				h = 1
			}
			return h
		}
	}
	if strings.HasSuffix(s, "d") {
		if n, err := strconv.Atoi(strings.TrimSuffix(s, "d")); err == nil && n > 0 {
			return n * 24
		}
	}
	return 24
}

func Load() *Config {
	_ = godotenv.Load() // .env optional, ignore error

	return &Config{
		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenvInt("DB_PORT", 3306),
		DBName:     getenv("DB_NAME", "radius"),
		DBUser:     getenv("DB_USER", "radius"),
		DBPassword: getenv("DB_PASSWORD", "radiuspass123!"),
		DBMaxConns: getenvInt("DB_CONNECTION_LIMIT", 10),

		Port:      getenvInt("PORT", 3000),
		APIPrefix: getenv("API_PREFIX", "/api/v1"),
		Env:       getenv("NODE_ENV", "development"),

		JWTSecret:   getenv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTExpiresH: parseExpiryHours(getenv("JWT_EXPIRES_IN", "24h")),

		RateWindowMs: getenvInt("RATE_LIMIT_WINDOW_MS", 15*60*1000),
		RateMax:      getenvInt("RATE_LIMIT_MAX_REQUESTS", 1000),

		CORSOrigin: getenv("CORS_ORIGIN", "*"),
		APIKey:     getenv("API_KEY", "freeradius-api-key-change-this"),

		AdminUsername: getenv("ADMIN_USERNAME", "admin"),
		AdminPassword: getenv("ADMIN_PASSWORD", "admin123!"),

		AppDBPath: getenv("APP_DB_PATH", "freeradius.db"),

		RadiusLogPath: getenv("RADIUS_LOG", "/var/log/freeradius/radius.log"),

		WebDir: resolveWebDir(getenv("WEB_DIR", "")),

		HTTPSEnabled:    getenvBool("HTTPS_ENABLED"),
		HTTPSPort:       getenvInt("HTTPS_PORT", 3443),
		RedirectToHTTPS: getenvBool("REDIRECT_HTTP_TO_HTTPS"),
		SSLCertPath:     os.Getenv("SSL_CERT_PATH"),
		SSLKeyPath:      os.Getenv("SSL_KEY_PATH"),
	}
}

func (c *Config) IsProduction() bool { return c.Env == "production" }
