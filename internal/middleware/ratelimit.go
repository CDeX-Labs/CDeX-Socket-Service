package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type RateLimiter struct {
	requests map[string]*clientLimit
	mu       sync.RWMutex
	limit    int
	window   time.Duration
	logger   zerolog.Logger
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration, logger zerolog.Logger) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientLimit),
		limit:    limit,
		window:   window,
		logger:   logger.With().Str("component", "ratelimit").Logger(),
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, cl := range rl.requests {
			if now.After(cl.resetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cl, exists := rl.requests[ip]

	if !exists || now.After(cl.resetTime) {
		rl.requests[ip] = &clientLimit{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if cl.count >= rl.limit {
		return false
	}

	cl.count++
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !rl.Allow(ip) {
			rl.logger.Warn().Str("ip", ip).Msg("Rate limit exceeded")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
