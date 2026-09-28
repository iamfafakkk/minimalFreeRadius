package router

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/config"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/handlers"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/middleware"
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
	r.Use(middleware.NewRateLimiter(cfg.RateWindowMs, cfg.RateMax).Middleware())
	r.Use(chimw.Recoverer)

	authH := handlers.NewAuthHandler(cfg)
	nasH := handlers.NewNASHandler()
	userH := handlers.NewUserHandler()

	// Static-ish endpoints (same paths as Node).
	r.Get("/", handlers.Root(cfg.APIPrefix))
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
	r.Route(p+"/auth", func(r chi.Router) {
		r.Post("/login", authH.Login)
		r.With(middleware.RequireToken(cfg)).Get("/verify", authH.Verify)
		r.Get("/info", authH.Info)
		r.Get("/health", authH.Health)
	})
	r.Route(p+"/nas", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg))
		r.Get("/", nasH.List)
		r.Get("/stats", nasH.Stats)
		r.Get("/{id}", nasH.Get)
		r.Post("/", nasH.Create)
		r.Put("/{id}", nasH.Update)
		r.Delete("/{id}", nasH.Delete)
	})
	r.Route(p+"/users", func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg))
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

	r.NotFound(handlers.NotFound(cfg.APIPrefix))
	return r
}

var _ = path.Join
