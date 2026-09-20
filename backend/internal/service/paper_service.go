package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// PaperService 论文服务：投稿/列表/详情/初审/终审/修稿/论文库检索。
type PaperService struct {
	store         repository.Store
	plagiarismSvc *PlagiarismService
	cfg           *config.Config
	logger        *slog.Logger
}

// NewPaperService 构造论文服务。
func NewPaperService(store repository.Store, plagiarismSvc *PlagiarismService, cfg *config.Config, logger *slog.Logger) *PaperService {
	return &PaperService{store: store, plagiarismSvc: plagiarismSvc, cfg: cfg, logger: logger}
}

// Create 创建投稿：事务内写论文 + 待查重记录，随后自动执行查重。
func (s *PaperService) Create(ctx context.Context, submitterID uint, req dto.CreatePaperRequest) (*model.Paper, error) {
	s.logger.Info(fmt.Sprintf(constants.LogPaperCreateStart, submitterID, req.Title))
	paper := &model.Paper{
		Title:       req.Title,
		Abstract:    req.Abstract,
		Keywords:    req.Keywords,
		Subject:     req.Subject,
		AuthorsMeta: req.AuthorsMeta,
		FileName:    req.FileName,
		FileKey:     req.FileKey,
		Status:      constants.PaperStatusSubmitted,
		Version:     1,
		SubmitterID: submitterID,
	}
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		if err := tx.PaperRepository().Create(ctx, paper); err != nil {
			return err
		}
		return tx.PlagiarismRepository().Create(ctx, &model.PlagiarismCheck{
			PaperID: paper.ID,
			Status:  constants.PlagiarismStatusPending,
		})
	})
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal,
			fmt.Sprintf("投稿失败：创建论文 %s 时系统内部错误", req.Title), err)
	}
	if _, err := s.plagiarismSvc.RunCheck(ctx, paper); err != nil {
		return nil, err
	}
	s.logger.Info(fmt.Sprintf(constants.LogPaperCreateOK, paper.ID, paper.Status))
	return paper, nil
}

// list 复用同一仓储 List 方法的统一列表查询（被 ListMine/ListByStatus/SearchLibrary 复用）。
func (s *PaperService) list(ctx context.Context, filter model.PaperFilter, page, size int) ([]model.Paper, int64, error) {
	items, total, err := s.store.PaperRepository().List(ctx, filter, page, size)
	if err != nil {
		return nil, 0, util.NewAppError(constants.ErrInternal, "查询论文列表失败：系统内部错误", err)
	}
	return items, total, nil
}

// ListMine 我的投稿列表。
func (s *PaperService) ListMine(ctx context.Context, submitterID uint, page, size int) ([]model.Paper, int64, error) {
	return s.list(ctx, model.PaperFilter{SubmitterID: submitterID}, page, size)
}

// ListByStatus 按状态列出论文（编辑初审队列；status 为空则列出全部）。
func (s *PaperService) ListByStatus(ctx context.Context, status string, page, size int) ([]model.Paper, int64, error) {
	return s.list(ctx, model.PaperFilter{Status: status}, page, size)
}

// SearchLibrary 论文库检索（仅已录用论文），复用 list。
func (s *PaperService) SearchLibrary(ctx context.Context, keyword, subject string, page, size int) ([]model.Paper, int64, error) {
	return s.list(ctx, model.PaperFilter{Status: constants.PaperStatusAccepted, Keyword: keyword, Subject: subject}, page, size)
}

// Detail 论文详情（含作者/审稿/修稿预加载）。
func (s *PaperService) Detail(ctx context.Context, id uint) (*model.Paper, error) {
	p, err := s.store.PaperRepository().FindByIDWithDetail(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrPaperNotFound,
				fmt.Sprintf("论文详情获取失败：论文 id=%d 不存在", id), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "论文详情获取失败：系统内部错误", err)
	}
	return p, nil
}

