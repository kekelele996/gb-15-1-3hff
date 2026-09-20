# PaperFlow 学术论文全流程管理平台

面向高校与期刊的学术论文全流程管理平台，覆盖论文提交、同行评审、修改反馈到最终收录的完整发表流程，服务作者、审稿人、期刊编辑三类角色。

## 快速启动（Docker Compose 一键部署，推荐）

```bash
cd 任意目录（含中文目录名均可）
docker compose up -d --build
```

启动后访问：

| 服务 | 地址 |
| --- | --- |
| 前端 | http://localhost:8009 |
| 后端 API | http://localhost:3009/api/v1 |
| 健康检查 | http://localhost:3009/healthz |
| MinIO 控制台 | http://localhost:9006 （minioadmin / minioadmin） |

演示账号（启动时自动 Seed）：

| 角色 | 用户名 | 密码 |
| --- | --- | --- |
| 管理员 | admin | admin123 |
| 编辑 | editor | editor123 |
| 审稿人 | reviewer | reviewer123 |
| 作者 | author | author123 |

停止并清理：

```bash
docker compose down -v --remove-orphans
```

## 主要功能

1. **作者投稿管理**：注册登录后创建投稿，上传 PDF/Word 论文文件，填写标题、摘要、关键词、学科分类、作者与单位，跟踪投稿状态（已提交→初审中→外审中→修改中→已录用/已拒稿）。
2. **编辑部初审**：编辑检查格式与选题，通过后分配给审稿人，不通过退回作者并附理由。
3. **同行评审管理**：审稿人接受/拒绝邀请，接受后按截止日期提交评审意见（录用/小修后录用/大修后重审/拒稿），支持匿名双盲（给编辑的保密意见仅编辑可见）。
4. **修稿与反馈**：作者按审稿意见修改并重新提交，上传逐条回复的修改说明，多轮迭代。
5. **查重检测**：投稿自动触发查重，展示重复率与重复段落标注，超过 30% 自动退回。
6. **论文库与检索**：已录用论文进入论文库，支持按标题/摘要/关键词/学科检索。
7. **数据统计**：编辑统计面板——投稿量趋势、学科分布、平均审稿周期、录用率、审稿人工作量排名。

## 本地开发（备选）

### 后端

```bash
cd backend
go mod tidy
go run ./cmd/server
```

构建命令：`go build ./...`

### 前端

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器代理 `/api` 到 `http://localhost:3009`。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript，使用 Element Plus 组件库，Vite 构建工具 |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 16 |
| 缓存/限流 | Redis 7 |
| 对象存储 | MinIO |
| 认证 | JWT + RBAC |
| 日志 | log/slog |
| 参数校验 | github.com/go-playground/validator/v10 |

## 项目目录结构

```
gb-15-1/
├── backend/
│   ├── cmd/server/main.go        # 入口：装配依赖、启动服务
│   ├── internal/
│   │   ├── config/               # 环境变量配置
│   │   ├── database/             # 连接、迁移、Seed
│   │   ├── model/                # 每个实体一个文件
│   │   ├── dto/                  # 每个实体一个 DTO 文件
│   │   ├── repository/           # 每个实体一个仓储文件
│   │   ├── service/              # 每个实体一个服务文件
│   │   ├── handler/              # 每个实体一个处理器文件
│   │   ├── router/               # 每个实体一个路由注册文件
│   │   ├── middleware/           # 认证/RBAC/审计/限流/请求ID/错误处理
│   │   ├── constants/            # 错误码/消息/日志模板/枚举
│   │   └── util/                 # jwt/password/logger/formatters/response
│   ├── migrations/               # 迁移说明
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── api/                  # 每个实体一个 API 文件
│   │   ├── components/           # StatusBadge / PaperStatusSteps / PaperInfoCard / EmptyState / EChart
│   │   ├── pages/                # auth/layout/author/editor/reviewer/library/audit
│   │   ├── stores/               # auth / paper / review
│   │   ├── hooks/                # useAuth / usePagination
│   │   ├── utils/                # request / format
│   │   ├── constants/            # 与后端对应枚举
│   │   └── router/
│   ├── Dockerfile
│   └── nginx.conf
├── database/init.sql
├── docker-compose.yml
├── .env
├── .env.example
└── README.md
```

