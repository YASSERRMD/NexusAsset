package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"nexusasset/backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler serves all /api/v1/users routes (admin only).
type Handler struct {
	svc *Service
}

// NewHandler constructs a users handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /api/v1/users
//
// Query params: page, limit, search, role
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := ListParams{
		Search: r.URL.Query().Get("search"),
		Role:   r.URL.Query().Get("role"),
	}
	p.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	p.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}

	users, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		response.InternalError(w, "failed to list users")
		return
	}
	response.OKList(w, users, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

// GetByID handles GET /api/v1/users/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "user not found")
			return
		}
		response.InternalError(w, err.Error())
		return
	}
	response.OK(w, u)
}

// Create handles POST /api/v1/users
//
// Request body:
//
//	{ "username": "jdoe", "email": "j@doe.com", "password": "s3cr3t!!", "role": "contributor" }
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	u, err := h.svc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalError(w, "failed to create user")
		return
	}
	response.Created(w, u)
}

// Update handles PUT /api/v1/users/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	u, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "user not found")
			return
		}
		response.InternalError(w, "failed to update user")
		return
	}
	response.OK(w, u)
}

// Delete handles DELETE /api/v1/users/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.InternalError(w, "failed to delete user")
		return
	}
	response.NoContent(w)
}

// UpdateRole handles PATCH /api/v1/users/{id}/role
//
// Request body:
//
//	{ "role": "admin" }
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Role string `json:"role" validate:"required,oneof=admin contributor reader"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(body); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	u, err := h.svc.UpdateRole(r.Context(), id, body.Role)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalError(w, "failed to update role")
		return
	}
	response.OK(w, u)
}
