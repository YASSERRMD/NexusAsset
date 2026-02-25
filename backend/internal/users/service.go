package users

import (
	"context"
	"errors"
	"fmt"

	"nexusasset/backend/internal/auth"
)

// Service implements users business logic.
type Service struct {
	repo Repository
}

// NewService creates a users Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ErrNotFound is returned when a user is not found.
var ErrNotFound = errors.New("user not found")

// ErrDuplicateEmail is returned when email already exists.
var ErrDuplicateEmail = errors.New("email already exists")

// ErrDuplicateUsername is returned when username already exists.
var ErrDuplicateUsername = errors.New("username already exists")

// ErrInvalidRole is returned for unknown roles.
var ErrInvalidRole = errors.New("role must be admin, contributor, or reader")

var validRoles = map[string]bool{"admin": true, "contributor": true, "reader": true}

// CreateUserInput is the request body for user creation.
type CreateUserInput struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email"    validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	Role     string  `json:"role"     validate:"required,oneof=admin contributor reader"`
	PersonID *string `json:"person_id,omitempty"`
}

// UpdateUserInput is the request body for user updates.
type UpdateUserInput struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty"    validate:"omitempty,email"`
	Role     *string `json:"role,omitempty"     validate:"omitempty,oneof=admin contributor reader"`
	IsActive *bool   `json:"is_active,omitempty"`
	PersonID *string `json:"person_id,omitempty"`
}

// List returns a paginated list of users.
func (s *Service) List(ctx context.Context, p ListParams) ([]User, int64, error) {
	return s.repo.List(ctx, p)
}

// GetByID returns a user by ID or ErrNotFound.
func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotFound, err)
	}
	return u, nil
}

// Create hashes the password and creates a user.
func (s *Service) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	if !validRoles[input.Role] {
		return nil, ErrInvalidRole
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	u, err := s.repo.Create(ctx, CreateInput{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hash,
		Role:         input.Role,
		PersonID:     input.PersonID,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Update applies partial updates to an existing user.
func (s *Service) Update(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	if input.Role != nil && !validRoles[*input.Role] {
		return nil, ErrInvalidRole
	}
	u, err := s.repo.Update(ctx, id, UpdateInput{
		Username: input.Username,
		Email:    input.Email,
		Role:     input.Role,
		IsActive: input.IsActive,
		PersonID: input.PersonID,
	})
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

// Delete deactivates a user (soft delete).
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// UpdateRole changes only a user's role.
func (s *Service) UpdateRole(ctx context.Context, id, role string) (*User, error) {
	if !validRoles[role] {
		return nil, ErrInvalidRole
	}
	u, err := s.repo.UpdateRole(ctx, id, role)
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}
	return u, nil
}
