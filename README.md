# Fungi Wiki 微生物百科

Fungi Wiki 1.0 是一个从微生物百科逐步演进到合成生物设计辅助平台的可交付 MVP。

```text
微生物百科 → 功能菌数据库 → 菌种智能搜索 → AI 菌种推荐 → 合成生物设计辅助平台
```

当前版本已经具备菌种数据维护、功能标签、培养条件、文献证据、多条件搜索和审核发布能力。

## 技术栈

- 用户端：React、TypeScript、Vite
- 运营端：React、TypeScript、Vite
- 后端：Go、Gin、pgx
- 数据库：PostgreSQL 17
- 本地环境：npm workspaces、Docker

## 项目结构

```text
apps/
  web/                 用户前端
  admin/               运营管理端
  api/                 Go API 服务
    cmd/server/        服务入口
    internal/          业务模块
    migrations/        PostgreSQL 脚本
packages/
  ui/                  前端共享组件
  shared/              共享类型与工具
  config/              共享配置
docs/                  产品、架构和开发文档
scripts/               开发和数据库脚本
```

## 已实现功能

- 菌种百科列表和详情
- 2–3 个已发布菌种的功能、培养条件、安全与证据横向对比
- 菌种、功能标签运营管理
- 菌种与功能标签关联
- 菌种别名、历史名称和同义词维护与搜索
- 培养基、温度、pH、盐度和氧需求维护
- 文献来源、实验结论、证据等级和评分维护
- 草稿、待审核、已发布、已归档状态管理
- 提交审核、审核通过和驳回流程
- 按关键词、功能、温度、pH、安全等级和来源环境联合筛选
- 基于 PostgreSQL trigram 的拼写容错搜索与相关度排序
- 搜索结果分页、排序、搜索日志和运营端搜索分析
- 基于功能、培养条件、数据质量与文献证据的可解释菌种推荐
- 从自然语言需求中识别功能、温度、pH、安全等级和来源环境
- 推荐零命中时诊断阻断条件并给出放宽建议
- 双功能组合菌多候选配对、培养温度/pH 交集计算、兼容性优先排序和安全提示
- 组合菌需求、候选快照、算法版本和风险等级持久化审计
- 组合菌建议有用性反馈与运营端逐条质量跟踪
- 组合菌候选级共培养实验结果录入与历史追踪（菌种快照、结论、温度、pH、备注）
- 基于历史共培养实验的组合推荐加权、降级排序和验证状态展示
- 推荐结果关联可核验的文献证据链并保存快照
- 菌种数据完整度自动评分，关联数据变更后实时刷新
- 运营端数据质量明细和缺失项补全清单
- 数据质量总览、完整度分布、高频缺失项和低分优先列表
- 推荐有用性反馈与运营端推荐质量看板
- CSV / Excel 批量导入、逐行错误反馈和导入批次记录
- 运营端登录、JWT 会话和运营/专家/管理员角色权限
- 公开接口仅展示已发布数据

## 环境要求

- Go 1.22 或更高版本
- Node.js 20 或更高版本
- npm 10 或更高版本
- Docker Desktop（含 Compose v2）或其他可用的 Docker daemon + Compose

## 快速启动

安装依赖：

```bash
npm install
```

启动全部开发服务：

```bash
./scripts/dev.sh
```

也可以分别启动：

```bash
./scripts/db-up.sh
npm run dev:api
npm run dev:web
npm run dev:admin
```

默认地址：

| 服务 | 地址 |
|---|---|
| 用户端 | http://localhost:5173 |
| 运营端 | http://localhost:5174 |
| API | http://localhost:8080 |
| 健康检查 | http://localhost:8080/healthz |
| 就绪检查 | http://localhost:8080/readyz |
| PostgreSQL | localhost:55432 |

默认数据库连接：

```text
postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable
```

## Docker 一键部署

完整交付环境包含 PostgreSQL、Go API、用户端和运营端：

