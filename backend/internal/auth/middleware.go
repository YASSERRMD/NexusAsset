package auth

import (
	"context"
	"net/http"
	"strings"

	"nexusasset/backend/pkg/response"
)

// contextKey is a private type for context keys in this package.
type contextKey string

const claimsKey contextKey = "auth_claims"

// Authenticate is a chi middleware that validates the Bearer token or cookie.
// It injects the parsed Claims into the request context.
// Returns 401 if token is missing or invalid.
func (s *JWTService) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			response.Unauthorized(w, "authentication required")
			return
		}

		claims, err := s.ParseToken(tokenStr)
		if err != nil {
			response.Unauthorized(w, "invalid or expired token")
			return
		}

		if claims.Kind != AccessToken {
			response.Unauthorized(w, "access token required")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a chi middleware that enforces role-based access.
// Returns 403 if the authenticated user's role is not in the allowed list.
//
// Example:
//
//	r.With(jwtSvc.RequireRole("admin")).Delete("/users/{id}", handler)
func (s *JWTService) RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				response.Unauthorized(w, "authentication required")
				return
			}
			if !allowed[claims.Role] {
				response.Forbidden(w, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromContext retrieves JWT Claims from a request context.
// Returns nil if not found (i.e., unauthenticated request).
func ClaimsFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsKey).(*Claims)
	return v
}

// extractToken pulls the JWT string from Authorization header or nexus_token cookie.
func extractToken(r *http.Request) string {
	// 1. Authorization: Bearer <token>
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	// 2. Cookie fallback: nexus_access_token
	if c, err := r.Cookie("nexus_access_token"); err == nil {
		return c.Value
	}
	return ""
}
