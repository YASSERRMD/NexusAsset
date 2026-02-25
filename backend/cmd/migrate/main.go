package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"nexusasset/backend/db/seed"
	"nexusasset/backend/internal/config"
	dbpkg "nexusasset/backend/internal/database"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Parse direction from args: "up" or "down"
	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	migrationsPath := "file://db/migrations"
	if len(os.Args) > 2 {
		migrationsPath = fmt.Sprintf("file://%s", os.Args[2])
	}

	m, err := migrate.New(migrationsPath, cfg.Database.URL)
	if err != nil {
		log.Fatalf("migrate.New: %v", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("Migrations applied successfully.")

		// Run seed after migrations
		if len(os.Args) > 3 && os.Args[3] == "--seed" {
			pool, err := dbpkg.NewPool(context.Background(), dbpkg.Config{URL: cfg.Database.URL})
			if err != nil {
				log.Fatalf("db pool for seed: %v", err)
			}
			defer pool.Close()
			if err := seed.Run(context.Background(), pool); err != nil {
				log.Fatalf("seed: %v", err)
			}
		}

	case "down":
		if err := m.Steps(-1); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("Rolled back 1 migration.")

	case "drop":
		if err := m.Drop(); err != nil {
			log.Fatalf("migrate drop: %v", err)
		}
		log.Println("All migrations dropped.")

	default:
		log.Fatalf("unknown direction: %s (use: up, down, drop)", direction)
	}
}
