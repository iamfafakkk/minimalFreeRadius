package handlers

import (
	"net/http"
	"time"
)

// Root mirrors GET / in server.js.
func Root(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"success": true, "message": "Welcome to FreeRADIUS API", "version": "1.0.0",
			"documentation": map[string]interface{}{
				"swagger_ui":         "GET /api-docs",
				"swagger_ui_custom":  "GET /swagger",
				"swagger_json":       "GET /swagger.json",
				"api_documentation":  "GET /docs/API_DOCUMENTATION.md",
				"installation_guide": "GET /docs/INSTALLATION_GUIDE.md",
				"endpoints": map[string]string{
					"health": "GET /health", "api_info": "GET " + prefix + "/auth/info",
					"login": "POST " + prefix + "/auth/login",
					"nas":   "GET " + prefix + "/nas", "users": "GET " + prefix + "/users",
				},
				"authentication": map[string]string{
					"jwt": "Use Bearer token in Authorization header", "api_key": "Use X-API-Key header",
				},
			},
		})
	}
}

// GlobalHealth mirrors GET /health in server.js.
func GlobalHealth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "FreeRADIUS API is running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).Seconds(),
	})
}

// NotFound mirrors the Express 404 handler.
func NotFound(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false, "message": "Endpoint not found",
			"path": r.URL.Path, "method": r.Method,
			"available_endpoints": map[string]string{
				"health": "GET /health", "api_info": "GET " + prefix + "/auth/info",
				"documentation": "GET /",
			},
		})
	}
}
