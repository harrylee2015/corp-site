# 用户中心与项目管理设计文档

> 版本：v4 · 更新日期：2026-06  
> 本文档描述用户身份体系、分类映射、用户中心、公司信息、项目管理、企业认证、管理端统计/导出及公开页隐私的设计与实现约定。

---

## 1. 设计背景

平台从「单一信息发布」升级为「**身份驱动 + 公司信息 + 项目**」模式：

- 用户注册时选定**业务身份**（需求方 / 设备供应商 / 资金方），身份注册后**不可修改**
- 三种身份共用同一套用户中心模块（公司信息 + 添加/管理项目），表单字段按身份动态展示
- 所有企业用户须完成**企业认证**后方可发布项目
- 用户登录后进入**用户中心**（左侧导航），管理公司信息与项目
- 首页展示已审核**项目（projects）**；旧版 posts 仍可通过 `/posts/:id` 访问

---

## 2. 概念区分

| 概念 | 字段/值 | 说明 |
|------|---------|------|
| 系统角色 | `users.role` | `user` / `admin`，控制登录权限 |
| 业务身份 | `users.identity` | `demander` / `supplier` / `funder`，注册时选定，**只读** |
| 行业 | 一级分类 | categories 树中 `parent_id IS NULL` 的节点 |
| 类别 | 二级分类 | 一级下的叶子分类；「其他类」无二级，本身为叶子 |

Go 模型与表名一致：`Company` → `companies`、`Project` → `projects`。已有库启动时自动将旧表 `shops`/`products` 重命名。

---

## 3. 地区规则（公司信息 + 项目均适用）

| 身份 | 公司信息「可做区域」 | 项目「所在/服务区域」 | UI 控件 |
|------|---------------------|----------------------|---------|
| **资金方** | 多选省份 | 多选省份 | Checkbox 列表 |
| **需求方** | 单选省市 | 单选省市 | 省 Dropdown → 市 Dropdown |
| **设备供应商** | 单选省市 | 单选省市 | 同上 |

**存储格式**（`companies.regions` / `projects.regions`，TEXT JSON）：

```json
// 资金方
["江苏省", "浙江省"]

// 需求方 / 设备供应商
{"province": "江苏省", "city": "南京市"}
```

实现：`internal/data/regions.go`（`RegionMode`、`BuildRegionsJSON`、`FormatRegionsDisplay`）。启动时 `MigrateRegionFormats()` 会清空非资金方用户的旧省份数组数据，提示用户重新选择。

---

## 4. 用户中心导航

```
用户首页          GET  /my
项目管理
  ├ 公司信息      GET  /my/company          POST /api/my/company
  ├ 添加项目      GET  /my/projects/new     POST /api/my/projects
  └ 项目列表      GET  /my/projects         POST /api/my/projects/:id/delete
基本信息          GET  /my/profile
```

**旧路由兼容（302 重定向）**：

| 旧路径 | 新路径 |
|--------|--------|
| `/my/shop` | `/my/company` |
| `/my/products` | `/my/projects` |
| `/my/products/new` | `/my/projects/new` |
| `/products/:id` | `/projects/:id` |
| `/admin/product-review` | `/admin/project-review` |

POST API 旧路径（`/api/my/shop`、`/api/my/products` 等）仍可用。

---

## 5. 公司信息

每用户一条记录（`companies.user_id` 唯一），UI 称「公司信息」。

| 字段 | 说明 |
|------|------|
| shop_name | 公司名称（可预填 `users.company`） |
| established_at | 成立时间，`YYYY-MM`，`<input type="month">` |
| regions | 按身份规则存储（见 §3） |
| category_ids | 可做类型（按身份过滤后多选） |
| contact / phone / tel / address / intro / banner_path | 联系信息与介绍 |

模板：`web/templates/user/company.html`

---

## 6. 添加项目

前置条件：`users.verify_status = approved`

### 6.1 通用字段（三种身份）

| 字段 | 说明 |
|------|------|
| name | 项目名称 |
| category_id | 叶子分类（按身份过滤） |
| regions | 地区（规则见 §3） |
| intro | 项目介绍 |
| image_path | 项目图片（可选） |
| status | 提交后 `pending` |

### 6.2 资金方专属（仅 `identity=funder` 显示且必填）

| 字段 | 说明 |
|------|------|
| amount_wan | 额度（万元） |
| rate_type | `daily` / `yearly` |
| rate_percent | 利率（%） |
| period_count | 还款期数 |
| repay_method | 等额本息 / 等额本金 |

非资金方上述字段为 NULL，表单不展示。

模板：`web/templates/user/project_create.html`

---

## 7. 项目列表

路径：`/my/projects`

按身份展示列：资金方含额度/利率/期数；需求方/供应商含区域。支持状态筛选与删除（待审核、已驳回）。

---

## 8. 企业认证

上传企业证明照片后**即时** `verify_status = approved`，无需管理员审核。

---

## 9. 公开页

| 路径 | 说明 |
|------|------|
| GET `/` | 已审核项目列表（分类筛选 + 搜索 + 无限滚动） |
| GET `/projects/:id` | 项目详情；资金方展示金融字段，其他身份隐藏 |
| GET `/api/projects/list` | 分页 JSON（兼容 `/api/products/list`） |

模板：`web/templates/public/index.html`、`project_detail.html`

---

## 10. 管理端

| 功能 | 路径 |
|------|------|
| 管理首页 | GET `/admin`（汇总 + 一级/二级分类信息数、**项目数**） |
| **项目审核** | GET `/admin/project-review`，POST `/api/admin/projects/:id/review` |
| 信息审核 | GET `/admin/review` |
| 用户导出 | GET `/api/admin/users/export` |

模板：`web/templates/admin/project_review.html`

---

## 11. 数据模型摘要

### companies（Company）

新增 `established_at VARCHAR(7)`；`regions` 格式按身份区分。

### projects（Project）

金融字段（`amount_wan`、`rate_type`、`rate_percent`、`period_count`、`repay_method`）改为**可空**；非资金方发布时为 NULL。

---

## 12. 代码索引

| 模块 | 路径 |
|------|------|
| 省市区与校验 | `internal/data/regions.go` |
| 公司/项目模型 | `internal/model/company.go`、`project.go` |
| 地区格式迁移 | `internal/database/migrate_regions.go` |
| 表名迁移 | `internal/database/migrate_tables.go` |
| 用户中心 | `internal/handler/user_center.go` |
| 公开页/API | `internal/handler/product_public.go` |
| 分类统计 | `internal/handler/stats.go` |
| 路由 | `cmd/server/main.go` |

---

## 13. 验收标准

- [x] 三种身份在用户中心看到「项目管理 / 公司信息 / 添加项目 / 项目列表」
- [x] 资金方：公司信息与项目均可多选省；表单含金融字段
- [x] 需求方、设备供应商：公司信息与项目均为单选省市；无金融字段
- [x] 公司信息可保存成立时间（YYYY-MM）
- [x] 审核通过的项目在首页按身份正确展示
- [x] 旧 `/my/products`、`/products/:id` 链接可跳转