## API 清单（前缀 `/api/v1`）

统一响应体：`{ "code": 0, "message": "ok", "data": ... }`。

### 认证

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | /auth/register | 注册 | 公开 |
| POST | /auth/login | 登录，返回 JWT | 公开 |

### 用户

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | /users/me | 当前用户 | 登录 |
| PUT | /users/me | 更新资料 | 登录 |
| GET | /users/reviewers | 审稿人列表 | 编辑/管理员 |

### 论文

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | /papers | 新建投稿（自动触发查重） | 作者/管理员 |
| GET | /papers/mine | 我的投稿 | 作者/管理员 |
| GET | /papers?status= | 按状态列论文（编辑初审队列） | 登录 |
| GET | /papers/:id | 论文详情 | 登录 |
| PUT | /papers/:id | 更新元信息 | 作者本人 |
| POST | /papers/:id/initial-review | 初审通过/退回 | 编辑/管理员 |
| POST | /papers/:id/final-decision | 终审录用/拒稿 | 编辑/管理员 |
| POST | /papers/:id/revise | 修稿重投 | 作者本人 |
| GET | /library/papers?keyword=&subject= | 论文库检索 | 登录 |
| GET | /papers/:id/revisions | 修稿记录 | 登录 |

### 审稿

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | /reviews/mine?status= | 我的审稿任务 | 审稿人/管理员 |
| GET | /reviews/paper/:paperID | 论文审稿记录 | 登录 |
| POST | /reviews/:id/respond | 接受/拒绝邀请 | 审稿人本人 |
| POST | /reviews/:id/submit | 提交评审意见 | 审稿人本人 |
| POST | /papers/:id/reviewers | 追加分配审稿人 | 编辑/管理员 |

### 查重 / 文件 / 统计 / 审计

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | /papers/:id/plagiarism | 查重结果 | 登录 |
| POST | /papers/:id/plagiarism/rerun | 重跑查重 | 编辑/管理员 |
| POST | /files/upload | 上传论文文件到 MinIO | 登录 |
| GET | /statistics/overview | 统计概览 | 编辑/管理员 |
| GET | /statistics/trend?days=30 | 投稿量趋势 | 编辑/管理员 |
| GET | /statistics/subjects | 学科分布 | 编辑/管理员 |
| GET | /statistics/reviewers | 审稿人工作量 | 编辑/管理员 |
| GET | /audit-logs?page=&size= | 审计日志 | 编辑/管理员 |
| GET | /healthz | 健康检查 | 公开 |

### 复用关系说明

- `GET /papers/mine`、`GET /papers`、`GET /library/papers` 三个接口复用 `PaperService.list` 与 `PaperRepository.List`（同一 repository 方法）。
- 「论文创建自动查重」与 `POST /papers/:id/plagiarism/rerun` 复用 `PlagiarismService.RunCheck`（同一 service 方法）。
- `POST /papers/:id/initial-review` 与 `POST /papers/:id/reviewers` 复用 `ReviewRepository.Create` 创建审稿邀请。

### curl 调用示例（含 JWT）

```bash
# 登录获取 token
TOKEN=$(curl -sS -X POST http://localhost:3009/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"editor","password":"editor123"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')

# 携带 JWT 调用
curl -sS http://localhost:3009/api/v1/papers?status=submitted \
  -H "Authorization: Bearer $TOKEN"

# 健康检查
curl -sS http://localhost:3009/healthz
```

## 环境变量说明

见 `.env.example`，关键项：`COMPOSE_PROJECT_NAME`、`FRONTEND_PORT`、`BACKEND_PORT`、`DB_PORT`、`REDIS_PORT`、`MINIO_API_PORT`、`MINIO_CONSOLE_PORT`、`DB_NAME`、`DB_USER`、`DB_PASSWORD`、`JWT_SECRET`、`MINIO_ROOT_USER`、`MINIO_ROOT_PASSWORD`。

