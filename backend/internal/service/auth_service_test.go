package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/util"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:           "test-secret",
		JWTExpireHours:      72,
		SimilarityThreshold: 30,
	}
}

func TestAuthServiceRegister(t *testing.T) {
	store := newFakeStore()
	svc := NewAuthService(store, newTestConfig(), newTestLogger())
	ctx := context.Background()

	tests := []struct {
		name     string
		req      dto.RegisterRequest
		wantErr  bool
		wantCode int
	}{
		{
			name: "valid author register",
			req: dto.RegisterRequest{
				Username: "alice", Password: "secret123", Email: "a@b.com",
				RealName: "爱丽丝", Role: "author",
			},
		},
		{
			name: "duplicate username rejected",
			req: dto.RegisterRequest{
				Username: "alice", Password: "secret123", RealName: "爱丽丝2",
			},
			wantErr:  true,
			wantCode: constants.ErrUserExists,
		},
		{
			name: "invalid role rejected",
			req: dto.RegisterRequest{
				Username: "bob", Password: "secret123", RealName: "鲍勃", Role: "admin",
			},
			wantErr:  true,
			wantCode: constants.ErrRoleNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := svc.Register(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				var appErr *util.AppError
				if !errors.As(err, &appErr) {
					t.Fatalf("expected AppError, got %v", err)
				}
				if appErr.Code != tt.wantCode {
					t.Fatalf("expected code %d, got %d", tt.wantCode, appErr.Code)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Role != "author" {
				t.Errorf("expected role author, got %s", user.Role)
			}
			if user.Password == tt.req.Password {
				t.Errorf("password must be hashed")
			}
		})
	}
}

func TestAuthServiceLogin(t *testing.T) {
	store := newFakeStore()
	svc := NewAuthService(store, newTestConfig(), newTestLogger())
	ctx := context.Background()

	_, err := svc.Register(ctx, dto.RegisterRequest{
		Username: "alice", Password: "secret123", RealName: "爱丽丝",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		name     string
		req      dto.LoginRequest
		wantErr  bool
		wantCode int
	}{
		{name: "correct credentials", req: dto.LoginRequest{Username: "alice", Password: "secret123"}},
		{name: "wrong password", req: dto.LoginRequest{Username: "alice", Password: "wrong"}, wantErr: true, wantCode: constants.ErrInvalidCredential},
		{name: "unknown user", req: dto.LoginRequest{Username: "nobody", Password: "x"}, wantErr: true, wantCode: constants.ErrInvalidCredential},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := svc.Login(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				var appErr *util.AppError
				if !errors.As(err, &appErr) || appErr.Code != tt.wantCode {
					t.Fatalf("expected code %d, got %v", tt.wantCode, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token == "" || user == nil {
				t.Fatalf("expected token and user")
			}
			claims, err := util.ParseToken("test-secret", token)
			if err != nil {
				t.Fatalf("parse token: %v", err)
			}
			if claims.Username != "alice" {
				t.Errorf("expected username alice, got %s", claims.Username)
			}
		})
	}
}
