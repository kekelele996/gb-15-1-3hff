package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
)

// RegisterUserRoutes 用户相关路由。
func RegisterUserRoutes(g *gin.RouterGroup, h *handler.UserHandler) {
	g.GET("/users/me", h.Me)
	g.PUT("/users/me", h.UpdateProfile)
	g.GET("/users/reviewers", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.ListReviewers)
}
