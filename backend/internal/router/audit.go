package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
)

// RegisterAuditRoutes 审计日志路由。
func RegisterAuditRoutes(g *gin.RouterGroup, h *handler.AuditHandler) {
	g.GET("/audit-logs", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.List)
}
