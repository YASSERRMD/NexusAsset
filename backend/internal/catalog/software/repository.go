package software

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// ─── Domain models ───────────────────────────────────────────────────────────

type Software struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	DisplayName  string     `json:"display_name"`
	Description  *string    `json:"description,omitempty"`
	Version      *string    `json:"version,omitempty"`
	SoftwareKind string     `json:"software_kind"` // "inhouse" | "vendor"
	CategoryID   *string    `json:"category_id,omitempty"`
	TypeID       *string    `json:"type_id,omitempty"`
	Criticality  string     `json:"criticality"`
	Status       string     `json:"status"`
	Architecture *string    `json:"architecture,omitempty"`
	OwnerTeamID  *string    `json:"owner_team_id,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedBy    *string    `json:"created_by,omitempty"`
	UpdatedBy    *string    `json:"updated_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type InhouseDetails struct {
	SoftwareID       string     `json:"software_id"`
	CicdPlatform     *string    `json:"cicd_platform,omitempty"`
	CicdPipelineURL  *string    `json:"cicd_pipeline_url,omitempty"`
	LastReleasedAt   *time.Time `json:"last_released_at,omitempty"`
	NextReleaseAt    *time.Time `json:"next_release_at,omitempty"`
	DocumentationURL *string    `json:"documentation_url,omitempty"`
	InternalNotes    *string    `json:"internal_notes,omitempty"`
}

type VendorDetails struct {
	SoftwareID             string  `json:"software_id"`
	VendorID               *string `json:"vendor_id,omitempty"`
	LicenseType            *string `json:"license_type,omitempty"`
	LicenseKey             *string `json:"license_key,omitempty"`
	LicenseExpiryDate      *string `json:"license_expiry_date,omitempty"`
	SupportContractRef     *string `json:"support_contract_ref,omitempty"`
	InstalledVersion       *string `json:"installed_version,omitempty"`
	LatestAvailableVersion *string `json:"latest_available_version,omitempty"`
	EolDate                *string `json:"eol_date,omitempty"`
	PurchaseDate           *string `json:"purchase_date,omitempty"`
}

type Responsibility struct {
	ID         string    `json:"id"`
	SoftwareID string    `json:"software_id"`
	PersonID   *string   `json:"person_id,omitempty"`
	RoleID     *string   `json:"role_id,omitempty"`
	IsPrimary  bool      `json:"is_primary"`
	AssignedAt time.Time `json:"assigned_at"`
	Notes      *string   `json:"notes,omitempty"`
}

