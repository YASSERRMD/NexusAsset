package vendor

import (
	"fmt"
	"time"

	"encoding/json"
	"errors"
	"net/http"
	"nexusasset/backend/pkg/response"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var ErrNotFound = errors.New("vendor not found")
var validate = validator.New()

// Vendor domain model.
type Vendor struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	DisplayName  string     `json:"display_name"`
	Website      *string    `json:"website,omitempty"`
	ContactEmail *string    `json:"contact_email,omitempty"`
	ContactPhone *string    `json:"contact_phone,omitempty"`
	Address      *string    `json:"address,omitempty"`
	Region       *string    `json:"region,omitempty"`
	IsActive     bool       `json:"is_active"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type ListParams struct {
	Page, Limit    int
	Search, Region string
}

type CreateInput struct {
	Name         string  `json:"name"         validate:"required"`
	DisplayName  string  `json:"display_name" validate:"required"`
	Website      *string `json:"website,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	Address      *string `json:"address,omitempty"`
	Region       *string `json:"region,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

type UpdateInput struct {
	Name         *string `json:"name,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	Website      *string `json:"website,omitempty"`
	ContactEmail *string `json:"contact_email,omitempty" validate:"omitempty,email"`
	ContactPhone *string `json:"contact_phone,omitempty"`
	Address      *string `json:"address,omitempty"`
	Region       *string `json:"region,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

const vendorCols = `id, name, display_name, website, contact_email, contact_phone, address, region, is_active, notes, created_at, updated_at, deleted_at`

func scanVendor(row interface{ Scan(...any) error }) (*Vendor, error) {
	v := &Vendor{}
	return v, row.Scan(&v.ID, &v.Name, &v.DisplayName, &v.Website, &v.ContactEmail, &v.ContactPhone,
		&v.Address, &v.Region, &v.IsActive, &v.Notes, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt)
}

// Repository + Service collapsed for brevity into handler wrapping db directly.

type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
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
		where += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d)", idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	var total int64
	h.db.QueryRow(r.Context(), "SELECT COUNT(*) FROM vendors "+where, args...).Scan(&total)
	args = append(args, limit, offset)
	rows, _ := h.db.Query(r.Context(), fmt.Sprintf("SELECT %s FROM vendors %s ORDER BY name ASC LIMIT $%d OFFSET $%d", vendorCols, where, idx, idx+1), args...)
	defer rows.Close()
	var vendors []Vendor
	for rows.Next() {
		v, _ := scanVendor(rows)
		vendors = append(vendors, *v)
	}
	response.OKList(w, vendors, response.Meta{Page: page, Limit: limit, Total: total})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	v, err := scanVendor(h.db.QueryRow(r.Context(), fmt.Sprintf("SELECT %s FROM vendors WHERE id = $1 AND deleted_at IS NULL", vendorCols), chi.URLParam(r, "id")))
	if err != nil {
		response.NotFound(w, "vendor not found")
		return
	}
	response.OK(w, v)
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
	v, err := scanVendor(h.db.QueryRow(r.Context(),
		fmt.Sprintf("INSERT INTO vendors (name, display_name, website, contact_email, contact_phone, address, region, notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING %s", vendorCols),
		input.Name, input.DisplayName, input.Website, input.ContactEmail, input.ContactPhone, input.Address, input.Region, input.Notes))
	if err != nil {
		response.InternalError(w, "failed to create vendor")
		return
	}
	response.Created(w, v)
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
	if input.DisplayName != nil {
		set += fmt.Sprintf(", display_name = $%d", idx)
		args = append(args, *input.DisplayName)
		idx++
	}
	if input.Website != nil {
		set += fmt.Sprintf(", website = $%d", idx)
		args = append(args, *input.Website)
		idx++
	}
	if input.ContactEmail != nil {
		set += fmt.Sprintf(", contact_email = $%d", idx)
		args = append(args, *input.ContactEmail)
		idx++
	}
	if input.ContactPhone != nil {
		set += fmt.Sprintf(", contact_phone = $%d", idx)
		args = append(args, *input.ContactPhone)
		idx++
	}
	if input.Address != nil {
		set += fmt.Sprintf(", address = $%d", idx)
		args = append(args, *input.Address)
		idx++
	}
	if input.Region != nil {
		set += fmt.Sprintf(", region = $%d", idx)
		args = append(args, *input.Region)
		idx++
	}
	if input.IsActive != nil {
		set += fmt.Sprintf(", is_active = $%d", idx)
		args = append(args, *input.IsActive)
		idx++
	}
	if input.Notes != nil {
		set += fmt.Sprintf(", notes = $%d", idx)
		args = append(args, *input.Notes)
		idx++
	}
	args = append(args, chi.URLParam(r, "id"))
	v, err := scanVendor(h.db.QueryRow(r.Context(),
		fmt.Sprintf("UPDATE vendors SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s", set, idx, vendorCols), args...))
	if err != nil {
		response.NotFound(w, "vendor not found")
		return
	}
	response.OK(w, v)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	h.db.Exec(r.Context(), "UPDATE vendors SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL", chi.URLParam(r, "id"))
	response.NoContent(w)
}
