package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"nexusasset/backend/pkg/response"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler wires auth HTTP routes.
type Handler struct {
	svc    *Service
	jwtSvc *JWTService
}

// NewHandler creates an auth Handler.
func NewHandler(svc *Service, jwtSvc *JWTService) *Handler {
	return &Handler{svc: svc, jwtSvc: jwtSvc}
}

// Login handles POST /api/v1/auth/login
//
// Request body:
//
//	{ "email": "admin@nexus.local", "password": "Admin1234!" }
//
// Response:
//
//	{ "success": true, "data": { "access_token": "...", "refresh_token": "...", "role": "admin" } }
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(w, "invalid email or password")
			return
		}
		response.InternalError(w, "login failed")
		return
	}

	// Set httpOnly cookies for web clients
	setAuthCookies(w, resp.AccessToken, resp.RefreshToken)

	response.OK(w, resp)
}

// Refresh handles POST /api/v1/auth/refresh
//
// Request body:
//
//	{ "refresh_token": "..." }
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	// Support both JSON body and cookie
	refreshToken := ""
	if c, err := r.Cookie("nexus_refresh_token"); err == nil {
		refreshToken = c.Value
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.RefreshToken != "" {
		refreshToken = body.RefreshToken
	}
	if refreshToken == "" {
		response.BadRequest(w, "refresh_token is required")
		return
	}

	resp, err := h.svc.Refresh(r.Context(), refreshToken)
	if err != nil {
		response.Unauthorized(w, "invalid or expired refresh token")
		return
	}

	// Rotate access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "nexus_access_token",
		Value:    resp.AccessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  resp.ExpiresAt,
	})

	response.OK(w, resp)
}

// Logout handles POST /api/v1/auth/logout
// Clears both auth cookies. Token revocation is client-side only in Phase 1.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookies(w)
	response.OK(w, map[string]string{"message": "logged out successfully"})
}

// Me handles GET /api/v1/auth/me — returns current user info from JWT.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "not authenticated")
		return
	}
	response.OK(w, map[string]any{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"email":    claims.Email,
		"role":     claims.Role,
	})
}

// setAuthCookies sets httpOnly cookies for access and refresh tokens.
func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "nexus_access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "nexus_refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

// clearAuthCookies expires both auth cookies.
func clearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{"nexus_access_token", "nexus_refresh_token"} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
}

// extractBearerToken is a utility for tests that build Authorization headers.
func extractBearerToken(authHeader string) string {
	return strings.TrimPrefix(authHeader, "Bearer ")
}
