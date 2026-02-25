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
	"nexusasset/backend/internal/config"
	dbpkg "nexusasset/backend/internal/database"
	"nexusasset/backend/internal/logger"
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