```bash
JWT_SECRET='请替换为高强度随机值' \
ADMIN_PASSWORD='请替换管理员密码' \
docker compose up -d --build
```

查看状态和停止服务：

```bash
docker compose ps
docker compose down
```

PostgreSQL 数据保存在 `postgres_data` volume 中，`docker compose down` 不会删除数据。仅在明确需要清空本地数据时使用 `docker compose down -v`。

### 生产环境部署

生产环境使用不含开发默认密码的独立 Compose 配置。先准备环境变量：

```bash
cp .env.production.example .env.production
```

将 `.env.production` 中的密码、JWT 密钥、管理员账号和正式域名全部替换后启动：

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

生产配置不会将 PostgreSQL 和 API 端口直接暴露到宿主机。用户端默认监听宿主机 `80`，管理端默认监听 `8081`，两者均通过容器内 Nginx 将 `/api/*` 请求转发给 API。公网部署时应在它们前面配置 TLS 反向代理，并限制管理端的网络访问范围。

## 数据库迁移

Go API 启动时会自动按文件名顺序执行 `apps/api/migrations` 下尚未应用的 SQL，无需再手动运行迁移命令。

迁移运行器会：

- 使用数据库 advisory lock 防止多个 API 实例重复迁移。
- 在 `schema_migrations` 表记录文件名、SHA-256 校验值和执行时间。
- 自动识别并兼容迁移运行器加入前已经创建的 001–004 表结构。
- 拒绝执行被修改过的历史迁移文件。

新增结构时应创建新的递增 SQL 文件，例如 `009_next_change.sql`，不要修改已经应用的迁移。

## 数据库备份

本地 PostgreSQL 容器运行时，可生成自带校验的 custom-format 备份：

```bash
./scripts/db-backup.sh
```

备份默认写入不纳入 Git 的 `backups/` 目录。脚本会在完成后调用 `pg_restore --list` 验证归档可读，并仅在校验成功后清理旧备份，默认保留当前数据库最近 14 份。可通过 `POSTGRES_CONTAINER`、`POSTGRES_DB`、`POSTGRES_USER`、`BACKUP_DIR` 和 `BACKUP_KEEP_COUNT` 覆盖默认值；设置 `BACKUP_KEEP_COUNT=0` 可关闭自动清理。恢复属于覆盖性操作，应在独立数据库完成恢复演练后再用于现有环境。

可在独立临时数据库中自动执行恢复演练：

```bash
./scripts/db-restore-drill.sh
# 或指定备份
./scripts/db-restore-drill.sh backups/fungi_wiki-YYYYMMDDTHHMMSSZ.dump
```

脚本会比较原库与恢复库的公共表清单和逐表行数，并在成功、失败或中断时删除临时数据库，不覆盖现有数据。

## 运行健康巡检

完整本机环境启动后，可一次检查 PostgreSQL、API、公共数据接口、用户端和管理端：

```bash
./scripts/health-check.sh
```

默认检查 `localhost:8080`、`localhost:5173` 和 `localhost:5174`。远程或自定义环境可通过 `API_URL`、`WEB_URL`、`ADMIN_URL`、`POSTGRES_CONTAINER` 和 `HEALTH_CHECK_TIMEOUT` 覆盖。

本地数据库内置两个已发布示例菌种，并包含用于功能验收的别名、功能标签、培养条件和演示文献。演示 DOI 与链接不是真实科学依据，上线前必须替换或删除 `source = demo-seed` 及 `10.0000/fungi.*.demo` 数据。

## 数据发布流程

```text
运营录入草稿
→ 维护功能、培养条件和文献证据
→ 提交审核
→ 审核通过并发布 / 驳回为草稿
→ 用户端展示已发布数据
```

菌种不能通过普通新增或编辑接口直接发布。编辑已发布菌种后，数据会重新变为草稿。

提交审核和最终批准都会执行发布质量门禁：数据质量至少 60 分，并且安全等级、菌种摘要、功能标签、培养条件和文献证据必须齐全。未通过时接口返回 `422` 及具体缺失项。

