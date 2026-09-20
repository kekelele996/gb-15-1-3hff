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

// StatisticsHandler 统计面板处理器。
type StatisticsHandler struct {
	statSvc *service.StatisticsService
	logger  *slog.Logger
}

// NewStatisticsHandler 构造统计处理器。
func NewStatisticsHandler(statSvc *service.StatisticsService, logger *slog.Logger) *StatisticsHandler {
	return &StatisticsHandler{statSvc: statSvc, logger: logger}
}

// Overview 统计概览。
func (h *StatisticsHandler) Overview(c *gin.Context) {
	o, err := h.statSvc.Overview(c.Request.Context())
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, o)
}

// Trend 投稿量趋势。
func (h *StatisticsHandler) Trend(c *gin.Context) {
	var q dto.TrendQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "查询失败：days 参数不合法")
		return
	}
	rows, err := h.statSvc.Trend(c.Request.Context(), q.Days)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, rows)
}

// Subjects 学科分布。
func (h *StatisticsHandler) Subjects(c *gin.Context) {
	rows, err := h.statSvc.Subjects(c.Request.Context())
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, rows)
}

// ReviewerWorkload 审稿人工作量排名。
func (h *StatisticsHandler) ReviewerWorkload(c *gin.Context) {
	rows, err := h.statSvc.ReviewerWorkload(c.Request.Context())
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, rows)
}

// wrapError 统计处理器错误包装。
func (h *StatisticsHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}
