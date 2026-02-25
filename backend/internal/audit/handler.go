package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	authpkg "nexusasset/backend/internal/auth"
	"nexusasset/backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Log is a single audit entry.
type Log struct {
	ID         string                     `json:"id"`
	UserID     *string                    `json:"user_id,omitempty"`
	EntityType string                     `json:"entity_type"`
	EntityID   string                     `json:"entity_id"`
	Action     string                     `json:"action"`
	Changes    map[string]json.RawMessage `json:"changes,omitempty"`
	IPAddress  *string                    `json:"ip_address,omitempty"`
	UserAgent  *string                    `json:"user_agent,omitempty"`
	CreatedAt  time.Time                  `json:"created_at"`
}

// Entry is a small struct used when recording from middleware.
type Entry struct {
	UserID     *string
	EntityType string
	EntityID   string
	Action     string
	Changes    map[string]interface{}
	IPAddress  string
	UserAgent  string
}

// Repository is the persistence layer for audit logs.
type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// Record inserts a new audit entry (fire-and-forget safe — errors are swallowed).
func (r *Repository) Record(ctx context.Context, e Entry) {
	var changes []byte
	if e.Changes != nil {
		changes, _ = json.Marshal(e.Changes)
	}
	r.db.Exec(ctx, //nolint:errcheck
		`INSERT INTO audit_logs (user_id, entity_type, entity_id, action, changes, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''))`,
		e.UserID, e.EntityType, e.EntityID, e.Action, changes, e.IPAddress, e.UserAgent)
}

// Handler serves GET /api/v1/audit.
type Handler struct {
	db   *pgxpool.Pool
	repo *Repository
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db, repo: NewRepository(db)}
}

// Repo exposes the repo so middleware can use it.
func (h *Handler) Repo() *Repository { return h.repo }

// List handles GET /api/v1/audit
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	entityType := q.Get("entity_type")
	entityID := q.Get("entity_id")
	action := q.Get("action")
	userID := q.Get("user_id")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	where := "WHERE 1=1"
	args := []any{}
	idx := 1
	if entityType != "" {
		where += fmt.Sprintf(" AND entity_type = $%d", idx)
		args = append(args, entityType)
		idx++
	}
	if entityID != "" {
		where += fmt.Sprintf(" AND entity_id = $%d", idx)
		args = append(args, entityID)
		idx++
	}
	if action != "" {
		where += fmt.Sprintf(" AND action = $%d", idx)
		args = append(args, action)
		idx++
	}
	if userID != "" {
		where += fmt.Sprintf(" AND user_id::text = $%d", idx)
		args = append(args, userID)
		idx++
	}

	var total int64
	h.db.QueryRow(r.Context(), "SELECT COUNT(*) FROM audit_logs "+where, args...).Scan(&total)

	args2 := append(args, limit, offset)
	rows, err := h.db.Query(r.Context(),
		fmt.Sprintf("SELECT id, user_id, entity_type, entity_id, action, changes, ip_address, user_agent, created_at FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
			where, idx, idx+1), args2...)
	if err != nil {
		response.InternalError(w, "failed to list audit logs")
		return
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		var changesRaw []byte
		if err := rows.Scan(&l.ID, &l.UserID, &l.EntityType, &l.EntityID, &l.Action,
			&changesRaw, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			continue
		}
		if changesRaw != nil {
			json.Unmarshal(changesRaw, &l.Changes)
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []Log{}
	}
	response.OKList(w, logs, response.Meta{Page: page, Limit: limit, Total: total})
}

// GetByEntity handles GET /api/v1/audit/{entityType}/{entityID}
func (h *Handler) GetByEntity(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityID")
	rows, _ := h.db.Query(r.Context(),
		`SELECT id, user_id, entity_type, entity_id, action, changes, ip_address, user_agent, created_at
		 FROM audit_logs WHERE entity_type = $1 AND entity_id = $2 ORDER BY created_at DESC LIMIT 100`,
		entityType, entityID)
	var logs []Log
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var l Log
			var changesRaw []byte
			rows.Scan(&l.ID, &l.UserID, &l.EntityType, &l.EntityID, &l.Action,
				&changesRaw, &l.IPAddress, &l.UserAgent, &l.CreatedAt)
			if changesRaw != nil {
				json.Unmarshal(changesRaw, &l.Changes)
			}
			logs = append(logs, l)
		}
	}
	if logs == nil {
		logs = []Log{}
	}
	response.OK(w, logs)
}

// ─── Middleware helper ─────────────────────────────────────────────────────────

// AutoLog is a middleware that automatically logs all mutation requests (POST, PUT, PATCH, DELETE).
func AutoLog(repo *Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			// Only log mutations
			m := r.Method
			if m != http.MethodPost && m != http.MethodPut && m != http.MethodPatch && m != http.MethodDelete {
				return
			}

			// Run in background to not block response
			go func() {
				var userID *string
				if claims := authpkg.ClaimsFromContext(r.Context()); claims != nil {
					id := claims.UserID
					userID = &id
				}

				// Basic heuristics for entity type/id from path
				// e.g. /api/v1/software/123 -> entityType="software", entityID="123"
				entityType := r.URL.Path
				entityID := chi.URLParam(r, "id")
				if entityID == "" {
					entityID = "unknown_or_new"
				}

				repo.Record(context.Background(), Entry{
					UserID:     userID,
					EntityType: entityType,
					EntityID:   entityID,
					Action:     m,
					Changes:    nil, // Full body diffing is complex, just log the action for now
					IPAddress:  r.RemoteAddr,
					UserAgent:  r.UserAgent(),
				})
			}()
		})
	}
}
