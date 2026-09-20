package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/util"
)

// AuthService 认证服务：注册/登录/令牌签发。
type AuthService struct {
	store  repository.Store
	cfg    *config.Config
	logger *slog.Logger
}

// NewAuthService 构造认证服务。
func NewAuthService(store repository.Store, cfg *config.Config, logger *slog.Logger) *AuthService {
	return &AuthService{store: store, cfg: cfg, logger: logger}
}

// Register 注册用户，默认角色为 author。
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*model.User, error) {
	s.logger.Info(fmt.Sprintf(constants.LogUserRegisterStart, req.Username))
	role := req.Role
	if role == "" {
		role = constants.RoleAuthor
	}
	if !s.validRole(role) {
		return nil, util.NewAppError(constants.ErrRoleNotAllowed,
			fmt.Sprintf("注册失败：角色 %s 不允许注册", role), nil)
	}
	_, err := s.store.UserRepository().FindByUsername(ctx, req.Username)
	if err == nil {
		return nil, util.NewAppError(constants.ErrUserExists,
			fmt.Sprintf("注册失败：用户名 %s 已被占用", req.Username), nil)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(constants.ErrInternal,
			fmt.Sprintf("注册失败：查询用户 %s 时系统内部错误", req.Username), err)
	}
	hashed, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, util.NewAppError(constants.ErrInternal, "注册失败：密码加密时系统内部错误", err)
	}
	user := &model.User{
		Username:    req.Username,
		Password:    hashed,
		Email:       req.Email,
		RealName:    req.RealName,
		Institution: req.Institution,
		Role:        role,
	}
	if err := s.store.UserRepository().Create(ctx, user); err != nil {
		return nil, util.NewAppError(constants.ErrInternal,
			fmt.Sprintf("注册失败：创建用户 %s 时系统内部错误", req.Username), err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserRegisterOK, user.Username, user.Role))
	return user, nil
}

// Login 登录并签发 JWT。
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*model.User, string, error) {
	user, err := s.store.UserRepository().FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.logger.Warn(fmt.Sprintf(constants.LogUserLoginFail, req.Username, "user not found"))
			return nil, "", util.NewAppError(constants.ErrInvalidCredential,
				fmt.Sprintf("登录失败：用户 %s 不存在或密码错误", req.Username), nil)
		}
		return nil, "", util.NewAppError(constants.ErrInternal,
			fmt.Sprintf("登录失败：查询用户 %s 时系统内部错误", req.Username), err)
	}
	if !util.CheckPassword(user.Password, req.Password) {
		s.logger.Warn(fmt.Sprintf(constants.LogUserLoginFail, req.Username, "bad password"))
		return nil, "", util.NewAppError(constants.ErrInvalidCredential,
			fmt.Sprintf("登录失败：用户 %s 不存在或密码错误", req.Username), nil)
	}
	token, err := util.GenerateToken(s.cfg.JWTSecret, s.cfg.JWTExpireHours, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", util.NewAppError(constants.ErrInternal, "登录失败：签发令牌时系统内部错误", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTokenIssue, user.ID, user.Username))
	s.logger.Info(fmt.Sprintf(constants.LogUserLoginOK, user.Username, user.Role))
	return user, token, nil
}

func (s *AuthService) validRole(role string) bool {
	switch role {
	case constants.RoleAuthor, constants.RoleReviewer, constants.RoleEditor:
		return true
	}
	return false
}
