package dashboard

import (
	"fmt"
	"net/http"

	"nexusasset/backend/pkg/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Stats is the top-level dashboard summary returned by GET /api/v1/dashboard.
type Stats struct {
	TotalSoftware      int64          `json:"total_software"`
	InhouseSoftware    int64          `json:"inhouse_software"`
	VendorSoftware     int64          `json:"vendor_software"`
	ActiveSoftware     int64          `json:"active_software"`
	DeprecatedSoftware int64          `json:"deprecated_software"`
	CriticalSoftware   int64          `json:"critical_software"`
	TotalServers       int64          `json:"total_servers"`
	ActiveServers      int64          `json:"active_servers"`
	TotalVendors       int64          `json:"total_vendors"`
	TotalDatabases     int64          `json:"total_databases"`
	TotalIntegrations  int64          `json:"total_integrations"`
	TotalPersons       int64          `json:"total_persons"`
	TotalTeams         int64          `json:"total_teams"`
	ExpiringLicenses   []ExpiryItem   `json:"expiring_licenses"`
	EolSoftware        []ExpiryItem   `json:"eol_software"`
	RecentAuditLogs    []AuditSummary `json:"recent_audit_logs"`
}

type ExpiryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	DaysLeft int    `json:"days_left"`
	ExpiryOn string `json:"expiry_on"`
}

type AuditSummary struct {
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	UserID     string `json:"user_id,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type CriticalityBreakdown struct {
	Criticality string `json:"criticality"`
	Count       int64  `json:"count"`
}

type StatusBreakdown struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type ChartData struct {
	SoftwareByCriticality []CriticalityBreakdown `json:"software_by_criticality"`
	SoftwareByStatus      []StatusBreakdown      `json:"software_by_status"`
	ServersByType         []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	} `json:"servers_by_type"`
}

type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

// Summary handles GET /api/v1/dashboard
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats := Stats{}

	// ─── Scalar counts ────────────────────────────────────────────────────────
	queries := []struct {
		target *int64
		query  string
	}{
		{&stats.TotalSoftware, "SELECT COUNT(*) FROM software WHERE deleted_at IS NULL"},
		{&stats.InhouseSoftware, "SELECT COUNT(*) FROM software WHERE deleted_at IS NULL AND software_kind = 'inhouse'"},
		{&stats.VendorSoftware, "SELECT COUNT(*) FROM software WHERE deleted_at IS NULL AND software_kind = 'vendor'"},
		{&stats.ActiveSoftware, "SELECT COUNT(*) FROM software WHERE deleted_at IS NULL AND status = 'active'"},
		{&stats.DeprecatedSoftware, "SELECT COUNT(*) FROM software WHERE deleted_at IS NULL AND status IN ('deprecated','eol')"},
		{&stats.CriticalSoftware, "SELECT COUNT(*) FROM software WHERE deleted_at IS NULL AND criticality = 'critical'"},
		{&stats.TotalServers, "SELECT COUNT(*) FROM servers WHERE deleted_at IS NULL"},
		{&stats.ActiveServers, "SELECT COUNT(*) FROM servers WHERE deleted_at IS NULL AND status = 'active'"},
		{&stats.TotalVendors, "SELECT COUNT(*) FROM vendors WHERE deleted_at IS NULL"},
		{&stats.TotalDatabases, "SELECT COUNT(*) FROM database_instances WHERE deleted_at IS NULL"},
		{&stats.TotalIntegrations, "SELECT COUNT(*) FROM integrations WHERE deleted_at IS NULL"},
		{&stats.TotalPersons, "SELECT COUNT(*) FROM persons WHERE deleted_at IS NULL"},
		{&stats.TotalTeams, "SELECT COUNT(*) FROM teams WHERE deleted_at IS NULL"},
	}
	for _, q := range queries {
		h.db.QueryRow(ctx, q.query).Scan(q.target) //nolint:errcheck
	}

	// ─── Expiring licenses (next 90 days) ─────────────────────────────────────
	rows, err := h.db.Query(ctx, `
		SELECT s.id, s.name, svd.license_expiry_date::date,
		       (svd.license_expiry_date::date - CURRENT_DATE)::int AS days_left
		FROM software s
		JOIN software_vendor_details svd ON svd.software_id = s.id
		WHERE svd.license_expiry_date IS NOT NULL
		  AND svd.license_expiry_date::date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '90 days'
		  AND s.deleted_at IS NULL
		ORDER BY svd.license_expiry_date ASC
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item ExpiryItem
			rows.Scan(&item.ID, &item.Name, &item.ExpiryOn, &item.DaysLeft)
			stats.ExpiringLicenses = append(stats.ExpiringLicenses, item)
		}
	}
	if stats.ExpiringLicenses == nil {
		stats.ExpiringLicenses = []ExpiryItem{}
	}

	// ─── EOL software (eol_date within 180 days) ───────────────────────────────
	rows2, err := h.db.Query(ctx, `
		SELECT s.id, s.name, svd.eol_date::date,
		       (svd.eol_date::date - CURRENT_DATE)::int AS days_left
		FROM software s
		JOIN software_vendor_details svd ON svd.software_id = s.id
		WHERE svd.eol_date IS NOT NULL
		  AND svd.eol_date::date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '180 days'
		  AND s.deleted_at IS NULL
		ORDER BY svd.eol_date ASC
		LIMIT 10`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var item ExpiryItem
			rows2.Scan(&item.ID, &item.Name, &item.ExpiryOn, &item.DaysLeft)
			stats.EolSoftware = append(stats.EolSoftware, item)
		}
	}
	if stats.EolSoftware == nil {
		stats.EolSoftware = []ExpiryItem{}
	}

	// ─── Recent audit logs ─────────────────────────────────────────────────────
	auditRows, _ := h.db.Query(ctx, `
		SELECT action, entity_type, entity_id, COALESCE(user_id::text,''), created_at::text
		FROM audit_logs ORDER BY created_at DESC LIMIT 5`)
	if auditRows != nil {
		defer auditRows.Close()
		for auditRows.Next() {
			var a AuditSummary
			auditRows.Scan(&a.Action, &a.EntityType, &a.EntityID, &a.UserID, &a.CreatedAt)
			stats.RecentAuditLogs = append(stats.RecentAuditLogs, a)
		}
	}
	if stats.RecentAuditLogs == nil {
		stats.RecentAuditLogs = []AuditSummary{}
	}

	response.OK(w, stats)
}

