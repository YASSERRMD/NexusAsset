package server

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// Server is the full server catalog record.
type Server struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Hostname      *string    `json:"hostname,omitempty"`
	IPAddress     *string    `json:"ip_address,omitempty"`
	ServerType    string     `json:"server_type"`
	OS            *string    `json:"os,omitempty"`
	OSVersion     *string    `json:"os_version,omitempty"`
	CPUCores      *int       `json:"cpu_cores,omitempty"`
	RAMGB         *int       `json:"ram_gb,omitempty"`
	DiskGB        *int       `json:"disk_gb,omitempty"`
	Datacenter    *string    `json:"datacenter,omitempty"`
	Region        *string    `json:"region,omitempty"`
	CloudProvider *string    `json:"cloud_provider,omitempty"`
	EnvironmentID *string    `json:"environment_id,omitempty"`
	OwnerTeamID   *string    `json:"owner_team_id,omitempty"`
	ManagedBy     *string    `json:"managed_by,omitempty"`
	Status        string     `json:"status"`
	Tags          []string   `json:"tags,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedBy     *string    `json:"created_by,omitempty"`
	UpdatedBy     *string    `json:"updated_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// ListParams holds query filters for servers.
type ListParams struct {
	Page          int
	Limit         int
	Search        string
	ServerType    string
	Status        string
	CloudProvider string
	EnvironmentID string
}

// CreateServerInput is the validated create payload.
type CreateServerInput struct {
	Name          string   `json:"name"         validate:"required"`
	Hostname      *string  `json:"hostname,omitempty"`
	IPAddress     *string  `json:"ip_address,omitempty"`
	ServerType    string   `json:"server_type"  validate:"required,oneof=bare_metal vm container cloud_instance"`
	OS            *string  `json:"os,omitempty"`
	OSVersion     *string  `json:"os_version,omitempty"`
	CPUCores      *int     `json:"cpu_cores,omitempty"`
	RAMGB         *int     `json:"ram_gb,omitempty"`
	DiskGB        *int     `json:"disk_gb,omitempty"`
	Datacenter    *string  `json:"datacenter,omitempty"`
	Region        *string  `json:"region,omitempty"`
	CloudProvider *string  `json:"cloud_provider,omitempty"`
	EnvironmentID *string  `json:"environment_id,omitempty"`
	OwnerTeamID   *string  `json:"owner_team_id,omitempty"`
	ManagedBy     *string  `json:"managed_by,omitempty"`
	Status        string   `json:"status"       validate:"omitempty,oneof=active decommissioned maintenance reserved"`
	Tags          []string `json:"tags,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
	CreatedBy     *string  `json:"-"` // set from JWT claims
}

// UpdateServerInput allows partial updates.
type UpdateServerInput struct {
	Name          *string  `json:"name,omitempty"`
	Hostname      *string  `json:"hostname,omitempty"`
	IPAddress     *string  `json:"ip_address,omitempty"`
	ServerType    *string  `json:"server_type,omitempty"  validate:"omitempty,oneof=bare_metal vm container cloud_instance"`
	OS            *string  `json:"os,omitempty"`
	OSVersion     *string  `json:"os_version,omitempty"`
	CPUCores      *int     `json:"cpu_cores,omitempty"`
	RAMGB         *int     `json:"ram_gb,omitempty"`
	DiskGB        *int     `json:"disk_gb,omitempty"`
	Datacenter    *string  `json:"datacenter,omitempty"`
	Region        *string  `json:"region,omitempty"`
	CloudProvider *string  `json:"cloud_provider,omitempty"`
	EnvironmentID *string  `json:"environment_id,omitempty"`
	OwnerTeamID   *string  `json:"owner_team_id,omitempty"`
	ManagedBy     *string  `json:"managed_by,omitempty"`
	Status        *string  `json:"status,omitempty"       validate:"omitempty,oneof=active decommissioned maintenance reserved"`
	Tags          []string `json:"tags,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
	UpdatedBy     *string  `json:"-"` // set from JWT claims
}

