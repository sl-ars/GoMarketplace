package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/ratelimit"
)

// RateLimitConfig holds rate limiter dependencies
type RateLimitConfig struct {
	Limiter        *ratelimit.RedisLimiter
	Logger         *logger.Logger
	TrustedProxies *TrustedProxies
}

// TrustedProxies manages a list of trusted proxy IPs/CIDRs
// Only requests from these addresses will have X-Forwarded-For/X-Real-IP headers trusted
type TrustedProxies struct {
	ips   map[string]bool
	cidrs []*net.IPNet
	mu    sync.RWMutex
}

// NewTrustedProxies creates a TrustedProxies instance from a list of IPs/CIDRs
// Accepts formats: "10.0.0.1", "192.168.0.0/16", "::1"
func NewTrustedProxies(proxies []string) *TrustedProxies {
	tp := &TrustedProxies{
		ips:   make(map[string]bool),
		cidrs: make([]*net.IPNet, 0),
	}

	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		if proxy == "" {
			continue
		}

		// Check if it's a CIDR
		if strings.Contains(proxy, "/") {
			_, cidr, err := net.ParseCIDR(proxy)
			if err == nil {
				tp.cidrs = append(tp.cidrs, cidr)
			}
		} else {
			// It's a single IP
			ip := net.ParseIP(proxy)
			if ip != nil {
				tp.ips[ip.String()] = true
			}
		}
	}

	return tp
}

// IsTrusted checks if the given IP is in the trusted list
func (tp *TrustedProxies) IsTrusted(ipStr string) bool {
	if tp == nil {
		return false
	}

	tp.mu.RLock()
	defer tp.mu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check exact IP match
	if tp.ips[ip.String()] {
		return true
	}

	// Check CIDR ranges
	for _, cidr := range tp.cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// HasTrustedProxies returns true if any trusted proxies are configured
func (tp *TrustedProxies) HasTrustedProxies() bool {
	if tp == nil {
		return false
	}
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return len(tp.ips) > 0 || len(tp.cidrs) > 0
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
			key := generateRateLimitKey(r, cfg.TrustedProxies)

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
		key := generateRateLimitKey(r, cfg.TrustedProxies)

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
			key := generateRateLimitKeyWithEndpoint(r, endpoint, cfg.TrustedProxies)

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
func generateRateLimitKey(r *http.Request, tp *TrustedProxies) string {
	// Try to get user ID from context (authenticated request)
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return ratelimit.KeyFromUserID(userID, "global")
	}

	// Fall back to IP address for unauthenticated requests
	ip := getClientIP(r, tp)
	return ratelimit.KeyFromIP(ip, "global")
}

// generateRateLimitKeyWithEndpoint generates a rate limit key with endpoint
func generateRateLimitKeyWithEndpoint(r *http.Request, endpoint string, tp *TrustedProxies) string {
	if userID, ok := r.Context().Value("user_id").(int64); ok {
		return ratelimit.KeyFromUserID(userID, endpoint)
	}

	ip := getClientIP(r, tp)
	return ratelimit.KeyFromIP(ip, endpoint)
}

// getClientIP extracts the client IP from the request securely
//
// SECURITY: X-Forwarded-For and X-Real-IP headers can be spoofed by attackers.
// We ONLY trust these headers if the direct connection (RemoteAddr) comes from
// a configured trusted proxy. Otherwise, we use RemoteAddr directly.
//
// This prevents attackers from bypassing rate limiting by setting fake headers.
func getClientIP(r *http.Request, tp *TrustedProxies) string {
	// Get the direct connection IP (cannot be spoofed)
	remoteIP := extractIPFromRemoteAddr(r.RemoteAddr)

	// Only trust forwarding headers if the direct connection is from a trusted proxy
	if tp != nil && tp.IsTrusted(remoteIP) {
		// Check X-Forwarded-For header
		// Format: client, proxy1, proxy2
		// We take the rightmost IP that is NOT a trusted proxy (the real client)
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			ips := strings.Split(xff, ",")
			// Walk backwards to find the rightmost untrusted IP
			for i := len(ips) - 1; i >= 0; i-- {
				ip := strings.TrimSpace(ips[i])
				if ip != "" && !tp.IsTrusted(ip) {
					return ip
				}
			}
		}

		// Check X-Real-IP header (set by nginx/reverse proxies)
		xri := r.Header.Get("X-Real-IP")
		if xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	// Use the direct connection IP (most secure, cannot be spoofed)
	return remoteIP
}

// extractIPFromRemoteAddr extracts the IP address from RemoteAddr
// RemoteAddr format is typically "IP:port" or "[IPv6]:port"
func extractIPFromRemoteAddr(remoteAddr string) string {
	// Handle IPv6 format [::1]:port
	if strings.HasPrefix(remoteAddr, "[") {
		if idx := strings.LastIndex(remoteAddr, "]"); idx != -1 {
			return remoteAddr[1:idx]
		}
	}

	// Handle IPv4 format 127.0.0.1:port
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If SplitHostPort fails, it might be just an IP without port
		return remoteAddr
	}
	return host
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
