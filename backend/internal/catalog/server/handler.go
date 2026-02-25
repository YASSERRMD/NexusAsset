package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	authpkg "nexusasset/backend/internal/auth"
	"nexusasset/backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler serves all /api/v1/servers routes.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List handles GET /api/v1/servers
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := ListParams{
		Search:        r.URL.Query().Get("search"),
		ServerType:    r.URL.Query().Get("server_type"),
		Status:        r.URL.Query().Get("status"),
		CloudProvider: r.URL.Query().Get("cloud_provider"),
		EnvironmentID: r.URL.Query().Get("environment_id"),
	}
	p.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	p.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	servers, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		response.InternalError(w, "failed to list servers")
		return
	}
	response.OKList(w, servers, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

// GetByID handles GET /api/v1/servers/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "server not found")
			return
		}
		response.InternalError(w, err.Error())
		return
	}
	response.OK(w, s)
}

// Create handles POST /api/v1/servers
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateServerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	claims := authpkg.ClaimsFromContext(r.Context())
	if claims != nil {
		input.CreatedBy = &claims.UserID
	}
	s, err := h.svc.Create(r.Context(), input)
	if err != nil {
		response.InternalError(w, "failed to create server")
		return
	}
	response.Created(w, s)
}

// Update handles PUT /api/v1/servers/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdateServerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	claims := authpkg.ClaimsFromContext(r.Context())
	if claims != nil {
		input.UpdatedBy = &claims.UserID
	}
	s, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "server not found")
			return
		}
		response.InternalError(w, "failed to update server")
		return
	}
	response.OK(w, s)
}

// Delete handles DELETE /api/v1/servers/{id}  (admin only)
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		response.InternalError(w, "failed to delete server")
		return
	}
	response.NoContent(w)
}
