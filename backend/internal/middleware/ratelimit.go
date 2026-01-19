package middleware

import (
	"net/http"
	"time"

	"github.com/auction-cards/backend/internal/cache"
)

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
	KeyFunc  func(r *http.Request) string
}

func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc: func(r *http.Request) string {
			return cache.RateLimitKeyIP(getClientIP(r))
		},
	}
}

func AuthRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Requests: 5,
		Window:   time.Minute,
		KeyFunc: func(r *http.Request) string {
			return cache.RateLimitKeyAuth(getClientIP(r))
		},
	}
}

func BidRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		KeyFunc: func(r *http.Request) string {
			userID := GetUserID(r.Context())
			return cache.RateLimitKeyBid(userID)
		},
	}
}

func RateLimit(redisCache *cache.RedisCache, config *RateLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redisCache == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := config.KeyFunc(r)
			count, err := redisCache.IncrementRateLimit(r.Context(), key, config.Window)
			if err != nil {
				// On error, allow the request
				next.ServeHTTP(w, r)
				return
			}

			if count > int64(config.Requests) {
				respondError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
