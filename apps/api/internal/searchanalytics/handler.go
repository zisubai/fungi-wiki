package searchanalytics

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func RegisterAdminRoutes(r *gin.RouterGroup, repo Repository) {
	r.GET("", func(c *gin.Context) {
		days, _ := strconv.Atoi(c.Query("days"))
		report, err := repo.Report(c.Request.Context(), days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}
		c.JSON(200, report)
	})
}
