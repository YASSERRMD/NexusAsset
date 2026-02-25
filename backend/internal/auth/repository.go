package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User represents a user record as returned from the DB.
type User struct {
	ID           string
	PersonID     *string
	Username     string
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Repository defines the data access interface for auth.
type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateLastLogin(ctx context.Context, id string, t time.Time) error
}

// pgRepository is the PostgreSQL implementation.
type pgRepository struct {
	db *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed auth repository.
func NewRepository(db *pgxpool.Pool) Repository {
	return &pgRepository{db: db}
}

// GetUserByEmail fetches an active user by email.
//
// Example query:
//
//	SELECT id, person_id, username, email, password_hash, role, is_active, last_login_at
//	FROM users WHERE email = $1 AND is_active = true
func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
		SELECT id, person_id, username, email, password_hash, role, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
		LIMIT 1`

	u := &User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.PersonID, &u.Username, &u.Email,
		&u.PasswordHash, &u.Role, &u.IsActive,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail: %w", err)
	}
	return u, nil
}

// GetUserByID fetches an active user by UUID.
func (r *pgRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	const q = `
		SELECT id, person_id, username, email, password_hash, role, is_active,
		       last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1`

	u := &User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.PersonID, &u.Username, &u.Email,
		&u.PasswordHash, &u.Role, &u.IsActive,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

// UpdateLastLogin records the login timestamp for the given user ID.
func (r *pgRepository) UpdateLastLogin(ctx context.Context, id string, t time.Time) error {
	const q = `UPDATE users SET last_login_at = $2, updated_at = now() WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, id, t); err != nil {
		return fmt.Errorf("UpdateLastLogin: %w", err)
	}
	return nil
}
