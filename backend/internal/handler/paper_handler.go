package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// PaperHandler 论文处理器。
type PaperHandler struct {
	paperSvc  *service.PaperService
	auditSvc  *service.AuditLogService
	logger    *slog.Logger
}

// NewPaperHandler 构造论文处理器。
func NewPaperHandler(paperSvc *service.PaperService, auditSvc *service.AuditLogService, logger *slog.Logger) *PaperHandler {
	return &PaperHandler{paperSvc: paperSvc, auditSvc: auditSvc, logger: logger}
}

// Create 新建投稿（作者）。
func (h *PaperHandler) Create(c *gin.Context) {
	var req dto.CreatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "投稿失败：请求参数不合法（title/abstract/keywords/file_key 等字段校验未通过）")
		return
	}
	paper, err := h.paperSvc.Create(c.Request.Context(), util.GetUserID(c), req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionCreatePaper, "paper", fmt.Sprint(paper.ID), "新建投稿："+paper.Title, c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit create paper failed", "error", e)
	}
	util.OK(c, paper)
}

// ListMine 我的投稿列表（作者）。
func (h *PaperHandler) ListMine(c *gin.Context) {
	var q dto.PaperQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "查询失败：分页参数不合法")
		return
	}
	page, size := pageSize(q.Page, q.Size)
	items, total, err := h.paperSvc.ListMine(c.Request.Context(), util.GetUserID(c), page, size)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	h.logger.Info(fmt.Sprintf(constants.LogPaperList, util.GetRole(c), q.Status, page))
	util.OK(c, dto.PageResult{Total: total, Page: page, Size: size, Items: items})
}

// ListAll 按状态列出全部论文（编辑/管理员）。
func (h *PaperHandler) ListAll(c *gin.Context) {
	var q dto.PaperQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "查询失败：分页参数不合法")
		return
	}
	page, size := pageSize(q.Page, q.Size)
	items, total, err := h.paperSvc.ListByStatus(c.Request.Context(), q.Status, page, size)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	h.logger.Info(fmt.Sprintf(constants.LogPaperList, util.GetRole(c), q.Status, page))
	util.OK(c, dto.PageResult{Total: total, Page: page, Size: size, Items: items})
}

// Detail 论文详情。
func (h *PaperHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	paper, err := h.paperSvc.Detail(c.Request.Context(), id)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	h.logger.Info(fmt.Sprintf(constants.LogPaperDetail, id, util.GetRole(c)))
	util.OK(c, paper)
}

// Update 更新论文元信息（作者本人）。
func (h *PaperHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "更新失败：请求参数不合法")
		return
	}
	paper, err := h.paperSvc.Detail(c.Request.Context(), id)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if paper.SubmitterID != util.GetUserID(c) && util.GetRole(c) != constants.RoleAdmin {
		util.Fail(c, http.StatusForbidden, constants.ErrPermissionDenied,
			fmt.Sprintf("更新失败：论文 id=%d 仅作者本人可修改", id))
		return
	}
	paper, err = h.paperSvc.Update(c.Request.Context(), id, req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionUpdatePaper, "paper", fmt.Sprint(id), "更新论文元信息", c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit update paper failed", "error", e)
	}
	util.OK(c, paper)
}

// InitialReview 编辑部初审（编辑/管理员）。
func (h *PaperHandler) InitialReview(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.InitialReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "初审失败：请求参数不合法")
		return
	}
	paper, err := h.paperSvc.InitialReview(c.Request.Context(), util.GetUserID(c), id, req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionInitialReview, "paper", fmt.Sprint(id), fmt.Sprintf("初审结果：pass=%v", req.Pass), c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit initial review failed", "error", e)
	}
	util.OK(c, paper)
}

// FinalDecision 终审决定（编辑/管理员）。
func (h *PaperHandler) FinalDecision(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.FinalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "终审失败：请求参数不合法")
		return
	}
	paper, err := h.paperSvc.FinalDecision(c.Request.Context(), util.GetUserID(c), id, req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionFinalDecision, "paper", fmt.Sprint(id), "终审决定："+req.Decision, c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit final decision failed", "error", e)
	}
	util.OK(c, paper)
}

// Revise 修稿重投（作者本人）。
func (h *PaperHandler) Revise(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ReviseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "修稿失败：请求参数不合法")
		return
	}
	paper, err := h.paperSvc.Detail(c.Request.Context(), id)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if paper.SubmitterID != util.GetUserID(c) && util.GetRole(c) != constants.RoleAdmin {
		util.Fail(c, http.StatusForbidden, constants.ErrPermissionDenied,
			fmt.Sprintf("修稿失败：论文 id=%d 仅作者本人可修改", id))
		return
	}
	paper, err = h.paperSvc.Revise(c.Request.Context(), util.GetUserID(c), id, req)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	if e := h.auditSvc.Record(c.Request.Context(), util.GetUserID(c), util.GetUsername(c),
		constants.AuditActionRevisePaper, "paper", fmt.Sprint(id), "提交修稿 V"+strconv.Itoa(paper.Version), c.ClientIP(), util.GetRequestID(c)); e != nil {
		h.logger.Error("audit revise failed", "error", e)
	}
	util.OK(c, paper)
}

// LibrarySearch 论文库检索。
func (h *PaperHandler) LibrarySearch(c *gin.Context) {
	var q dto.PaperQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "检索失败：分页参数不合法")
		return
	}
	page, size := pageSize(q.Page, q.Size)
	items, total, err := h.paperSvc.SearchLibrary(c.Request.Context(), q.Keyword, q.Subject, page, size)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, dto.PageResult{Total: total, Page: page, Size: size, Items: items})
}

// parseID 解析路径 id 参数。
func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "请求失败：路径参数 id 不合法")
		return 0, false
	}
	return uint(id), true
}

// pageSize 规范化分页参数。
func pageSize(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

// wrapError 论文处理器错误包装。
func (h *PaperHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}
