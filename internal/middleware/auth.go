package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"go-app-marketplace/pkg/auth"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"net/http"
	"strings"
)

func AuthMiddleware(secret []byte, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				log.WithRequest(r.Method, r.URL.Path, r.UserAgent(), nil).WithOperation("auth_middleware").Warn("Missing or invalid Authorization header")
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing or invalid Authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := auth.ParseToken(tokenStr, secret)
			if err != nil || !token.Valid {
				log.WithError(err).WithRequest(r.Method, r.URL.Path, r.UserAgent(), nil).WithOperation("auth_middleware").Warn("Invalid or expired token")
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.WithRequest(r.Method, r.URL.Path, r.UserAgent(), nil).WithOperation("auth_middleware").Warn("Invalid token claims")
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid token claims")
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				log.WithRequest(r.Method, r.URL.Path, r.UserAgent(), nil).WithOperation("auth_middleware").Warn("Missing user_id in token")
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing user_id in token")
				return
			}
			userID := int64(userIDFloat)

			role, ok := claims["role"].(string)
			if !ok {
				log.WithUser(userID).WithRequest(r.Method, r.URL.Path, r.UserAgent(), userID).WithOperation("auth_middleware").Warn("Missing role in token")
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing role in token")
				return
			}

			log.WithUser(userID).WithFields(logger.Fields{
				"role": role,
				"path": r.URL.Path,
			}).WithOperation("auth_middleware").Debug("User authenticated successfully")

			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "role", role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
