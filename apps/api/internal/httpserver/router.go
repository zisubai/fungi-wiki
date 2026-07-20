package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fungi-wiki/apps/api/internal/audit"
	"fungi-wiki/apps/api/internal/auth"
	"fungi-wiki/apps/api/internal/config"
	"fungi-wiki/apps/api/internal/culturecondition"
	"fungi-wiki/apps/api/internal/dataquality"
	"fungi-wiki/apps/api/internal/evidence"
	"fungi-wiki/apps/api/internal/functiontag"
	"fungi-wiki/apps/api/internal/health"
	"fungi-wiki/apps/api/internal/importjob"
	"fungi-wiki/apps/api/internal/recommendation"
	"fungi-wiki/apps/api/internal/searchanalytics"
	"fungi-wiki/apps/api/internal/species"
	"fungi-wiki/apps/api/internal/speciesalias"
	"fungi-wiki/apps/api/internal/speciesfunction"
)

func NewRouter(cfg config.Config, pool *pgxpool.Pool) (*gin.Engine, error) {
	router := gin.New()
	trustedProxies := splitCommaSeparated(cfg.TrustedProxies)
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("configure trusted proxies: %w", err)
	}
	router.Use(requestIDMiddleware(), securityHeadersMiddleware(), requestLoggerMiddleware(), gin.Recovery(), corsMiddleware(cfg.CORSOrigins))

	router.GET("/healthz", health.Handle(cfg))
	router.GET("/readyz", health.Ready(cfg, pool))

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
	speciesAliasRepo := speciesalias.NewPostgresRepository(pool)
	recommendationRepo := recommendation.NewPostgresRepository(pool)
	dataQualityRepo := dataquality.NewPostgresRepository(pool)

	api := router.Group("/api")
	{
		auth.RegisterRoutes(api.Group("/auth"), authRepo, tokenService)
		species.RegisterPublicRoutes(api.Group("/species"), speciesRepo)
		publishedSpecies := api.Group("/species", publishedSpeciesOnly(pool))
		speciesfunction.RegisterPublicRoutes(publishedSpecies, speciesFunctionRepo)
		culturecondition.RegisterPublicRoutes(publishedSpecies, cultureRepo)
		evidence.RegisterPublicRoutes(publishedSpecies, evidenceRepo)
		speciesalias.RegisterPublicRoutes(publishedSpecies, speciesAliasRepo)
		functiontag.RegisterPublicRoutes(api.Group("/function-tags"), functionTagRepo)
		recommendation.RegisterPublicRoutes(api.Group("/recommendations"), recommendationRepo)
		recommendation.RegisterFeedbackRoutes(api.Group("/recommendations"), recommendationRepo)

		admin := api.Group("/admin", auth.Authenticate(tokenService), auth.AuthorizeAdmin())
		{
			species.RegisterAdminRoutes(admin.Group("/species"), speciesRepo)
			speciesfunction.RegisterAdminRoutes(admin.Group("/species"), speciesFunctionRepo)
			culturecondition.RegisterAdminRoutes(admin.Group("/species"), cultureRepo)
			evidence.RegisterAdminRoutes(admin.Group("/species"), evidenceRepo)
			speciesalias.RegisterAdminRoutes(admin.Group("/species"), speciesAliasRepo)
			functiontag.RegisterAdminRoutes(admin.Group("/function-tags"), functionTagRepo)
			audit.RegisterAdminRoutes(admin.Group("/audits"), auditRepo)
			importjob.RegisterAdminRoutes(admin.Group("/imports"), importRepo)
			auth.RegisterAdminRoutes(admin.Group("/users"), authRepo)
			searchanalytics.RegisterAdminRoutes(admin.Group("/search-analytics"), searchAnalyticsRepo)
			recommendation.RegisterAdminRoutes(admin.Group("/recommendations"), recommendationRepo)
			dataquality.RegisterAdminRoutes(admin.Group("/data-quality"), dataQualityRepo)
		}
	}

	return router, nil
}

func splitCommaSeparated(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func requestIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader("X-Request-ID")
		if !requestIDPattern.MatchString(requestID) {
			buffer := make([]byte, 16)
			if _, err := rand.Read(buffer); err != nil {
				requestID = "request-id-unavailable"
			} else {
				requestID = hex.EncodeToString(buffer)
			}
		}
		ctx.Set("requestID", requestID)
		ctx.Header("X-Request-ID", requestID)
		ctx.Next()
	}
}

func securityHeadersMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "DENY")
		ctx.Header("Referrer-Policy", "no-referrer")
		ctx.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		ctx.Next()
	}
}

func requestLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		requestID, _ := params.Keys["requestID"].(string)
		return fmt.Sprintf("request_id=%s method=%s path=%s status=%d latency=%s client_ip=%s error=%q\n",
			requestID,
			params.Method,
			params.Path,
			params.StatusCode,
			params.Latency,
			params.ClientIP,
			params.ErrorMessage,
		)
	})
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

func corsMiddleware(origins string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{})
	for _, origin := range strings.Split(origins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" {
			_, allowed := allowedOrigins[origin]
			_, wildcard := allowedOrigins["*"]
			if !allowed && !wildcard {
				if ctx.Request.Method == http.MethodOptions {
					ctx.AbortWithStatus(http.StatusForbidden)
					return
				}
				ctx.Next()
				return
			}
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Vary", "Origin")
		}
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
		ctx.Header("Access-Control-Expose-Headers", "X-Request-ID")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
