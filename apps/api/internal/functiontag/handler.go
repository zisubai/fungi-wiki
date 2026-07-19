package functiontag

import (
	"errors"
	"net/http"
	"strconv"

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
	router.GET("", handler.List)
	router.GET("/:id", handler.Get)
}

func RegisterAdminRoutes(router *gin.RouterGroup, repo Repository) {
	handler := NewHandler(repo)
	router.GET("", handler.List)
	router.POST("", handler.Create)
	router.GET("/:id", handler.Get)
	router.PUT("/:id", handler.Update)
	router.DELETE("/:id", handler.Delete)
}

func (handler *Handler) List(ctx *gin.Context) {
	items, err := handler.repo.List(ctx.Request.Context(), ListParams{
		Query:  ctx.Query("q"),
		Limit:  parseInt(ctx.Query("limit"), 100),
		Offset: parseInt(ctx.Query("offset"), 0),
	})
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (handler *Handler) Get(ctx *gin.Context) {
	item, err := handler.repo.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
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
		ctx.JSON(http.StatusNotFound, gin.H{"message": "function tag not found"})
	case errors.Is(err, ErrDuplicateCode):
		ctx.JSON(http.StatusConflict, gin.H{"message": "function tag code already exists"})
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
