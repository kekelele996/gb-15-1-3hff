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

// UserHandler 用户处理器。
type UserHandler struct {
	userSvc *service.UserService
	logger  *slog.Logger
}

// NewUserHandler 构造用户处理器。
func NewUserHandler(userSvc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{userSvc: userSvc, logger: logger}
}

// Me 当前用户信息。
func (h *UserHandler) Me(c *gin.Context) {
	uid := util.GetUserID(c)
	user, err := h.userSvc.Me(c.Request.Context(), uid)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, toUserSummary(user))
}

// UpdateProfile 更新当前用户资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "更新资料失败：请求参数不合法")
		return
	}
	user, err := h.userSvc.UpdateProfile(c.Request.Context(), util.GetUserID(c), req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, toUserSummary(user))
}

// ListReviewers 列出全部审稿人（编辑分配审稿人使用）。
func (h *UserHandler) ListReviewers(c *gin.Context) {
	items, err := h.userSvc.ListReviewers(c.Request.Context())
	if err != nil {
		h.wrapError(c, err)
		return
	}
	out := make([]*dto.UserSummary, 0, len(items))
	for i := range items {
		out = append(out, toUserSummary(&items[i]))
	}
	util.OK(c, out)
}

// wrapError 用户处理器错误包装。
func (h *UserHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}
