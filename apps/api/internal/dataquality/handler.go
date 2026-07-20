package dataquality

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.RouterGroup, repo Repository) {
	router.GET("", func(ctx *gin.Context) {
		report, err := repo.Report(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, report)
	})
}
