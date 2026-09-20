package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/util"
)

// ErrorHandler 全局错误处理：panic 恢复 + 将 c.Error 转换为统一 JSON 响应。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered", "panic", rec, "stack", string(debug.Stack()),
					"path", c.Request.URL.Path, "request_id", util.GetRequestID(c))
				util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误：服务处理异常")
			}
		}()
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr *util.AppError
			if errors.As(err, &appErr) {
				util.Fail(c, httpStatusForAppError(appErr.Code), appErr.Code, appErr.Message)
				return
			}
			logger.Error("unhandled error", "error", err, "path", c.Request.URL.Path,
				"request_id", util.GetRequestID(c))
			util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
		}
	}
}

func httpStatusForAppError(code int) int {
	switch code {
	case constants.ErrInvalidCredential, constants.ErrInvalidToken:
		return http.StatusUnauthorized
	case constants.ErrPermissionDenied, constants.ErrRoleNotAllowed:
		return http.StatusForbidden
	case constants.ErrUserNotFound, constants.ErrPaperNotFound, constants.ErrReviewNotFound,
		constants.ErrRevisionNotFound, constants.ErrPlagiarismNotFound, constants.ErrAuditNotFound:
		return http.StatusNotFound
	case constants.ErrUserExists, constants.ErrPaperTitleExists, constants.ErrPaperStatusNotAllowed,
		constants.ErrReviewNotAllowed:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

