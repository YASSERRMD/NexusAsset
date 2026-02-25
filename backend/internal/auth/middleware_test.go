package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nexusasset/backend/internal/auth"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticate_MissingToken(t *testing.T) {
	svc := newTestJWT()
	handler := svc.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthenticate_ValidToken(t *testing.T) {
	svc := newTestJWT()

	pair, _ := svc.GenerateTokenPair("u1", "user1", "u1@test.com", "admin")

	var capturedClaims *auth.Claims
	handler := svc.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = auth.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotNil(t, capturedClaims)
	assert.Equal(t, "admin", capturedClaims.Role)
}

func TestAuthenticate_RefreshTokenRejected(t *testing.T) {
	svc := newTestJWT()

	pair, _ := svc.GenerateTokenPair("u1", "user1", "u1@test.com", "admin")

	handler := svc.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Refresh token should be rejected as access token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.RefreshToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireRole_AllowedRole(t *testing.T) {
	svc := newTestJWT()
	ctx, _ := injectClaimsViaMiddleware(svc, "admin")

	handler := svc.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_ForbiddenRole(t *testing.T) {
	svc := newTestJWT()
	ctx, _ := injectClaimsViaMiddleware(svc, "reader")

	// reader tries to hit admin-only route
	handler := svc.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_MultipleRolesAllowed(t *testing.T) {
	svc := newTestJWT()

	for _, role := range []string{"admin", "contributor"} {
		ctx, _ := injectClaimsViaMiddleware(svc, role)

		handler := svc.RequireRole("admin", "contributor")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "role %s should be allowed", role)
	}
}

func TestAuthenticate_CookieFallback(t *testing.T) {
	svc := newTestJWT()
	pair, _ := svc.GenerateTokenPair("u1", "u1", "u1@test.com", "admin")

	handler := svc.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "nexus_access_token", Value: pair.AccessToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClaimsFromContext_NilWhenAbsent(t *testing.T) {
	claims := auth.ClaimsFromContext(context.Background())
	assert.Nil(t, claims)
}

// All RequireRole tests inject claims via injectClaimsViaMiddleware, which uses
// the real Authenticate middleware so the context key matches the production code.

// Override: use Authenticate's key by going through the public middleware API
func injectClaimsViaMiddleware(svc *auth.JWTService, role string) (context.Context, string) {
	pair, _ := svc.GenerateTokenPair("u1", "u1", "u1@test.com", role)

	var capturedCtx context.Context
	handler := svc.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return capturedCtx, pair.AccessToken
}

func TestRequireRole_ViaAuthenticate(t *testing.T) {
	svc := newTestJWT()

	tests := []struct {
		role           string
		allowedRoles   []string
		expectedStatus int
	}{
		{"admin", []string{"admin"}, http.StatusOK},
		{"contributor", []string{"admin"}, http.StatusForbidden},
		{"reader", []string{"admin", "contributor", "reader"}, http.StatusOK},
		{"reader", []string{"admin", "contributor"}, http.StatusForbidden},
	}

	for _, tt := range tests {
		ctx, _ := injectClaimsViaMiddleware(svc, tt.role)

		handler := svc.RequireRole(tt.allowedRoles...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, tt.expectedStatus, rec.Code,
			"role=%s allowed=%v", tt.role, tt.allowedRoles)
	}
}

// Ensure claims injected via Authenticate expire correctly
func TestAuthenticate_ExpiredToken(t *testing.T) {
	expiredSvc := auth.NewJWTService(testSecret, -1*time.Second, -1*time.Second)
	validSvc := newTestJWT()

	pair, _ := expiredSvc.GenerateTokenPair("u1", "u1", "u1@test.com", "admin")

	handler := validSvc.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