## Docker 部署说明

- 端口映射：前端 `${FRONTEND_PORT:-8009}:80`，后端 `${BACKEND_PORT:-3009}:8080`，数据库 `${DB_PORT:-44009}:5432`，Redis `${REDIS_PORT:-46309}:6379`，MinIO `${MINIO_API_PORT:-9005}:9000` / `${MINIO_CONSOLE_PORT:-9006}:9001`。
- 数据卷：`paperflow_pgdata`（PostgreSQL）、`paperflow_redisdata`（Redis）、`paperflow_miniodata`（MinIO），均为命名卷，不依赖目录名。
- 数据库/后端均配置 healthcheck，后端通过 `depends_on: { db: { condition: service_healthy } }` 等待数据库就绪。
- 常见问题：
  - 端口冲突：修改 `.env` 中对应端口后重新 `docker compose up -d`。
  - 依赖镜像拉取慢：配置 Docker 镜像加速源。
  - 首次构建较慢：等待 `docker compose ps` 全部 healthy 后再访问。

## 横切关注点（触达文件层）

1. **JWT 认证 + RBAC 权限**：数据库 `users.role` 字段 → `backend/internal/middleware/auth.go`、`backend/internal/middleware/rbac.go`、`backend/internal/util/jwt.go` → 前端 `frontend/src/router/index.ts`（路由守卫）与 `frontend/src/pages/layout/MainLayout.vue`（菜单显隐）、`frontend/src/utils/request.ts`（令牌注入）。
2. **操作审计日志**：数据库 `audit_logs` 表 → `backend/internal/middleware/audit.go` → service 埋点（`backend/internal/service/audit_log_service.go`，handler 关键动作埋点）→ 前端 `frontend/src/pages/audit/AuditLogs.vue`。
3. **全局错误处理与请求追踪**：`backend/internal/middleware/request_id.go`、`backend/internal/middleware/error_handler.go`、`backend/internal/util/app_error.go`、`backend/internal/constants/error_codes.go` → 前端 `frontend/src/utils/request.ts` 拦截器。

## 共享枚举出现位置清单

### 1. 角色 RoleType（admin / editor / reviewer / author）

- 后端 model：`backend/internal/model/user.go`
- 后端 constants：`backend/internal/constants/roles.go`
- 后端 DTO：`backend/internal/dto/user_dto.go`（注册角色校验 oneof）
- 后端 service 状态机：`backend/internal/service/auth_service.go`（validRole）、`backend/internal/service/paper_service.go`（审稿人角色校验）、`backend/internal/service/review_service.go`（Assign）
- 后端 handler 校验：`backend/internal/handler/user_handler.go`（ListReviewers）
- 后端 middleware：`backend/internal/middleware/rbac.go`（RequireRoles）
- 后端 formatters：`backend/internal/util/formatters.go`（FormatRole）
- 后端日志模板：`backend/internal/constants/log_templates.go`（LogUserRegisterOK 等）
- 前端 constants：`frontend/src/constants/index.ts`（ROLE_MAP）
- 前端页面：`frontend/src/pages/layout/MainLayout.vue`（菜单过滤）、`frontend/src/router/index.ts`（roles 守卫）、`frontend/src/pages/auth/Register.vue`

### 2. 论文状态 PaperStatus（submitted / initial_review / external_review / revision / accepted / rejected）