type Repository struct {
	ID            string     `json:"id"`
	SoftwareID    string     `json:"software_id"`
	Name          string     `json:"name"`
	RepoType      string     `json:"repo_type"`
	URL           string     `json:"url"`
	DefaultBranch string     `json:"default_branch"`
	PlatformID    *string    `json:"platform_id,omitempty"`
	IsPrivate     bool       `json:"is_private"`
	LastCommitAt  *time.Time `json:"last_commit_at,omitempty"`
	PrimaryDevID  *string    `json:"primary_dev_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Deployment struct {
	ID              string     `json:"id"`
	SoftwareID      string     `json:"software_id"`
	ServerID        *string    `json:"server_id,omitempty"`
	EnvironmentID   *string    `json:"environment_id,omitempty"`
	Version         *string    `json:"version,omitempty"`
	DeployedURL     *string    `json:"deployed_url,omitempty"`
	Port            *int       `json:"port,omitempty"`
	DeployPath      *string    `json:"deploy_path,omitempty"`
	DeploymentType  *string    `json:"deployment_type,omitempty"`
	Status          string     `json:"status"`
	HealthStatus    string     `json:"health_status"`
	DeployedBy      *string    `json:"deployed_by,omitempty"`
	DeployedAt      *time.Time `json:"deployed_at,omitempty"`
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TechStack struct {
	ID             string  `json:"id"`
	SoftwareID     string  `json:"software_id"`
	Technology     string  `json:"technology"`
	TechCategoryID *string `json:"tech_category_id,omitempty"`
	Version        *string `json:"version,omitempty"`
}

type DatabaseLink struct {
	ID                 string  `json:"id"`
	SoftwareID         string  `json:"software_id"`
	DatabaseInstanceID *string `json:"database_instance_id,omitempty"`
	AccessType         string  `json:"access_type"`
	SchemaName         *string `json:"schema_name,omitempty"`
	Notes              *string `json:"notes,omitempty"`
}

// ─── List params ──────────────────────────────────────────────────────────────

type ListParams struct {
	Page        int
	Limit       int
	Search      string
	Kind        string
	CategoryID  string
	TypeID      string
	Status      string
	Criticality string
}

// ─── Input types ─────────────────────────────────────────────────────────────

type CreateSoftwareInput struct {
	Name         string   `json:"name"          validate:"required"`
	DisplayName  string   `json:"display_name"  validate:"required"`
	Description  *string  `json:"description,omitempty"`
	Version      *string  `json:"version,omitempty"`
	SoftwareKind string   `json:"software_kind" validate:"required,oneof=inhouse vendor"`
	CategoryID   *string  `json:"category_id,omitempty"`
	TypeID       *string  `json:"type_id,omitempty"`
	Criticality  string   `json:"criticality"   validate:"omitempty,oneof=critical high medium low"`
	Status       string   `json:"status"        validate:"omitempty,oneof=active in_development deprecated eol on_hold decommissioned"`
	Architecture *string  `json:"architecture,omitempty"`
	OwnerTeamID  *string  `json:"owner_team_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
	CreatedBy    *string  `json:"-"`

	// Optional details
	InhouseDetails *InhouseDetails `json:"inhouse_details,omitempty"`
	VendorDetails  *VendorDetails  `json:"vendor_details,omitempty"`
}

type UpdateSoftwareInput struct {
	Name         *string  `json:"name,omitempty"`
	DisplayName  *string  `json:"display_name,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Version      *string  `json:"version,omitempty"`
	CategoryID   *string  `json:"category_id,omitempty"`
	TypeID       *string  `json:"type_id,omitempty"`
	Criticality  *string  `json:"criticality,omitempty"  validate:"omitempty,oneof=critical high medium low"`
	Status       *string  `json:"status,omitempty"       validate:"omitempty,oneof=active in_development deprecated eol on_hold decommissioned"`
	Architecture *string  `json:"architecture,omitempty"`
	OwnerTeamID  *string  `json:"owner_team_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
	UpdatedBy    *string  `json:"-"`
}

type AddResponsibilityInput struct {
	PersonID  *string `json:"person_id,omitempty"`
	RoleID    *string `json:"role_id,omitempty"`
	IsPrimary bool    `json:"is_primary"`
	Notes     *string `json:"notes,omitempty"`
}

type AddRepoInput struct {
	Name          string  `json:"name"      validate:"required"`
	RepoType      string  `json:"repo_type" validate:"required"`
	URL           string  `json:"url"       validate:"required,url"`
	DefaultBranch string  `json:"default_branch"`
	PlatformID    *string `json:"platform_id,omitempty"`
	IsPrivate     bool    `json:"is_private"`
	PrimaryDevID  *string `json:"primary_dev_id,omitempty"`
}

