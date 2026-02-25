package dbinstances

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"nexusasset/backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

var validate = validator.New()

type DBInstance struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Engine        string     `json:"engine"`
	EngineVersion *string    `json:"engine_version,omitempty"`
	Hostname      *string    `json:"hostname,omitempty"`
	Port          *int       `json:"port,omitempty"`
	DatabaseName  *string    `json:"database_name,omitempty"`
	EnvironmentID *string    `json:"environment_id,omitempty"`
	ServerID      *string    `json:"server_id,omitempty"`
	SizeGB        *int       `json:"size_gb,omitempty"`
	Status        string     `json:"status"`
	IsManaged     bool       `json:"is_managed"`
	CloudProvider *string    `json:"cloud_provider,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type CreateInput struct {
	Name          string  `json:"name"    validate:"required"`
	Engine        string  `json:"engine"  validate:"required"`
	EngineVersion *string `json:"engine_version,omitempty"`
	Hostname      *string `json:"hostname,omitempty"`
	Port          *int    `json:"port,omitempty"`
	DatabaseName  *string `json:"database_name,omitempty"`
	EnvironmentID *string `json:"environment_id,omitempty"`
	ServerID      *string `json:"server_id,omitempty"`
	SizeGB        *int    `json:"size_gb,omitempty"`
	Status        string  `json:"status"  validate:"omitempty,oneof=active decommissioned maintenance"`
	IsManaged     bool    `json:"is_managed"`
	CloudProvider *string `json:"cloud_provider,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

type UpdateInput struct {
	Name          *string `json:"name,omitempty"`
	Engine        *string `json:"engine,omitempty"`
	EngineVersion *string `json:"engine_version,omitempty"`
	Hostname      *string `json:"hostname,omitempty"`
	Port          *int    `json:"port,omitempty"`
	DatabaseName  *string `json:"database_name,omitempty"`
	EnvironmentID *string `json:"environment_id,omitempty"`
	ServerID      *string `json:"server_id,omitempty"`
	SizeGB        *int    `json:"size_gb,omitempty"`
	Status        *string `json:"status,omitempty"`
	IsManaged     *bool   `json:"is_managed,omitempty"`
	CloudProvider *string `json:"cloud_provider,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

const cols = `id, name, engine, engine_version, hostname, port, database_name, environment_id, server_id,
    size_gb, status, is_managed, cloud_provider, notes, created_at, updated_at, deleted_at`

func scan(row interface{ Scan(...any) error }) (*DBInstance, error) {
	d := &DBInstance{}
	return d, row.Scan(&d.ID, &d.Name, &d.Engine, &d.EngineVersion, &d.Hostname, &d.Port, &d.DatabaseName,
		&d.EnvironmentID, &d.ServerID, &d.SizeGB, &d.Status, &d.IsManaged, &d.CloudProvider, &d.Notes,
		&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt)
}

type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := r.URL.Query().Get("search")
	engine := r.URL.Query().Get("engine")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	idx := 1
	if search != "" {
		where += fmt.Sprintf(" AND name ILIKE $%d", idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	if engine != "" {
		where += fmt.Sprintf(" AND engine = $%d", idx)
		args = append(args, engine)
		idx++
	}
	var total int64
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM database_instances "+where, args...).Scan(&total)
	args2 := append(args, limit, offset)
	rows, _ := h.db.Query(ctx, fmt.Sprintf("SELECT %s FROM database_instances %s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		cols, where, idx, idx+1), args2...)
	defer rows.Close()
	var items []DBInstance
	for rows.Next() {
		d, _ := scan(rows)
		items = append(items, *d)
	}
	response.OKList(w, items, response.Meta{Page: page, Limit: limit, Total: total})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	d, err := scan(h.db.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM database_instances WHERE id = $1 AND deleted_at IS NULL", cols), chi.URLParam(r, "id")))
	if err != nil {
		response.NotFound(w, "database instance not found")
		return
	}
	response.OK(w, d)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	if input.Status == "" {
		input.Status = "active"
	}
	d, err := scan(h.db.QueryRow(r.Context(),
		fmt.Sprintf(`INSERT INTO database_instances (name,engine,engine_version,hostname,port,database_name,environment_id,server_id,size_gb,status,is_managed,cloud_provider,notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING %s`, cols),
		input.Name, input.Engine, input.EngineVersion, input.Hostname, input.Port, input.DatabaseName,
		input.EnvironmentID, input.ServerID, input.SizeGB, input.Status, input.IsManaged, input.CloudProvider, input.Notes))
	if err != nil {
		response.InternalError(w, "failed to create database instance")
		return
	}
	response.Created(w, d)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	set := "updated_at = now()"
	args := []any{}
	idx := 1
	if input.Name != nil {
		set += fmt.Sprintf(", name = $%d", idx)
		args = append(args, *input.Name)
		idx++
	}
	if input.Engine != nil {
		set += fmt.Sprintf(", engine = $%d", idx)
		args = append(args, *input.Engine)
		idx++
	}
	if input.EngineVersion != nil {
		set += fmt.Sprintf(", engine_version = $%d", idx)
		args = append(args, *input.EngineVersion)
		idx++
	}
	if input.Hostname != nil {
		set += fmt.Sprintf(", hostname = $%d", idx)
		args = append(args, *input.Hostname)
		idx++
	}
	if input.Port != nil {
		set += fmt.Sprintf(", port = $%d", idx)
		args = append(args, *input.Port)
		idx++
	}
	if input.Status != nil {
		set += fmt.Sprintf(", status = $%d", idx)
		args = append(args, *input.Status)
		idx++
	}
	if input.Notes != nil {
		set += fmt.Sprintf(", notes = $%d", idx)
		args = append(args, *input.Notes)
		idx++
	}
	args = append(args, chi.URLParam(r, "id"))

	var ignored *string
	d, err := scan(h.db.QueryRow(r.Context(),
		fmt.Sprintf("UPDATE database_instances SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, cols), args...))
	_ = ignored
	if err != nil {
		response.NotFound(w, "database instance not found")
		return
	}
	response.OK(w, d)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	h.db.Exec(r.Context(), "UPDATE database_instances SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL", chi.URLParam(r, "id"))
	response.NoContent(w)
}

// Silence unused var
var _ = errors.New
