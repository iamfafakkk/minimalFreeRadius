package handlers

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ServeSPA menyajikan hasil build frontend (SvelteKit adapter-static).
// File statis (js/css/gambar) diserve langsung; route SPA (/login,
// /dashboard/...) fallback ke index.html. webDir kosong = mode API-only.
func ServeSPA(webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if webDir == "" {
			WriteErr(w, http.StatusNotFound, "Endpoint not found")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			WriteErr(w, http.StatusNotFound, "Endpoint not found")
			return
		}
		p := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		// Blokir path traversal / file tersembunyi.
		if strings.Contains(p, "..") {
			WriteErr(w, http.StatusNotFound, "Endpoint not found")
			return
		}
		full := filepath.Join(webDir, filepath.FromSlash(p))
		if st, err := os.Stat(full); err == nil && !st.IsDir() {
			http.ServeFile(w, r, full)
			return
		}
		// Fallback SPA: semua route non-file ke index.html.
		index := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(index); err == nil {
			// Jangan cache shell SPA agar deploy baru langsung kepakai.
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, index)
			return
		}
		WriteErr(w, http.StatusNotFound, "Endpoint not found")
	}
}
