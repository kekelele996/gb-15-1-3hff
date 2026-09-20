package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// AuthMiddleware JWT 认证中间件。
type AuthMiddleware struct {
	userRepo repository.UserRepository
	secret   string
	logger   *slog.Logger
}

// NewAuth 构造认证中间件。
func NewAuth(userRepo repository.UserRepository, secret string, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{userRepo: userRepo, secret: secret, logger: logger}
}

// RequireAuth 校验 Bearer Token 并加载用户写入上下文。
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			util.Fail(c, http.StatusUnauthorized, constants.ErrInvalidToken, "认证失败：缺少访问令牌")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := util.ParseToken(m.secret, tokenStr)
		if err != nil {
			m.logger.Warn("parse token failed", "error", err)
			util.Fail(c, http.StatusUnauthorized, constants.ErrInvalidToken, "认证失败：访问令牌无效或已过期")
			return
		}
		user, err := m.userRepo.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				util.Fail(c, http.StatusUnauthorized, constants.ErrInvalidToken, "认证失败：用户不存在")
				return
			}
			m.logger.Error("load user failed", "error", err)
			util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "认证失败：系统内部错误")
			return
		}
		util.SetAuthUser(c, user.ID, user.Username, user.Role)
		c.Next()
	}
}
