package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/appdb"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/config"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/database"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/middleware"
)

var startTime = time.Now()

type AuthHandler struct{ cfg *config.Config }

func NewAuthHandler(cfg *config.Config) *AuthHandler { return &AuthHandler{cfg: cfg} }

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteErr(w, http.StatusBadRequest, "Invalid JSON in request body")
		return
	}
	if body.Username == "" || body.Password == "" {
		errs := []map[string]string{}
		if body.Username == "" {
			errs = append(errs, map[string]string{"field": "username", "message": "Username is required"})
		}
		if body.Password == "" {
			errs = append(errs, map[string]string{"field": "password", "message": "Password is required"})
		}
		WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false, "message": "Validation error", "errors": errs,
		})
		return
	}
	if body.Username != h.cfg.AdminUsername {
		_ = appdb.RecordLogin(0, body.Username, middleware.ClientIP(r), r.UserAgent(), false)
		WriteErr(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	uid, role, ok := appdb.Authenticate(body.Username, body.Password)
	if !ok {
		_ = appdb.RecordLogin(0, body.Username, middleware.ClientIP(r), r.UserAgent(), false)
		WriteErr(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	_ = appdb.RecordLogin(uid, body.Username, middleware.ClientIP(r), r.UserAgent(), true)
	token, err := middleware.GenerateToken(h.cfg, body.Username, role)
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	// Set cookie sesi untuk frontend SPA (same-origin). fr_token httpOnly
	// agar JS tidak bisa membaca JWT; fr_user readable untuk display name.
	secure := r.TLS != nil
	maxAge := h.cfg.JWTExpiresH * 3600
	if maxAge < 3600 {
		maxAge = 24 * 3600
	}
	http.SetCookie(w, &http.Cookie{
		Name: middleware.CookieToken, Value: token, Path: "/",
		MaxAge: maxAge, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name: middleware.CookieUser, Value: body.Username, Path: "/",
		MaxAge: maxAge, HttpOnly: false, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"data": map[string]interface{}{
			"token":      token,
			"user":       map[string]interface{}{"username": body.Username, "role": role},
			"expires_in": os.Getenv("JWT_EXPIRES_IN"),
		},
	})
}

// Logout menghapus cookie sesi (dipakai frontend SPA).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	secure := r.TLS != nil
	for _, name := range []string{middleware.CookieToken, middleware.CookieUser} {
		http.SetCookie(w, &http.Cookie{
			Name: name, Value: "", Path: "/",
			MaxAge: -1, HttpOnly: name == middleware.CookieToken,
			Secure: secure, SameSite: http.SameSiteLaxMode,
		})
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Logout successful",
	})
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	c := middleware.ClaimsFrom(r)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Token is valid",
		"data":    map[string]interface{}{"user": c, "valid": true},
	})
}

func (h *AuthHandler) Info(w http.ResponseWriter, r *http.Request) {
	prefix := h.cfg.APIPrefix
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "API information retrieved successfully",
		"data": map[string]interface{}{
			"name":        "FreeRADIUS API",
			"version":     "1.0.0",
			"description": "REST API for FreeRADIUS management with NAS and user CRUD operations",
			"endpoints": map[string]interface{}{
				"authentication": map[string]string{
					"login": "POST " + prefix + "/auth/login", "verify": "GET " + prefix + "/auth/verify",
				},
				"nas": map[string]string{
					"list": "GET " + prefix + "/nas", "get": "GET " + prefix + "/nas/:id",
					"create": "POST " + prefix + "/nas", "update": "PUT " + prefix + "/nas/:id",
					"delete": "DELETE " + prefix + "/nas/:id", "stats": "GET " + prefix + "/nas/stats",
				},
				"users": map[string]string{
					"list": "GET " + prefix + "/users", "get": "GET " + prefix + "/users/:username",
					"create": "POST " + prefix + "/users", "update": "PUT " + prefix + "/users/:username",
					"delete": "DELETE " + prefix + "/users/:username", "stats": "GET " + prefix + "/users/stats",
					"attributes":       "GET " + prefix + "/users/:username/attributes",
					"reply_attributes": "GET " + prefix + "/users/:username/reply-attributes",
					"add_attribute":    "POST " + prefix + "/users/:username/attributes",
					"remove_attribute": "DELETE " + prefix + "/users/:username/attributes",
				},
			},
			"authentication_methods": []string{"JWT Token (Bearer)", "API Key (X-API-Key header)"},
			"supported_formats":      []string{"JSON"},
			"rate_limiting":          map[string]interface{}{"window": "15 minutes", "max_requests": 100},
		},
	})
}

func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := database.DB.Ping(); err != nil {
		WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"message": "API is unhealthy",
			"data": map[string]interface{}{
				"status": "unhealthy", "timestamp": time.Now().UTC().Format(time.RFC3339),
				"database": "disconnected", "error": err.Error(),
			},
		})
		return
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "API is healthy",
		"data": map[string]interface{}{
			"status": "healthy", "timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime": time.Since(startTime).Seconds(), "database": "connected",
			"memory_usage": map[string]interface{}{"alloc_bytes": m.Alloc, "sys_bytes": m.Sys},
			"go_version":   runtime.Version(),
		},
	})
}
