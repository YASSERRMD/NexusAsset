package people

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Person represents a person record.
type Person struct {
	ID         string     `json:"id"`
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	Phone      *string    `json:"phone,omitempty"`
	Title      *string    `json:"title,omitempty"`
	Department *string    `json:"department,omitempty"`
	TeamID     *string    `json:"team_id,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// Team represents a team record.
type Team struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Department string     `json:"department"`
	TeamEmail  *string    `json:"team_email,omitempty"`
	ManagerID  *string    `json:"manager_id,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// ListParams holds pagination + filter query params.
type ListParams struct {
	Page       int
	Limit      int
	Search     string
	Department string
	TeamID     string
	IsActive   *bool
}

// PersonRepository defines the data access interface for persons.
type PersonRepository interface {
	List(ctx context.Context, p ListParams) ([]Person, int64, error)
	GetByID(ctx context.Context, id string) (*Person, error)
	Create(ctx context.Context, input CreatePersonInput) (*Person, error)
	Update(ctx context.Context, id string, input UpdatePersonInput) (*Person, error)
	Delete(ctx context.Context, id string) error
}

// TeamRepository defines the data access interface for teams.
type TeamRepository interface {
	List(ctx context.Context, p ListParams) ([]Team, int64, error)
	GetByID(ctx context.Context, id string) (*Team, error)
	Create(ctx context.Context, input CreateTeamInput) (*Team, error)
	Update(ctx context.Context, id string, input UpdateTeamInput) (*Team, error)
	Delete(ctx context.Context, id string) error
}

// ─── Input types ─────────────────────────────────────────────────────────────

// CreatePersonInput is the validated input for creating a person.
type CreatePersonInput struct {
	FullName   string  `json:"full_name"   validate:"required,min=2"`
	Email      string  `json:"email"       validate:"required,email"`
	Phone      *string `json:"phone,omitempty"`
	Title      *string `json:"title,omitempty"`
	Department *string `json:"department,omitempty"`
	TeamID     *string `json:"team_id,omitempty"`
}

// UpdatePersonInput allows partial updates.
type UpdatePersonInput struct {
	FullName   *string `json:"full_name,omitempty"   validate:"omitempty,min=2"`
	Email      *string `json:"email,omitempty"       validate:"omitempty,email"`
	Phone      *string `json:"phone,omitempty"`
	Title      *string `json:"title,omitempty"`
	Department *string `json:"department,omitempty"`
	TeamID     *string `json:"team_id,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// CreateTeamInput is the validated input for creating a team.
type CreateTeamInput struct {
	Name       string  `json:"name"        validate:"required,min=2"`
	Department string  `json:"department"  validate:"required"`
	TeamEmail  *string `json:"team_email,omitempty"  validate:"omitempty,email"`
	ManagerID  *string `json:"manager_id,omitempty"`
}

// UpdateTeamInput allows partial updates.
type UpdateTeamInput struct {
	Name       *string `json:"name,omitempty"        validate:"omitempty,min=2"`
	Department *string `json:"department,omitempty"`
	TeamEmail  *string `json:"team_email,omitempty"  validate:"omitempty,email"`
	ManagerID  *string `json:"manager_id,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// ─── pgPersonRepository ───────────────────────────────────────────────────────

type pgPersonRepository struct{ db *pgxpool.Pool }

// NewPersonRepository creates a PostgreSQL-backed person repository.
func NewPersonRepository(db *pgxpool.Pool) PersonRepository {
	return &pgPersonRepository{db: db}
}

const personCols = `id, full_name, email, phone, title, department, team_id,
                    is_active, created_at, updated_at, deleted_at`

func scanPerson(row interface{ Scan(...any) error }) (*Person, error) {
	p := &Person{}
	return p, row.Scan(
		&p.ID, &p.FullName, &p.Email, &p.Phone, &p.Title, &p.Department,
		&p.TeamID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
}

func (r *pgPersonRepository) List(ctx context.Context, p ListParams) ([]Person, int64, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.Limit

	where := "WHERE deleted_at IS NULL"
	args := []any{}
	idx := 1

	if p.Search != "" {
		where += fmt.Sprintf(" AND (full_name ILIKE $%d OR email ILIKE $%d)", idx, idx)
		args = append(args, "%"+p.Search+"%")
		idx++
	}
	if p.Department != "" {
		where += fmt.Sprintf(" AND department ILIKE $%d", idx)
		args = append(args, "%"+p.Department+"%")
		idx++
	}
	if p.TeamID != "" {
		where += fmt.Sprintf(" AND team_id = $%d", idx)
		args = append(args, p.TeamID)
		idx++
	}
	if p.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", idx)
		args = append(args, *p.IsActive)
		idx++
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM persons "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("persons.List count: %w", err)
	}

	args = append(args, p.Limit, offset)
	q := fmt.Sprintf("SELECT %s FROM persons %s ORDER BY full_name ASC LIMIT $%d OFFSET $%d",
		personCols, where, idx, idx+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("persons.List: %w", err)
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("persons.List scan: %w", err)
		}
		persons = append(persons, *p)
	}
	return persons, total, rows.Err()
}

func (r *pgPersonRepository) GetByID(ctx context.Context, id string) (*Person, error) {
	q := fmt.Sprintf("SELECT %s FROM persons WHERE id = $1 AND deleted_at IS NULL", personCols)
	p, err := scanPerson(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("persons.GetByID: %w", err)
	}
	return p, nil
}

func (r *pgPersonRepository) Create(ctx context.Context, input CreatePersonInput) (*Person, error) {
	q := fmt.Sprintf(`INSERT INTO persons (full_name, email, phone, title, department, team_id)
	                  VALUES ($1,$2,$3,$4,$5,$6) RETURNING %s`, personCols)
	p, err := scanPerson(r.db.QueryRow(ctx, q,
		input.FullName, input.Email, input.Phone, input.Title, input.Department, input.TeamID))
	if err != nil {
		return nil, fmt.Errorf("persons.Create: %w", err)
	}
	return p, nil
}

func (r *pgPersonRepository) Update(ctx context.Context, id string, input UpdatePersonInput) (*Person, error) {
	set := "updated_at = now()"
	args := []any{}
	idx := 1
	if input.FullName != nil {
		set += fmt.Sprintf(", full_name = $%d", idx)
		args = append(args, *input.FullName)
		idx++
	}
	if input.Email != nil {
		set += fmt.Sprintf(", email = $%d", idx)
		args = append(args, *input.Email)
		idx++
	}
	if input.Phone != nil {
		set += fmt.Sprintf(", phone = $%d", idx)
		args = append(args, *input.Phone)
		idx++
	}
	if input.Title != nil {
		set += fmt.Sprintf(", title = $%d", idx)
		args = append(args, *input.Title)
		idx++
	}
	if input.Department != nil {
		set += fmt.Sprintf(", department = $%d", idx)
		args = append(args, *input.Department)
		idx++
	}
	if input.TeamID != nil {
		set += fmt.Sprintf(", team_id = $%d", idx)
		args = append(args, *input.TeamID)
		idx++
	}
	if input.IsActive != nil {
		set += fmt.Sprintf(", is_active = $%d", idx)
		args = append(args, *input.IsActive)
		idx++
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE persons SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, personCols)
	p, err := scanPerson(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		return nil, fmt.Errorf("persons.Update: %w", err)
	}
	return p, nil
}

func (r *pgPersonRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE persons SET deleted_at = now(), is_active = false WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

// ─── pgTeamRepository ────────────────────────────────────────────────────────

type pgTeamRepository struct{ db *pgxpool.Pool }

// NewTeamRepository creates a PostgreSQL-backed team repository.
func NewTeamRepository(db *pgxpool.Pool) TeamRepository {
	return &pgTeamRepository{db: db}
}

const teamCols = `id, name, department, team_email, manager_id, is_active, created_at, updated_at, deleted_at`

func scanTeam(row interface{ Scan(...any) error }) (*Team, error) {
	t := &Team{}
	return t, row.Scan(&t.ID, &t.Name, &t.Department, &t.TeamEmail, &t.ManagerID,
		&t.IsActive, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
}

func (r *pgTeamRepository) List(ctx context.Context, p ListParams) ([]Team, int64, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.Limit
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	idx := 1
	if p.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR department ILIKE $%d)", idx, idx)
		args = append(args, "%"+p.Search+"%")
		idx++
	}
	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM teams "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("teams.List count: %w", err)
	}
	args = append(args, p.Limit, offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf("SELECT %s FROM teams %s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		teamCols, where, idx, idx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, 0, err
		}
		teams = append(teams, *t)
	}
	return teams, total, rows.Err()
}

