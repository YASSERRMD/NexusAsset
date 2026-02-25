package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenKind distinguishes access vs refresh tokens.
type TokenKind string

const (
	AccessToken  TokenKind = "access"
	RefreshToken TokenKind = "refresh"
)

// Claims is the custom JWT claims struct.
type Claims struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Kind     TokenKind `json:"kind"`
	jwt.RegisteredClaims
}

// TokenPair holds both the access and refresh token strings.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JWTService handles token generation and parsing.
type JWTService struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTService constructs a JWTService.
func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secret:             []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// GenerateTokenPair creates a new access + refresh token pair for a user.
//
// Example:
//
//	pair, err := svc.GenerateTokenPair(userID, username, email, "admin")
func (s *JWTService) GenerateTokenPair(userID, username, email, role string) (*TokenPair, error) {
	access, err := s.generateToken(userID, username, email, role, AccessToken, s.accessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refresh, err := s.generateToken(userID, username, email, role, RefreshToken, s.refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *JWTService) generateToken(
	userID, username, email, role string,
	kind TokenKind,
	expiry time.Duration,
) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		Kind:     kind,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Issuer:    "nexusasset",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}

// ParseToken validates and parses a JWT string, returning its Claims.
// Returns an error if the token is expired, invalid, or has wrong signing method.
func (s *JWTService) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
