package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/handler"
)

// RegisterRevisionRoutes 修稿记录路由。
func RegisterRevisionRoutes(g *gin.RouterGroup, h *handler.RevisionHandler) {
	g.GET("/papers/:id/revisions", h.ListByPaper)
}
