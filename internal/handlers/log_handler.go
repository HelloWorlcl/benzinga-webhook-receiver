package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"webhook-receiver/internal/model"
	"webhook-receiver/internal/processor"
)

type LogHandler struct {
	bp  *processor.BatchProcessor
	log *zap.Logger
}

func NewLogHandler(bp *processor.BatchProcessor, log *zap.Logger) *LogHandler {
	return &LogHandler{
		bp:  bp,
		log: log,
	}
}

func (h *LogHandler) Handle(c *gin.Context) {
	var logEntry model.LogEntry

	if err := c.BindJSON(&logEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.bp.Add(logEntry)
	c.Status(http.StatusOK)
}
