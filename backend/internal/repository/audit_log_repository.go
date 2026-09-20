package repository

import (
	"context"
	"fmt"

	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
)

// AuditLogRepository 审计日志仓储接口。
type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, page, size int) ([]model.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository 构造审计日志仓储。
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *auditLogRepository) List(ctx context.Context, page, size int) ([]model.AuditLog, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	var items []model.AuditLog
	if err := r.db.WithContext(ctx).Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	return items, total, nil
}
