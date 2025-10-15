package middleware

import (
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/httpx"
	"net/http"
)

func RequireRoles(allowedRoles ...domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			roleVal := ctx.Value("role")
			if roleVal == nil {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Role not found in context")
				return
			}

			roleStr, ok := roleVal.(string)
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized", "Invalid role type")
				return
			}

			userRole := domain.UserRole(roleStr)

			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				httpx.WriteError(w, http.StatusForbidden, "Forbidden", "You don't have permission to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
