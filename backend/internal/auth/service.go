package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// LoginRequest is the input to the login endpoint.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Role         string    `json:"role"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Service handles auth business logic.
type Service struct {
	repo     Repository
	jwtSvc   *JWTService
	atExpiry time.Duration
}

// NewService creates an auth Service.
func NewService(repo Repository, jwtSvc *JWTService, accessExpiry time.Duration) *Service {
	return &Service{repo: repo, jwtSvc: jwtSvc, atExpiry: accessExpiry}
}

// ErrInvalidCredentials is returned when email/password do not match.
var ErrInvalidCredentials = errors.New("invalid email or password")

// Login validates credentials and returns a token pair.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}

	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	pair, err := s.jwtSvc.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	// Fire-and-forget login timestamp update
	go func() {
		_ = s.repo.UpdateLastLogin(context.Background(), user.ID, time.Now())
	}()

	return &LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		Role:         user.Role,
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		ExpiresAt:    time.Now().Add(s.atExpiry),
	}, nil
}

// RefreshResponse is returned on successful token refresh.
type RefreshResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Refresh validates a refresh token and issues a new access token.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, error) {
	claims, err := s.jwtSvc.ParseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	if claims.Kind != RefreshToken {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Re-fetch user to pick up any role changes
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("refresh: user lookup: %w", err)
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	pair, err := s.jwtSvc.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generating new access token: %w", err)
	}

	return &RefreshResponse{
		AccessToken: pair.AccessToken,
		ExpiresAt:   time.Now().Add(s.atExpiry),
	}, nil
}
