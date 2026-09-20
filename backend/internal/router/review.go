package router

import (
	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
)

// RegisterReviewRoutes 审稿相关路由。
func RegisterReviewRoutes(g *gin.RouterGroup, h *handler.ReviewHandler) {
	g.GET("/reviews/mine", middleware.RequireRoles(constants.RoleReviewer, constants.RoleAdmin), h.ListMine)
	g.GET("/reviews/paper/:paperID", h.ListByPaper)
	g.POST("/reviews/:id/respond", middleware.RequireRoles(constants.RoleReviewer, constants.RoleAdmin), h.Respond)
	g.POST("/reviews/:id/submit", middleware.RequireRoles(constants.RoleReviewer, constants.RoleAdmin), h.Submit)
	g.POST("/papers/:id/reviewers", middleware.RequireRoles(constants.RoleEditor, constants.RoleAdmin), h.Assign)
}
