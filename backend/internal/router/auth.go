package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/handler"
)

// RegisterAuthRoutes 认证相关路由（公开）。
func RegisterAuthRoutes(g *gin.RouterGroup, h *handler.AuthHandler) {
	g.POST("/auth/register", h.Register)
	g.POST("/auth/login", h.Login)
}
