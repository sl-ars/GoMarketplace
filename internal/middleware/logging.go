package middleware

import (
	"go-app-marketplace/pkg/logger"
	"net/http"
	"time"
)

// LoggingMiddleware создает middleware для логирования HTTP запросов
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем wrapper для ResponseWriter чтобы перехватить статус код
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Выполняем следующий handler
			next.ServeHTTP(wrapped, r)

			// Логируем запрос
			duration := time.Since(start)
			log.WithFields(logger.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"status":     wrapped.statusCode,
				"duration":   duration.String(),
				"userAgent":  r.UserAgent(),
				"remoteAddr": r.RemoteAddr,
			}).Info("HTTP request processed")
		})
	}
}

// responseWriter обертка для http.ResponseWriter для перехвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
