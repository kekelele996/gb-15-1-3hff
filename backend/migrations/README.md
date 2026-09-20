# Migrations

本项目使用 GORM AutoMigrate 在服务启动时自动建表/迁移（见 `backend/internal/database/database.go`）。

实体表：
- users
- papers
- reviews
- revisions
- plagiarism_checks
- audit_logs

如需手工执行 SQL 迁移，可参照 database/database.go 中的 DSN 与模型结构编写。
