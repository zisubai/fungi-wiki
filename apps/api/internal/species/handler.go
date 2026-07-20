package species

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func RegisterPublicRoutes(router *gin.RouterGroup, repo Repository) {
	handler := NewHandler(repo)
	router.GET("", handler.ListPublished)
	router.GET("/compare", handler.Compare)
	router.GET("/:id", handler.GetPublished)
}

func RegisterAdminRoutes(router *gin.RouterGroup, repo Repository) {
	handler := NewHandler(repo)
	router.GET("", handler.ListAll)
	router.POST("", handler.Create)
	router.GET("/:id/quality", handler.Quality)
	router.GET("/:id", handler.Get)
	router.PUT("/:id", handler.Update)
	router.DELETE("/:id", handler.Archive)
	router.DELETE("/:id/hard", handler.Delete)
}

func (handler *Handler) Quality(ctx *gin.Context) {
	report, err := handler.repo.Quality(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, report)
}

func (handler *Handler) ListPublished(ctx *gin.Context) {
	handler.list(ctx, "published")
}

func (handler *Handler) ListAll(ctx *gin.Context) {
	handler.list(ctx, ctx.Query("status"))
}

func (handler *Handler) Compare(ctx *gin.Context) {
	rawIDs := strings.Split(ctx.Query("ids"), ",")
	ids := make([]string, 0, len(rawIDs))
	seen := make(map[string]struct{})
	for _, id := range rawIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	if len(ids) < 2 || len(ids) > 3 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "select 2 to 3 unique species"})
		return
	}
	items, err := handler.repo.Compare(ctx.Request.Context(), ids)
	if err != nil {
		respondError(ctx, err)
		return
	}
	if len(items) != len(ids) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "one or more published species were not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (handler *Handler) list(ctx *gin.Context, status string) {
	temperature, err := parseOptionalFloat(ctx.Query("temperature"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "temperature must be a number"})
		return
	}
	ph, err := parseOptionalFloat(ctx.Query("ph"))
	if err != nil || ph != nil && (*ph < 0 || *ph > 14) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "ph must be a number between 0 and 14"})
		return
	}
	params := ListParams{
		Query:             ctx.Query("q"),
		Status:            status,
		FunctionTag:       ctx.Query("functionTag"),
		Temperature:       temperature,
		PH:                ph,
		SafetyLevel:       ctx.Query("safetyLevel"),
		SourceEnvironment: ctx.Query("sourceEnvironment"),
		Sort:              ctx.Query("sort"),
		Limit:             parseInt(ctx.Query("limit"), 20),
		Offset:            parseInt(ctx.Query("offset"), 0),
	}
	result, err := handler.repo.List(ctx.Request.Context(), params)
	if err != nil {
		respondError(ctx, err)
		return
	}

	if status == "published" && hasSearchCriteria(params) {
		if logger, ok := handler.repo.(interface {
			LogSearch(context.Context, ListParams, int) error
		}); ok {
			_ = logger.LogSearch(ctx.Request.Context(), params, result.Total)
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"items": result.Items, "total": result.Total, "limit": params.Limit, "offset": params.Offset})
}

func hasSearchCriteria(params ListParams) bool {
	return strings.TrimSpace(params.Query) != "" || params.FunctionTag != "" || params.Temperature != nil || params.PH != nil || params.SafetyLevel != "" || params.SourceEnvironment != ""
}

func parseOptionalFloat(value string) (*float64, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (handler *Handler) Get(ctx *gin.Context) {
	item, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (handler *Handler) GetPublished(ctx *gin.Context) {
	item, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil || item.Status != StatusPublished {
		if err == nil {
			err = ErrNotFound
		}
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (handler *Handler) Create(ctx *gin.Context) {
	var input CreateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	item, err := handler.repo.Create(ctx.Request.Context(), input)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, item)
}

func (handler *Handler) Update(ctx *gin.Context) {
	current, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		respondError(ctx, err)
		return
	}
	if current.Status == StatusPendingReview {
		ctx.JSON(http.StatusConflict, gin.H{"message": "species pending review cannot be edited"})
		return
	}
	var input UpdateInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	item, err := handler.repo.Update(ctx.Request.Context(), ctx.Param("id"), input)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (handler *Handler) Archive(ctx *gin.Context) {
	if err := handler.repo.Archive(ctx.Request.Context(), ctx.Param("id")); err != nil {
		respondError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (handler *Handler) Delete(ctx *gin.Context) {
	if err := handler.repo.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		respondError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func respondError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"message": "species not found"})
	case errors.Is(err, ErrDuplicateSlug):
		ctx.JSON(http.StatusConflict, gin.H{"message": "species slug already exists"})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
	}
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
