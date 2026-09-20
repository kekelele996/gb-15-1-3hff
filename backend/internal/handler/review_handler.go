package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
)

// ReviewHandler 审稿处理器。
type ReviewHandler struct {
	reviewSvc *service.ReviewService
	logger    *slog.Logger
}

// NewReviewHandler 构造审稿处理器。
func NewReviewHandler(reviewSvc *service.ReviewService, logger *slog.Logger) *ReviewHandler {
	return &ReviewHandler{reviewSvc: reviewSvc, logger: logger}
}

// ListMine 我的审稿任务列表（审稿人）。
func (h *ReviewHandler) ListMine(c *gin.Context) {
	var q dto.ReviewQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "查询失败：分页参数不合法")
		return
	}
	page, size := pageSize(q.Page, q.Size)
	items, total, err := h.reviewSvc.ListMine(c.Request.Context(), util.GetUserID(c), q.Status, page, size)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, dto.PageResult{Total: total, Page: page, Size: size, Items: items})
}

// ListByPaper 论文审稿记录列表。
func (h *ReviewHandler) ListByPaper(c *gin.Context) {
	paperID, err := strconv.ParseUint(c.Param("paperID"), 10, 64)
	if err != nil || paperID == 0 {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "请求失败：路径参数 paperID 不合法")
		return
	}
	items, err := h.reviewSvc.ListByPaper(c.Request.Context(), uint(paperID))
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, items)
}

// Respond 接受/拒绝审稿邀请（审稿人）。
func (h *ReviewHandler) Respond(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.RespondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "回应失败：请求参数不合法")
		return
	}
	if err := h.reviewSvc.Respond(c.Request.Context(), id, util.GetUserID(c), req.Accept); err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, gin.H{"review_id": id, "accepted": req.Accept})
}

// Submit 提交评审意见（审稿人）。
func (h *ReviewHandler) Submit(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "提交失败：请求参数不合法（decision/comments 字段校验未通过）")
		return
	}
	if err := h.reviewSvc.Submit(c.Request.Context(), id, util.GetUserID(c), req); err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, gin.H{"review_id": id, "decision": req.Decision})
}

// Assign 追加分配审稿人（编辑/管理员）。
func (h *ReviewHandler) Assign(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.AssignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, "分配失败：请求参数不合法")
		return
	}
	review, err := h.reviewSvc.Assign(c.Request.Context(), id, req.ReviewerID)
	if err != nil {
		h.wrapError(c, err)
		return
	}
	util.OK(c, gin.H{"review_id": review.ID, "paper_id": id, "reviewer_id": req.ReviewerID})
}

// wrapError 审稿处理器错误包装。
func (h *ReviewHandler) wrapError(c *gin.Context, err error) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		util.Fail(c, httpStatusForCode(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	util.Fail(c, http.StatusInternalServerError, constants.ErrInternal, "系统内部错误："+err.Error())
}