// Repository defines the data access interface.
type Repository interface {
	List(ctx context.Context, p ListParams) ([]Server, int64, error)
	GetByID(ctx context.Context, id string) (*Server, error)
	Create(ctx context.Context, input CreateServerInput) (*Server, error)
	Update(ctx context.Context, id string, input UpdateServerInput) (*Server, error)
	Delete(ctx context.Context, id string) error
}

const serverCols = `id, name, hostname, ip_address, server_type, os, os_version,
    cpu_cores, ram_gb, disk_gb, datacenter, region, cloud_provider, environment_id,
    owner_team_id, managed_by, status, tags, notes, created_by, updated_by,
    created_at, updated_at, deleted_at`

type pgRepository struct{ db *pgxpool.Pool }

// NewRepository creates a PostgreSQL-backed server repository.
func NewRepository(db *pgxpool.Pool) Repository { return &pgRepository{db: db} }

func scanServer(row interface{ Scan(...any) error }) (*Server, error) {
	s := &Server{}
	var tags pq.StringArray
	err := row.Scan(
		&s.ID, &s.Name, &s.Hostname, &s.IPAddress, &s.ServerType, &s.OS, &s.OSVersion,
		&s.CPUCores, &s.RAMGB, &s.DiskGB, &s.Datacenter, &s.Region, &s.CloudProvider,
		&s.EnvironmentID, &s.OwnerTeamID, &s.ManagedBy, &s.Status, &tags, &s.Notes,
		&s.CreatedBy, &s.UpdatedBy, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)
	s.Tags = []string(tags)
	return s, err
}

