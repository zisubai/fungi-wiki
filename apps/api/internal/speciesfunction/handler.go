package speciesfunction

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func RegisterPublicRoutes(router *gin.RouterGroup, repo Repository) {
	router.GET("/:id/functions", NewHandler(repo).List)
}

func RegisterAdminRoutes(router *gin.RouterGroup, repo Repository) {
	handler := NewHandler(repo)
	router.GET("/:id/functions", handler.List)
	router.PUT("/:id/functions", handler.Replace)
}

func (handler *Handler) List(ctx *gin.Context) {
	items, err := handler.repo.List(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (handler *Handler) Replace(ctx *gin.Context) {
	var input ReplaceInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	for _, item := range input.Items {
		if item.ConfidenceScore < 0 || item.ConfidenceScore > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "confidenceScore must be between 0 and 100"})
			return
		}
	}
	items, err := handler.repo.Replace(ctx.Request.Context(), ctx.Param("id"), input.Items)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func respondError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrSpeciesNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "species not found"})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}
