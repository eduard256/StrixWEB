package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type limiter struct {
	count    int
	resetAt  time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*limiter
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*limiter),
		limit:   limit,
		window:  window,
	}

	// cleanup expired entries every minute
	go func() {
		for range time.Tick(time.Minute) {
			rl.mu.Lock()
			now := time.Now()
			for ip, l := range rl.clients {
				if now.After(l.resetAt) {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	l, ok := rl.clients[ip]
	if !ok || now.After(l.resetAt) {
		rl.clients[ip] = &limiter{count: 1, resetAt: now.Add(rl.window)}
		return true
	}

	l.count++
	return l.count <= rl.limit
}

var (
	getLimiter  = newRateLimiter(60, time.Minute)
	postLimiter = newRateLimiter(5, time.Minute)
)

func rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)

		var rl *rateLimiter
		if r.Method == "POST" {
			rl = postLimiter
		} else {
			rl = getLimiter
		}

		if !rl.allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func limitBody(next http.HandlerFunc, maxBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next(w, r)
	}
}

func cors(next http.HandlerFunc, origins string) http.HandlerFunc {
	allowed := strings.Split(origins, ",")

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// check allowed origins, empty = allow all
		if origins == "" || origins == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			for _, a := range allowed {
				if strings.TrimSpace(a) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func realIP(r *http.Request) string {
	// X-Forwarded-For from Traefik
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// first IP in the chain is the client
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}

	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
