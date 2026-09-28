package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/appdb"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/config"
)

type SystemHandler struct{ cfg *config.Config }

func NewSystemHandler(cfg *config.Config) *SystemHandler { return &SystemHandler{cfg: cfg} }

func limitParam(r *http.Request, def int) int {
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// Users lists admin accounts stored in SQLite.
func (h *SystemHandler) Users(w http.ResponseWriter, r *http.Request) {
	us, err := appdb.ListUsers()
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	for i := range us {
		if last, ok := appdb.LastSuccessfulLogin(us[i].Username); ok {
			us[i].LastLoginAt = last
		}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "App users retrieved successfully", "data": us, "count": len(us),
	})
}

// LoginHistory lists recent login attempts.
func (h *SystemHandler) LoginHistory(w http.ResponseWriter, r *http.Request) {
	items, err := appdb.RecentLogins(limitParam(r, 50))
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Login history retrieved successfully", "data": items, "count": len(items),
	})
}

// Activity lists recent mutating actions (audit trail).
func (h *SystemHandler) Activity(w http.ResponseWriter, r *http.Request) {
	items, err := appdb.RecentActivity(limitParam(r, 100))
	if err != nil {
		WriteInternal(w, "Internal server error", err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "Activity log retrieved successfully", "data": items, "count": len(items),
	})
}

// DBStats returns counters for the SQLite store.
func (h *SystemHandler) DBStats(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true, "message": "App DB statistics retrieved successfully",
		"data": map[string]interface{}{"path": h.cfg.AppDBPath, "counts": appdb.Stats()},
	})
}

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
