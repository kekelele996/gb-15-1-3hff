package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PaperRepository 论文仓储接口。
type PaperRepository interface {
	Create(ctx context.Context, paper *model.Paper) error
	Update(ctx context.Context, paper *model.Paper) error
	FindByID(ctx context.Context, id uint) (*model.Paper, error)
	FindByIDForUpdate(ctx context.Context, id uint) (*model.Paper, error)
	FindByIDWithDetail(ctx context.Context, id uint) (*model.Paper, error)
	List(ctx context.Context, filter model.PaperFilter, page, size int) ([]model.Paper, int64, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
	CountBySubject(ctx context.Context) ([]model.SubjectCount, error)
	CountCreatedByDay(ctx context.Context, days int) ([]model.DayCount, error)
	CountAll(ctx context.Context) (int64, error)
}

type paperRepository struct {
	db *gorm.DB
}

// NewPaperRepository 构造论文仓储。
func NewPaperRepository(db *gorm.DB) PaperRepository {
	return &paperRepository{db: db}
}

func (r *paperRepository) Create(ctx context.Context, paper *model.Paper) error {
	if err := r.db.WithContext(ctx).Create(paper).Error; err != nil {
		return fmt.Errorf("create paper: %w", err)
	}
	return nil
}

func (r *paperRepository) Update(ctx context.Context, paper *model.Paper) error {
	if err := r.db.WithContext(ctx).Omit("Submitter", "Reviews", "Revisions").Save(paper).Error; err != nil {
		return fmt.Errorf("update paper: %w", err)
	}
	return nil
}

func (r *paperRepository) FindByID(ctx context.Context, id uint) (*model.Paper, error) {
	var p model.Paper
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find paper %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find paper %d: %w", id, err)
	}
	return &p, nil
}

func (r *paperRepository) FindByIDForUpdate(ctx context.Context, id uint) (*model.Paper, error) {
	var p model.Paper
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("lock paper %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("lock paper %d: %w", id, err)
	}
	return &p, nil
}

func (r *paperRepository) FindByIDWithDetail(ctx context.Context, id uint) (*model.Paper, error) {
	var p model.Paper
	if err := r.db.WithContext(ctx).
		Preload("Submitter").
		Preload("Reviews", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Reviewer")
		}).
		Preload("Revisions", func(db *gorm.DB) *gorm.DB {
			return db.Preload("SubmittedBy")
		}).
		First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find paper detail %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find paper detail %d: %w", id, err)
	}
	return &p, nil
}

func (r *paperRepository) List(ctx context.Context, filter model.PaperFilter, page, size int) ([]model.Paper, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Paper{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.SubmitterID > 0 {
		q = q.Where("submitter_id = ?", filter.SubmitterID)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		q = q.Where("title ILIKE ? OR abstract ILIKE ? OR keywords ILIKE ?", like, like, like)
	}
	if filter.Subject != "" {
		q = q.Where("subject = ?", filter.Subject)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count papers: %w", err)
	}
	var items []model.Paper
	if err := q.Preload("Submitter").Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list papers: %w", err)
	}
	return items, total, nil
}

func (r *paperRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).Model(&model.Paper{}).
		Select("status, count(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("count papers by status: %w", err)
	}
	result := make(map[string]int64, len(rows))
	for _, rw := range rows {
		result[rw.Status] = rw.Count
	}
	return result, nil
}

func (r *paperRepository) CountBySubject(ctx context.Context) ([]model.SubjectCount, error) {
	var rows []model.SubjectCount
	if err := r.db.WithContext(ctx).Model(&model.Paper{}).
		Select("subject, count(*) as count").Group("subject").Order("count DESC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("count papers by subject: %w", err)
	}
	return rows, nil
}

func (r *paperRepository) CountCreatedByDay(ctx context.Context, days int) ([]model.DayCount, error) {
	var rows []model.DayCount
	start := time.Now().AddDate(0, 0, -(days - 1))
	if err := r.db.WithContext(ctx).Model(&model.Paper{}).
		Select("to_char(created_at, 'YYYY-MM-DD') as day, count(*) as count").
		Where("created_at >= ?", start).
		Group("day").Order("day").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("count papers by day: %w", err)
	}
	return rows, nil
}

func (r *paperRepository) CountAll(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&model.Paper{}).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count all papers: %w", err)
	}
	return n, nil
}
