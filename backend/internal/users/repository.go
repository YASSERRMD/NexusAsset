package users

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User is the full user record including hashed password.
type User struct {
	ID          string     `json:"id"`
	PersonID    *string    `json:"person_id,omitempty"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListParams holds query parameters for listing users.
type ListParams struct {
	Page   int
	Limit  int
	Search string
	Role   string
}

// Repository defines the data access interface for users.
type Repository interface {
	List(ctx context.Context, p ListParams) ([]User, int64, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, input CreateInput) (*User, error)
	Update(ctx context.Context, id string, input UpdateInput) (*User, error)
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, id, role string) (*User, error)
}

// CreateInput is validated before being passed to the repository.
type CreateInput struct {
	Username     string
	Email        string
	PasswordHash string
	Role         string
	PersonID     *string
}

// UpdateInput allows partial updates.
type UpdateInput struct {
	Username *string
	Email    *string
	Role     *string
	IsActive *bool
	PersonID *string
}

type pgRepository struct {
	db *pgxpool.Pool
}

// NewRepository creates a users repository.
func NewRepository(db *pgxpool.Pool) Repository {
	return &pgRepository{db: db}
}

const userCols = `id, person_id, username, email, role, is_active, last_login_at, created_at, updated_at`

func scanUser(row interface {
	Scan(...any) error
}) (*User, error) {
	u := &User{}
	if err := row.Scan(
		&u.ID, &u.PersonID, &u.Username, &u.Email,
		&u.Role, &u.IsActive, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return u, nil
}

// List returns a paginated list of users with optional filters.
//
// Example:
//
//	users, total, err := repo.List(ctx, ListParams{Page: 1, Limit: 20})
func (r *pgRepository) List(ctx context.Context, p ListParams) ([]User, int64, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.Limit

	args := []any{}
	where := "WHERE 1=1"
	idx := 1

	if p.Role != "" {
		where += fmt.Sprintf(" AND role = $%d", idx)
		args = append(args, p.Role)
		idx++
	}
	if p.Search != "" {
		where += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", idx, idx)
		args = append(args, "%"+p.Search+"%")
		idx++
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("users.List count: %w", err)
	}

	args = append(args, p.Limit, offset)
	rows, err := r.db.Query(ctx,
		fmt.Sprintf("SELECT %s FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
			userCols, where, idx, idx+1),
		args...)
	if err != nil {
		return nil, 0, fmt.Errorf("users.List query: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("users.List scan: %w", err)
		}
		users = append(users, *u)
	}
	return users, total, rows.Err()
}

// GetByID returns a single user by primary key.
func (r *pgRepository) GetByID(ctx context.Context, id string) (*User, error) {
	q := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", userCols)
	u, err := scanUser(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("users.GetByID: %w", err)
	}
	return u, nil
}

// Create inserts a new user and returns the created record.
func (r *pgRepository) Create(ctx context.Context, input CreateInput) (*User, error) {
	q := fmt.Sprintf(`
		INSERT INTO users (person_id, username, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING %s`, userCols)
	u, err := scanUser(r.db.QueryRow(ctx, q,
		input.PersonID, input.Username, input.Email, input.PasswordHash, input.Role))
	if err != nil {
		return nil, fmt.Errorf("users.Create: %w", err)
	}
	return u, nil
}

// Update applies partial field updates to a user.
func (r *pgRepository) Update(ctx context.Context, id string, input UpdateInput) (*User, error) {
	set := "updated_at = now()"
	args := []any{}
	idx := 1

	if input.Username != nil {
		set += fmt.Sprintf(", username = $%d", idx)
		args = append(args, *input.Username)
		idx++
	}
	if input.Email != nil {
		set += fmt.Sprintf(", email = $%d", idx)
		args = append(args, *input.Email)
		idx++
	}
	if input.Role != nil {
		set += fmt.Sprintf(", role = $%d", idx)
		args = append(args, *input.Role)
		idx++
	}
	if input.IsActive != nil {
		set += fmt.Sprintf(", is_active = $%d", idx)
		args = append(args, *input.IsActive)
		idx++
	}
	if input.PersonID != nil {
		set += fmt.Sprintf(", person_id = $%d", idx)
		args = append(args, *input.PersonID)
		idx++
	}

	args = append(args, id)
	q := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING %s", set, idx, userCols)
	u, err := scanUser(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		return nil, fmt.Errorf("users.Update: %w", err)
	}
	return u, nil
}

// Delete soft-deletes a user by deactivating them (no hard delete on users).
func (r *pgRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET is_active = false, updated_at = now() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("users.Delete: %w", err)
	}
	return nil
}

// UpdateRole changes only the user's role.
func (r *pgRepository) UpdateRole(ctx context.Context, id, role string) (*User, error) {
	q := fmt.Sprintf(`UPDATE users SET role = $2, updated_at = now() WHERE id = $1 RETURNING %s`, userCols)
	u, err := scanUser(r.db.QueryRow(ctx, q, id, role))
	if err != nil {
		return nil, fmt.Errorf("users.UpdateRole: %w", err)
	}
	return u, nil
}
