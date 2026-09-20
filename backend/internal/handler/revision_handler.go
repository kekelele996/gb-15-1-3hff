package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// RevisionHandler 修稿记录处理器。
type RevisionHandler struct {
	revisionSvc *service.RevisionService
	logger      *slog.Logger
}

// NewRevisionHandler 构造修稿记录处理器。
func NewRevisionHandler(revisionSvc *service.RevisionService, logger *slog.Logger) *RevisionHandler {
	return &RevisionHandler{revisionSvc: revisionSvc, logger: logger}
}

// ListByPaper 论文修稿记录列表。
func (h *RevisionHandler) ListByPaper(c *gin.Context) {
	paperID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || paperID == 0 {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "请求失败：路径参数 id 不合法")
		return
	}
	items, err := h.revisionSvc.ListByPaper(c.Request.Context(), uint(paperID))
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, items)
}

// wrapError 修稿处理器错误包装。
func (h *RevisionHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}
