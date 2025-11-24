package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"webhook-receiver/internal/config"
	"webhook-receiver/internal/handlers"
	"webhook-receiver/internal/server/middleware"
)

func New(cfg *config.Config, logHandler *handlers.LogHandler) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.JSONLogMiddleware())

	r.GET("/healthz", HealthzHandler)
	r.POST("/log", logHandler.Handle)

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func HealthzHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