// Update 更新论文元信息（仅 submitted/revision 状态允许）。
func (s *PaperService) Update(ctx context.Context, id uint, req dto.UpdatePaperRequest) (*model.Paper, error) {
	paper, err := s.Detail(ctx, id)
	if err != nil {
		return nil, err
	}
	if paper.Status != constants.PaperStatusSubmitted && paper.Status != constants.PaperStatusRevision {
		return nil, util.NewAppError(constants.ErrPaperStatusNotAllowed,
			fmt.Sprintf("更新论文失败：论文 %s 当前状态 %s 不允许修改元信息",
				paper.Title, util.FormatPaperStatus(paper.Status)), nil)
	}
	if req.Title != "" {
		paper.Title = req.Title
	}
	if req.Abstract != "" {
		paper.Abstract = req.Abstract
	}
	if req.Keywords != "" {
		paper.Keywords = req.Keywords
	}
	if req.Subject != "" {
		paper.Subject = req.Subject
	}
	if req.AuthorsMeta != "" {
		paper.AuthorsMeta = req.AuthorsMeta
	}
	if err := s.store.PaperRepository().Update(ctx, paper); err != nil {
		return nil, util.NewAppError(constants.ErrInternal,
			fmt.Sprintf("更新论文失败：论文 id=%d 保存时系统内部错误", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPaperUpdateOK, id))
	return s.Detail(ctx, id)
}

// InitialReview 编辑部初审：通过则分配审稿人，不通过则退回作者。
func (s *PaperService) InitialReview(ctx context.Context, editorID uint, paperID uint, req dto.InitialReviewRequest) (*model.Paper, error) {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		paper, err := tx.PaperRepository().FindByIDForUpdate(ctx, paperID)
		if err != nil {
			return err
		}
		if paper.Status != constants.PaperStatusSubmitted {
			return util.NewAppError(constants.ErrPaperStatusNotAllowed,
				fmt.Sprintf("初审失败：论文 %s 当前状态 %s 不允许初审",
					paper.Title, util.FormatPaperStatus(paper.Status)), nil)
		}
		if !req.Pass {
			paper.Status = constants.PaperStatusRejected
			paper.FinalDecision = constants.ReviewDecisionReject
			paper.InitialReviewComment = req.Reason
			if err := tx.PaperRepository().Update(ctx, paper); err != nil {
				return err
			}
			s.logger.Info(fmt.Sprintf(constants.LogInitialReviewReject, paper.ID, req.Reason))
			return nil
		}
		if req.ReviewerID == 0 {
			return util.NewAppError(constants.ErrBadRequest, "初审失败：通过初审必须指定审稿人 reviewer_id", nil)
		}
		reviewer, err := tx.UserRepository().FindByID(ctx, req.ReviewerID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.ErrUserNotFound,
					fmt.Sprintf("初审失败：审稿人 id=%d 不存在", req.ReviewerID), nil)
			}
			return util.NewAppError(constants.ErrInternal, "初审失败：查询审稿人时系统内部错误", err)
		}
		if reviewer.Role != constants.RoleReviewer {
			return util.NewAppError(constants.ErrRoleNotAllowed,
				fmt.Sprintf("初审失败：用户 %s 角色为 %s，不是审稿人",
					reviewer.Username, util.FormatRole(reviewer.Role)), nil)
		}
		paper.Status = constants.PaperStatusInitialReview
		paper.InitialReviewComment = req.Reason
		if err := tx.PaperRepository().Update(ctx, paper); err != nil {
			return err
		}
		due := timeNow().AddDate(0, 0, 14)
		invite := &model.Review{
			PaperID:    paper.ID,
			ReviewerID: reviewer.ID,
			Status:     constants.ReviewStatusInvited,
			Round:      paper.Version,
			DueDate:    &due,
		}
		if err := tx.ReviewRepository().Create(ctx, invite); err != nil {
			return err
		}
		s.logger.Info(fmt.Sprintf(constants.LogAssignReviewer, paper.ID, reviewer.ID))
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, util.NewAppError(constants.ErrInternal, "初审失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogInitialReviewOK, paperID, req.Pass, req.ReviewerID), "editor_id", editorID)
	return s.Detail(ctx, paperID)
}

