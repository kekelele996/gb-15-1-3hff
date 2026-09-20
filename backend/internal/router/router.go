package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
	"github.com/paperflow/paperflow/internal/util"
)

// Router 路由装配器。
type Router struct {
	engine *gin.Engine
	h      *handler.Handlers
}

// New 构造 Router 并注册全部路由与中间件。
func New(engine *gin.Engine, h *handler.Handlers, auth *middleware.AuthMiddleware, audit *middleware.AuditMiddleware, rateLimit gin.HandlerFunc, logger *slog.Logger) *Router {
	r := &Router{engine: engine, h: h}
	engine.Use(middleware.RequestID(), middleware.ErrorHandler(logger), rateLimit)

	engine.GET("/healthz", h.Health.Healthz)

	api := engine.Group("/api/v1")
	RegisterAuthRoutes(api, h.Auth)

	// 受保护路由组：JWT 认证 + 审计
	protected := api.Group("", auth.RequireAuth(), audit.Audit())
	RegisterUserRoutes(protected, h.User)
	RegisterFileRoutes(protected, h.File)
	RegisterPaperRoutes(protected, h.Paper)
	RegisterReviewRoutes(protected, h.Review)
	RegisterRevisionRoutes(protected, h.Revision)
	RegisterPlagiarismRoutes(protected, h.Plagiarism)
	RegisterAuditRoutes(protected, h.Audit)
	RegisterStatisticsRoutes(protected, h.Statistics)
	return r
}

// Engine 返回 gin.Engine。
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// NoRoute 兜底处理。
func (r *Router) NoRoute() {
	r.engine.NoRoute(func(c *gin.Context) {
		util.Fail(c, http.StatusNotFound, constants.CodeNotFound, "接口不存在："+c.Request.URL.Path)
	})
}
