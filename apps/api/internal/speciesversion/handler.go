package speciesversion

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(group *gin.RouterGroup, repo Repository) {
	group.GET("/:id/versions", func(ctx *gin.Context) {
		limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "30"))
		items, err := repo.List(ctx.Request.Context(), ctx.Param("id"), limit)
		if errors.Is(err, ErrSpeciesNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"items": items})
	})
}
