package middleware

import (
	"log"
	"net/http"
	"time"
)

// SecurityHeaders is a small helmet-equivalent for the API + web panel.
// CSP mengizinkan 'unsafe-inline' untuk script/style karena frontend adalah
// SPA SvelteKit statis yang meng-inline bootstrap script dan CSS kecil
// (tidak ada nonce tanpa server-side rendering). Selain itu tetap ketat:
// same-origin saja, tanpa frame/object embedding.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data:; " +
	"font-src 'self' data:; " +
	"connect-src 'self'; " +
	"frame-ancestors 'none'; " +
	"object-src 'none'; " +
	"base-uri 'self'"

func SecurityHeaders(isProd bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
			if isProd {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000")
			}
			next.ServeHTTP(w, r)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Flush forwards to the underlying writer so SSE/streaming endpoints work
// through the logging wrapper (http.Flusher is a separate interface).
func (s *statusWriter) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// RequestLogger mirrors the Node request log line.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		log.Printf("%s - %s %s - IP: %s - %d (%s)", start.Format(time.RFC3339), r.Method, r.RequestURI, ClientIP(r), sw.status, time.Since(start).Round(time.Millisecond))
	})
}
