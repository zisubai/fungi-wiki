package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fungi-wiki/apps/api/internal/audit"
	"fungi-wiki/apps/api/internal/auth"
	"fungi-wiki/apps/api/internal/config"
	"fungi-wiki/apps/api/internal/culturecondition"
	"fungi-wiki/apps/api/internal/evidence"
	"fungi-wiki/apps/api/internal/functiontag"
	"fungi-wiki/apps/api/internal/health"
	"fungi-wiki/apps/api/internal/importjob"
	"fungi-wiki/apps/api/internal/searchanalytics"
	"fungi-wiki/apps/api/internal/species"
	"fungi-wiki/apps/api/internal/speciesfunction"
)

func NewRouter(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	router.GET("/healthz", health.Handle(cfg))

	speciesRepo := species.NewPostgresRepository(pool)
	functionTagRepo := functiontag.NewPostgresRepository(pool)
	speciesFunctionRepo := speciesfunction.NewPostgresRepository(pool)
	cultureRepo := culturecondition.NewPostgresRepository(pool)
	evidenceRepo := evidence.NewPostgresRepository(pool)
	auditRepo := audit.NewPostgresRepository(pool)
	importRepo := importjob.NewPostgresRepository(pool)
	authRepo := auth.NewPostgresRepository(pool)
	tokenService := auth.NewTokenService(cfg.JWTSecret)
	searchAnalyticsRepo := searchanalytics.NewPostgresRepository(pool)

	api := router.Group("/api")
	{
		auth.RegisterRoutes(api.Group("/auth"), authRepo, tokenService)
		species.RegisterPublicRoutes(api.Group("/species"), speciesRepo)
		publishedSpecies := api.Group("/species", publishedSpeciesOnly(pool))
		speciesfunction.RegisterPublicRoutes(publishedSpecies, speciesFunctionRepo)
		culturecondition.RegisterPublicRoutes(publishedSpecies, cultureRepo)
		evidence.RegisterPublicRoutes(publishedSpecies, evidenceRepo)
		functiontag.RegisterPublicRoutes(api.Group("/function-tags"), functionTagRepo)

		admin := api.Group("/admin", auth.Authenticate(tokenService), auth.AuthorizeAdmin())
		{
			species.RegisterAdminRoutes(admin.Group("/species"), speciesRepo)
			speciesfunction.RegisterAdminRoutes(admin.Group("/species"), speciesFunctionRepo)
			culturecondition.RegisterAdminRoutes(admin.Group("/species"), cultureRepo)
			evidence.RegisterAdminRoutes(admin.Group("/species"), evidenceRepo)
			functiontag.RegisterAdminRoutes(admin.Group("/function-tags"), functionTagRepo)
			audit.RegisterAdminRoutes(admin.Group("/audits"), auditRepo)
			importjob.RegisterAdminRoutes(admin.Group("/imports"), importRepo)
			auth.RegisterAdminRoutes(admin.Group("/users"), authRepo)
			searchanalytics.RegisterAdminRoutes(admin.Group("/search-analytics"), searchAnalyticsRepo)
		}
	}

	return router
}

func publishedSpeciesOnly(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var exists bool
		err := pool.QueryRow(ctx.Request.Context(), `SELECT EXISTS(SELECT 1 FROM species WHERE (id::text = $1 OR slug = $1) AND status = 'published')`, ctx.Param("id")).Scan(&exists)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "species not found"})
			return
		}
		ctx.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
