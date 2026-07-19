# Fungi Wiki 微生物百科

Fungi Wiki 是一个从微生物百科逐步演进到合成生物设计辅助平台的项目。

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
- 菌种、功能标签运营管理
- 菌种与功能标签关联
- 培养基、温度、pH、盐度和氧需求维护
- 文献来源、实验结论、证据等级和评分维护
- 草稿、待审核、已发布、已归档状态管理
- 提交审核、审核通过和驳回流程
- 按关键词、功能、温度、pH、安全等级和来源环境联合筛选
- 公开接口仅展示已发布数据

## 环境要求

- Go 1.22 或更高版本
- Node.js 20 或更高版本
- npm 10 或更高版本
- Docker Desktop 或其他可用的 Docker daemon

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
| PostgreSQL | localhost:55432 |

默认数据库连接：

```text
postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable
```

## 数据库迁移

新建数据库时，PostgreSQL 容器会按文件名顺序执行 `apps/api/migrations` 下的 SQL。

已有数据库需要手动应用新增迁移。例如应用搜索索引：

```bash
docker exec -i fungi-wiki-postgres \
  psql -U fungi -d fungi_wiki \
  < apps/api/migrations/002_search_indexes.sql
```

## 数据发布流程

```text
运营录入草稿
→ 维护功能、培养条件和文献证据
→ 提交审核
→ 审核通过并发布 / 驳回为草稿
→ 用户端展示已发布数据
```

菌种不能通过普通新增或编辑接口直接发布。编辑已发布菌种后，数据会重新变为草稿。

## 多条件搜索

示例：查询具备生防功能、适宜 30°C 和 pH 7、属于 BSL-1 且来源包含“土壤”的菌种。

```text
GET /api/species?functionTag=biocontrol&temperature=30&ph=7&safetyLevel=BSL-1&sourceEnvironment=土壤
```

多个条件采用 AND 逻辑。温度和 pH 按目标值是否落在培养条件范围内进行匹配。

## 构建与测试

```bash
npm run test:api
npm run build:api
npm run build:web
npm run build:admin
```

## 相关文档

- [需求规划](docs/需求规划.md)
- [系统分层设计](docs/系统分层设计.md)
- [数据库表结构设计](docs/数据库表结构设计.md)
- [本地开发启动说明](docs/本地开发启动说明.md)
- [Go API 使用说明](apps/api/README.md)

## 演进路线

1. 完善百科和功能菌结构化数据。
2. 增加搜索排序、分页、同义词和搜索日志。
3. 建设基于证据和规则的菌种推荐。
4. 引入向量检索和可解释 AI 推荐。
5. 扩展到底盘菌、代谢通路和合成生物设计辅助。
