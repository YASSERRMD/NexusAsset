package seed

import (
	"context"
	"fmt"
	"log"

	"nexusasset/backend/internal/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run executes the seed in a single transaction.
// Idempotent: uses INSERT ... ON CONFLICT DO NOTHING.
func Run(ctx context.Context, db *pgxpool.Pool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("seed: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	log.Println("Seeding software_categories...")
	categories := []struct{ name, color string }{
		{"ERP", "#6366f1"}, {"CRM", "#0ea5e9"}, {"HRMS", "#f59e0b"},
		{"Finance", "#22c55e"}, {"Analytics", "#a855f7"}, {"Security", "#ef4444"},
		{"Infrastructure", "#64748b"}, {"Communication", "#06b6d4"}, {"DevTools", "#f97316"},
		{"AI/ML", "#ec4899"},
	}
	for _, c := range categories {
		_, err := tx.Exec(ctx,
			`INSERT INTO software_categories (name, color) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			c.name, c.color)
		if err != nil {
			return fmt.Errorf("seed categories: %w", err)
		}
	}

	log.Println("Seeding software_types...")
	types := []struct{ name, icon string }{
		{"Web App", "globe"}, {"REST API", "zap"}, {"Batch Job", "clock"},
		{"Mobile App", "smartphone"}, {"CLI Tool", "terminal"}, {"ETL", "arrow-right-left"},
		{"AI Service", "brain"}, {"SDK/Library", "library"},
	}
	for _, t := range types {
		_, err := tx.Exec(ctx,
			`INSERT INTO software_types (name, icon) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			t.name, t.icon)
		if err != nil {
			return fmt.Errorf("seed types: %w", err)
		}
	}

	log.Println("Seeding responsibility_roles...")
	roles := []struct{ name, desc string }{
		{"product_owner", "Owns the product vision and backlog"},
		{"tech_lead", "Technical decision maker for the team"},
		{"dev_lead", "Leads day-to-day development"},
		{"qa_lead", "Responsible for quality assurance"},
		{"devops_owner", "Owns CI/CD, infra, and deployments"},
		{"security_owner", "Responsible for security compliance"},
		{"business_owner", "Business stakeholder / sponsor"},
		{"support_contact", "Primary point of contact for support"},
	}
	for _, ro := range roles {
		_, err := tx.Exec(ctx,
			`INSERT INTO responsibility_roles (name, description) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			ro.name, ro.desc)
		if err != nil {
			return fmt.Errorf("seed responsibility_roles: %w", err)
		}
	}

	log.Println("Seeding environments...")
	envs := []struct {
		name, display, color string
		order                int
	}{
		{"dev", "Development", "#94a3b8", 1},
		{"qa", "QA", "#60a5fa", 2},
		{"staging", "Staging", "#a78bfa", 3},
		{"uat", "UAT", "#fb923c", 4},
		{"pre_prod", "Pre-Production", "#f87171", 5},
		{"production", "Production", "#4ade80", 6},
		{"dr", "Disaster Recovery", "#facc15", 7},
		{"sandbox", "Sandbox", "#e2e8f0", 8},
	}
	for _, e := range envs {
		_, err := tx.Exec(ctx,
			`INSERT INTO environments (name, display_name, color, order_index) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (name) DO NOTHING`,
			e.name, e.display, e.color, e.order)
		if err != nil {
			return fmt.Errorf("seed environments: %w", err)
		}
	}

	log.Println("Seeding repo_platforms...")
	platforms := []struct{ name, display, icon string }{
		{"github", "GitHub", "github"}, {"gitlab", "GitLab", "gitlab"},
		{"bitbucket", "Bitbucket", "code"}, {"azure_devops", "Azure DevOps", "cloud"},
		{"gitea", "Gitea", "git-branch"}, {"local", "Local / Other", "hard-drive"},
	}
	for _, p := range platforms {
		_, err := tx.Exec(ctx,
			`INSERT INTO repo_platforms (name, display_name, icon) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`,
			p.name, p.display, p.icon)
		if err != nil {
			return fmt.Errorf("seed repo_platforms: %w", err)
		}
	}

	log.Println("Seeding tech_categories...")
	techCats := []string{"language", "framework", "database", "infrastructure", "messaging", "ai", "security", "monitoring"}
	for _, tc := range techCats {
		_, err := tx.Exec(ctx,
			`INSERT INTO tech_categories (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, tc)
		if err != nil {
			return fmt.Errorf("seed tech_categories: %w", err)
		}
	}

	log.Println("Seeding default admin user...")
	hash, err := auth.HashPassword("Admin1234!")
	if err != nil {
		return fmt.Errorf("seed: hashing admin password: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ('admin', 'admin@nexus.local', $1, 'admin')
		ON CONFLICT (email) DO NOTHING`, hash)
	if err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("seed: commit: %w", err)
	}

	log.Println("Seed completed successfully.")
	return nil
}
