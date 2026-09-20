package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// StatisticsService 期刊编辑统计面板服务。
type StatisticsService struct {
	store  repository.Store
	logger *slog.Logger
}

// NewStatisticsService 构造统计服务。
func NewStatisticsService(store repository.Store, logger *slog.Logger) *StatisticsService {
	return &StatisticsService{store: store, logger: logger}
}

// Overview 统计概览（复用 PaperRepository.CountByStatus 与 ReviewRepository.AvgDurationDays）。
func (s *StatisticsService) Overview(ctx context.Context) (*dto.OverviewResponse, error) {
	statusCount, err := s.store.PaperRepository().CountByStatus(ctx)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "统计失败：获取投稿状态分布时系统内部错误", err)
	}
	avgDays, err := s.store.ReviewRepository().AvgDurationDays(ctx)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "统计失败：计算平均审稿周期时系统内部错误", err)
	}
	o := &dto.OverviewResponse{AvgReviewDays: avgDays}
	for k, v := range statusCount {
		switch k {
		case constants.PaperStatusSubmitted:
			o.Submitted = v
		case constants.PaperStatusInitialReview:
			o.InitialReview = v
		case constants.PaperStatusExternalReview:
			o.ExternalReview = v
		case constants.PaperStatusRevision:
			o.Revision = v
		case constants.PaperStatusAccepted:
			o.Accepted = v
		case constants.PaperStatusRejected:
			o.Rejected = v
		}
		o.Total += v
	}
	if o.Total > 0 {
		o.AcceptanceRate = float64(o.Accepted) / float64(o.Total) * 100
	}
	s.logger.Info(constants.LogStatsOverview)
	return o, nil
}

// Trend 投稿量趋势（按天聚合）。
func (s *StatisticsService) Trend(ctx context.Context, days int) ([]model.DayCount, error) {
	if days <= 0 {
		days = 30
	}
	rows, err := s.store.PaperRepository().CountCreatedByDay(ctx, days)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "统计失败：获取投稿趋势时系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogStatsTrend, days))
	return rows, nil
}

// Subjects 学科投稿分布。
func (s *StatisticsService) Subjects(ctx context.Context) ([]model.SubjectCount, error) {
	rows, err := s.store.PaperRepository().CountBySubject(ctx)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "统计失败：获取学科分布时系统内部错误", err)
	}
	s.logger.Info(constants.LogStatsSubjects)
	return rows, nil
}

// ReviewerWorkload 审稿人工作量排名。
func (s *StatisticsService) ReviewerWorkload(ctx context.Context) ([]model.ReviewerLoad, error) {
	rows, err := s.store.ReviewRepository().CountCompletedByReviewer(ctx)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "统计失败：获取审稿人工作量时系统内部错误", err)
	}
	s.logger.Info(constants.LogStatsReviewers)
	return rows, nil
}
