package lookup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"nexusasset/backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Generic lookup row used by all 6 extensible lookup tables.
type Row struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Color       *string    `json:"color,omitempty"`
	Icon        *string    `json:"icon,omitempty"`
	DisplayName *string    `json:"display_name,omitempty"`
	OrderIndex  *int       `json:"order_index,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type UpsertInput struct {
	Name        string  `json:"name"        validate:"required"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	OrderIndex  *int    `json:"order_index,omitempty"`
}

type PatchActiveInput struct {
	IsActive bool `json:"is_active"`
}

// table metadata for each lookup endpoint
type tableSpec struct {
	table   string
	hasCols []string // extra columns beyond id,name,is_active
}

var tables = map[string]tableSpec{
	"software-categories":  {table: "software_categories", hasCols: []string{"description", "color", "created_at", "updated_at"}},
	"software-types":       {table: "software_types", hasCols: []string{"description", "icon", "created_at", "updated_at"}},
	"environments":         {table: "environments", hasCols: []string{"display_name", "color", "order_index", "created_at"}},
	"responsibility-roles": {table: "responsibility_roles", hasCols: []string{"description"}},
	"repo-platforms":       {table: "repo_platforms", hasCols: []string{"display_name", "icon"}},
	"tech-categories":      {table: "tech_categories", hasCols: []string{}},
}

// Handler serves all /api/v1/lookups/* routes.
type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

// List handles GET /api/v1/lookups/{table}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "table")
	spec, ok := tables[key]
	if !ok {
		response.NotFound(w, "unknown lookup table: "+key)
		return
	}

	rows, err := h.db.Query(r.Context(), fmt.Sprintf("SELECT id, name, is_active FROM %s ORDER BY name ASC", spec.table))
	if err != nil {
		response.InternalError(w, "failed to list lookup")
		return
	}
	defer rows.Close()

	var items []Row
	for rows.Next() {
		var row Row
		if err := rows.Scan(&row.ID, &row.Name, &row.IsActive); err != nil {
			response.InternalError(w, "scan error")
			return
		}
		items = append(items, row)
	}
	response.OK(w, items)
}

// Create handles POST /api/v1/lookups/{table}  (admin only)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "table")
	spec, ok := tables[key]
	if !ok {
		response.NotFound(w, "unknown lookup table: "+key)
		return
	}

	var input UpsertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if input.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	var id string
	err := h.db.QueryRow(r.Context(),
		fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id", spec.table),
		input.Name).Scan(&id)
	if err != nil {
		response.InternalError(w, fmt.Sprintf("failed to create %s", key))
		return
	}
	response.Created(w, map[string]string{"id": id, "name": input.Name})
}

// PatchActive handles PATCH /api/v1/lookups/{table}/{id}/active  (admin only)
func (h *Handler) PatchActive(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "table")
	id := chi.URLParam(r, "id")
	spec, ok := tables[key]
	if !ok {
		response.NotFound(w, "unknown lookup table: "+key)
		return
	}

	// tech_categories has no is_active column
	if key == "tech-categories" {
		response.BadRequest(w, "tech-categories does not support is_active")
		return
	}

	var input PatchActiveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}

	ct, err := h.db.Exec(r.Context(),
		fmt.Sprintf("UPDATE %s SET is_active = $1, updated_at = now() WHERE id = $2", spec.table),
		input.IsActive, id)
	if err != nil {
		response.InternalError(w, "failed to update")
		return
	}
	if ct.RowsAffected() == 0 {
		response.NotFound(w, "lookup entry not found")
		return
	}
	response.OK(w, map[string]bool{"is_active": input.IsActive})
}

// Update handles PUT /api/v1/lookups/{table}/{id}  (admin only)
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "table")
	id := chi.URLParam(r, "id")
	spec, ok := tables[key]
	if !ok {
		response.NotFound(w, "unknown lookup table: "+key)
		return
	}

	var input UpsertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if input.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	ct, err := h.db.Exec(r.Context(), fmt.Sprintf("UPDATE %s SET name = $1 WHERE id = $2", spec.table), input.Name, id)
	if err != nil {
		response.InternalError(w, "failed to update")
		return
	}
	if ct.RowsAffected() == 0 {
		response.NotFound(w, "entry not found")
		return
	}
	response.OK(w, map[string]string{"id": id, "name": input.Name})
}
