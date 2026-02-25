package integrations

import (
	"encoding/json"
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

type Integration struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	SourceSoftwareID *string    `json:"source_software_id,omitempty"`
	TargetSoftwareID *string    `json:"target_software_id,omitempty"`
	IntegrationKind  string     `json:"integration_kind"`
	Protocol         *string    `json:"protocol,omitempty"`
	DataFormat       *string    `json:"data_format,omitempty"`
	Frequency        *string    `json:"frequency,omitempty"`
	Status           string     `json:"status"`
	OwnerTeamID      *string    `json:"owner_team_id,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateInput struct {
	Name             string  `json:"name"             validate:"required"`
	Description      *string `json:"description,omitempty"`
	SourceSoftwareID *string `json:"source_software_id,omitempty"`
	TargetSoftwareID *string `json:"target_software_id,omitempty"`
	IntegrationKind  string  `json:"integration_kind" validate:"required"`
	Protocol         *string `json:"protocol,omitempty"`
	DataFormat       *string `json:"data_format,omitempty"`
	Frequency        *string `json:"frequency,omitempty"`
	Status           string  `json:"status"`
	OwnerTeamID      *string `json:"owner_team_id,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

type UpdateInput struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	IntegrationKind *string `json:"integration_kind,omitempty"`
	Protocol        *string `json:"protocol,omitempty"`
	DataFormat      *string `json:"data_format,omitempty"`
	Frequency       *string `json:"frequency,omitempty"`
	Status          *string `json:"status,omitempty"`
	OwnerTeamID     *string `json:"owner_team_id,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

const cols = `id, name, description, source_software_id, target_software_id, integration_kind,
    protocol, data_format, frequency, status, owner_team_id, notes, created_at, updated_at, deleted_at`

func scan(row interface{ Scan(...any) error }) (*Integration, error) {
	v := &Integration{}
	return v, row.Scan(&v.ID, &v.Name, &v.Description, &v.SourceSoftwareID, &v.TargetSoftwareID,
		&v.IntegrationKind, &v.Protocol, &v.DataFormat, &v.Frequency, &v.Status, &v.OwnerTeamID,
		&v.Notes, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt)
}

type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	kind := r.URL.Query().Get("kind")
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
	if kind != "" {
		where += fmt.Sprintf(" AND integration_kind = $%d", idx)
		args = append(args, kind)
		idx++
	}
	var total int64
	h.db.QueryRow(r.Context(), "SELECT COUNT(*) FROM integrations "+where, args...).Scan(&total)
	args2 := append(args, limit, offset)
	rows, _ := h.db.Query(r.Context(), fmt.Sprintf("SELECT %s FROM integrations %s ORDER BY name ASC LIMIT $%d OFFSET $%d",
		cols, where, idx, idx+1), args2...)
	defer rows.Close()
	var items []Integration
	for rows.Next() {
		d, _ := scan(rows)
		items = append(items, *d)
	}
	response.OKList(w, items, response.Meta{Page: page, Limit: limit, Total: total})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	d, err := scan(h.db.QueryRow(r.Context(), fmt.Sprintf("SELECT %s FROM integrations WHERE id = $1 AND deleted_at IS NULL", cols), chi.URLParam(r, "id")))
	if err != nil {
		response.NotFound(w, "integration not found")
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
		fmt.Sprintf(`INSERT INTO integrations (name, description, source_software_id, target_software_id, integration_kind, protocol, data_format, frequency, status, owner_team_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING %s`, cols),
		input.Name, input.Description, input.SourceSoftwareID, input.TargetSoftwareID,
		input.IntegrationKind, input.Protocol, input.DataFormat, input.Frequency,
		input.Status, input.OwnerTeamID, input.Notes))
	if err != nil {
		response.InternalError(w, "failed to create integration")
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
	if input.Description != nil {
		set += fmt.Sprintf(", description = $%d", idx)
		args = append(args, *input.Description)
		idx++
	}
	if input.IntegrationKind != nil {
		set += fmt.Sprintf(", integration_kind = $%d", idx)
		args = append(args, *input.IntegrationKind)
		idx++
	}
	if input.Protocol != nil {
		set += fmt.Sprintf(", protocol = $%d", idx)
		args = append(args, *input.Protocol)
		idx++
	}
	if input.DataFormat != nil {
		set += fmt.Sprintf(", data_format = $%d", idx)
		args = append(args, *input.DataFormat)
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
	d, err := scan(h.db.QueryRow(r.Context(),
		fmt.Sprintf("UPDATE integrations SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, cols), args...))
	if err != nil {
		response.NotFound(w, "integration not found")
		return
	}
	response.OK(w, d)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	h.db.Exec(r.Context(), "UPDATE integrations SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL", chi.URLParam(r, "id"))
	response.NoContent(w)
}
