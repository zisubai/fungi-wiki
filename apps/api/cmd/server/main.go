package main

import (
	"context"
	"log"
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

	router := httpserver.NewRouter(cfg, pool)

	log.Printf("fungi wiki api listening on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("api server stopped: %v", err)
	}
}
