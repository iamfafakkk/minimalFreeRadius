package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteErr(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]interface{}{"success": false, "message": msg})
}

func WriteInternal(w http.ResponseWriter, msg string, err error) {
	if os.Getenv("NODE_ENV") == "development" && err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false, "message": msg, "error": err.Error(),
		})
		return
	}
	WriteErr(w, http.StatusInternalServerError, msg)
}

func PathID(r *http.Request) (int64, bool) {
	return ParseID(chi.URLParam(r, "id"))
}

func ParseID(s string) (int64, bool) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

func QueryPageLimit(r *http.Request) (page, limit int) {
	page, limit = 1, 10
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			page = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}
	return page, limit
}

func Paginate[T any](items []T, page, limit int) ([]T, int, int) {
	total := len(items)
	if total == 0 {
		return []T{}, 0, 0
	}
	pages := (total + limit - 1) / limit
	if page > pages {
		return []T{}, total, pages
	}
	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total, pages
}
