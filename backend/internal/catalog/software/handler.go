package software

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

// Handler serves all /api/v1/software routes.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ─── Core software CRUD ───────────────────────────────────────────────────────

// List handles GET /api/v1/software
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p := ListParams{
		Search:      q.Get("search"),
		Kind:        q.Get("kind"),
		CategoryID:  q.Get("category"),
		TypeID:      q.Get("type"),
		Status:      q.Get("status"),
		Criticality: q.Get("criticality"),
	}
	p.Page, _ = strconv.Atoi(q.Get("page"))
	p.Limit, _ = strconv.Atoi(q.Get("limit"))
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}

	items, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		response.InternalError(w, "failed to list software")
		return
	}
	response.OKList(w, items, response.Meta{Page: p.Page, Limit: p.Limit, Total: total})
}

// GetByID handles GET /api/v1/software/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	sw, err := h.svc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "software not found")
			return
		}
		response.InternalError(w, err.Error())
		return
	}
	response.OK(w, sw)
}

// Create handles POST /api/v1/software
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateSoftwareInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	if claims := authpkg.ClaimsFromContext(r.Context()); claims != nil {
		input.CreatedBy = &claims.UserID
	}
	sw, err := h.svc.Create(r.Context(), input)
	if err != nil {
		response.InternalError(w, "failed to create software")
		return
	}
	response.Created(w, sw)
}

// Update handles PUT /api/v1/software/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdateSoftwareInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	if claims := authpkg.ClaimsFromContext(r.Context()); claims != nil {
		input.UpdatedBy = &claims.UserID
	}
	sw, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "software not found")
			return
		}
		response.InternalError(w, "failed to update software")
		return
	}
	response.OK(w, sw)
}

// Delete handles DELETE /api/v1/software/{id}  (admin only)
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		response.InternalError(w, "failed to delete software")
		return
	}
	response.NoContent(w)
}

// ─── Responsibilities ─────────────────────────────────────────────────────────

func (h *Handler) ListResponsibilities(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListResponsibilities(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.InternalError(w, "failed to list responsibilities")
		return
	}
	response.OK(w, items)
}

func (h *Handler) AddResponsibility(w http.ResponseWriter, r *http.Request) {
	var input AddResponsibilityInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	item, err := h.svc.AddResponsibility(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.InternalError(w, "failed to add responsibility")
		return
	}
	response.Created(w, item)
}

func (h *Handler) DeleteResponsibility(w http.ResponseWriter, r *http.Request) {
	err := h.svc.DeleteResponsibility(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "rid"))
	if err != nil {
		response.InternalError(w, "failed to delete responsibility")
		return
	}
	response.NoContent(w)
}

// ─── Repositories ─────────────────────────────────────────────────────────────

func (h *Handler) ListRepos(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListRepos(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.InternalError(w, "failed to list repositories")
		return
	}
	response.OK(w, items)
}

func (h *Handler) AddRepo(w http.ResponseWriter, r *http.Request) {
	var input AddRepoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	item, err := h.svc.AddRepo(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.InternalError(w, "failed to add repository")
		return
	}
	response.Created(w, item)
}

func (h *Handler) DeleteRepo(w http.ResponseWriter, r *http.Request) {
	err := h.svc.DeleteRepo(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "rid"))
	if err != nil {
		response.InternalError(w, "failed to delete repository")
		return
	}
	response.NoContent(w)
}

// ─── Deployments ──────────────────────────────────────────────────────────────

func (h *Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListDeployments(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.InternalError(w, "failed to list deployments")
		return
	}
	response.OK(w, items)
}

func (h *Handler) AddDeployment(w http.ResponseWriter, r *http.Request) {
	var input AddDeploymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	item, err := h.svc.AddDeployment(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.InternalError(w, "failed to add deployment")
		return
	}
	response.Created(w, item)
}

func (h *Handler) UpdateDeploymentHealth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Health string `json:"health"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	err := h.svc.UpdateDeploymentHealth(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "did"), body.Health)
	if err != nil {
		response.InternalError(w, "failed to update health")
		return
	}
	response.OK(w, map[string]string{"health_status": body.Health})
}

func (h *Handler) DeleteDeployment(w http.ResponseWriter, r *http.Request) {
	err := h.svc.DeleteDeployment(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "did"))
	if err != nil {
		response.InternalError(w, "failed to delete deployment")
		return
	}
	response.NoContent(w)
}

// ─── Tech Stack ───────────────────────────────────────────────────────────────

func (h *Handler) ListTechStack(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListTechStack(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.InternalError(w, "failed to list tech stack")
		return
	}
	response.OK(w, items)
}

func (h *Handler) AddTechStack(w http.ResponseWriter, r *http.Request) {
	var input AddTechStackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	item, err := h.svc.AddTechStack(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.InternalError(w, "failed to add tech stack")
		return
	}
	response.Created(w, item)
}

func (h *Handler) DeleteTechStack(w http.ResponseWriter, r *http.Request) {
	err := h.svc.DeleteTechStack(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "tid"))
	if err != nil {
		response.InternalError(w, "failed to delete tech stack")
		return
	}
	response.NoContent(w)
}

// ─── Database Links ───────────────────────────────────────────────────────────

func (h *Handler) ListDatabaseLinks(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListDatabaseLinks(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		response.InternalError(w, "failed to list database links")
		return
	}
	response.OK(w, items)
}

func (h *Handler) AddDatabaseLink(w http.ResponseWriter, r *http.Request) {
	var input AddDatabaseLinkInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := validate.Struct(input); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	item, err := h.svc.AddDatabaseLink(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		response.InternalError(w, "failed to add database link")
		return
	}
	response.Created(w, item)
}

func (h *Handler) DeleteDatabaseLink(w http.ResponseWriter, r *http.Request) {
	err := h.svc.DeleteDatabaseLink(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "lid"))
	if err != nil {
		response.InternalError(w, "failed to delete database link")
		return
	}
	response.NoContent(w)
}
