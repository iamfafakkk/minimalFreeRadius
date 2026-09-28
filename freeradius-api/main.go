package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/appdb"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/config"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/database"
	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/router"
	"github.com/rs/cors"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()
	_ = db
	log.Println("Database connected successfully")

	// SQLite app DB: admin users + login history + activity log. Schema and
	// admin seed are idempotent, so restarts never duplicate or overwrite.
	if err := appdb.Open(cfg.AppDBPath); err != nil {
		log.Fatalf("App database (SQLite) failed: %v", err)
	}
	defer appdb.Close()
	if err := appdb.SeedAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Printf("WARNING: failed to seed admin user: %v", err)
	}
	log.Printf("App database (SQLite) ready: %s", cfg.AppDBPath)

	handler := router.New(cfg)

	// CORS (mirrors Node corsOptions).
	var origins []string
	if cfg.CORSOrigin == "*" {
		origins = []string{"*"}
	} else {
		for _, o := range strings.Split(cfg.CORSOrigin, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins = append(origins, o)
			}
		}
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	})
	// rs/cors with "*" + credentials is invalid; fall back to permissive reflect.
	var h http.Handler = c.Handler(handler)
	if cfg.CORSOrigin == "*" {
		h = cors.AllowAll().Handler(handler)
	}

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("═══════════════════════════════════════════════════════════════")
		log.Println("                    FreeRADIUS API Server (Go)")
		log.Println("═══════════════════════════════════════════════════════════════")
		log.Printf("HTTP server running on port %d", cfg.Port)
		log.Printf("API Base URL: http://localhost:%d%s", cfg.Port, cfg.APIPrefix)
		log.Printf("Health Check: http://localhost:%d/health", cfg.Port)
		log.Printf("Environment: %s", cfg.Env)
		if cfg.WebDir != "" {
			log.Printf("Web panel: serving %s", cfg.WebDir)
		} else {
			log.Printf("Web panel: not found (API-only mode, set WEB_DIR)")
		}
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Optional HTTPS server (same env keys as Node).
	var httpsSrv *http.Server
	if cfg.HTTPSEnabled {
		if cfg.SSLCertPath == "" || cfg.SSLKeyPath == "" {
			log.Println("WARNING: HTTPS enabled but SSL_CERT_PATH or SSL_KEY_PATH is not set. Skipping HTTPS server.")
		} else {
			cert, err := tls.LoadX509KeyPair(cfg.SSLCertPath, cfg.SSLKeyPath)
			if err != nil {
				log.Printf("Failed to start HTTPS server: %v", err)
			} else {
				httpsSrv = &http.Server{
					Addr:              fmt.Sprintf(":%d", cfg.HTTPSPort),
					Handler:           h,
					ReadHeaderTimeout: 10 * time.Second,
					TLSConfig:         &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
				}
				go func() {
					log.Printf("HTTPS server running on port %d", cfg.HTTPSPort)
					if err := httpsSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
						log.Printf("HTTPS server error: %v", err)
					}
				}()
			}
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	if httpsSrv != nil {
		if err := httpsSrv.Shutdown(ctx); err != nil {
			log.Printf("HTTPS shutdown error: %v", err)
		}
	}
	log.Println("Database connections closed")
}
