package middleware

import (
	"net/http"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/appdb"
)

// actionFor maps an HTTP method to a coarse activity action.
func actionFor(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return method
	}
}

// ActivityLog records mutating API requests (POST/PUT/PATCH/DELETE) into the
// SQLite activity_logs table. Reads (GET/HEAD/OPTIONS) are skipped to keep the
// single-writer SQLite store light. Must run after Authenticate so the caller
// is known.
func ActivityLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		var id int64
		username := ""
		if c := ClaimsFrom(r); c != nil {
			username = c.Username
			id = appdb.UserID(username)
		}
		_ = appdb.RecordActivity(id, username, actionFor(r.Method), r.Method, r.URL.Path,
			sw.status, ClientIP(r), r.UserAgent())
	})
}
