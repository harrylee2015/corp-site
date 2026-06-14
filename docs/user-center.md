# 用户中心与产品管理设计文档

> 版本：v3 · 更新日期：2026-06  
> 本文档描述用户身份体系、分类映射、用户中心、店铺、产品管理、企业认证、管理端统计/导出及公开页隐私的设计与实现约定。

---

## 1. 设计背景

平台从「单一信息发布」升级为「**身份驱动 + 店铺 + 金融产品**」模式：

- 用户注册时选定**业务身份**（需求方 / 设备供应商 / 资金方），身份注册后**不可修改**
- 所有企业用户须完成**企业认证**后方可发布产品
- 用户登录后进入**用户中心**（左侧导航），管理店铺与产品
- 首页仍保留原有「信息广场」（posts），与新「产品（products）」并行；产品为主力业务流

---

## 2. 概念区分

| 概念 | 字段/值 | 说明 |
|------|---------|------|
| 系统角色 | `users.role` | `user` / `admin`，控制登录权限 |
| 业务身份 | `users.identity` | `demander` / `supplier` / `funder`，注册时选定，**只读** |
| 行业 | 一级分类 | categories 树中 `parent_id IS NULL` 的节点 |
| 类别 | 二级分类 | 一级下的叶子分类；「其他类」无二级，本身为叶子 |

### 2.1 业务身份枚举

| 值 | 显示名 | 定位 |
|----|--------|------|
| `demander` | 需求方 | 有项目/设备/融资需求 |
| `supplier` | 设备供应商 | 提供设备、工程及相关服务 |
| `funder` | 资金方 | 提供融资、租赁、收购等资金服务 |

配置位置：`internal/identity/identity.go`

---

## 3. 身份 → 分类映射（行业 / 类别）

用户身份决定**店铺**和**产品**中可选的分类范围。一级分类作为「可做行业」，二级分类作为「可做的类型」，均支持多选（店铺）或单选（产品）。

| 身份 | 可选一级分类（行业） | 说明 |
|------|---------------------|------|
| 需求方 | 新能源项目、企业类项目、其他类 | 发布需求类信息 |
| 设备供应商 | 新能源项目、企业类项目、电站出售方、其他类 | 供应设备，可参与电站出售 |
| 资金方 | 租赁公司、企业类项目、电站收购方、其他类 | 提供资金、租赁、收购服务 |

### 3.1 完整分类树

```
全部
├── 新能源项目
│   ├── 光伏发电、储能电站、风力发电、垃圾发电
│   ├── 水利发电、充电桩、光储充一体化项目
├── 企业类项目
│   ├── 央国企设备租赁、上市公司设备租赁、中小微企业设备租赁
├── 电站出售方（二级同新能源项目 7 项）
├── 电站收购方（二级同新能源项目 7 项）
├── 租赁公司
│   ├── 金租、商租、外资
└── 其他类（无二级，本身为叶子）
```

### 3.2 首页导航

- 顶部导航栏：**全部 + 6 个一级分类**，同一行分段等宽布局
- 有二级的一级：鼠标悬停展开下拉，点击二级筛选 `/?category_id={叶子ID}`
- 「其他类」无下拉，直接点击筛选
- 分类数据：`categories.parent_id` 两级树；启动时自动从旧 6 分类迁移

旧分类自动映射规则见 `internal/database/categories.go`。

---

## 4. 用户注册

### 4.1 注册字段

| 字段 | 必填 | 校验 |
|------|------|------|
| 手机号 | 是 | 11 位，唯一 |
| 短信验证码 | 是 | 6 位，5 分钟有效 |
| 登录密码 | 是 | 8–20 位，含字母+数字 |
| 确认密码 | 是 | 与密码一致 |
| 真实姓名 | 是 | 2–20 字符，写入 `real_name` |
| 企业名称 | 是 | 2–100 字符，写入 `company` |
| 用户身份 | 是 | 三选一，写入 `identity` |

### 4.2 规则

- 身份注册后**不可修改**（基本信息页只读展示）
- 对外展示优先使用**真实姓名**（`real_name`），不再依赖昵称
- 注册成功自动登录，跳转 `/my` 用户中心
- 初始企业认证状态：`verify_status = none`

---

## 5. 企业认证

