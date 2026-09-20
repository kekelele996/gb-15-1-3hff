package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
)

// PlagiarismRepository 查重仓储接口。
type PlagiarismRepository interface {
	Create(ctx context.Context, check *model.PlagiarismCheck) error
	Update(ctx context.Context, check *model.PlagiarismCheck) error
	FindByPaper(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error)
}

type plagiarismRepository struct {
	db *gorm.DB
}

// NewPlagiarismRepository 构造查重仓储。
func NewPlagiarismRepository(db *gorm.DB) PlagiarismRepository {
	return &plagiarismRepository{db: db}
}

func (r *plagiarismRepository) Create(ctx context.Context, check *model.PlagiarismCheck) error {
	if err := r.db.WithContext(ctx).Create(check).Error; err != nil {
		return fmt.Errorf("create plagiarism check: %w", err)
	}
	return nil
}

func (r *plagiarismRepository) Update(ctx context.Context, check *model.PlagiarismCheck) error {
	if err := r.db.WithContext(ctx).Omit("Paper").Save(check).Error; err != nil {
		return fmt.Errorf("update plagiarism check: %w", err)
	}
	return nil
}

func (r *plagiarismRepository) FindByPaper(ctx context.Context, paperID uint) (*model.PlagiarismCheck, error) {
	var check model.PlagiarismCheck
	if err := r.db.WithContext(ctx).Where("paper_id = ?", paperID).First(&check).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find plagiarism by paper %d: %w", paperID, ErrNotFound)
		}
		return nil, fmt.Errorf("find plagiarism by paper %d: %w", paperID, err)
	}
	return &check, nil
}
