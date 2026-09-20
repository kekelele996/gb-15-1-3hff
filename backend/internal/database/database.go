package database

import (
	"context"
	"fmt"
	"time"

	"github.com/paperflow/paperflow/internal/config"
	"github.com/paperflow/paperflow/internal/model"
	"github.com/paperflow/paperflow/internal/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect 建立 PostgreSQL 连接。
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

// Migrate 自动迁移全部模型。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Paper{},
		&model.Review{},
		&model.Revision{},
		&model.PlagiarismCheck{},
		&model.AuditLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}

// Seed 初始化演示账号与论文库样例数据。
func Seed(db *gorm.DB) error {
	ctx := context.Background()
	users := []model.User{
		{Username: "admin", Password: hash("admin123"), Email: "admin@paperflow.dev", RealName: "系统管理员", Role: "admin"},
		{Username: "editor", Password: hash("editor123"), Email: "editor@paperflow.dev", RealName: "王编辑", Institution: "PaperFlow 期刊编辑部", Role: "editor"},
		{Username: "reviewer", Password: hash("reviewer123"), Email: "reviewer@paperflow.dev", RealName: "李审稿", Institution: "某大学计算机学院", Role: "reviewer"},
		{Username: "author", Password: hash("author123"), Email: "author@paperflow.dev", RealName: "张作者", Institution: "某大学信息学院", Role: "author"},
	}
	for i := range users {
		var count int64
		if err := db.WithContext(ctx).Model(&model.User{}).Where("username = ?", users[i].Username).Count(&count).Error; err != nil {
			return fmt.Errorf("count user: %w", err)
		}
		if count == 0 {
			if err := db.WithContext(ctx).Create(&users[i]).Error; err != nil {
				return fmt.Errorf("seed user %s: %w", users[i].Username, err)
			}
			util.Info("seeded user", "username", users[i].Username, "role", users[i].Role)
		}
	}

	var author model.User
	if err := db.WithContext(ctx).Where("username = ?", "author").First(&author).Error; err != nil {
		return fmt.Errorf("find author: %w", err)
	}
	var cnt int64
	if err := db.WithContext(ctx).Model(&model.Paper{}).Where("status = ?", "accepted").Count(&cnt).Error; err != nil {
		return fmt.Errorf("count accepted papers: %w", err)
	}
	if cnt == 0 {
		samples := []struct {
			title, abs, kw, subject, authors string
		}{
			{
				"基于深度学习的学术论文自动摘要研究",
				"针对海量学术文献，提出一种基于深度学习的自动摘要方法，在公开数据集上取得优于基线模型的性能表现。",
				"深度学习,自动摘要,自然语言处理",
				"computer",
				`[{"name":"张作者","institution":"某大学信息学院"},{"name":"李学者","institution":"某研究院"}]`,
			},
			{
				"高校实验室安全管理信息系统设计",
				"设计并实现一套高校实验室安全管理信息系统，覆盖危化品台账、安全检查与隐患整改闭环管理。",
				"实验室安全,信息系统,高校",
				"management",
				`[{"name":"张作者","institution":"某大学信息学院"}]`,
			},
			{
				"基于大语言模型的智能教学辅助系统",
				"探索大语言模型在智能教学辅助领域的应用，构建习题生成、学情分析与个性化推荐一体化系统。",
				"大语言模型,智能教学,教育技术",
				"education",
				`[{"name":"张作者","institution":"某大学信息学院"},{"name":"王老师","institution":"某师范大学"}]`,
			},
		}
		for _, s := range samples {
			paper := model.Paper{
				Title:        s.title,
				Abstract:     s.abs,
				Keywords:     s.kw,
				Subject:      s.subject,
				AuthorsMeta:  s.authors,
				Status:       "accepted",
				Version:      3,
				SubmitterID:  author.ID,
				FileName:     s.title + ".pdf",
				FileKey:      "demo/" + s.title + ".pdf",
				Similarity:   9.5,
				FinalDecision: "accepted",
			}
			if err := db.WithContext(ctx).Create(&paper).Error; err != nil {
				return fmt.Errorf("seed paper: %w", err)
			}
			now := time.Now()
			check := model.PlagiarismCheck{
				PaperID:    paper.ID,
				Similarity: 9.5,
				Status:     "completed",
				CheckedAt:  &now,
				Report:     `[{"source":"公开文献库","paragraph":"摘要内容与公开文献存在部分相似表达","similarity":9.5}]`,
			}
			if err := db.WithContext(ctx).Create(&check).Error; err != nil {
				return fmt.Errorf("seed plagiarism: %w", err)
			}
		}
	}
	return nil
}

func hash(pwd string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}
