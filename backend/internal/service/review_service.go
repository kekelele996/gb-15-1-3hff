package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// ReviewService 同行评审服务。
type ReviewService struct {
	store  repository.Store
	logger *slog.Logger
}

// NewReviewService 构造评审服务。
func NewReviewService(store repository.Store, logger *slog.Logger) *ReviewService {
	return &ReviewService{store: store, logger: logger}
}

// sweepExpired 惰性清理：把超过截止日期的开放任务（invited/accepted）置为 expired 失效。
func (s *ReviewService) sweepExpired(ctx context.Context, repo repository.ReviewRepository) {
	affected, err := repo.ExpireOverdue(ctx, timeNow())
	if err != nil {
		s.logger.Error("review expire sweep failed", "error", err)
		return
	}
	if affected > 0 {
		s.logger.Info(fmt.Sprintf(constants.LogReviewExpireSweep, affected))
	}
}

// ListMine 当前审稿人的审稿任务列表。
func (s *ReviewService) ListMine(ctx context.Context, reviewerID uint, status string, page, size int) ([]model.Review, int64, error) {
	s.sweepExpired(ctx, s.store.ReviewRepository())
	items, total, err := s.store.ReviewRepository().ListByReviewer(ctx, reviewerID, status, page, size)
	if err != nil {
		return nil, 0, util.NewAppError(constants.ErrInternal, "审稿任务列表获取失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviewList, reviewerID, status))
	return items, total, nil
}

// ListByPaper 论文的审稿记录列表（含历史轮次，旧意见只读保留）。
func (s *ReviewService) ListByPaper(ctx context.Context, paperID uint) ([]model.Review, error) {
	s.sweepExpired(ctx, s.store.ReviewRepository())
	items, err := s.store.ReviewRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "审稿列表获取失败：系统内部错误", err)
	}
	return items, nil
}

// SummarizeReviews 汇总当前轮次审稿进度（终审门禁与终审页人数统计共用）。
func SummarizeReviews(paperID uint, reviews []model.Review) dto.ReviewSummaryResponse {
	summary := dto.ReviewSummaryResponse{PaperID: paperID, Round: 0}
	for _, r := range reviews {
		if r.Round > summary.Round {
			summary.Round = r.Round
		}
	}
	for _, r := range reviews {
		if r.Round != summary.Round {
			continue
		}
		switch r.Status {
		case constants.ReviewStatusCompleted:
			summary.Completed++
		case constants.ReviewStatusInvited:
			summary.Pending++
		case constants.ReviewStatusAccepted:
			summary.InProgress++
		case constants.ReviewStatusDeclined:
			summary.Declined++
		case constants.ReviewStatusExpired:
			summary.Expired++
		}
	}
	summary.CanFinalize, summary.BlockReason = finalizeGate(&summary)
	return summary
}

// finalizeGate 终审门禁：有效评审记录至少两位完成且没有待处理邀请。
func finalizeGate(summary *dto.ReviewSummaryResponse) (bool, string) {
	if summary.Completed < 2 {
		return false, fmt.Sprintf("当前第 %d 轮有效评审记录仅 %d 位完成（至少 2 位）", summary.Round, summary.Completed)
	}
	if summary.Pending > 0 {
		return false, fmt.Sprintf("当前第 %d 轮仍有 %d 条待处理邀请未回应", summary.Round, summary.Pending)
	}
	return true, ""
}

// Summary 论文当前轮次审稿进度汇总。
func (s *ReviewService) Summary(ctx context.Context, paperID uint) (*dto.ReviewSummaryResponse, error) {
	s.sweepExpired(ctx, s.store.ReviewRepository())
	items, err := s.store.ReviewRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "审稿进度汇总失败：系统内部错误", err)
	}
	summary := SummarizeReviews(paperID, items)
	s.logger.Info(fmt.Sprintf(constants.LogReviewSummary, paperID, summary.Round))
	return &summary, nil
}

// Assign 编辑补邀审稿人（与初审分配共用 ReviewRepository.Create）。
// 同一人同一轮不能同时存在两条有效任务；被拒绝/已超期的旧记录失效后可再次邀请。
func (s *ReviewService) Assign(ctx context.Context, paperID, reviewerID uint) (*model.Review, error) {
	var created *model.Review
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		if _, err := tx.PaperRepository().FindByIDForUpdate(ctx, paperID); err != nil {
			return err
		}
		s.sweepExpired(ctx, tx.ReviewRepository())
		reviewer, err := tx.UserRepository().FindByID(ctx, reviewerID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.ErrUserNotFound,
					fmt.Sprintf("分配审稿人失败：审稿人 id=%d 不存在", reviewerID), nil)
			}
			return util.NewAppError(constants.ErrInternal, "分配审稿人失败：查询审稿人时系统内部错误", err)
		}
		if reviewer.Role != constants.RoleReviewer {
			return util.NewAppError(constants.ErrRoleNotAllowed,
				fmt.Sprintf("分配审稿人失败：用户 %s 角色为 %s，不是审稿人",
					reviewer.Username, util.FormatRole(reviewer.Role)), nil)
		}
		round, err := tx.ReviewRepository().MaxRoundByPaper(ctx, paperID)
		if err != nil {
			return util.NewAppError(constants.ErrInternal, "分配审稿人失败：查询审稿轮次时系统内部错误", err)
		}
		if round == 0 {
			round = 1
		}
		if _, err := tx.ReviewRepository().FindActiveByPaperReviewer(ctx, paperID, reviewerID, round); err == nil {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("分配审稿人失败：审稿人 %s 在论文 id=%d 第 %d 轮已存在有效审稿任务，同一人不能同时存在两条有效任务",
					reviewer.Username, paperID, round), nil)
		} else if !errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.ErrInternal, "分配审稿人失败：查询已有邀请时系统内部错误", err)
		}
		due := timeNow().AddDate(0, 0, 14)
		invite := &model.Review{
			PaperID:    paperID,
			ReviewerID: reviewerID,
			Status:     constants.ReviewStatusInvited,
			Round:      round,
			DueDate:    &due,
		}
		if err := tx.ReviewRepository().Create(ctx, invite); err != nil {
			return err
		}
		created = invite
		s.logger.Info(fmt.Sprintf(constants.LogAssignReviewer, paperID, reviewerID))
		s.logger.Info(fmt.Sprintf(constants.LogReviewReinvite, paperID, reviewerID, round))
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, util.NewAppError(constants.ErrInternal, "分配审稿人失败：系统内部错误", err)
	}
	return created, nil
}