**所有企业用户**须上传企业证明照片后方可**添加产品**。上传后**即时完成认证**，无需管理员审核。

| 状态 | 值 | 说明 |
|------|-----|------|
| 未认证 | `none` | 初始状态 |
| 已认证 | `approved` | 上传照片后自动设置 |

- 用户入口：用户中心 → **基本信息** → 上传企业证明照片（仅图片：jpg/png/gif/webp）
- 证明材料路径：`users.verify_doc_path`
- 管理端审核接口已废弃，不再使用

---

## 6. 用户中心

### 6.1 布局

- 模板：`web/templates/layout/user.html`
- 登录后顶部简栏 + **左侧竖向导航** + 右侧内容区
- 公开页仍使用 `layout/base.html`（含分类导航栏、友情链接、页脚）

### 6.2 导航结构

```
用户首页          GET  /my
产品管理
  ├ 店铺信息      GET  /my/shop          POST /api/my/shop
  ├ 添加产品      GET  /my/products/new  POST /api/my/products
  └ 产品列表      GET  /my/products      POST /api/my/products/:id/delete
基本信息          GET  /my/profile
                  POST /api/my/password
                  POST /api/my/verify
```

### 6.3 用户首页（/my）

- 统计：全部产品数、已发布数、待审核数
- 提示：企业认证状态、店铺是否已完善
- 快捷入口：完善店铺、添加产品、产品列表

---

## 7. 店铺信息

每用户一条店铺记录（`shops.user_id` 唯一）。

| 字段 | 说明 |
|------|------|
| shop_name | 店铺名称 |
| regions | JSON 数组，可做区域（省份多选） |
| category_ids | JSON 数组，可做的二级分类 ID（按身份过滤后多选） |
| contact / phone / tel | 联系人、手机、电话 |
| address | 公司地址 |
| intro | 公司介绍 |
| banner_path | Banner 图路径（上传至 /uploads） |

- 可做区域数据源：`internal/data/regions.go`（省级列表）
- 可做类型：根据 `users.identity` 动态展示可选分类树
- 操作：**重置** / **确认保存**

---

## 8. 添加产品

前置条件：`users.verify_status = approved`

| 字段 | 说明 |
|------|------|
| name | 产品名称 |
| category_id | 二级分类（按身份过滤） |
| image_path | 产品图片 |
| amount_wan | 额度，单位：万元 |
| rate_type | `daily` 日利率 / `yearly` 年利率 |
| rate_percent | 利率数值（%） |
| period_count | 还款期数（按月，如 12 = 1 年 12 期） |
| repay_method | `equal_installment` 等额本息 / `equal_principal` 等额本金 |
| regions | JSON 数组，可做区域（省份多选） |
| intro | 产品介绍 |
| status | 提交后 `pending`，管理员审核 |

---

## 9. 产品列表

路径：`/my/products`

| 列 | 说明 |
|----|------|
| 产品名称 | |
| 分类 | 一级 · 二级 |
| 额度(万) | |
| 利率 | 日/年 + 数值 |
| 期数 | |
| 状态 | 待审核 / 已发布 / 已驳回 / 已下架 |
| 时间 | 创建时间 |
| 操作 | 待审核、已驳回可删除 |

支持按状态筛选。

---

## 10. 基本信息

| 模块 | 说明 |
|------|------|
| 账号信息 | 手机、真实姓名、企业名称、身份（只读） |
| 企业认证 | 状态展示 + 上传照片（上传即 `approved`） |
| 修改密码 | 原密码 + 新密码 + 确认 |

---

## 11. 管理端扩展

| 功能 | 路径 |
|------|------|
| 管理首页统计 | GET `/admin`（汇总 + **一级/二级分类**信息数、产品数明细） |
| 信息审核（posts） | GET `/admin/review` |
| **产品审核** | GET `/admin/product-review`，POST `/api/admin/products/:id/review` |
| 信息导出 | GET `/api/admin/export` |
| **用户导出** | GET `/api/admin/users/export`（用户管理页入口，支持 keyword 筛选） |
| 用户管理 | GET `/admin/users` |

产品审核流程与信息审核一致：通过 → `approved`；驳回 → `rejected` + `reject_reason`。

---

## 12. 数据模型

