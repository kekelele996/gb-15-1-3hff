package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// AuthHandler 认证处理器。
type AuthHandler struct {
	authSvc  *service.AuthService
	auditSvc *service.AuditLogService
	logger   *slog.Logger
}

// NewAuthHandler 构造认证处理器。
func NewAuthHandler(authSvc *service.AuthService, auditSvc *service.AuditLogService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, auditSvc: auditSvc, logger: logger}
}

// Register 注册接口。
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("register request validation failed", "error", err)
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "注册失败：请求参数不合法（username/password/real_name 等字段校验未通过）")
		return
	}
	user, err := h.authSvc.Register(c.Request.Context(), req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), user.ID, user.Username,
		constants.AuditActionRegister, "user", fmt.Sprint(user.ID), "注册账号", c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit register failed", "error", e)
	}
	util.OK(c, toUserSummary(user))
}

// Login 登录接口。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("login request validation failed", "error", err)
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "登录失败：请求参数不合法")
		return
	}
	user, token, err := h.authSvc.Login(c.Request.Context(), req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), user.ID, user.Username,
		constants.AuditActionLogin, "user", fmt.Sprint(user.ID), "用户登录", c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit login failed", "error", e)
	}
	util.OK(c, dto.LoginResponse{Token: token, User: *toUserSummary(user)})
}

// wrapError 再次包装 service 错误并转换为统一响应。
func (h *AuthHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}

// toUserSummary 转换为用户摘要。
func toUserSummary(user *model.User) *dto.UserSummary {
	return &dto.UserSummary{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		RealName:    user.RealName,
		Institution: user.Institution,
		Role:        user.Role,
	}
}

// httpStatusForCode 业务错误码到 HTTP 状态码映射。
func httpStatusForCode(code int) int {
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
	case constants.ErrSubjectNotAllowed:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadRequest
	}
}
