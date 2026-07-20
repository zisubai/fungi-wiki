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

API 启动时自动执行尚未应用的迁移：

```text
apps/api/migrations/*.sql
```

执行记录保存在 `schema_migrations`，历史迁移通过 SHA-256 校验，已有 001–004 数据库会自动建立基线。新增迁移后重启 API 即可。

## 启动 API

```bash
DATABASE_URL="postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable" go run ./cmd/server
```

首次启动会根据以下环境变量创建或更新初始化管理员密码：

```text
JWT_SECRET=replace-with-a-long-random-secret
ADMIN_EMAIL=admin@fungi.local
ADMIN_PASSWORD=admin123456
```

默认值仅供本地开发，部署时必须修改 `JWT_SECRET` 和 `ADMIN_PASSWORD`。

## 登录与角色权限

```text
POST /api/auth/login
GET  /api/auth/me
GET  /api/admin/users
POST /api/admin/users
```

- `operator`：维护菌种、标签、培养条件、文献证据和批量导入。
- `expert`：查看待审核数据并执行通过或驳回。
- `admin`：拥有全部权限，并可创建账号。

所有 `/api/admin/*` 请求必须携带 `Authorization: Bearer <token>`。

也可以在项目根目录执行：

```bash
npm run dev:api
```

## 用户端接口

```text
GET /healthz
GET /readyz
GET /api/species
GET /api/species?q=促生
GET /api/species/bacillus-subtilis
```

用户端列表支持多条件联合筛选：

```text
GET /api/species?functionTag=biocontrol&temperature=30&ph=7.0&safetyLevel=BSL-1&sourceEnvironment=土壤
```

- `functionTag`：功能标签编码或 UUID。
- `temperature`：目标培养温度；匹配温度范围覆盖该值的菌种。
- `ph`：目标 pH，范围为 0–14；匹配 pH 范围覆盖该值的菌种。
- `safetyLevel`：安全等级，精确匹配且不区分大小写。
- `sourceEnvironment`：来源环境，模糊匹配。
- `q`：名称、Slug 或摘要关键词。

多个参数同时提供时使用 AND 联合筛选。

分页和排序参数：

- `limit`：每页数量，默认 20，最大 100。
- `offset`：结果偏移量。
- `sort`：`relevance`、`updated`、`name`、`quality` 或 `oldest`。`relevance` 在有关键词时按拉丁名、中文名、Slug 和别名的相似度排序。

列表响应包含：

```json
{
  "items": [],
  "total": 0,
  "limit": 20,
  "offset": 0
}
```

带关键词或筛选条件的公开搜索会写入 `search_logs`，记录查询、筛选条件和结果数量。

关键词搜索同时匹配菌种 Slug、拉丁名、中文名、摘要以及别名/同义词。

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

## 菌种别名接口

```text
GET /api/species/{idOrSlug}/aliases
GET /api/admin/species/{idOrSlug}/aliases
PUT /api/admin/species/{idOrSlug}/aliases
```

更新示例：

```json
{
  "items": [
    { "name": "历史拉丁名", "type": "former_name", "source": "文献来源" },
    { "name": "常用中文名", "type": "common_name" }
  ]
}
```

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

## CSV / Excel 批量导入

```text
POST /api/admin/imports/species
GET  /api/admin/imports?limit=20
```

上传接口使用 `multipart/form-data`，文件字段名为 `file`，支持 CSV、XLSX 和 XLSM，单次最多 10MB、5000 行。

必需表头为 `slug` 和 `latin_name`，也支持中文表头“标识”和“拉丁名”。可选表头：

```text
chinese_name,aliases,strain_number,source_environment,safety_level,is_model_organism,
summary,function_tags,medium_name,temperature_min,temperature_max,
ph_min,ph_max,oxygen_requirement,culture_time
```

合法数据会直接进入 `pending_review` 并生成审核记录；重复 Slug、缺少必填字段、范围错误或功能标签不存在的数据会作为失败行返回，不影响其他合法行。

## 搜索分析

```text
GET /api/admin/search-analytics?days=30
```

返回搜索次数、不同关键词数、无结果搜索数、热门关键词和无结果关键词。`days` 支持 1–365 天。

## 可解释菌种推荐

```text
POST /api/recommendations
```

请求示例：

```json
{
  "requirement": "寻找适合 30°C、中性环境的土壤生防菌",
  "functionTag": "biocontrol",
  "temperature": 30,
  "ph": 7,
  "safetyLevel": "BSL-1",
  "sourceEnvironment": "土壤",
  "limit": 5
}
```

当前版本为 `rules-v1`：从已发布菌种中按结构化条件召回，结合功能匹配、数据质量、证据数量和证据评分进行排序，并返回推荐理由与生物安全提示。若未传 `functionTag`，会尝试从需求文字中识别已有功能标签。

每次推荐都会写入 `recommendation_records`，保存需求、解析意图、候选结果、模型版本和风险等级。推荐仅用于候选初筛，不替代专家判断和实验验证。

推荐反馈与质量接口：

```text
POST /api/recommendations/{recordId}/feedback
GET  /api/admin/recommendations
```

反馈请求示例：

```json
{
  "feedbackType": "helpful",
  "content": "推荐理由清晰，候选符合预期"
}
```

`feedbackType` 支持 `helpful` 和 `unhelpful`。运营端质量接口返回推荐总量、有帮助/无帮助数量、风险等级、候选结果及单条推荐反馈统计。