### 12.1 users（扩展字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| real_name | VARCHAR(50) | 真实姓名 |
| identity | VARCHAR(20) | demander / supplier / funder |
| verify_status | VARCHAR(15) | none / approved |
| verify_doc_path | VARCHAR(500) | 企业证明材料路径 |

### 12.2 categories（扩展）

| 字段 | 类型 | 说明 |
|------|------|------|
| parent_id | INT NULL | NULL=一级导航；有值=二级 |
| name | VARCHAR(50) | 同级唯一（parent_id + name 联合唯一） |
| sort_order | INT | 排序 |

### 12.3 shops

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID PK | |
| user_id | UUID UNIQUE | 所属用户 |
| shop_name | VARCHAR(100) | 店铺名称 |
| regions | TEXT | JSON 省份数组 |
| category_ids | TEXT | JSON 分类 ID 数组 |
| contact, phone, tel, address | | 联系信息 |
| intro | TEXT | 介绍 |
| banner_path | VARCHAR(500) | Banner |

### 12.4 products

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID PK | |
| user_id | UUID FK | 发布人 |
| category_id | INT FK | 叶子分类 |
| name | VARCHAR(200) | 产品名称 |
| image_path | VARCHAR(500) | 产品图 |
| amount_wan | FLOAT | 额度（万元） |
| rate_type | VARCHAR(10) | daily / yearly |
| rate_percent | FLOAT | 利率 % |
| period_count | INT | 期数 |
| period_unit | VARCHAR(10) | 默认 month |
| repay_method | VARCHAR(30) | 还款方式 |
| regions | TEXT | JSON 省份数组 |
| intro | TEXT | 介绍 |
| status | VARCHAR(15) | pending / approved / rejected / delisted |
| reject_reason, reviewed_by, reviewed_at | | 审核信息 |

---

## 13. 业务流程

```
注册（选身份）
  → 登录 → 用户中心 /my
  → 基本信息：上传企业证明照片（即时认证）
  → 店铺信息：完善店铺
  → 添加产品 → 管理员产品审核 → 已发布
```

### 13.1 产品状态机

```
pending → approved（管理员通过）
       → rejected（管理员驳回，用户可见原因，可删除重发）
approved → delisted（下架，预留）
```

---

## 14. 公开页隐私保护

用户端对外展示的信息（首页、信息详情）遵循以下规则：

| 项目 | 规则 |
|------|------|
| 联系人 | 匿名：`MaskName`（如 李四 → 李*） |
| 联系电话 | 匿名：`MaskPhone`（138****1234） |
| 发布者 | 公开页匿名；本人与管理员可见完整信息 |
| 附件 | 公开页不展示、不可下载；仅发布者本人与管理员可查看 |
| 上传文件 | `/uploads/*` 鉴权访问，他人无法查看认证照、Banner、产品图等 |

实现：`internal/handler/upload_serve.go`、`MaskName`/`MaskPhone`（`internal/handler/admin.go`）

---

## 15. 与原有功能的关系

| 模块 | 状态 | 说明 |
|------|------|------|
| posts 信息发布 | 保留 | `/my/posts` 旧流程仍可用 |
| products 产品 | 新增 | 用户中心主流程 |
| 首页列表 | posts | 产品审核通过后首页展示待后续迭代 |
| 文章管理 | 未做 | 需求中未纳入本期范围 |

---

## 16. 代码索引

| 模块 | 路径 |
|------|------|
| 身份与分类映射 | `internal/identity/identity.go` |
| 分类迁移/Seed | `internal/database/categories.go` |
| 分类统计 | `internal/handler/stats.go` |
| 用户/信息导出 | `internal/handler/export.go` |
| 上传鉴权 | `internal/handler/upload_serve.go` |
| 用户中心 Handler | `internal/handler/user_center.go` |
| 用户中心布局 | `web/templates/layout/user.html` |
| 店铺/产品/资料页 | `web/templates/user/` |
| 产品审核 | `web/templates/admin/product_review.html` |
| 管理首页统计 | `web/templates/admin/dashboard.html` |
| 省份数据 | `internal/data/regions.go` |

---

## 17. 待办（后续迭代）

- [ ] 已审核产品展示到首页广场
- [ ] 店铺/产品在公开页的详情页
- [ ] 文章管理模块（如需要）
