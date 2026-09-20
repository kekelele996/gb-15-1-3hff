package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// UserService 用户服务。
type UserService struct {
	store  repository.Store
	logger *slog.Logger
}

// NewUserService 构造用户服务。
func NewUserService(store repository.Store, logger *slog.Logger) *UserService {
	return &UserService{store: store, logger: logger}
}

// Me 获取当前用户。
func (s *UserService) Me(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.store.UserRepository().FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.ErrUserNotFound,
				fmt.Sprintf("获取用户失败：用户 id=%d 不存在", id), nil)
		}
		return nil, util.NewAppError(constants.ErrInternal, "获取用户失败：系统内部错误", err)
	}
	return user, nil
}

// UpdateProfile 更新当前用户资料。
func (s *UserService) UpdateProfile(ctx context.Context, id uint, req dto.UpdateProfileRequest) (*model.User, error) {
	user, err := s.Me(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Institution != "" {
		user.Institution = req.Institution
	}
	if err := s.store.UserRepository().Update(ctx, user); err != nil {
		return nil, util.NewAppError(constants.ErrInternal,
			fmt.Sprintf("更新用户失败：用户 id=%d 资料保存时系统内部错误", id), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogProfileUpdate, id))
	return user, nil
}

// ListReviewers 列出全部审稿人（编辑分配审稿人时使用）。
func (s *UserService) ListReviewers(ctx context.Context) ([]model.User, error) {
	items, err := s.store.UserRepository().ListByRole(ctx, constants.RoleReviewer)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "获取审稿人列表失败：系统内部错误", err)
	}
	return items, nil
}