type AddDeploymentInput struct {
	ServerID       *string `json:"server_id,omitempty"`
	EnvironmentID  *string `json:"environment_id,omitempty"`
	Version        *string `json:"version,omitempty"`
	DeployedURL    *string `json:"deployed_url,omitempty"`
	Port           *int    `json:"port,omitempty"`
	DeployPath     *string `json:"deploy_path,omitempty"`
	DeploymentType *string `json:"deployment_type,omitempty"`
	DeployedBy     *string `json:"deployed_by,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}

type AddTechStackInput struct {
	Technology     string  `json:"technology" validate:"required"`
	TechCategoryID *string `json:"tech_category_id,omitempty"`
	Version        *string `json:"version,omitempty"`
}

type AddDatabaseLinkInput struct {
	DatabaseInstanceID *string `json:"database_instance_id,omitempty"`
	AccessType         string  `json:"access_type" validate:"required,oneof=read write read_write admin"`
	SchemaName         *string `json:"schema_name,omitempty"`
	Notes              *string `json:"notes,omitempty"`
}

// ─── Repository interface ─────────────────────────────────────────────────────

type SoftwareRepo interface {
	List(ctx context.Context, p ListParams) ([]Software, int64, error)
	GetByID(ctx context.Context, id string) (*Software, error)
	Create(ctx context.Context, input CreateSoftwareInput) (*Software, error)
	Update(ctx context.Context, id string, input UpdateSoftwareInput) (*Software, error)
	Delete(ctx context.Context, id string) error

	GetInhouseDetails(ctx context.Context, id string) (*InhouseDetails, error)
	UpsertInhouseDetails(ctx context.Context, id string, d InhouseDetails) error
	GetVendorDetails(ctx context.Context, id string) (*VendorDetails, error)
	UpsertVendorDetails(ctx context.Context, id string, d VendorDetails) error

	ListResponsibilities(ctx context.Context, softwareID string) ([]Responsibility, error)
	AddResponsibility(ctx context.Context, softwareID string, input AddResponsibilityInput) (*Responsibility, error)
	DeleteResponsibility(ctx context.Context, softwareID, respID string) error

	ListRepos(ctx context.Context, softwareID string) ([]Repository, error)
	AddRepo(ctx context.Context, softwareID string, input AddRepoInput) (*Repository, error)
	DeleteRepo(ctx context.Context, softwareID, repoID string) error

	ListDeployments(ctx context.Context, softwareID string) ([]Deployment, error)
	AddDeployment(ctx context.Context, softwareID string, input AddDeploymentInput) (*Deployment, error)
	UpdateDeploymentHealth(ctx context.Context, softwareID, depID, health string) error
	DeleteDeployment(ctx context.Context, softwareID, depID string) error

	ListTechStack(ctx context.Context, softwareID string) ([]TechStack, error)
	AddTechStack(ctx context.Context, softwareID string, input AddTechStackInput) (*TechStack, error)
	DeleteTechStack(ctx context.Context, softwareID, techID string) error

	ListDatabaseLinks(ctx context.Context, softwareID string) ([]DatabaseLink, error)
	AddDatabaseLink(ctx context.Context, softwareID string, input AddDatabaseLinkInput) (*DatabaseLink, error)
	DeleteDatabaseLink(ctx context.Context, softwareID, linkID string) error
}

// ─── PostgreSQL implementation ────────────────────────────────────────────────

type pgRepo struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) SoftwareRepo { return &pgRepo{db: db} }

const softwareCols = `id, name, display_name, description, version, software_kind,
    category_id, type_id, criticality, status, architecture, owner_team_id,
    tags, notes, created_by, updated_by, created_at, updated_at, deleted_at`

func scanSoftware(row interface{ Scan(...any) error }) (*Software, error) {
	s := &Software{}
	var tags pq.StringArray
	err := row.Scan(
		&s.ID, &s.Name, &s.DisplayName, &s.Description, &s.Version, &s.SoftwareKind,
		&s.CategoryID, &s.TypeID, &s.Criticality, &s.Status, &s.Architecture,
		&s.OwnerTeamID, &tags, &s.Notes, &s.CreatedBy, &s.UpdatedBy,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)
	s.Tags = []string(tags)
	return s, err
}

func (r *pgRepo) List(ctx context.Context, p ListParams) ([]Software, int64, error) {
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
		where += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d OR description ILIKE $%d)", idx, idx, idx)
		args = append(args, "%"+p.Search+"%")
		idx++
	}
	if p.Kind != "" {
		where += fmt.Sprintf(" AND software_kind = $%d", idx)
		args = append(args, p.Kind)
		idx++
	}
	if p.CategoryID != "" {
		where += fmt.Sprintf(" AND category_id = $%d", idx)
		args = append(args, p.CategoryID)
		idx++
	}
	if p.TypeID != "" {
		where += fmt.Sprintf(" AND type_id = $%d", idx)
		args = append(args, p.TypeID)
		idx++
	}
	if p.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, p.Status)
		idx++
	}
	if p.Criticality != "" {
		where += fmt.Sprintf(" AND criticality = $%d", idx)
		args = append(args, p.Criticality)
		idx++
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM software "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("software.List count: %w", err)
	}
	args = append(args, p.Limit, offset)
	rows, err := r.db.Query(ctx, fmt.Sprintf("SELECT %s FROM software %s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		softwareCols, where, idx, idx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []Software
	for rows.Next() {
		s, err := scanSoftware(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *s)
	}
	return items, total, rows.Err()
}

func (r *pgRepo) GetByID(ctx context.Context, id string) (*Software, error) {
	return scanSoftware(r.db.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM software WHERE id = $1 AND deleted_at IS NULL", softwareCols), id))
}

func (r *pgRepo) Create(ctx context.Context, input CreateSoftwareInput) (*Software, error) {
	if input.Criticality == "" {
		input.Criticality = "medium"
	}
	if input.Status == "" {
		input.Status = "active"
	}
	q := fmt.Sprintf(`INSERT INTO software
	    (name, display_name, description, version, software_kind, category_id, type_id,
	     criticality, status, architecture, owner_team_id, tags, notes, created_by, updated_by)
	    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14) RETURNING %s`, softwareCols)
	s, err := scanSoftware(r.db.QueryRow(ctx, q,
		input.Name, input.DisplayName, input.Description, input.Version, input.SoftwareKind,
		input.CategoryID, input.TypeID, input.Criticality, input.Status, input.Architecture,
		input.OwnerTeamID, pq.StringArray(input.Tags), input.Notes, input.CreatedBy))
	if err != nil {
		return nil, fmt.Errorf("software.Create: %w", err)
	}

	// Insert extension tables if provided
	if s.SoftwareKind == "inhouse" && input.InhouseDetails != nil {
		input.InhouseDetails.SoftwareID = s.ID
		_ = r.UpsertInhouseDetails(ctx, s.ID, *input.InhouseDetails)
	}
	if s.SoftwareKind == "vendor" && input.VendorDetails != nil {
		input.VendorDetails.SoftwareID = s.ID
		_ = r.UpsertVendorDetails(ctx, s.ID, *input.VendorDetails)
	}
	return s, nil
}

func (r *pgRepo) Update(ctx context.Context, id string, input UpdateSoftwareInput) (*Software, error) {
	set := "updated_at = now()"
	args := []any{}
	idx := 1
	if input.Name != nil {
		set += fmt.Sprintf(", name = $%d", idx)
		args = append(args, *input.Name)
		idx++
	}
	if input.DisplayName != nil {
		set += fmt.Sprintf(", display_name = $%d", idx)
		args = append(args, *input.DisplayName)
		idx++
	}
	if input.Description != nil {
		set += fmt.Sprintf(", description = $%d", idx)
		args = append(args, *input.Description)
		idx++
	}
	if input.Version != nil {
		set += fmt.Sprintf(", version = $%d", idx)
		args = append(args, *input.Version)
		idx++
	}
	if input.CategoryID != nil {
		set += fmt.Sprintf(", category_id = $%d", idx)
		args = append(args, *input.CategoryID)
		idx++
	}
	if input.TypeID != nil {
		set += fmt.Sprintf(", type_id = $%d", idx)
		args = append(args, *input.TypeID)
		idx++
	}
	if input.Criticality != nil {
		set += fmt.Sprintf(", criticality = $%d", idx)
		args = append(args, *input.Criticality)
		idx++
	}
	if input.Status != nil {
		set += fmt.Sprintf(", status = $%d", idx)
		args = append(args, *input.Status)
		idx++
	}
	if input.Architecture != nil {
		set += fmt.Sprintf(", architecture = $%d", idx)
		args = append(args, *input.Architecture)
		idx++
	}
	if input.OwnerTeamID != nil {
		set += fmt.Sprintf(", owner_team_id = $%d", idx)
		args = append(args, *input.OwnerTeamID)
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
	return scanSoftware(r.db.QueryRow(ctx,
		fmt.Sprintf("UPDATE software SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, softwareCols), args...))
}

func (r *pgRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE software SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

// ─── Inhouse / Vendor details ─────────────────────────────────────────────────

func (r *pgRepo) GetInhouseDetails(ctx context.Context, id string) (*InhouseDetails, error) {
	d := &InhouseDetails{}
	err := r.db.QueryRow(ctx,
		`SELECT software_id, cicd_platform, cicd_pipeline_url, last_released_at, next_release_at, documentation_url, internal_notes
		 FROM software_inhouse_details WHERE software_id = $1`, id).Scan(
		&d.SoftwareID, &d.CicdPlatform, &d.CicdPipelineURL, &d.LastReleasedAt, &d.NextReleaseAt, &d.DocumentationURL, &d.InternalNotes)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *pgRepo) UpsertInhouseDetails(ctx context.Context, id string, d InhouseDetails) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO software_inhouse_details (software_id, cicd_platform, cicd_pipeline_url, last_released_at, next_release_at, documentation_url, internal_notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (software_id) DO UPDATE SET
		   cicd_platform=EXCLUDED.cicd_platform, cicd_pipeline_url=EXCLUDED.cicd_pipeline_url,
		   last_released_at=EXCLUDED.last_released_at, next_release_at=EXCLUDED.next_release_at,
		   documentation_url=EXCLUDED.documentation_url, internal_notes=EXCLUDED.internal_notes`,
		id, d.CicdPlatform, d.CicdPipelineURL, d.LastReleasedAt, d.NextReleaseAt, d.DocumentationURL, d.InternalNotes)
	return err
}

