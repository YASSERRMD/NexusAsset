package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nexusasset/backend/db/seed"
	authpkg "nexusasset/backend/internal/auth"
	catdb "nexusasset/backend/internal/catalog/dbinstances"
	catintegrations "nexusasset/backend/internal/catalog/integrations"
	catserver "nexusasset/backend/internal/catalog/server"
	catsoftware "nexusasset/backend/internal/catalog/software"
	catvendor "nexusasset/backend/internal/catalog/vendor"
	"nexusasset/backend/internal/config"
	dbpkg "nexusasset/backend/internal/database"
	"nexusasset/backend/internal/logger"
	"nexusasset/backend/internal/lookup"
	"nexusasset/backend/internal/people"
	"nexusasset/backend/internal/users"
	appmiddleware "nexusasset/backend/pkg/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func main() {
	// 1. Load config
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize logger
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// 3. Connect to database
	ctx := context.Background()
	pool, err := dbpkg.NewPool(ctx, dbpkg.Config{
		URL:          cfg.Database.URL,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})
	if err != nil {
		log.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()
	log.Info("database connected")

	// 4. Auto-seed on first run (only if --seed flag provided)
	if len(os.Args) > 1 && os.Args[1] == "--seed" {
		if err := seed.Run(ctx, pool); err != nil {
			log.Fatal("seed failed", zap.Error(err))
		}
	}

	// 5. Build services
	jwtSvc := authpkg.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	authRepo := authpkg.NewRepository(pool)
	authSvc := authpkg.NewService(authRepo, jwtSvc, cfg.JWT.AccessTokenExpiry)
	authHandler := authpkg.NewHandler(authSvc, jwtSvc)

	usersRepo := users.NewRepository(pool)
	usersSvc := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersSvc)

	// People
	personRepo := people.NewPersonRepository(pool)
	personSvc := people.NewPersonService(personRepo)
	personHandler := people.NewPersonHandler(personSvc)

	teamRepo := people.NewTeamRepository(pool)
	teamSvc := people.NewTeamService(teamRepo)
	teamHandler := people.NewTeamHandler(teamSvc)

	// Server catalog
	serverRepo := catserver.NewRepository(pool)
	serverSvc := catserver.NewService(serverRepo)
	serverHandler := catserver.NewHandler(serverSvc)

	// Software catalog — Phase 3
	softwareRepo := catsoftware.NewRepository(pool)
	softwareSvc := catsoftware.NewService(softwareRepo)
	softwareHandler := catsoftware.NewHandler(softwareSvc)

	// Vendor catalog — Phase 3
	vendorHandler := catvendor.NewHandler(pool)

	// Database instances — Phase 3
	dbHandler := catdb.NewHandler(pool)

	// Integrations — Phase 3
	integrationsHandler := catintegrations.NewHandler(pool)

	// Lookups
	lookupHandler := lookup.NewHandler(pool)

	// 6. Build chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.RequestLogger(log))
	r.Use(appmiddleware.Recoverer(log))
	r.Use(appmiddleware.Cors([]string{"http://localhost:3000", "http://127.0.0.1:3000"}))
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// --- Public routes (no auth) ---
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		// --- Authenticated routes ---
		r.Group(func(r chi.Router) {
			r.Use(jwtSvc.Authenticate)

			// Auth: me
			r.Get("/auth/me", authHandler.Me)

			// Users — admin only
			r.Route("/users", func(r chi.Router) {
				r.Use(jwtSvc.RequireRole("admin"))
				r.Get("/", usersHandler.List)
				r.Post("/", usersHandler.Create)
				r.Get("/{id}", usersHandler.GetByID)
				r.Put("/{id}", usersHandler.Update)
				r.Delete("/{id}", usersHandler.Delete)
				r.Patch("/{id}/role", usersHandler.UpdateRole)
			})

			// People — admin + contributor
			r.Route("/persons", func(r chi.Router) {
				r.Use(jwtSvc.RequireRole("admin", "contributor"))
				r.Get("/", personHandler.List)
				r.Post("/", personHandler.Create)
				r.Get("/{id}", personHandler.GetByID)
				r.Put("/{id}", personHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", personHandler.Delete)
			})

			r.Route("/teams", func(r chi.Router) {
				r.Use(jwtSvc.RequireRole("admin", "contributor"))
				r.Get("/", teamHandler.List)
				r.Post("/", teamHandler.Create)
				r.Get("/{id}", teamHandler.GetByID)
				r.Put("/{id}", teamHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", teamHandler.Delete)
			})

			// Servers — read: all; write: admin+contributor; delete: admin
			r.Route("/servers", func(r chi.Router) {
				r.Get("/", serverHandler.List)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/", serverHandler.Create)
				r.Get("/{id}", serverHandler.GetByID)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Put("/{id}", serverHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", serverHandler.Delete)
			})

			// Software catalog — read: all; write: admin+contributor; delete: admin
			r.Route("/software", func(r chi.Router) {
				r.Get("/", softwareHandler.List)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/", softwareHandler.Create)
				r.Get("/{id}", softwareHandler.GetByID)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Put("/{id}", softwareHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", softwareHandler.Delete)
				// Sub-resources
				r.Get("/{id}/responsibilities", softwareHandler.ListResponsibilities)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/{id}/responsibilities", softwareHandler.AddResponsibility)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}/responsibilities/{rid}", softwareHandler.DeleteResponsibility)
				r.Get("/{id}/repositories", softwareHandler.ListRepos)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/{id}/repositories", softwareHandler.AddRepo)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}/repositories/{rid}", softwareHandler.DeleteRepo)
				r.Get("/{id}/deployments", softwareHandler.ListDeployments)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/{id}/deployments", softwareHandler.AddDeployment)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Patch("/{id}/deployments/{did}/health", softwareHandler.UpdateDeploymentHealth)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}/deployments/{did}", softwareHandler.DeleteDeployment)
				r.Get("/{id}/tech-stack", softwareHandler.ListTechStack)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/{id}/tech-stack", softwareHandler.AddTechStack)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}/tech-stack/{tid}", softwareHandler.DeleteTechStack)
				r.Get("/{id}/database-links", softwareHandler.ListDatabaseLinks)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/{id}/database-links", softwareHandler.AddDatabaseLink)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}/database-links/{lid}", softwareHandler.DeleteDatabaseLink)
			})

			// Vendors — read: all; write: admin+contributor; delete: admin
			r.Route("/vendors", func(r chi.Router) {
				r.Get("/", vendorHandler.List)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/", vendorHandler.Create)
				r.Get("/{id}", vendorHandler.GetByID)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Put("/{id}", vendorHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", vendorHandler.Delete)
			})

			// Database instances — read: all; write: admin+contributor; delete: admin
			r.Route("/databases", func(r chi.Router) {
				r.Get("/", dbHandler.List)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/", dbHandler.Create)
				r.Get("/{id}", dbHandler.GetByID)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Put("/{id}", dbHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", dbHandler.Delete)
			})

			// Integrations — read: all; write: admin+contributor; delete: admin
			r.Route("/integrations", func(r chi.Router) {
				r.Get("/", integrationsHandler.List)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Post("/", integrationsHandler.Create)
				r.Get("/{id}", integrationsHandler.GetByID)
				r.With(jwtSvc.RequireRole("admin", "contributor")).Put("/{id}", integrationsHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Delete("/{id}", integrationsHandler.Delete)
			})

			// Lookup tables — GET: all; POST/PUT/PATCH: admin only
			r.Route("/lookups/{table}", func(r chi.Router) {
				r.Get("/", lookupHandler.List)
				r.With(jwtSvc.RequireRole("admin")).Post("/", lookupHandler.Create)
				r.With(jwtSvc.RequireRole("admin")).Put("/{id}", lookupHandler.Update)
				r.With(jwtSvc.RequireRole("admin")).Patch("/{id}/active", lookupHandler.PatchActive)
			})
		})
	})

	// 7. Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Server error channel
	serverErr := make(chan error, 1)
	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr))
		serverErr <- srv.ListenAndServe()
	}()

	// Wait for termination signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatal("server error", zap.Error(err))
	case sig := <-quit:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	// Graceful shutdown with 30s timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
		os.Exit(1)
	}
	log.Info("server stopped gracefully")
}
