package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"fungi-wiki/apps/api/internal/config"
)

type DatabasePinger interface {
	Ping(context.Context) error
}

type Response struct {
	App       string    `json:"app"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func Ready(cfg config.Config, database DatabasePinger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := "ready"
		statusCode := http.StatusOK
		if err := database.Ping(ctx.Request.Context()); err != nil {
			status = "unavailable"
			statusCode = http.StatusServiceUnavailable
		}
		ctx.JSON(statusCode, Response{App: cfg.AppName, Status: status, Timestamp: time.Now().UTC()})
	}
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
