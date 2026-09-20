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

// ListMine 当前审稿人的审稿任务列表。
func (s *ReviewService) ListMine(ctx context.Context, reviewerID uint, status string, page, size int) ([]model.Review, int64, error) {
	expireOverdueReviews(ctx, s.store.ReviewRepository(), s.logger)
	items, total, err := s.store.ReviewRepository().ListByReviewer(ctx, reviewerID, status, page, size)
	if err != nil {
		return nil, 0, util.NewAppError(constants.ErrInternal, "审稿任务列表获取失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviewList, reviewerID, status))
	return items, total, nil
}

// ListByPaper 论文的审稿记录列表。
func (s *ReviewService) ListByPaper(ctx context.Context, paperID uint) ([]model.Review, error) {
	expireOverdueReviews(ctx, s.store.ReviewRepository(), s.logger)
	items, err := s.store.ReviewRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "审稿列表获取失败：系统内部错误", err)
	}
	return items, nil
}

// Summarize 论文外审进度汇总：各状态人数统计与终审门槛判定（终审页展示）。
func (s *ReviewService) Summarize(ctx context.Context, paperID uint) (*dto.ReviewSummary, error) {
	expireOverdueReviews(ctx, s.store.ReviewRepository(), s.logger)
	reviews, err := s.store.ReviewRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "外审进度汇总失败：系统内部错误", err)
	}
	paper, err := s.store.PaperRepository().FindByID(ctx, paperID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrPaperNotFound,
				fmt.Sprintf("外审进度汇总失败：论文 id=%d 不存在", paperID), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "外审进度汇总失败：系统内部错误", err)
	}
	summary := buildReviewSummary(paperID, paper.Version, reviews)
	s.logger.Info(fmt.Sprintf(constants.LogReviewSummary, paperID, summary.CanFinalize))
	return summary, nil
}

// Assign 编辑追加分配审稿人（与初审分配共用 ReviewRepository.Create）。
// 同一审稿人存在有效任务（待接受/审稿中）时禁止重复邀请；已拒绝/已超期的旧记录失效，可补邀。
func (s *ReviewService) Assign(ctx context.Context, paperID, reviewerID uint) (*model.Review, error) {
	var invite *model.Review
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		expireOverdueReviews(ctx, tx.ReviewRepository(), s.logger)
		paper, err := tx.PaperRepository().FindByIDForUpdate(ctx, paperID)
		if err != nil {
			return err
		}
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
		if old, err := tx.ReviewRepository().FindActiveByPaperReviewer(ctx, paperID, reviewerID); err == nil {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("分配审稿人失败：审稿人 %s 在论文 id=%d 上已存在有效审稿任务 id=%d（状态 %s），同一审稿人不能同时存在两条有效任务",
					reviewer.Username, paperID, old.ID, util.FormatReviewStatus(old.Status)), nil)
		} else if !errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.ErrInternal, "分配审稿人失败：查询已有邀请时系统内部错误", err)
		}
		due := timeNow().AddDate(0, 0, 14)
		invite = &model.Review{
			PaperID:    paperID,
			ReviewerID: reviewerID,
			Status:     constants.ReviewStatusInvited,
			Round:      paper.Version,
			DueDate:    &due,
		}
		if err := tx.ReviewRepository().Create(ctx, invite); err != nil {
			return err
		}
		s.logger.Info(fmt.Sprintf(constants.LogAssignReviewer, paperID, reviewerID))
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, util.NewAppError(constants.ErrInternal, "分配审稿人失败：系统内部错误", err)
	}
	return invite, nil
}

// Respond 审稿人接受/拒绝邀请。已超期的旧任务失效，不得再回应。
func (s *ReviewService) Respond(ctx context.Context, reviewID, reviewerID uint, accept bool) error {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		expireOverdueReviews(ctx, tx.ReviewRepository(), s.logger)
		review, err := tx.ReviewRepository().FindByIDForUpdate(ctx, reviewID)
		if err != nil {
			return err
		}
		if review.ReviewerID != reviewerID {
			return util.NewAppError(constants.ErrPermissionDenied,
				fmt.Sprintf("审稿回应失败：审稿 id=%d 不属于当前审稿人 id=%d", reviewID, reviewerID), nil)
		}
		if review.Status == constants.ReviewStatusExpired {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("审稿回应失败：审稿 id=%d 已超过截止日期，旧任务失效不得再回应", reviewID), nil)
		}
		if review.Status != constants.ReviewStatusInvited {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("审稿回应失败：审稿 id=%d 当前状态 %s 不可回应",
					reviewID, util.FormatReviewStatus(review.Status)), nil)
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

