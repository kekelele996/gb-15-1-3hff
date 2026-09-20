package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/handler"
)

// RegisterFileRoutes 文件上传路由。
func RegisterFileRoutes(g *gin.RouterGroup, h *handler.FileHandler) {
	g.POST("/files/upload", h.Upload)
}