// Charts handles GET /api/v1/dashboard/charts
func (h *Handler) Charts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chart := ChartData{}

	// software by criticality
	rows, _ := h.db.Query(ctx, `SELECT criticality, COUNT(*) FROM software WHERE deleted_at IS NULL GROUP BY criticality ORDER BY criticality`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var b CriticalityBreakdown
			rows.Scan(&b.Criticality, &b.Count)
			chart.SoftwareByCriticality = append(chart.SoftwareByCriticality, b)
		}
	}
	if chart.SoftwareByCriticality == nil {
		chart.SoftwareByCriticality = []CriticalityBreakdown{}
	}

	// software by status
	rows2, _ := h.db.Query(ctx, `SELECT status, COUNT(*) FROM software WHERE deleted_at IS NULL GROUP BY status ORDER BY status`)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var b StatusBreakdown
			rows2.Scan(&b.Status, &b.Count)
			chart.SoftwareByStatus = append(chart.SoftwareByStatus, b)
		}
	}
	if chart.SoftwareByStatus == nil {
		chart.SoftwareByStatus = []StatusBreakdown{}
	}

	// servers by type
	rows3, _ := h.db.Query(ctx, `SELECT server_type, COUNT(*) FROM servers WHERE deleted_at IS NULL GROUP BY server_type`)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var b struct {
				Type  string `json:"type"`
				Count int64  `json:"count"`
			}
			rows3.Scan(&b.Type, &b.Count)
			chart.ServersByType = append(chart.ServersByType, b)
		}
	}
	if chart.ServersByType == nil {
		chart.ServersByType = []struct {
			Type  string `json:"type"`
			Count int64  `json:"count"`
		}{}
	}

	// silence unused fmt
	_ = fmt.Sprintf

	response.OK(w, chart)
}