## 多条件搜索

示例：查询具备生防功能、适宜 30°C 和 pH 7、属于 BSL-1 且来源包含“土壤”的菌种。

```text
GET /api/species?functionTag=biocontrol&temperature=30&ph=7&safetyLevel=BSL-1&sourceEnvironment=土壤
```

多个条件采用 AND 逻辑。温度和 pH 按目标值是否落在培养条件范围内进行匹配。

关键词搜索支持少量拼写误差，并可通过 `sort=relevance` 按 Slug、拉丁名、中文名和别名的相似度排序。

## 数据质量评分

每个菌种按 100 分评估资料完整度：基础身份 20 分、来源与安全 20 分、摘要 15 分、别名 5 分、功能关联 15 分、培养条件 10 分、文献证据 15 分。菌种或关联数据变更时由 PostgreSQL 自动重算，运营端将评分标记为“完整”、“待补充”或“不完整”。

## 构建与测试

```bash
npm run test:api
npm run test:admin
npm run test:admin:coverage
npm run test:web
npm run test:web:coverage
npm run build:api
npm run vet:api
npm run build:web
npm run build:admin
npm run verify
npm run release:check
```

`npm run verify` 会执行用户端和运营端覆盖率门禁、前后端构建及 Go 测试，适合提交代码前一次性检查。CI 会将 HTML 报告保存为 `web-coverage` 和 `admin-coverage` 构建产物。

后端 CI 还会启动 PostgreSQL 17，自动执行迁移与管理员初始化，再运行 `scripts/smoke-api.sh` 检查健康接口和公共菌种接口。

Go API 配置了请求读写超时，并在收到 `SIGINT` 或 `SIGTERM` 时最多等待 10 秒完成优雅关闭。`/healthz` 用于进程存活检查，`/readyz` 会访问 PostgreSQL，用于判断实例是否可以接收流量。

API 响应包含 `X-Request-ID` 和基础安全响应头；客户端也可以传入合法的 `X-Request-ID` 以关联网关、后端和前端日志。

`CORS_ALLOWED_ORIGINS` 使用逗号分隔允许访问 API 的前端来源；生产环境应填写实际用户端和运营端域名，不建议配置为 `*`。

`TRUSTED_PROXIES` 使用逗号分隔可信反向代理 IP 或 CIDR；只有来自这些代理的转发 IP 头会影响访问日志中的客户端 IP，配置非法时 API 会拒绝启动。

API 访问日志包含 `request_id`、HTTP 方法、路径、状态码、耗时和客户端 IP，可使用响应头中的 Request ID 检索对应请求。

用户端和运营端会在请求发出前生成 `X-Request-ID`，并在接口或网络错误消息中显示该值；即使没有收到后端响应，也能使用请求 ID 检索网关和 API 日志。

GitHub Actions 会在每次 push 和 pull request 时自动执行运营端测试、Go 测试以及三端生产构建，配置位于 `.github/workflows/ci.yml`。

## 相关文档

- [需求规划](docs/需求规划.md)
- [系统分层设计](docs/系统分层设计.md)
- [数据库表结构设计](docs/数据库表结构设计.md)
- [本地开发启动说明](docs/本地开发启动说明.md)
- [Go API 使用说明](apps/api/README.md)
- [MVP 交付验收](docs/MVP交付验收.md)

## 演进路线

1. 完善百科和功能菌结构化数据。
2. 增加搜索排序、分页、同义词和搜索日志。
3. 建设基于证据和规则的菌种推荐。
4. 引入向量检索和可解释 AI 推荐。
5. 扩展到底盘菌、代谢通路和合成生物设计辅助。

当前代码已完成第 1–3 阶段及基于规则和实验反馈的第 4 阶段 MVP。第 5 阶段属于后续产品建设范围，不以未经验证的自动化设计替代专家和实验审核。
