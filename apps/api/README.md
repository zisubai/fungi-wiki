# Fungi Wiki API

Go 后端服务，当前已接入 PostgreSQL，并提供菌种 CRUD API。

## 本地启动数据库

优先使用项目脚本，避免和本机已有 PostgreSQL 的 5432 端口冲突：

```bash
./scripts/db-up.sh
```

如果本机支持 Docker Compose，也可以执行：

```bash
docker compose up -d postgres
```

数据库默认连接：

```text
postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable
```

首次启动会自动执行：

```text
apps/api/migrations/001_init_schema.sql
```

## 启动 API

```bash
DATABASE_URL="postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable" go run ./cmd/server
```

也可以在项目根目录执行：

```bash
npm run dev:api
```

## 用户端接口

```text
GET /healthz
GET /api/species
GET /api/species?q=促生
GET /api/species/bacillus-subtilis
```

用户端列表默认只返回 `published` 状态的菌种。

## 运营端接口

```text
GET    /api/admin/species
GET    /api/admin/species?status=draft
POST   /api/admin/species
GET    /api/admin/species/{idOrSlug}
PUT    /api/admin/species/{idOrSlug}
DELETE /api/admin/species/{idOrSlug}
DELETE /api/admin/species/{idOrSlug}/hard
```

`DELETE /api/admin/species/{idOrSlug}` 默认归档数据，不物理删除；`/hard` 用于物理删除。

## 创建菌种示例

```bash
curl -X POST http://localhost:8080/api/admin/species \
  -H 'Content-Type: application/json' \
  -d '{
    "slug": "pseudomonas-putida",
    "latinName": "Pseudomonas putida",
    "chineseName": "恶臭假单胞菌",
    "safetyLevel": "BSL-1",
    "isModelOrganism": true,
    "summary": "常用于环境污染物降解和代谢工程研究。",
    "status": "published"
  }'
```

## 功能标签接口

用户端接口：

```text
GET /api/function-tags
GET /api/function-tags/{idOrCode}
```

运营端接口：

```text
GET    /api/admin/function-tags
GET    /api/admin/function-tags?q=促生
POST   /api/admin/function-tags
GET    /api/admin/function-tags/{idOrCode}
PUT    /api/admin/function-tags/{idOrCode}
DELETE /api/admin/function-tags/{idOrCode}
```

创建功能标签示例：

```bash
curl -X POST http://localhost:8080/api/admin/function-tags \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "促生",
    "code": "plant-growth-promotion",
    "description": "促进植物生长，包括促根、促苗、提高养分吸收等。",
    "sortOrder": 10
  }'
```

## 菌种功能关联接口

用户端查看某个菌种的已关联功能：

```text
GET /api/species/{idOrSlug}/functions
```

运营端读取或整体更新关联功能：

```text
GET /api/admin/species/{idOrSlug}/functions
PUT /api/admin/species/{idOrSlug}/functions
```

更新请求示例：

```json
{
  "items": [
    {
      "functionTagId": "功能标签 UUID",
      "description": "促进根系生长",
      "functionStrength": "high",
      "verificationMethod": "盆栽试验",
      "applicableEnvironment": "农田土壤",
      "confidenceScore": 85
    }
  ]
}
```

`PUT` 使用事务整体替换当前菌种的功能关联；传入空数组可以清空关联。

## 培养条件接口

```text
GET /api/species/{idOrSlug}/culture-conditions
GET /api/admin/species/{idOrSlug}/culture-conditions
PUT /api/admin/species/{idOrSlug}/culture-conditions
```

`PUT` 请求使用 `{ "items": [...] }`，每项可包含 `mediumName`、`temperatureMin`、`temperatureMax`、`phMin`、`phMax`、`salinityMin`、`salinityMax`、`oxygenRequirement`、`cultureTime` 和 `notes`。

## 文献证据接口

```text
GET    /api/species/{idOrSlug}/evidences
GET    /api/admin/species/{idOrSlug}/evidences
POST   /api/admin/species/{idOrSlug}/evidences
DELETE /api/admin/species/{idOrSlug}/evidences/{evidenceId}
```

新增证据时同时提交文献标题、作者、期刊、年份、DOI/PMID、来源链接、实验结论、证据等级和证据分。

## 审核发布接口

```text
GET  /api/admin/audits?status=pending
POST /api/admin/audits/species/{idOrSlug}/submit
POST /api/admin/audits/{auditId}/approve
POST /api/admin/audits/{auditId}/reject
```

状态流转为：

```text
draft → pending_review → published
                       ↘ draft（驳回）
```

菌种创建后固定为草稿；编辑已发布菌种会退回草稿。待审核数据不可编辑，只有审核通过接口可以发布。