// FinalDecision 终审决定：录用或拒稿。
// 终审门槛：完成评审的有效审稿人不少于 2 人，且没有待接受邀请与审稿中任务，否则阻止提交。
func (s *PaperService) FinalDecision(ctx context.Context, editorID uint, paperID uint, req dto.FinalDecisionRequest) (*model.Paper, error) {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		expireOverdueReviews(ctx, tx.ReviewRepository(), s.logger)
		paper, err := tx.PaperRepository().FindByIDForUpdate(ctx, paperID)
		if err != nil {
			return err
		}
		if paper.Status != constants.PaperStatusRevision &&
			paper.Status != constants.PaperStatusExternalReview &&
			paper.Status != constants.PaperStatusInitialReview {
			return util.NewAppError(constants.ErrPaperStatusNotAllowed,
				fmt.Sprintf("终审失败：论文 %s 当前状态 %s 不允许终审",
					paper.Title, util.FormatPaperStatus(paper.Status)), nil)
		}
		reviews, err := tx.ReviewRepository().ListByPaper(ctx, paperID)
		if err != nil {
			return util.NewAppError(constants.ErrInternal, "终审失败：查询审稿记录时系统内部错误", err)
		}
		summary := buildReviewSummary(paperID, paper.Version, reviews)
		if !summary.CanFinalize {
			return util.NewAppError(constants.ErrReviewNotAllowed,
				fmt.Sprintf("终审失败：论文 %s 外审进度不足（%s）",
					paper.Title, strings.Join(summary.BlockReasons, "；")), nil)
		}
		paper.Status = req.Decision
		paper.FinalDecision = req.Decision
		paper.FinalComment = req.Comment
		if err := tx.PaperRepository().Update(ctx, paper); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, util.NewAppError(constants.ErrInternal, "终审失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogFinalDecision, paperID, req.Decision), "editor_id", editorID)
	return s.Detail(ctx, paperID)
}

// Revise 作者修稿重投：事务内写修稿记录并推进论文版本与状态。
// 同时按仍在有效期内的审稿人开启新一轮匿名评审：旧的有效任务（待接受/审稿中）失效，
// 已完成的旧意见只读保留，为上述审稿人创建新一轮邀请；已拒绝/已超期的审稿人不自动续邀，编辑可补邀。
func (s *PaperService) Revise(ctx context.Context, authorID uint, paperID uint, req dto.ReviseRequest) (*model.Paper, error) {
	err := s.store.Transaction(ctx, func(tx repository.Store) error {
		expireOverdueReviews(ctx, tx.ReviewRepository(), s.logger)
		paper, err := tx.PaperRepository().FindByIDForUpdate(ctx, paperID)
		if err != nil {
			return err
		}
		if paper.Status != constants.PaperStatusRevision {
			return util.NewAppError(constants.ErrPaperStatusNotAllowed,
				fmt.Sprintf("修稿失败：论文 %s 当前状态 %s 不允许修稿",
					paper.Title, util.FormatPaperStatus(paper.Status)), nil)
		}
		version := paper.Version + 1
		revision := &model.Revision{
			PaperID:        paperID,
			Version:        version,
			FileName:       req.FileName,
			FileKey:        req.FileKey,
			ResponseLetter: req.ResponseLetter,
			SubmittedByID:  authorID,
		}
		if err := tx.RevisionRepository().Create(ctx, revision); err != nil {
			return err
		}
		paper.Version = version
		paper.FileName = req.FileName
		paper.FileKey = req.FileKey
		paper.Status = constants.PaperStatusExternalReview
		if err := tx.PaperRepository().Update(ctx, paper); err != nil {
			return err
		}
		invites, err := s.openNewReviewRound(ctx, tx, paperID, version)
		if err != nil {
			return err
		}
		s.logger.Info(fmt.Sprintf(constants.LogReviewNewRound, paperID, version, invites))
		return nil
	})
	if err != nil {
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, util.NewAppError(constants.ErrInternal, "修稿失败：系统内部错误", err)
	}
	paper, err := s.Detail(ctx, paperID)
	if err != nil {
		return nil, err
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviseSubmit, paperID, paper.Version))
	return paper, nil
}

// openNewReviewRound 修稿后为仍在有效期内的审稿人开启新一轮匿名评审，返回新建邀请数。
// 每个审稿人取其最新一条任务判定：待接受/审稿中/已完成视为仍在有效期内，为其创建新一轮邀请；
// 待接受/审稿中的旧任务同时失效（同一审稿人不能同时存在两条有效任务）；已拒绝/已超期不续邀。
func (s *PaperService) openNewReviewRound(ctx context.Context, tx repository.Store, paperID uint, version int) (int, error) {
	reviews, err := tx.ReviewRepository().ListByPaper(ctx, paperID)
	if err != nil {
		return 0, util.NewAppError(constants.ErrInternal, "修稿失败：查询审稿记录时系统内部错误", err)
	}
	latest := map[uint]*model.Review{}
	for i := range reviews {
		r := reviews[i]
		if old, ok := latest[r.ReviewerID]; !ok || r.ID > old.ID {
			rr := r
			latest[r.ReviewerID] = &rr
		}
	}
	invites := 0
	for _, r := range latest {
		switch r.Status {
		case constants.ReviewStatusInvited, constants.ReviewStatusAccepted:
			r.Status = constants.ReviewStatusExpired
			if err := tx.ReviewRepository().Update(ctx, r); err != nil {
				return 0, util.NewAppError(constants.ErrInternal, "修稿失败：失效旧审稿任务时系统内部错误", err)
			}
		case constants.ReviewStatusCompleted:
			// 已完成的旧意见只读保留，无需变更。
		default:
			continue
		}
		due := timeNow().AddDate(0, 0, 14)
		invite := &model.Review{
			PaperID:    paperID,
			ReviewerID: r.ReviewerID,
			Status:     constants.ReviewStatusInvited,
			Round:      version,
			DueDate:    &due,
		}
		if err := tx.ReviewRepository().Create(ctx, invite); err != nil {
			return 0, util.NewAppError(constants.ErrInternal, "修稿失败：创建新一轮审稿邀请时系统内部错误", err)
		}
		invites++
	}
	return invites, nil
}
