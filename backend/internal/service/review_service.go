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
	items, total, err := s.store.ReviewRepository().ListByReviewer(ctx, reviewerID, status, page, size)
	if err != nil {
		return nil, 0, util.NewAppError(constants.ErrInternal, "审稿任务列表获取失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviewList, reviewerID, status))
	return items, total, nil
}

// ListByPaper 论文的审稿记录列表。
func (s *ReviewService) ListByPaper(ctx context.Context, paperID uint) ([]model.Review, error) {
	items, err := s.store.ReviewRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "审稿列表获取失败：系统内部错误", err)
	}
	return items, nil
}

// Assign 编辑追加分配审稿人（与初审分配共用 ReviewRepository.Create）。
func (s *ReviewService) Assign(ctx context.Context, paperID, reviewerID uint) (*model.Review, error) {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		if _, err := tx.PaperRepository().FindByIDForUpdate(ctx, paperID); err != nil {
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
		if _, err := tx.ReviewRepository().FindInviteByPaperReviewer(ctx, paperID, reviewerID); err == nil {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("分配审稿人失败：论文 id=%d 已邀请审稿人 id=%d", paperID, reviewerID), nil)
		} else if !errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.ErrInternal, "分配审稿人失败：查询已有邀请时系统内部错误", err)
		}
		due := timeNow().AddDate(0, 0, 14)
		invite := &model.Review{
			PaperID:    paperID,
			ReviewerID: reviewerID,
			Status:     constants.ReviewStatusInvited,
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
	return s.store.ReviewRepository().FindInviteByPaperReviewer(ctx, paperID, reviewerID)
}

// Respond 审稿人接受/拒绝邀请。
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

// Submit 提交评审意见并推进论文状态。
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
