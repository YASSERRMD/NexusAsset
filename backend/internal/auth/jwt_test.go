package auth_test

import (
	"testing"
	"time"

	"nexusasset/backend/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-that-is-sufficiently-long-32+"

func newTestJWT() *auth.JWTService {
	return auth.NewJWTService(testSecret, 15*time.Minute, 7*24*time.Hour)
}

func TestGenerateTokenPair(t *testing.T) {
	svc := newTestJWT()

	pair, err := svc.GenerateTokenPair("user-123", "johndoe", "john@test.com", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)
}

func TestParseAccessToken(t *testing.T) {
	svc := newTestJWT()

	pair, err := svc.GenerateTokenPair("user-456", "janesmith", "jane@test.com", "contributor")
	require.NoError(t, err)

	claims, err := svc.ParseToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.UserID)
	assert.Equal(t, "janesmith", claims.Username)
	assert.Equal(t, "jane@test.com", claims.Email)
	assert.Equal(t, "contributor", claims.Role)
	assert.Equal(t, auth.AccessToken, claims.Kind)
}

func TestParseRefreshToken(t *testing.T) {
	svc := newTestJWT()

	pair, err := svc.GenerateTokenPair("user-789", "readuser", "read@test.com", "reader")
	require.NoError(t, err)

	claims, err := svc.ParseToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, auth.RefreshToken, claims.Kind)
	assert.Equal(t, "reader", claims.Role)
}

func TestParseToken_InvalidSignature(t *testing.T) {
	svc := newTestJWT()
	otherSvc := auth.NewJWTService("different-secret-that-is-also-long-enough!", 15*time.Minute, time.Hour)

	pair, err := svc.GenerateTokenPair("u1", "u1", "u1@test.com", "admin")
	require.NoError(t, err)

	_, err = otherSvc.ParseToken(pair.AccessToken)
	assert.Error(t, err, "should reject token signed with different secret")
}

func TestParseToken_Expired(t *testing.T) {
	// Create a service with near-zero expiry to produce immediately expired tokens
	svc := auth.NewJWTService(testSecret, -1*time.Millisecond, -1*time.Millisecond)

	pair, err := svc.GenerateTokenPair("u1", "u1", "u1@test.com", "admin")
	require.NoError(t, err)

	_, err = newTestJWT().ParseToken(pair.AccessToken)
	assert.Error(t, err, "should reject expired token")
}

func TestBcryptHashAndCheck(t *testing.T) {
	password := "Super$ecret99!"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	assert.NoError(t, auth.CheckPassword(password, hash))
}

func TestBcryptCheck_WrongPassword(t *testing.T) {
	hash, _ := auth.HashPassword("correct-password")
	err := auth.CheckPassword("wrong-password", hash)
	assert.Error(t, err)
}
