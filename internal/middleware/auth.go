package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"go-app-marketplace/pkg/auth"
	"go-app-marketplace/pkg/httpx"
	"net/http"
	"strings"
)

func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing or invalid Authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := auth.ParseToken(tokenStr, secret)
			if err != nil || !token.Valid {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid token claims")
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing user_id in token")
				return
			}
			userID := int64(userIDFloat)

			role, ok := claims["role"].(string)
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Missing role in token")
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "role", role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
