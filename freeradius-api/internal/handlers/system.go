package handlers

import (
	"net/http"
	"time"
)

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
				"web_panel": "GET /",
			},
		})
	}
}
