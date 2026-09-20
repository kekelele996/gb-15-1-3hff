package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
)

// RegisterPaperRoutes 论文相关路由。
func RegisterPaperRoutes(g *gin.RouterGroup, h *handler.PaperHandler) {
	g.POST("/papers", middleware.RequireRoles(constants.RoleAuthor, constants.RoleAdmin), h.Create)
	g.GET("/papers/mine", middleware.RequireRoles(constants.RoleAuthor, constants.RoleAdmin), h.ListMine)
	g.GET("/papers", h.ListAll)
	g.GET("/papers/:id", h.Detail)
	g.PUT("/papers/:id", middleware.RequireRoles(constants.RoleAuthor, constants.RoleAdmin), h.Update)
	g.POST("/papers/:id/initial-review", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.InitialReview)
	g.POST("/papers/:id/final-decision", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.FinalDecision)
	g.POST("/papers/:id/revise", middleware.RequireRoles(constants.RoleAuthor, constants.RoleAdmin), h.Revise)
	g.GET("/library/papers", h.LibrarySearch)
}