// Respond 审稿人接受/拒绝邀请（已超期失效的旧任务不得再回应）。
func (s *ReviewService) Respond(ctx context.Context, reviewID, reviewerID uint, accept bool) error {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		review, err := tx.ReviewRepository().FindByIDForUpdate(ctx, reviewID)
		if err != nil {
			return err
		}
		if review.ReviewerID != reviewerID {
			return util.NewAppError(constants.ErrPermissionDenied,
				fmt.Sprintf("审稿回应失败：审稿 id=%d 不属于当前审稿人 id=%d", reviewID, reviewerID), nil)
		}
		if review.Status != constants.ReviewStatusInvited {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("审稿回应失败：审稿 id=%d 当前状态 %s 不可回应",
					reviewID, util.FormatReviewStatus(review.Status)), nil)
		}
		if review.DueDate != nil && review.DueDate.Before(timeNow()) {
			review.Status = constants.ReviewStatusExpired
			if err := tx.ReviewRepository().Update(ctx, review); err != nil {
				return err
			}
			return util.NewAppError(constants.ErrReviewExpired,
				fmt.Sprintf("审稿回应失败：审稿 id=%d 已超过截止日期 %s，旧任务失效不得再回应",
					reviewID, util.FormatTime(review.DueDate)), nil)
		}
		if accept {
			review.Status = constants.ReviewStatusAccepted
		} else {
			review.Status = constants.ReviewStatusDeclined
		}
		if err := tx.ReviewRepository().Update(ctx, review); err != nil {
			return err
		}
		if accept {
			paper, err := tx.PaperRepository().FindByIDForUpdate(ctx, review.PaperID)
			if err != nil {
				return err
			}
			if paper.Status == constants.PaperStatusInitialReview {
				paper.Status = constants.PaperStatusExternalReview
				if err := tx.PaperRepository().Update(ctx, paper); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return util.NewAppError(constants.ErrInternal, "审稿回应失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviewRespond, reviewID, reviewerID, accept))
	return nil
}

// Submit 提交评审意见并推进论文状态（已超期失效的旧任务不得再提交）。
func (s *ReviewService) Submit(ctx context.Context, reviewID, reviewerID uint, req dto.SubmitReviewRequest) error {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		review, err := tx.ReviewRepository().FindByIDForUpdate(ctx, reviewID)
		if err != nil {
			return err
		}
		if review.ReviewerID != reviewerID {
			return util.NewAppError(constants.ErrPermissionDenied,
				fmt.Sprintf("提交审稿失败：审稿 id=%d 不属于当前审稿人 id=%d", reviewID, reviewerID), nil)
		}
		if review.Status != constants.ReviewStatusAccepted {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("提交审稿失败：审稿 id=%d 当前状态 %s 不可提交",
					reviewID, util.FormatReviewStatus(review.Status)), nil)
		}
		if review.DueDate != nil && review.DueDate.Before(timeNow()) {
			review.Status = constants.ReviewStatusExpired
			if err := tx.ReviewRepository().Update(ctx, review); err != nil {
				return err
			}
			return util.NewAppError(constants.ErrReviewExpired,
				fmt.Sprintf("提交审稿失败：审稿 id=%d 已超过截止日期 %s，旧任务失效不得再提交",
					reviewID, util.FormatTime(review.DueDate)), nil)
		}
		review.Decision = req.Decision
		review.Comments = req.Comments
		review.ConfidentialComments = req.ConfidentialComments
		review.Status = constants.ReviewStatusCompleted
		now := timeNow()
		review.CompletedAt = &now
		if err := tx.ReviewRepository().Update(ctx, review); err != nil {
			return err
		}
		paper, err := tx.PaperRepository().FindByIDForUpdate(ctx, review.PaperID)
		if err != nil {
			return err
		}
		switch req.Decision {
		case constants.ReviewDecisionAccept, constants.ReviewDecisionReject:
			paper.Status = constants.PaperStatusExternalReview
		case constants.ReviewDecisionMinorRevision, constants.ReviewDecisionMajorRevision:
			paper.Status = constants.PaperStatusRevision
		}
		if err := tx.PaperRepository().Update(ctx, paper); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return appErr
		}
		return util.NewAppError(constants.ErrInternal, "提交审稿失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviewSubmit, reviewID, req.Decision))
	return nil
}

// timeNow 便于测试替换。
var timeNow = time.Now
