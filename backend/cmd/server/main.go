package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/database"
	"github.com/paperflow/paperflow/internal/handler"
	"github.com/paperflow/paperflow/internal/middleware"
	"github.com/paperflow/paperflow/internal/repository"
	"github.com/paperflow/paperflow/internal/router"
	"github.com/paperflow/paperflow/internal/service"
	"github.com/paperflow/paperflow/internal/util"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	logger := util.NewLogger(cfg.LogLevel)
	util.SetLogger(logger)

	ctx := context.Background()

	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("connect database failed", "error", err)
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		logger.Error("migrate database failed", "error", err)
		os.Exit(1)
	}
	if err := database.Seed(db); err != nil {
		logger.Error("seed database failed", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, DB: 0})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("redis ping failed, rate limit will fail open", "error", err)
	}

	storage, err := service.NewStorageService(cfg, logger)
	if err != nil {
		logger.Error("init storage failed", "error", err)
		os.Exit(1)
	}
	if err := storage.EnsureBucket(ctx); err != nil {
		logger.Error("ensure bucket failed", "error", err)
		os.Exit(1)
	}

	store := repository.NewStore(db)

	authSvc := service.NewAuthService(store, cfg, logger)
	userSvc := service.NewUserService(store, logger)
	plagiarismSvc := service.NewPlagiarismService(store, cfg, logger)
	paperSvc := service.NewPaperService(store, plagiarismSvc, cfg, logger)
	reviewSvc := service.NewReviewService(store, logger)
	revisionSvc := service.NewRevisionService(store, logger)
	auditSvc := service.NewAuditLogService(store, logger)
	statSvc := service.NewStatisticsService(store, logger)

	authMw := middleware.NewAuth(store.UserRepository(), cfg.JWTSecret, logger)
	auditMw := middleware.NewAudit(auditSvc, logger)
	rateLimit := middleware.NewRateLimit(rdb, logger, 300, 60).Limit()

	handlers := &handler.Handlers{
		Auth:       handler.NewAuthHandler(authSvc, auditSvc, logger),
		User:       handler.NewUserHandler(userSvc, logger),
		Paper:      handler.NewPaperHandler(paperSvc, auditSvc, logger),
		Review:     handler.NewReviewHandler(reviewSvc, logger),
		Revision:   handler.NewRevisionHandler(revisionSvc, logger),
		Plagiarism: handler.NewPlagiarismHandler(plagiarismSvc, auditSvc, logger),
		Audit:      handler.NewAuditHandler(auditSvc, logger),
		Statistics: handler.NewStatisticsHandler(statSvc, logger),
		File:       handler.NewFileHandler(storage, auditSvc, logger),
		Health:     handler.NewHealthHandler(),
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	r := router.New(engine, handlers, authMw, auditMw, rateLimit, logger)
	r.NoRoute()

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: engine,
	}

	go func() {
		logger.Info("paperflow backend started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

