package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"time"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// PlagiarismService 查重检测服务。
type PlagiarismService struct {
	store  repository.Store
	cfg    *config.Config
	logger *slog.Logger
}

// NewPlagiarismService 构造查重服务。
func NewPlagiarismService(store repository.Store, cfg *config.Config, logger *slog.Logger) *PlagiarismService {
	return &PlagiarismService{store: store, cfg: cfg, logger: logger}
}

// RunCheck 执行查重检测（确定性模拟），超过阈值自动退回论文。
// 该 service 方法同时被「论文创建自动查重」与「编辑手动重跑查重」两个接口复用。
func (s *PlagiarismService) RunCheck(ctx context.Context, paper *model.Paper) (*model.PlagiarismCheck, error) {
	sim := computeSimilarity(paper.Title + "|" + paper.Keywords)
	now := time.Now()
	check, err := s.store.PlagiarismRepository().FindByPaper(ctx, paper.ID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrInternal, "查重检测失败：查询已有查重记录时系统内部错误", err)
		}
		check = &model.PlagiarismCheck{PaperID: paper.ID}
	}
	check.Similarity = sim
	check.Status = constants.PlagiarismStatusCompleted
	check.CheckedAt = &now
	check.Report = buildPlagiarismReport(sim, paper.Keywords)
	paper.Similarity = sim
	if sim > s.cfg.SimilarityThreshold {
		paper.Status = constants.PaperStatusRejected
		paper.FinalDecision = constants.ReviewDecisionReject
		paper.FinalComment = fmt.Sprintf("查重检测未通过：相似度 %s 超过阈值 %s，系统自动退回",
			util.FormatPercent(sim), util.FormatPercent(s.cfg.SimilarityThreshold))
		s.logger.Warn(fmt.Sprintf(constants.LogPlagiarismAutoReject, paper.ID, sim))
	}
	err = s.store.Transaction(ctx, func(tx repository.Store) error {
		if err := tx.PlagiarismRepository().Update(ctx, check); err != nil {
			return err
		}
		return tx.PaperRepository().Update(ctx, paper)
	})
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "查重检测失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPlagiarismRun, paper.ID, sim))
	return check, nil
}

// GetByPaper 获取论文查重结果。
func (s *PlagiarismService) GetByPaper(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error) {
	check, err := s.store.PlagiarismRepository().FindByPaper(ctx, paperID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrPlagiarismNotFound,
				fmt.Sprintf("查重结果获取失败：论文 id=%d 暂无查重记录", paperID), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "查重结果获取失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPlagiarismGet, paperID))
	return check, nil
}

// Rerun 重新执行查重。
func (s *PlagiarismService) Rerun(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error) {
	paper, err := s.store.PaperRepository().FindByID(ctx, paperID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrPaperNotFound,
				fmt.Sprintf("查重重跑失败：论文 id=%d 不存在", paperID), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "查重重跑失败：系统内部错误", err)
	}
	return s.RunCheck(ctx, paper)
}

// computeSimilarity 确定性模拟相似度：同一论文始终得到相同结果。
func computeSimilarity(text string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	return 8 + float64(h.Sum32()%1800)/100.0 // 8.00 ~ 25.99
}

// buildPlagiarismReport 生成重复段落标注 JSON。
func buildPlagiarismReport(sim float64, keywords string) string {
	type reportItem struct {
		Source     string  `json:"source"`
		Paragraph  string  `json:"paragraph"`
		Similarity float64 `json:"similarity"`
	}
	kws := util.SplitKeywords(keywords)
	if len(kws) > 3 {
		kws = kws[:3]
	}
	items := make([]reportItem, 0, len(kws))
	for _, kw := range kws {
		items = append(items, reportItem{
			Source:     "公开文献库",
			Paragraph:  fmt.Sprintf("摘要中与「%s」相关的表述在公开文献中存在相似表达", kw),
			Similarity: sim / float64(len(kws)),
		})
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(b)
}
