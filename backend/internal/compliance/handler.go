package compliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"nexusasset/backend/pkg/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ISO27001Control maps a software record to a control family.
type ISO27001Control struct {
	ControlID    string `json:"control_id"`
	ControlName  string `json:"control_name"`
	SoftwareID   string `json:"software_id"`
	SoftwareName string `json:"software_name"`
	Status       string `json:"status"`
	Criticality  string `json:"criticality"`
	Owner        string `json:"owner_team,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

// ComplianceReport is the top-level export.
type ComplianceReport struct {
	GeneratedAt string            `json:"generated_at"`
	TotalItems  int               `json:"total_items"`
	Controls    []ISO27001Control `json:"controls"`
}

// Handler is the compliance export handler.
type Handler struct{ db *pgxpool.Pool }

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

// Export handles GET /api/v1/compliance/export?format=json|csv
// Returns a JSON report mapping all software to ISO 27001 Annex A control A.8.3
// (information asset management).
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT s.id, s.name, s.status, s.criticality,
		       COALESCE(t.name, ''), COALESCE(s.notes, '')
		FROM software s
		LEFT JOIN teams t ON t.id = s.owner_team_id
		WHERE s.deleted_at IS NULL
		ORDER BY s.criticality DESC, s.name ASC`)
	if err != nil {
		response.InternalError(w, "failed to query software")
		return
	}
	defer rows.Close()

	var controls []ISO27001Control
	for rows.Next() {
		var c ISO27001Control
		// All software maps to A.8.3 — Information assets are identified and inventory is maintained.
		c.ControlID = "A.8.3"
		c.ControlName = "Information asset management"
		rows.Scan(&c.SoftwareID, &c.SoftwareName, &c.Status, &c.Criticality, &c.Owner, &c.Notes)
		controls = append(controls, c)
	}
	if controls == nil {
		controls = []ISO27001Control{}
	}

	report := ComplianceReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		TotalItems:  len(controls),
		Controls:    controls,
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="compliance_%s.csv"`, time.Now().Format("20060102")))
		fmt.Fprintln(w, "control_id,control_name,software_id,software_name,status,criticality,owner_team,notes")
		for _, c := range controls {
			fmt.Fprintf(w, "%s,%s,%s,%q,%s,%s,%q,%q\n",
				c.ControlID, c.ControlName, c.SoftwareID, c.SoftwareName,
				c.Status, c.Criticality, c.Owner, c.Notes)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="compliance_%s.json"`, time.Now().Format("20060102")))
		json.NewEncoder(w).Encode(report)
	}
}

// Summary handles GET /api/v1/compliance/summary — returns counts by criticality + status
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	rows, _ := h.db.Query(r.Context(), `
		SELECT criticality, status, COUNT(*) as cnt
		FROM software WHERE deleted_at IS NULL
		GROUP BY criticality, status ORDER BY criticality, status`)
	type row struct {
		Criticality string `json:"criticality"`
		Status      string `json:"status"`
		Count       int64  `json:"count"`
	}
	var items []row
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var rr row
			rows.Scan(&rr.Criticality, &rr.Status, &rr.Count)
			items = append(items, rr)
		}
	}
	if items == nil {
		items = []row{}
	}
	response.OK(w, items)
}
