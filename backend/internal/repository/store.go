package repository

import (
	"context"

	"gorm.io/gorm"
)

// Store 仓储集合，提供事务能力与各实体仓储访问。
type Store interface {
	Transaction(ctx context.Context, fn func(Store) error) error
	UserRepository() UserRepository
	PaperRepository() PaperRepository
	ReviewRepository() ReviewRepository
	RevisionRepository() RevisionRepository
	PlagiarismRepository() PlagiarismRepository
	AuditLogRepository() AuditLogRepository
}

type store struct {
	db             *gorm.DB
	userRepo       UserRepository
	paperRepo      PaperRepository
	reviewRepo     ReviewRepository
	revisionRepo   RevisionRepository
	plagiarismRepo PlagiarismRepository
	auditRepo      AuditLogRepository
}

// NewStore 构造 Store，所有仓储绑定同一数据库连接。
func NewStore(db *gorm.DB) Store {
	return newStore(db)
}

func newStore(db *gorm.DB) *store {
	return &store{
		db:             db,
		userRepo:       NewUserRepository(db),
		paperRepo:      NewPaperRepository(db),
		reviewRepo:     NewReviewRepository(db),
		revisionRepo:   NewRevisionRepository(db),
		plagiarismRepo: NewPlagiarismRepository(db),
		auditRepo:      NewAuditLogRepository(db),
	}
}

// Transaction 在数据库事务中执行 fn，fn 内获得事务作用域的新 Store。
func (s *store) Transaction(ctx context.Context, fn func(Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(newStore(tx))
	})
}

func (s *store) UserRepository() UserRepository     { return s.userRepo }
func (s *store) PaperRepository() PaperRepository   { return s.paperRepo }
func (s *store) ReviewRepository() ReviewRepository { return s.reviewRepo }
func (s *store) RevisionRepository() RevisionRepository {
	return s.revisionRepo
}
func (s *store) PlagiarismRepository() PlagiarismRepository {
	return s.plagiarismRepo
}
func (s *store) AuditLogRepository() AuditLogRepository {
	return s.auditRepo
}
