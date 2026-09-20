package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// AuditHandler 审计日志处理器。
type AuditHandler struct {
	auditSvc *service.AuditLogService
	logger   *slog.Logger
}

// NewAuditHandler 构造审计日志处理器。
func NewAuditHandler(auditSvc *service.AuditLogService, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc, logger: logger}
}

// List 分页查询审计日志（编辑/管理员）。
func (h *AuditHandler) List(c *gin.Context) {
	var q dto.AuditQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "查询失败：分页参数不合法")
		return
	}
	page, size := pageSize(q.Page, q.Size)
	items, total, err := h.auditSvc.List(c.Request.Context(), page, size)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, dto.PageResult{Total: total, Page: page, Size: size, Items: items})
}

// wrapError 审计处理器错误包装。
func (h *AuditHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}
