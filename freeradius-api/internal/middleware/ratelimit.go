package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple fixed-window per-IP limiter (stdlib only).
type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	window time.Duration
	max    int
}

func NewRateLimiter(windowMs, max int) *RateLimiter {
	if windowMs <= 0 {
		windowMs = 15 * 60 * 1000
	}
	if max <= 0 {
		max = 100
	}
	rl := &RateLimiter{
		hits:   map[string][]time.Time{},
		window: time.Duration(windowMs) * time.Millisecond,
		max:    max,
	}
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for range t.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, ts := range rl.hits {
				kept := ts[:0]
				for _, x := range ts {
					if now.Sub(x) < rl.window {
						kept = append(kept, x)
					}
				}
				if len(kept) == 0 {
					delete(rl.hits, ip)
				} else {
					rl.hits[ip] = kept
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		for i, c := range fwd {
			if c == ',' {
				return trimSpace(fwd[:i])
			}
		}
		return trimSpace(fwd)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func trimSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	return s[i:j]
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			now := time.Now()
			rl.mu.Lock()
			ts := rl.hits[ip]
			kept := ts[:0]
			for _, x := range ts {
				if now.Sub(x) < rl.window {
					kept = append(kept, x)
				}
			}
			kept = append(kept, now)
			rl.hits[ip] = kept
			over := len(kept) > rl.max
			rl.mu.Unlock()
			if over {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"message":     "Too many requests from this IP, please try again later.",
					"retry_after": int(rl.window.Seconds()),
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
