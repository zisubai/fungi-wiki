package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"fungi-wiki/apps/api/internal/auth"
	"fungi-wiki/apps/api/internal/config"
	"fungi-wiki/apps/api/internal/httpserver"
	"fungi-wiki/apps/api/migrations"
	"fungi-wiki/apps/api/pkg/database"
)

func main() {
	cfg := config.Load()

	connectCtx, cancelConnect := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelConnect()

	pool, err := database.Connect(connectCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()
	migrationCtx, cancelMigration := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelMigration()
	if err := migrations.Run(migrationCtx, pool); err != nil {
		log.Fatalf("run database migrations: %v", err)
	}
	authRepo := auth.NewPostgresRepository(pool)
	if err := authRepo.BootstrapAdmin(migrationCtx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("bootstrap admin: %v (apply migration 004_users_and_roles.sql first)", err)
	}

	router, err := httpserver.NewRouter(cfg, pool)
	if err != nil {
		log.Fatalf("configure HTTP router: %v", err)
	}

	server := newHTTPServer(cfg.HTTPAddr, router)
	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Printf("fungi wiki api listening on %s", cfg.HTTPAddr)
	if err := runHTTPServer(runCtx, server, 10*time.Second); err != nil {
		log.Fatalf("api server stopped: %v", err)
	}
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func runHTTPServer(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