// Submit 提交评审意见并推进论文状态。已超期的旧任务失效，不得再提交。
func (s *ReviewService) Submit(ctx context.Context, reviewID, reviewerID uint, req dto.SubmitReviewRequest) error {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		expireOverdueReviews(ctx, tx.ReviewRepository(), s.logger)
		review, err := tx.ReviewRepository().FindByIDForUpdate(ctx, reviewID)
		if err != nil {
			return err
		}
		if review.ReviewerID != reviewerID {
			return util.NewAppError(constants.ErrPermissionDenied,
				fmt.Sprintf("提交审稿失败：审稿 id=%d 不属于当前审稿人 id=%d", reviewID, reviewerID), nil)
		}
		if review.Status == constants.ReviewStatusExpired {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("提交审稿失败：审稿 id=%d 已超过截止日期，旧任务失效不得再提交", reviewID), nil)
		}
		if review.Status != constants.ReviewStatusAccepted {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("提交审稿失败：审稿 id=%d 当前状态 %s 不可提交",
					reviewID, util.FormatReviewStatus(review.Status)), nil)
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
		if paper.Status != constants.PaperStatusExternalReview && paper.Status != constants.PaperStatusInitialReview {
			return util.NewAppError(constants.ErrPaperStatusNotAllowed,
				fmt.Sprintf("提交审稿失败：论文 id=%d 当前状态 %s 不允许提交评审意见",
					paper.ID, util.FormatPaperStatus(paper.Status)), nil)
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

// expireOverdueReviews 惰性过期清扫：超过截止日期的待接受/审稿中任务置为已超期（旧记录失效）。
// 在审稿任务读写路径上调用，保证过期任务不可再回应且统计口径一致。
func expireOverdueReviews(ctx context.Context, repo repository.ReviewRepository, logger *slog.Logger) {
	expired, err := repo.ExpireOverdue(ctx, timeNow())
	if err != nil {
		logger.Error("review expire sweep failed", "error", err)
		return
	}
	if expired > 0 {
		logger.Info(fmt.Sprintf(constants.LogReviewExpireSweep, expired))
	}
}

// buildReviewSummary 汇总一篇论文的外审进度并判定终审门槛。
// 终审条件：完成评审的有效审稿人不少于 FinalDecisionMinCompletedReviewers 人，
// 且不存在待接受邀请与审稿中任务；已拒绝/已超期的旧记录不参与计数。
func buildReviewSummary(paperID uint, round int, reviews []model.Review) *dto.ReviewSummary {
	summary := &dto.ReviewSummary{PaperID: paperID, Round: round, BlockReasons: []string{}}
	completedReviewers := map[uint]bool{}
	for _, r := range reviews {
		switch r.Status {
		case constants.ReviewStatusCompleted:
			summary.Completed++
			completedReviewers[r.ReviewerID] = true
		case constants.ReviewStatusInvited:
			summary.Invited++
		case constants.ReviewStatusAccepted:
			summary.Accepted++
		case constants.ReviewStatusDeclined:
			summary.Declined++
		case constants.ReviewStatusExpired:
			summary.Expired++
		}
	}
	summary.CompletedReviewers = len(completedReviewers)
	if summary.CompletedReviewers < constants.FinalDecisionMinCompletedReviewers {
		summary.BlockReasons = append(summary.BlockReasons,
			fmt.Sprintf("完成评审的有效审稿人不足 %d 人（当前 %d 人）",
				constants.FinalDecisionMinCompletedReviewers, summary.CompletedReviewers))
	}
	if summary.Invited > 0 {
		summary.BlockReasons = append(summary.BlockReasons,
			fmt.Sprintf("仍有 %d 条待接受邀请未处理", summary.Invited))
	}
	if summary.Accepted > 0 {
		summary.BlockReasons = append(summary.BlockReasons,
			fmt.Sprintf("仍有 %d 位审稿人审稿中（已接受未提交）", summary.Accepted))
	}
	summary.CanFinalize = len(summary.BlockReasons) == 0
	return summary
}
