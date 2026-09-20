package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/util"
)

// HealthHandler 健康检查。
type HealthHandler struct{}

// NewHealthHandler 构造健康检查处理器。
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Healthz 健康检查接口。
func (h *HealthHandler) Healthz(c *gin.Context) {
	util.OK(c, gin.H{
		"status":  "ok",
		"service": "paperflow-backend",
		"time":    time.Now().Format(time.RFC3339),
	})
}
