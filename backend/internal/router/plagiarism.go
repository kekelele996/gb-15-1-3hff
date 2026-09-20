package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
)

// RegisterPlagiarismRoutes 查重相关路由。
func RegisterPlagiarismRoutes(g *gin.RouterGroup, h *handler.PlagiarismHandler) {
	g.GET("/papers/:id/plagiarism", h.GetByPaper)
	g.POST("/papers/:id/plagiarism/rerun", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.Rerun)
}
