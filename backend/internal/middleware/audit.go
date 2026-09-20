package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// AuditMiddleware 操作审计日志中间件：记录受保护接口的写操作。
type AuditMiddleware struct {
	auditSvc *service.AuditLogService
	logger   *slog.Logger
}

// NewAudit 构造审计中间件。
func NewAudit(auditSvc *service.AuditLogService, logger *slog.Logger) *AuditMiddleware {
	return &AuditMiddleware{auditSvc: auditSvc, logger: logger}
}

// Audit 在请求完成后记录审计日志（仅记录 2xx/3xx 写请求）。
func (m *AuditMiddleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() >= 200 && c.Writer.Status() < 400 && c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			uid := util.GetUserID(c)
			username := util.GetUsername(c)
			action := c.Request.Method + " " + c.Request.URL.Path
			entity := parseEntity(c.Request.URL.Path)
			entityID := c.Param("id")
			detail := c.Request.URL.RawQuery
			if err := m.auditSvc.Record(c.Request.Context(), uid, username, action, entity, entityID, detail, c.ClientIP(), util.GetRequestID(c)); err != nil {
				m.logger.Error("audit middleware record failed", "error", err)
			}
		}
	}
}

// parseEntity 依据路径前缀解析审计实体。
func parseEntity(path string) string {
	p := strings.TrimPrefix(path, "/api/v1/")
	seg := strings.Split(p, "/")[0]
	switch {
	case strings.HasPrefix(seg, "paper"):
		return "paper"
	case strings.HasPrefix(seg, "review"):
		return "review"
	case strings.HasPrefix(seg, "revision"):
		return "revision"
	case strings.HasPrefix(seg, "user"):
		return "user"
	case strings.HasPrefix(seg, "audit"):
		return "audit"
	case strings.HasPrefix(seg, "file"):
		return "file"
	default:
		return seg
	}
}
