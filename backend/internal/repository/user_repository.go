package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户仓储接口。
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	ListByRole(ctx context.Context, role string) ([]model.User, error)
	CountByRole(ctx context.Context, role string) (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find user %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("find user %d: %w", id, err)
	}
	return &u, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find user by username %s: %w", username, ErrNotFound)
		}
		return nil, fmt.Errorf("find user by username %s: %w", username, err)
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *userRepository) ListByRole(ctx context.Context, role string) ([]model.User, error) {
	var items []model.User
	if err := r.db.WithContext(ctx).Where("role = ?", role).Order("id ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list users by role %s: %w", role, err)
	}
	return items, nil
}

func (r *userRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("role = ?", role).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("count users by role %s: %w", role, err)
	}
	return n, nil
}