func (r *pgRepo) GetVendorDetails(ctx context.Context, id string) (*VendorDetails, error) {
	d := &VendorDetails{}
	err := r.db.QueryRow(ctx,
		`SELECT software_id, vendor_id, license_type, license_key, license_expiry_date, support_contract_ref,
		        installed_version, latest_available_version, eol_date, purchase_date
		 FROM software_vendor_details WHERE software_id = $1`, id).Scan(
		&d.SoftwareID, &d.VendorID, &d.LicenseType, &d.LicenseKey, &d.LicenseExpiryDate,
		&d.SupportContractRef, &d.InstalledVersion, &d.LatestAvailableVersion, &d.EolDate, &d.PurchaseDate)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *pgRepo) UpsertVendorDetails(ctx context.Context, id string, d VendorDetails) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO software_vendor_details (software_id, vendor_id, license_type, license_key, license_expiry_date,
		   support_contract_ref, installed_version, latest_available_version, eol_date, purchase_date)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (software_id) DO UPDATE SET
		   vendor_id=EXCLUDED.vendor_id, license_type=EXCLUDED.license_type, license_key=EXCLUDED.license_key,
		   license_expiry_date=EXCLUDED.license_expiry_date, support_contract_ref=EXCLUDED.support_contract_ref,
		   installed_version=EXCLUDED.installed_version, latest_available_version=EXCLUDED.latest_available_version,
		   eol_date=EXCLUDED.eol_date, purchase_date=EXCLUDED.purchase_date`,
		id, d.VendorID, d.LicenseType, d.LicenseKey, d.LicenseExpiryDate,
		d.SupportContractRef, d.InstalledVersion, d.LatestAvailableVersion, d.EolDate, d.PurchaseDate)
	return err
}

// ─── Responsibilities ─────────────────────────────────────────────────────────

func (r *pgRepo) ListResponsibilities(ctx context.Context, softwareID string) ([]Responsibility, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, software_id, person_id, role_id, is_primary, assigned_at, notes
		 FROM software_responsibilities WHERE software_id = $1 ORDER BY is_primary DESC, assigned_at ASC`, softwareID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Responsibility
	for rows.Next() {
		var item Responsibility
		if err := rows.Scan(&item.ID, &item.SoftwareID, &item.PersonID, &item.RoleID, &item.IsPrimary, &item.AssignedAt, &item.Notes); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *pgRepo) AddResponsibility(ctx context.Context, softwareID string, input AddResponsibilityInput) (*Responsibility, error) {
	item := &Responsibility{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO software_responsibilities (software_id, person_id, role_id, is_primary, notes)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, software_id, person_id, role_id, is_primary, assigned_at, notes`,
		softwareID, input.PersonID, input.RoleID, input.IsPrimary, input.Notes).Scan(
		&item.ID, &item.SoftwareID, &item.PersonID, &item.RoleID, &item.IsPrimary, &item.AssignedAt, &item.Notes)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *pgRepo) DeleteResponsibility(ctx context.Context, softwareID, respID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM software_responsibilities WHERE id = $1 AND software_id = $2", respID, softwareID)
	return err
}

// ─── Repositories ─────────────────────────────────────────────────────────────

func (r *pgRepo) ListRepos(ctx context.Context, softwareID string) ([]Repository, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, software_id, name, repo_type, url, default_branch, platform_id, is_private, last_commit_at, primary_dev_id, created_at
		 FROM software_repositories WHERE software_id = $1 ORDER BY created_at ASC`, softwareID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Repository
	for rows.Next() {
		var item Repository
		if err := rows.Scan(&item.ID, &item.SoftwareID, &item.Name, &item.RepoType, &item.URL, &item.DefaultBranch,
			&item.PlatformID, &item.IsPrivate, &item.LastCommitAt, &item.PrimaryDevID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *pgRepo) AddRepo(ctx context.Context, softwareID string, input AddRepoInput) (*Repository, error) {
	if input.DefaultBranch == "" {
		input.DefaultBranch = "main"
	}
	item := &Repository{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO software_repositories (software_id, name, repo_type, url, default_branch, platform_id, is_private, primary_dev_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, software_id, name, repo_type, url, default_branch, platform_id, is_private, last_commit_at, primary_dev_id, created_at`,
		softwareID, input.Name, input.RepoType, input.URL, input.DefaultBranch, input.PlatformID, input.IsPrivate, input.PrimaryDevID).Scan(
		&item.ID, &item.SoftwareID, &item.Name, &item.RepoType, &item.URL, &item.DefaultBranch,
		&item.PlatformID, &item.IsPrivate, &item.LastCommitAt, &item.PrimaryDevID, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *pgRepo) DeleteRepo(ctx context.Context, softwareID, repoID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM software_repositories WHERE id = $1 AND software_id = $2", repoID, softwareID)
	return err
}

// ─── Deployments ──────────────────────────────────────────────────────────────

func (r *pgRepo) ListDeployments(ctx context.Context, softwareID string) ([]Deployment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, software_id, server_id, environment_id, version, deployed_url, port, deploy_path,
		        deployment_type, status, health_status, deployed_by, deployed_at, last_health_check, notes, created_at, updated_at
		 FROM software_deployments WHERE software_id = $1 ORDER BY created_at DESC`, softwareID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Deployment
	for rows.Next() {
		var d Deployment
		if err := rows.Scan(&d.ID, &d.SoftwareID, &d.ServerID, &d.EnvironmentID, &d.Version, &d.DeployedURL,
			&d.Port, &d.DeployPath, &d.DeploymentType, &d.Status, &d.HealthStatus, &d.DeployedBy,
			&d.DeployedAt, &d.LastHealthCheck, &d.Notes, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *pgRepo) AddDeployment(ctx context.Context, softwareID string, input AddDeploymentInput) (*Deployment, error) {
	d := &Deployment{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO software_deployments (software_id, server_id, environment_id, version, deployed_url, port, deploy_path, deployment_type, deployed_by, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING id, software_id, server_id, environment_id, version, deployed_url, port, deploy_path, deployment_type, status, health_status, deployed_by, deployed_at, last_health_check, notes, created_at, updated_at`,
		softwareID, input.ServerID, input.EnvironmentID, input.Version, input.DeployedURL,
		input.Port, input.DeployPath, input.DeploymentType, input.DeployedBy, input.Notes).Scan(
		&d.ID, &d.SoftwareID, &d.ServerID, &d.EnvironmentID, &d.Version, &d.DeployedURL,
		&d.Port, &d.DeployPath, &d.DeploymentType, &d.Status, &d.HealthStatus, &d.DeployedBy,
		&d.DeployedAt, &d.LastHealthCheck, &d.Notes, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *pgRepo) UpdateDeploymentHealth(ctx context.Context, softwareID, depID, health string) error {
	_, err := r.db.Exec(ctx,
		"UPDATE software_deployments SET health_status = $1, last_health_check = now(), updated_at = now() WHERE id = $2 AND software_id = $3",
		health, depID, softwareID)
	return err
}

func (r *pgRepo) DeleteDeployment(ctx context.Context, softwareID, depID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM software_deployments WHERE id = $1 AND software_id = $2", depID, softwareID)
	return err
}

// ─── Tech Stack ───────────────────────────────────────────────────────────────

func (r *pgRepo) ListTechStack(ctx context.Context, softwareID string) ([]TechStack, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, software_id, technology, tech_category_id, version
		 FROM software_tech_stack WHERE software_id = $1 ORDER BY technology ASC`, softwareID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TechStack
	for rows.Next() {
		var t TechStack
		if err := rows.Scan(&t.ID, &t.SoftwareID, &t.Technology, &t.TechCategoryID, &t.Version); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *pgRepo) AddTechStack(ctx context.Context, softwareID string, input AddTechStackInput) (*TechStack, error) {
	t := &TechStack{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO software_tech_stack (software_id, technology, tech_category_id, version) VALUES ($1,$2,$3,$4)
		 RETURNING id, software_id, technology, tech_category_id, version`,
		softwareID, input.Technology, input.TechCategoryID, input.Version).Scan(
		&t.ID, &t.SoftwareID, &t.Technology, &t.TechCategoryID, &t.Version)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *pgRepo) DeleteTechStack(ctx context.Context, softwareID, techID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM software_tech_stack WHERE id = $1 AND software_id = $2", techID, softwareID)
	return err
}

// ─── Database Links ───────────────────────────────────────────────────────────

func (r *pgRepo) ListDatabaseLinks(ctx context.Context, softwareID string) ([]DatabaseLink, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, software_id, database_instance_id, access_type, schema_name, notes
		 FROM software_database_links WHERE software_id = $1`, softwareID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DatabaseLink
	for rows.Next() {
		var l DatabaseLink
		if err := rows.Scan(&l.ID, &l.SoftwareID, &l.DatabaseInstanceID, &l.AccessType, &l.SchemaName, &l.Notes); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *pgRepo) AddDatabaseLink(ctx context.Context, softwareID string, input AddDatabaseLinkInput) (*DatabaseLink, error) {
	l := &DatabaseLink{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO software_database_links (software_id, database_instance_id, access_type, schema_name, notes) VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, software_id, database_instance_id, access_type, schema_name, notes`,
		softwareID, input.DatabaseInstanceID, input.AccessType, input.SchemaName, input.Notes).Scan(
		&l.ID, &l.SoftwareID, &l.DatabaseInstanceID, &l.AccessType, &l.SchemaName, &l.Notes)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *pgRepo) DeleteDatabaseLink(ctx context.Context, softwareID, linkID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM software_database_links WHERE id = $1 AND software_id = $2", linkID, softwareID)
	return err
}
