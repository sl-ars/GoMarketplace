package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/ratelimit"
)

// RateLimitConfig holds rate limiter dependencies
type RateLimitConfig struct {
	Limiter *ratelimit.RedisLimiter
	Logger  *logger.Logger
}

// RateLimitMiddleware creates a rate limiting middleware with the specified config
// This is the primary way to add rate limiting to routes
//
// Usage:
//
//	router.Use(middleware.RateLimitMiddleware(rlConfig, ratelimit.StandardLimit))
func RateLimitMiddleware(cfg *RateLimitConfig, limit ratelimit.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate rate limit key based on authentication status
			key := generateRateLimitKey(r)

			result, err := cfg.Limiter.Allow(r.Context(), key, limit)
			if err != nil {
				cfg.Logger.WithError(err).Error("Rate limit check failed")
				// On error, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			setRateLimitHeaders(w, limit, result)

			if !result.Allowed {
				cfg.Logger.WithField("key", key).Warn("Rate limit exceeded")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
				httpx.WriteError(w, http.StatusTooManyRequests,
					"Too many requests",
					apperror.ErrTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitHandler wraps a single handler with rate limiting
// Useful for applying different limits to specific endpoints
//
// Usage:
//
//	router.HandleFunc("/api/auth/login",
//	    middleware.RateLimitHandler(rlConfig, ratelimit.AuthLimit, loginHandler))
func RateLimitHandler(cfg *RateLimitConfig, limit ratelimit.Config, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := generateRateLimitKey(r)

		result, err := cfg.Limiter.Allow(r.Context(), key, limit)
		if err != nil {
			cfg.Logger.WithError(err).Error("Rate limit check failed")
			handler(w, r)
			return
		}

		setRateLimitHeaders(w, limit, result)

		if !result.Allowed {
			cfg.Logger.WithField("key", key).Warn("Rate limit exceeded")
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
			httpx.WriteError(w, http.StatusTooManyRequests,
				"Too many requests",
				apperror.ErrTooManyRequests)
			return
		}

		handler(w, r)
	}
}

// RateLimitByEndpoint creates middleware that uses endpoint-specific keys
// This provides more granular rate limiting per endpoint
func RateLimitByEndpoint(cfg *RateLimitConfig, limit ratelimit.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			endpoint := normalizeEndpoint(r.URL.Path)
			key := generateRateLimitKeyWithEndpoint(r, endpoint)

			result, err := cfg.Limiter.Allow(r.Context(), key, limit)
			if err != nil {
				cfg.Logger.WithError(err).Error("Rate limit check failed")
				next.ServeHTTP(w, r)
				return
			}

			setRateLimitHeaders(w, limit, result)

			if !result.Allowed {
				cfg.Logger.WithFields(map[string]interface{}{
					"key":      key,
					"endpoint": endpoint,
				}).Warn("Rate limit exceeded")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
				httpx.WriteError(w, http.StatusTooManyRequests,
					"Too many requests",
					apperror.ErrTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// generateRateLimitKey generates a rate limit key from the request
func generateRateLimitKey(r *http.Request) string {
	// Try to get user ID from context (authenticated request)
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return ratelimit.KeyFromUserID(userID, "global")
	}

	// Fall back to IP address for unauthenticated requests
	ip := getClientIP(r)
	return ratelimit.KeyFromIP(ip, "global")
}

// generateRateLimitKeyWithEndpoint generates a rate limit key with endpoint
func generateRateLimitKeyWithEndpoint(r *http.Request, endpoint string) string {
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return ratelimit.KeyFromUserID(userID, endpoint)
	}

	ip := getClientIP(r)
	return ratelimit.KeyFromIP(ip, endpoint)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}

	return ip
}

// normalizeEndpoint normalizes the endpoint path for rate limiting
// Removes variable parts like IDs to group similar endpoints
func normalizeEndpoint(path string) string {
	parts := strings.Split(path, "/")
	normalized := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}
		// Replace numeric IDs with placeholder
		if isNumeric(part) {
			normalized = append(normalized, ":id")
		} else {
			normalized = append(normalized, part)
		}
	}

	return "/" + strings.Join(normalized, "/")
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// setRateLimitHeaders sets standard rate limit headers
func setRateLimitHeaders(w http.ResponseWriter, limit ratelimit.Config, result *ratelimit.Result) {
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit.Requests))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))
}
