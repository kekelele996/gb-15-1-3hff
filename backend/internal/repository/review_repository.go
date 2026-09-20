package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReviewRepository 审稿仓储接口。
type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	Update(ctx context.Context, review *model.Review) error
	FindByID(ctx context.Context, id uint) (*model.Review, error)
	FindByIDForUpdate(ctx context.Context, id uint) (*model.Review, error)
	ListByReviewer(ctx context.Context, reviewerID uint, status string, page, size int) ([]model.Review, int64, error)
	ListByPaper(ctx context.Context, paperID uint) ([]model.Review, error)
	FindInviteByPaperReviewer(ctx context.Context, paperID, reviewerID uint) (*model.Review, error)
	CountCompletedByReviewer(ctx context.Context) ([]model.ReviewerLoad, error)
	CountCompleted(ctx context.Context) (int64, error)
	AvgDurationDays(ctx context.Context) (float64, error)
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 构造审稿仓储。
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *model.Review) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *reviewRepository) Update(ctx context.Context, review *model.Review) error {
	if err := r.db.WithContext(ctx).Omit("Paper", "Reviewer").Save(review).Error; err != nil {
		return fmt.Errorf("update review: %w", err)
	}
	return nil
}

func (r *reviewRepository) FindByID(ctx context.Context, id uint) (*model.Review, error) {
	var v model.Review
	if err := r.db.WithContext(ctx).First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find review %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find review %d: %w", id, err)
	}
	return &v, nil
}

func (r *reviewRepository) FindByIDForUpdate(ctx context.Context, id uint) (*model.Review, error) {
	var v model.Review
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("lock review %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("lock review %d: %w", id, err)
	}
	return &v, nil
}

func (r *reviewRepository) ListByReviewer(ctx context.Context, reviewerID uint, status string, page, size int) ([]model.Review, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Review{}).Where("reviewer_id = ?", reviewerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}
	var items []model.Review
	if err := q.Preload("Paper").Preload("Reviewer").
		Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list reviews: %w", err)
	}
	return items, total, nil
}

func (r *reviewRepository) ListByPaper(ctx context.Context, paperID uint) ([]model.Review, error) {
	var items []model.Review
	if err := r.db.WithContext(ctx).Preload("Reviewer").
		Where("paper_id = ?", paperID).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list reviews by paper %d: %w", paperID, err)
	}
	return items, nil
}

func (r *reviewRepository) FindInviteByPaperReviewer(ctx context.Context, paperID, reviewerID uint) (*model.Review, error) {
	var v model.Review
	if err := r.db.WithContext(ctx).Where("paper_id = ? AND reviewer_id = ?", paperID, reviewerID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find invite paper %d reviewer %d: %w", paperID, reviewerID, ErrNotFound)
		}
		return nil, fmt.Errorf("find invite paper %d reviewer %d: %w", paperID, reviewerID, err)
	}
	return &v, nil
}

func (r *reviewRepository) CountCompletedByReviewer(ctx context.Context) ([]model.ReviewerLoad, error) {
	var rows []model.ReviewerLoad
	completed := constants.ReviewStatusCompleted
	if err := r.db.WithContext(ctx).Model(&model.Review{}).
		Select("reviewer_id, u.real_name as reviewer_name, count(*) as total, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as completed", completed).
		Joins("JOIN users u ON u.id = reviews.reviewer_id").
		Group("reviewer_id, u.real_name").Order("total DESC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("count completed reviews by reviewer: %w", err)
	}
	return rows, nil
}

func (r *reviewRepository) CountCompleted(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&model.Review{}).
		Where("status = ?", constants.ReviewStatusCompleted).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count completed reviews: %w", err)
	}
	return n, nil
}

func (r *reviewRepository) AvgDurationDays(ctx context.Context) (float64, error) {
	var avg float64
	if err := r.db.WithContext(ctx).Model(&model.Review{}).
		Where("status = ?", constants.ReviewStatusCompleted).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0), 0)").
		Scan(&avg).Error; err != nil {
		return 0, fmt.Errorf("avg review duration: %w", err)
	}
	return avg, nil
}
