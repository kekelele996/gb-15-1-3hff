package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
)

// RegisterStatisticsRoutes 统计面板路由。
func RegisterStatisticsRoutes(g *gin.RouterGroup, h *handler.StatisticsHandler) {
	g.GET("/statistics/overview", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.Overview)
	g.GET("/statistics/trend", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.Trend)
	g.GET("/statistics/subjects", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.Subjects)
	g.GET("/statistics/reviewers", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.ReviewerWorkload)
}
