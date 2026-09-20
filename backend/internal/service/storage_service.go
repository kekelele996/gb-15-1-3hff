package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/constants"
)

// StorageService MinIO 对象存储服务。
type StorageService struct {
	client *minio.Client
	bucket string
	logger *slog.Logger
}

// NewStorageService 构造 MinIO 客户端。
func NewStorageService(cfg *config.Config, logger *slog.Logger) (*StorageService, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &StorageService{client: client, bucket: cfg.MinIOBucket, logger: logger}, nil
}

// EnsureBucket 确保桶存在。
func (s *StorageService) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return nil
}

// Upload 上传对象。
func (s *StorageService) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("put object %s: %w", objectName, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogFileUploadOK, objectName, size))
	return nil
}

// Bucket 返回桶名。
func (s *StorageService) Bucket() string { return s.bucket }