func (r *pgRepository) List(ctx context.Context, p ListParams) ([]Server, int64, error) {
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
		where += fmt.Sprintf(" AND (name ILIKE $%d OR hostname ILIKE $%d OR ip_address ILIKE $%d)", idx, idx, idx)
		args = append(args, "%"+p.Search+"%")
		idx++
	}
	if p.ServerType != "" {
		where += fmt.Sprintf(" AND server_type = $%d", idx)
		args = append(args, p.ServerType)
		idx++
	}
	if p.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, p.Status)
		idx++
	}
	if p.CloudProvider != "" {
		where += fmt.Sprintf(" AND cloud_provider = $%d", idx)
		args = append(args, p.CloudProvider)
		idx++
	}
	if p.EnvironmentID != "" {
		where += fmt.Sprintf(" AND environment_id = $%d", idx)
		args = append(args, p.EnvironmentID)
		idx++
	}
	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM servers "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("servers.List count: %w", err)
	}
	args = append(args, p.Limit, offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf("SELECT %s FROM servers %s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		serverCols, where, idx, idx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var servers []Server
	for rows.Next() {
		s, err := scanServer(rows)
		if err != nil {
			return nil, 0, err
		}
		servers = append(servers, *s)
	}
	return servers, total, rows.Err()
}

func (r *pgRepository) GetByID(ctx context.Context, id string) (*Server, error) {
	q := fmt.Sprintf("SELECT %s FROM servers WHERE id = $1 AND deleted_at IS NULL", serverCols)
	return scanServer(r.db.QueryRow(ctx, q, id))
}

func (r *pgRepository) Create(ctx context.Context, input CreateServerInput) (*Server, error) {
	if input.Status == "" {
		input.Status = "active"
	}
	q := fmt.Sprintf(`INSERT INTO servers
	    (name, hostname, ip_address, server_type, os, os_version, cpu_cores, ram_gb, disk_gb,
	     datacenter, region, cloud_provider, environment_id, owner_team_id, managed_by,
	     status, tags, notes, created_by, updated_by)
	    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$19)
	    RETURNING %s`, serverCols)
	return scanServer(r.db.QueryRow(ctx, q,
		input.Name, input.Hostname, input.IPAddress, input.ServerType, input.OS, input.OSVersion,
		input.CPUCores, input.RAMGB, input.DiskGB, input.Datacenter, input.Region, input.CloudProvider,
		input.EnvironmentID, input.OwnerTeamID, input.ManagedBy, input.Status,
		pq.StringArray(input.Tags), input.Notes, input.CreatedBy))
}

func (r *pgRepository) Update(ctx context.Context, id string, input UpdateServerInput) (*Server, error) {
	set := "updated_at = now()"
	args := []any{}
	idx := 1
	if input.Name != nil {
		set += fmt.Sprintf(", name = $%d", idx)
		args = append(args, *input.Name)
		idx++
	}
	if input.Hostname != nil {
		set += fmt.Sprintf(", hostname = $%d", idx)
		args = append(args, *input.Hostname)
		idx++
	}
	if input.IPAddress != nil {
		set += fmt.Sprintf(", ip_address = $%d", idx)
		args = append(args, *input.IPAddress)
		idx++
	}
	if input.ServerType != nil {
		set += fmt.Sprintf(", server_type = $%d", idx)
		args = append(args, *input.ServerType)
		idx++
	}
	if input.OS != nil {
		set += fmt.Sprintf(", os = $%d", idx)
		args = append(args, *input.OS)
		idx++
	}
	if input.OSVersion != nil {
		set += fmt.Sprintf(", os_version = $%d", idx)
		args = append(args, *input.OSVersion)
		idx++
	}
	if input.CPUCores != nil {
		set += fmt.Sprintf(", cpu_cores = $%d", idx)
		args = append(args, *input.CPUCores)
		idx++
	}
	if input.RAMGB != nil {
		set += fmt.Sprintf(", ram_gb = $%d", idx)
		args = append(args, *input.RAMGB)
		idx++
	}
	if input.DiskGB != nil {
		set += fmt.Sprintf(", disk_gb = $%d", idx)
		args = append(args, *input.DiskGB)
		idx++
	}
	if input.Datacenter != nil {
		set += fmt.Sprintf(", datacenter = $%d", idx)
		args = append(args, *input.Datacenter)
		idx++
	}
	if input.Region != nil {
		set += fmt.Sprintf(", region = $%d", idx)
		args = append(args, *input.Region)
		idx++
	}
	if input.CloudProvider != nil {
		set += fmt.Sprintf(", cloud_provider = $%d", idx)
		args = append(args, *input.CloudProvider)
		idx++
	}
	if input.EnvironmentID != nil {
		set += fmt.Sprintf(", environment_id = $%d", idx)
		args = append(args, *input.EnvironmentID)
		idx++
	}
	if input.OwnerTeamID != nil {
		set += fmt.Sprintf(", owner_team_id = $%d", idx)
		args = append(args, *input.OwnerTeamID)
		idx++
	}
	if input.ManagedBy != nil {
		set += fmt.Sprintf(", managed_by = $%d", idx)
		args = append(args, *input.ManagedBy)
		idx++
	}
	if input.Status != nil {
		set += fmt.Sprintf(", status = $%d", idx)
		args = append(args, *input.Status)
		idx++
	}
	if input.Tags != nil {
		set += fmt.Sprintf(", tags = $%d", idx)
		args = append(args, pq.StringArray(input.Tags))
		idx++
	}
	if input.Notes != nil {
		set += fmt.Sprintf(", notes = $%d", idx)
		args = append(args, *input.Notes)
		idx++
	}
	if input.UpdatedBy != nil {
		set += fmt.Sprintf(", updated_by = $%d", idx)
		args = append(args, *input.UpdatedBy)
		idx++
	}
	args = append(args, id)
	return scanServer(r.db.QueryRow(ctx,
		fmt.Sprintf("UPDATE servers SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, serverCols), args...))
}

func (r *pgRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE servers SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}
