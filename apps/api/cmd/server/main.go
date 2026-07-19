package main

import (
	"context"
	"log"
	"time"

	"fungi-wiki/apps/api/internal/config"
	"fungi-wiki/apps/api/internal/httpserver"
	"fungi-wiki/apps/api/pkg/database"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	router := httpserver.NewRouter(cfg, pool)

	log.Printf("fungi wiki api listening on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("api server stopped: %v", err)
	}
}
