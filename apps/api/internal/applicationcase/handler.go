package applicationcase

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo Repository }

func RegisterPublicRoutes(group *gin.RouterGroup, repo Repository) {
	group.GET("/:id/application-cases", (&Handler{repo: repo}).list)
}

func RegisterAdminRoutes(group *gin.RouterGroup, repo Repository) {
	handler := &Handler{repo: repo}
	group.GET("/:id/application-cases", handler.list)
	group.POST("/:id/application-cases", handler.create)
	group.PUT("/:id/application-cases/:caseId", handler.update)
	group.DELETE("/:id/application-cases/:caseId", handler.delete)
}

func (h *Handler) list(ctx *gin.Context) {
	items, err := h.repo.List(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) create(ctx *gin.Context) {
	var input Input
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	item, err := h.repo.Create(ctx.Request.Context(), ctx.Param("id"), input)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, item)
}

func (h *Handler) update(ctx *gin.Context) {
	var input Input
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	item, err := h.repo.Update(ctx.Request.Context(), ctx.Param("id"), ctx.Param("caseId"), input)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, item)
}

func (h *Handler) delete(ctx *gin.Context) {
	if err := h.repo.Delete(ctx.Request.Context(), ctx.Param("id"), ctx.Param("caseId")); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func respondError(ctx *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrSpeciesNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
}
