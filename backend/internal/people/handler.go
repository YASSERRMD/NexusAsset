package people

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

// ─── Persons Handler ──────────────────────────────────────────────────────────

type PersonHandler struct{ svc *PersonService }

func NewPersonHandler(svc *PersonService) *PersonHandler { return &PersonHandler{svc: svc} }

// List handles GET /api/v1/persons
func (h *PersonHandler) List(w http.ResponseWriter, r *http.Request) {
	p := ListParams{Search: r.URL.Query().Get("search"), Department: r.URL.Query().Get("department"), TeamID: r.URL.Query().Get("team_id")}
	p.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	p.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	persons, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		response.InternalError(w, "failed to list persons")
		return
	}
	response.OKList(w, persons, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

// GetByID handles GET /api/v1/persons/{id}
func (h *PersonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "person not found")
			return
		}
		response.InternalError(w, err.Error())
		return
	}
	response.OK(w, p)
}

// Create handles POST /api/v1/persons
func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreatePersonInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	p, err := h.svc.Create(r.Context(), input)
	if err != nil {
		response.InternalError(w, "failed to create person")
		return
	}
	response.Created(w, p)
}

// Update handles PUT /api/v1/persons/{id}
func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdatePersonInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	p, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "person not found")
			return
		}
		response.InternalError(w, "failed to update person")
		return
	}
	response.OK(w, p)
}

// Delete handles DELETE /api/v1/persons/{id}
func (h *PersonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		response.InternalError(w, "failed to delete person")
		return
	}
	response.NoContent(w)
}

// ─── Teams Handler ────────────────────────────────────────────────────────────

type TeamHandler struct{ svc *TeamService }

func NewTeamHandler(svc *TeamService) *TeamHandler { return &TeamHandler{svc: svc} }

// List handles GET /api/v1/teams
func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	p := ListParams{Search: r.URL.Query().Get("search")}
	p.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	p.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	teams, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		response.InternalError(w, "failed to list teams")
		return
	}
	response.OKList(w, teams, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

// GetByID handles GET /api/v1/teams/{id}
func (h *TeamHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "team not found")
			return
		}
		response.InternalError(w, err.Error())
		return
	}
	response.OK(w, t)
}

// Create handles POST /api/v1/teams
func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateTeamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	t, err := h.svc.Create(r.Context(), input)
	if err != nil {
		response.InternalError(w, "failed to create team")
		return
	}
	response.Created(w, t)
}

// Update handles PUT /api/v1/teams/{id}
func (h *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdateTeamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	t, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "team not found")
			return
		}
		response.InternalError(w, "failed to update team")
		return
	}
	response.OK(w, t)
}

// Delete handles DELETE /api/v1/teams/{id}
func (h *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		response.InternalError(w, "failed to delete team")
		return
	}
	response.NoContent(w)
}
