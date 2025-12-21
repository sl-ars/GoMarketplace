package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-app-marketplace/pkg/metrics"

	"github.com/gorilla/mux"
)

// metricsResponseWriter wraps http.ResponseWriter to capture the status code for metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{w, http.StatusOK}
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// PrometheusMiddleware collects HTTP metrics for Prometheus
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip metrics endpoint to avoid recursion
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Track in-flight requests
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		// Start timer
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := newMetricsResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get route pattern (normalized endpoint)
		endpoint := getRoutePattern(r)

		// Record metrics
		metrics.RecordHTTPRequest(
			r.Method,
			endpoint,
			strconv.Itoa(wrapped.statusCode),
			duration,
		)
	})
}

// getRoutePattern extracts the route pattern from the request for metrics
// This prevents high cardinality by using patterns like /api/users/{id}
// instead of actual values like /api/users/123
func getRoutePattern(r *http.Request) string {
	// Try to get the route pattern from mux
	route := mux.CurrentRoute(r)
	if route != nil {
		if pattern, err := route.GetPathTemplate(); err == nil {
			return pattern
		}
	}

	// Fallback: normalize the path manually
	path := r.URL.Path

	// Common patterns to normalize
	parts := strings.Split(path, "/")
	normalized := make([]string, len(parts))

	for i, part := range parts {
		// Check if this looks like an ID (numeric or UUID-like)
		if isIDLike(part) {
			normalized[i] = "{id}"
		} else {
			normalized[i] = part
		}
	}

	return strings.Join(normalized, "/")
}

// isIDLike checks if a string looks like an ID
func isIDLike(s string) bool {
	if s == "" {
		return false
	}

	// Check if it's numeric
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return true
	}

	// Check if it looks like a UUID (36 chars with dashes)
	if len(s) == 36 && strings.Count(s, "-") == 4 {
		return true
	}

	// Check if it's a hex string (common for tokens)
	if len(s) >= 32 && isHexString(s) {
		return true
	}

	return false
}

// isHexString checks if a string contains only hex characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