func (r *pgTeamRepository) GetByID(ctx context.Context, id string) (*Team, error) {
	q := fmt.Sprintf("SELECT %s FROM teams WHERE id = $1 AND deleted_at IS NULL", teamCols)
	t, err := scanTeam(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("teams.GetByID: %w", err)
	}
	return t, nil
}

func (r *pgTeamRepository) Create(ctx context.Context, input CreateTeamInput) (*Team, error) {
	q := fmt.Sprintf(`INSERT INTO teams (name, department, team_email, manager_id)
	                  VALUES ($1,$2,$3,$4) RETURNING %s`, teamCols)
	t, err := scanTeam(r.db.QueryRow(ctx, q, input.Name, input.Department, input.TeamEmail, input.ManagerID))
	if err != nil {
		return nil, fmt.Errorf("teams.Create: %w", err)
	}
	return t, nil
}

func (r *pgTeamRepository) Update(ctx context.Context, id string, input UpdateTeamInput) (*Team, error) {
	set := "updated_at = now()"
	args := []any{}
	idx := 1
	if input.Name != nil {
		set += fmt.Sprintf(", name = $%d", idx)
		args = append(args, *input.Name)
		idx++
	}
	if input.Department != nil {
		set += fmt.Sprintf(", department = $%d", idx)
		args = append(args, *input.Department)
		idx++
	}
	if input.TeamEmail != nil {
		set += fmt.Sprintf(", team_email = $%d", idx)
		args = append(args, *input.TeamEmail)
		idx++
	}
	if input.ManagerID != nil {
		set += fmt.Sprintf(", manager_id = $%d", idx)
		args = append(args, *input.ManagerID)
		idx++
	}
	if input.IsActive != nil {
		set += fmt.Sprintf(", is_active = $%d", idx)
		args = append(args, *input.IsActive)
		idx++
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE teams SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, teamCols)
	t, err := scanTeam(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		return nil, fmt.Errorf("teams.Update: %w", err)
	}
	return t, nil
}

func (r *pgTeamRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE teams SET deleted_at = now(), is_active = false WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}
