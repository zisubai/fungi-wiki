package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"fungi-wiki/apps/api/internal/config"
)

type Response struct {
	App       string    `json:"app"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func Handle(cfg config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, Response{
			App:       cfg.AppName,
			Status:    "ok",
			Timestamp: time.Now().UTC(),
		})
	}
}
