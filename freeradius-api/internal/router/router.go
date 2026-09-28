package router

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/config"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/handlers"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/middleware"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/radiox"
)

func readFirstExisting(paths []string) ([]byte, bool) {
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			return data, true
		}
	}
	return nil, false
}

func candidatePaths(name string) []string {
	exe, _ := os.Executable()
	return []string{
		name,
		filepath.Join("freeradius-api", name),
		filepath.Join(filepath.Dir(exe), name),
	}
}

func New(cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger)
	r.Use(middleware.SecurityHeaders(cfg.IsProduction()))
	r.Use(chimw.Recoverer)
	// NB: rate limiter TIDAK global — file statis panel (html/js/css) tidak
	// dihitung, hanya request API. Satu load halaman bisa belasan request.

	authH := handlers.NewAuthHandler(cfg)
	nasH := handlers.NewNASHandler()
	userH := handlers.NewUserHandler()
	systemH := handlers.NewSystemHandler(cfg)

	// RADIUS live log: tailer fills a ring buffer in the background; the
	// handler serves the recent snapshot plus an SSE stream of new lines.
	radioxHub := radiox.NewHub()
	radioxTail := radiox.NewTailer(cfg.RadiusLogPath, radioxHub, 200)
	go radioxTail.Run(context.Background())
	radiusLogH := handlers.NewRadiusLogHandler(radioxHub, radioxTail, cfg.RadiusLogPath)

	// Static-ish endpoints (same paths as Node).
	// NB: "/" sengaja TIDAK didaftarkan sebagai JSON agar index.html
	// frontend SPA (ServeSPA fallback) yang tampil di root.
	r.Get("/health", handlers.GlobalHealth)
	r.Get("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if data, ok := readFirstExisting(candidatePaths("swagger.json")); ok {
			w.Write(data)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "swagger.json not found"})
	})
	r.Get("/api-docs", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>FreeRADIUS API</h1><p>Swagger UI is not bundled in the Go build. See <a href="/swagger.json">/swagger.json</a> or <a href="/swagger">/swagger</a>.</p></body></html>`))
	})
	r.Get("/swagger", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>FreeRADIUS API</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script><script>window.onload=function(){SwaggerUIBundle({url:'/swagger.json',dom_id:'#swagger-ui'})}</script></body></html>`))
	})
	// Serve docs/ like Express static (from disk).
	for _, d := range []string{"docs", "freeradius-api/docs"} {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir(d))))
			break
		}
	}

	p := cfg.APIPrefix
	// Semua endpoint API berbagi satu rate limiter per-IP.
	r.Route(p, func(r chi.Router) {
		r.Use(middleware.NewRateLimiter(cfg.RateWindowMs, cfg.RateMax).Middleware())
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authH.Login)
			r.Post("/logout", authH.Logout)
			r.With(middleware.RequireToken(cfg), middleware.ActivityLog).Get("/verify", authH.Verify)
			r.Get("/info", authH.Info)
			r.Get("/health", authH.Health)
		})
		r.Route("/nas", func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg))
			r.Use(middleware.ActivityLog)
			r.Get("/", nasH.List)
			r.Get("/stats", nasH.Stats)
			r.Get("/{id}", nasH.Get)
			r.Post("/", nasH.Create)
			r.Put("/{id}", nasH.Update)
			r.Delete("/{id}", nasH.Delete)
		})
		r.Route("/users", func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg))
			r.Use(middleware.ActivityLog)
			r.Get("/", userH.List)
			r.Get("/stats", userH.Stats)
			// Specific routes must be registered before :username.
			r.Get("/id/{id}", userH.GetByID)
			r.Put("/id/{id}", userH.UpdateByID)
			r.Get("/{username}/attributes", userH.Attributes)
			r.Post("/{username}/attributes", userH.AddAttribute)
			r.Delete("/{username}/attributes", userH.RemoveAttribute)
			r.Get("/{username}/reply-attributes", userH.ReplyAttributes)
			r.Get("/{username}", userH.GetByUsername)
			r.Put("/{username}", userH.Update)
			r.Delete("/{username}", userH.Delete)
			r.Post("/", userH.Create)
		})
		// Read-only views over the SQLite app DB (admin login users + logs).
		r.Route("/system", func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg))
			r.Get("/users", systemH.Users)
			r.Get("/login-history", systemH.LoginHistory)
			r.Get("/activity", systemH.Activity)
			r.Get("/db-stats", systemH.DBStats)
		})
		// Live FreeRADIUS log (snapshot + SSE). Cookie-authenticated because
		// EventSource cannot set an Authorization header.
		r.Route("/radius", func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg))
			r.Get("/log", radiusLogH.Recent)
			r.Get("/log/stream", radiusLogH.Stream)
		})
	})

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		// API yang tidak dikenal tetap JSON 404.
		if strings.HasPrefix(req.URL.Path, p+"/") || strings.HasPrefix(req.URL.Path, "/api/") {
			handlers.NotFound(cfg.APIPrefix)(w, req)
			return
		}
		// Selain itu: serve frontend SPA (atau JSON 404 bila WebDir kosong).
		handlers.ServeSPA(cfg.WebDir)(w, req)
	})
	return r
}
