package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// AuditLogService 操作审计日志服务。
type AuditLogService struct {
	store  repository.Store
	logger *slog.Logger
}

// NewAuditLogService 构造审计日志服务。
func NewAuditLogService(store repository.Store, logger *slog.Logger) *AuditLogService {
	return &AuditLogService{store: store, logger: logger}
}

// Record 写入一条审计日志（middleware 与关键业务埋点共用）。
func (s *AuditLogService) Record(ctx context.Context, userID uint, username, action, entity, entityID, detail, ip, requestID string) error {
	if username == "" {
		username = "anonymous"
	}
	log := &model.AuditLog{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Detail:    detail,
		IP:        ip,
		RequestID: requestID,
	}
	if err := s.store.AuditLogRepository().Create(ctx, log); err != nil {
		s.logger.Error("write audit log failed", "action", action, "error", err)
		return fmt.Errorf("record audit log: %w", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogAuditRecord, action, entity, entityID))
	return nil
}

// List 分页查询审计日志。
func (s *AuditLogService) List(ctx context.Context, page, size int) ([]model.AuditLog, int64, error) {
	items, total, err := s.store.AuditLogRepository().List(ctx, page, size)
	if err != nil {
		return nil, 0, util.NewAppError(constants.ErrInternal, "审计日志获取失败：系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogAuditList, page, size))
	return items, total, nil
}