- 后端 model：`backend/internal/model/paper.go`
- 后端 constants：`backend/internal/constants/paper_status.go`
- 后端 DTO：`backend/internal/dto/paper_dto.go`（PaperQuery oneof 校验）
- 后端 service 状态机：`backend/internal/service/paper_service.go`（Create/InitialReview/FinalDecision/Revise）、`backend/internal/service/plagiarism_service.go`（自动退回）、`backend/internal/service/review_service.go`（接受/提交推进状态）
- 后端 handler 校验：`backend/internal/handler/paper_handler.go`
- 后端 repository：`backend/internal/repository/paper_repository.go`（按状态统计）
- 后端 formatters：`backend/internal/util/formatters.go`（FormatPaperStatus）
- 后端日志模板：`backend/internal/constants/log_templates.go`（LogPaperCreateOK、LogInitialReviewOK 等）
- 前端 constants：`frontend/src/constants/index.ts`（PAPER_STATUS_MAP / PAPER_STATUS_ORDER）
- 前端组件/页面：`frontend/src/components/StatusBadge.vue`、`frontend/src/components/PaperStatusSteps.vue`、`frontend/src/pages/author/PaperList.vue`、`frontend/src/pages/editor/InitialReview.vue`

### 3. 审稿状态 ReviewStatus（invited / accepted / declined / completed）

- 后端 model：`backend/internal/model/review.go`
- 后端 constants：`backend/internal/constants/review_status.go`
- 后端 service 状态机：`backend/internal/service/review_service.go`（Respond/Submit）
- 后端 repository：`backend/internal/repository/review_repository.go`（CountCompleted）
- 后端 formatters：`backend/internal/util/formatters.go`（FormatReviewStatus）
- 后端日志模板：`backend/internal/constants/log_templates.go`（LogReviewRespond、LogReviewSubmit）
- 前端 constants：`frontend/src/constants/index.ts`（REVIEW_STATUS_MAP）
- 前端页面：`frontend/src/pages/reviewer/MyReviews.vue`、`frontend/src/pages/editor/PaperManage.vue`

### 4. 评审等级 ReviewDecision（accept / minor_revision / major_revision / reject）

- 后端 model：`backend/internal/model/review.go`
- 后端 constants：`backend/internal/constants/review_decision.go`
- 后端 DTO：`backend/internal/dto/review_dto.go`（SubmitReviewRequest oneof）
- 后端 service：`backend/internal/service/review_service.go`（Submit 状态推进）、`backend/internal/service/plagiarism_service.go`（自动拒稿 FinalDecision）
- 后端 formatters：`backend/internal/util/formatters.go`（FormatReviewDecision）
- 后端日志模板：`backend/internal/constants/log_templates.go`（LogReviewSubmit）
- 前端 constants：`frontend/src/constants/index.ts`（REVIEW_DECISION_MAP）
- 前端组件/页面：`frontend/src/components/StatusBadge.vue`、`frontend/src/pages/reviewer/ReviewDetail.vue`、`frontend/src/pages/editor/PaperManage.vue`

### 5. 查重状态 PlagiarismStatus（pending / completed / failed）

- 后端 model：`backend/internal/model/plagiarism_check.go`
- 后端 constants：`backend/internal/constants/plagiarism_status.go`
- 后端 service：`backend/internal/service/plagiarism_service.go`
- 后端 formatters：`backend/internal/util/formatters.go`（FormatPlagiarismStatus）
- 前端 constants：`frontend/src/constants/index.ts`（PLAGIARISM_STATUS_MAP）
- 前端页面：`frontend/src/pages/author/PaperDetail.vue`

## 共享前端组件 / hooks / utils

- 共享组件（≥3）：`StatusBadge.vue`、`PaperStatusSteps.vue`、`PaperInfoCard.vue`、`EmptyState.vue`、`EChart.vue`（PaperInfoCard 被作者详情页与编辑管理页复用）。
- 共享 hooks（≥2）：`useAuth.ts`、`usePagination.ts`。
- 共享 utils（≥2）：`request.ts`、`format.ts`。

## 后端中间件

- `middleware/auth.go`：JWT 认证。
- `middleware/rbac.go`：RBAC 角色校验。
- `middleware/request_id.go`：请求追踪 ID。
- `middleware/error_handler.go`：panic 恢复与统一错误响应。
- `middleware/audit.go`：操作审计。
- `middleware/rate_limit.go`：Redis 限流。

## 测试

```bash
cd backend
go test ./...
```

service 与 repository 均有表驱动单元测试（`*_test.go`）。

## License

MIT
