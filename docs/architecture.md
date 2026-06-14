# 系统架构文档

## 1. 项目概述

构建一个信息收集与发布平台（类似猪八戒），品牌 **金筹设备租赁**。支持：

- 用户按**业务身份**（需求方 / 设备供应商 / 资金方）注册，进入**用户中心**管理公司信息与项目
- **企业认证**：上传企业证明照片即可发布项目（无需后台审核）
- **管理统计**：按一级/二级分类统计信息与项目数量
- **数据导出**：信息 Excel 导出 + 用户 Excel 导出
- **公开页隐私**：联系人匿名、附件不可下载、上传文件鉴权访问
- 保留原有 **posts 信息发布**流程，与新 **projects 项目**体系并行
- 两级**分类导航**（6 个一级 + 悬停下拉二级），启动时自动从旧分类迁移
- 管理员审核、Excel 导出、用户管理

> 用户中心详细设计见 [user-center.md](./user-center.md)

## 2. 技术选型

| 层面 | 技术 | 版本 | 说明 |
|------|------|------|------|
| **后端语言** | Go | 1.22+ | 编译型、高性能、部署简单 |
| **Web 框架** | Gin | v1.9+ | 高性能 HTTP 框架，中文社区活跃 |
| **ORM** | GORM | v1.25+ | Go 生态最成熟 ORM |
| **数据库** | PostgreSQL | 15+ | 支持 JSON 字段、丰富索引、全文检索 |
| **认证** | JWT (access + refresh) | v5+ | 双 token 机制，用户端和管理端独立签发 |
| **密码加密** | bcrypt | - | golang.org/x/crypto |
| **Excel** | excelize | v2.8+ | 纯 Go 实现，读写 .xlsx |
| **短信** | 腾讯云 SMS SDK | latest | 注册/登录验证码 |
| **配置** | viper | v1.18+ | YAML 配置文件加载 |
| **前端渲染** | Go html/template | 标准库 | 服务端渲染(SSR) |
| **CSS 框架** | Tailwind CSS | v3.4+ | 原子化 CSS，响应式优先 |
| **UUID** | google/uuid | v1.6+ | 主键生成 |

## 3. 系统架构图

```
┌─────────────────────────────────────────────────────────┐
│                      用户浏览器 / 手机浏览器               │
├──────────────────────────┬──────────────────────────────┤
│       用户端 (/)          │      管理端 (/admin)          │
├──────────────────────────┴──────────────────────────────┤
│                      Nginx (可选)                         │
├─────────────────────────────────────────────────────────┤
│                   Go Application (Gin)                    │
│  ┌──────────┬──────────┬──────────┬──────────────────┐  │
│  │ Handler  │ Service  │ Model    │ Middleware        │  │
│  │  Layer   │  Layer   │  Layer   │  Layer            │  │
│  │          │          │          │  JWT / Role /     │  │
│  │  auth    │  auth    │  user    │  RateLimit /      │  │
│  │  post    │  post    │  category│  CSRF             │  │
│  │  admin   │  admin   │  post    │                   │  │
│  │  user_   │  export  │  shop    │                   │  │
│  │  center  │  export  │  product │                   │  │
│  │  export  │  sms     │  sms_log │                   │  │
│  │  upload  │  sms     │  attach  │                   │  │
│  └──────────┴──────────┴──────────┴──────────────────┘  │
│  ┌──────────────────────────────────────────────────┐   │
│  │              Template Engine (html/template)       │   │
│  │       SSR 渲染，内嵌 tailwind.min.css + 静态资源    │   │
│  └──────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────┤
│                    PostgreSQL 15+                        │
│  ┌────────┬────────┬──────┬──────┬────────┬───────────┐     │
│  │ users  │ posts  │ cats │ sms  │ shops  │ products  │     │
│  └────────┴────────┴──────┴──────┴────────┴───────────┘     │
├─────────────────────────────────────────────────────────┤
│                    文件系统 (/uploads)                    │
└─────────────────────────────────────────────────────────┘
```

## 4. 分层架构

```
┌─────────────────────────────────────────────┐
│  Handler Layer (HTTP 请求处理)               │
│  - 参数绑定与校验                             │
│  - 调用 Service 层                           │
│  - 返回 JSON 响应 或 渲染 HTML 模板           │
├─────────────────────────────────────────────┤
│  Service Layer (业务逻辑)                     │
│  - 业务规则实现                               │
│  - 事务管理                                  │
│  - 多表操作协调                               │
├─────────────────────────────────────────────┤
│  Model Layer (数据模型)                       │
│  - GORM 模型定义                              │
│  - 数据库表映射                               │
│  - 关联关系定义                               │
├─────────────────────────────────────────────┤
│  Database Layer (数据访问)                    │
│  - GORM 连接管理                              │
│  - AutoMigrate 自动建表                       │
│  - 初始数据填充                               │
└─────────────────────────────────────────────┘
```

## 5. 路由架构

```
# ===== 公开路由(无需认证) =====
GET    /                          # 首页(已审核信息列表+分类导航筛选+搜索)
GET    /posts/:id                 # 信息详情页
GET    /login                     # 用户登录页
GET    /register                  # 用户注册页
POST   /api/sms/send              # 发送短信验证码
POST   /api/auth/register         # 用户注册
POST   /api/auth/login            # 用户登录

# ===== 用户中心(需 JWT, role=user) =====
GET    /my                        # 用户中心首页
GET    /my/company                # 公司信息
POST   /api/my/company            # 保存公司信息
GET    /my/projects/new           # 添加项目
GET    /my/projects               # 项目列表
POST   /api/my/projects           # 创建项目
POST   /api/my/projects/:id/delete # 删除项目
GET    /projects/:id              # 项目详情（公开）
GET    /api/projects/list         # 项目分页 API
# 旧路径 /my/shop、/my/products/*、/products/:id 302 重定向至新路径
GET    /my/profile                # 基本信息
POST   /api/my/password           # 修改密码
POST   /api/my/verify             # 提交企业认证

# ===== 用户端-旧版 posts(需 JWT, role=user) =====
GET    /my/posts                  # 我的发布列表
GET    /my/posts/new              # 发布新信息表单
POST   /api/posts                 # 提交新信息
GET    /api/posts/:id             # 信息详情(我的)
DELETE /api/posts/:id             # 删除信息(pending状态)
POST   /api/upload                # 上传附件
GET    /logout                    # 登出

# ===== 管理端页面 =====
GET    /admin/login               # 管理员登录页

# ===== 管理端(需 JWT, role=admin) =====
GET    /admin                     # 管理后台首页(汇总统计+分类明细统计)
GET    /admin/review              # 信息待审核列表
GET    /admin/project-review      # 项目待审核列表
GET    /admin/posts               # 全部信息管理
GET    /admin/export              # 导出筛选页
GET    /admin/users               # 用户列表
POST   /api/admin/login           # 管理员登录
POST   /api/admin/posts/:id/review   # 信息审核
POST   /api/admin/projects/:id/review # 项目审核
GET    /api/admin/users/export   # 导出用户 Excel
POST   /api/admin/export          # 导出信息 Excel
PUT    /api/admin/users/:id/status # 启用/禁用用户

# ===== 上传与静态 =====
GET    /uploads/*filepath        # 鉴权访问上传文件（OptionalAuth + 所有权校验）
GET    /static/*filepath          # 静态资源
```

## 6. 认证架构

```
┌──────────────┐     ┌──────────────┐
│   用户登录    │     │  管理员登录   │
└──────┬───────┘     └──────┬───────┘
       │                    │
       ▼                    ▼
┌──────────────────────────────────────┐
│           POST /api/auth/login        │
│           POST /api/admin/login       │
│  - 校验手机号 + 密码                   │
│  - 查询 users 表，检查 role 字段       │
│  - 用户登录仅允许 role=user            │
│  - 管理员登录仅允许 role=admin         │
│  - 生成 JWT{ user_id, role, exp }     │
│  - 返回 access_token + refresh_token  │
└──────────────────┬───────────────────┘
                   │
                   ▼
┌──────────────────────────────────────┐
│         JWT Middleware                │
│  - 解析 Authorization: Bearer <token> │
│  - 验证签名 + 过期时间                 │
│  - 注入 user_id / role 到 Context     │
└──────────────────┬───────────────────┘
                   │
                   ▼
┌──────────────────────────────────────┐
│         Role Middleware               │
│  - 检查 Context 中的 role 字段        │
│  - 匹配所需角色，否则返回 403          │
└──────────────────────────────────────┘
```

### Token 配置

| 参数 | 值 | 说明 |
|------|-----|------|
| access_token 有效期 | 2h | 短时效，频繁使用 |
| refresh_token 有效期 | 168h (7天) | 长时效，换取新 access_token |
| 签名算法 | HS256 | HMAC-SHA256 |

## 7. 数据库 ER 图

```
┌──────────────┐       ┌──────────────┐
│    users     │       │  categories  │
├──────────────┤       ├──────────────┤
│ id       PK  │       │ id       PK  │
│ phone        │       │ parent_id FK │──┐ 自引用(两级树)
│ password_hash│       │ name         │  │
│ role         │       │ sort_order   │  │
│ nickname     │       │ created_at   │  │
│ real_name    │       └──────┬───────┘  │
│ company      │              │◄─────────┘
│ identity     │              │
│ verify_status│              │
│ verify_doc   │              │
│ status       │              │
│ created_at   │              │
│ updated_at   │              │
└──────┬───────┘              │
       │                      │
       │ 1:1                  │
       ▼                      │
┌──────────────┐              │
│    shops     │              │
├──────────────┤              │
│ id       PK  │              │
│ user_id  FK  │              │
│ shop_name    │              │
│ regions JSON │              │
│ category_ids │              │
│ contact...   │              │
│ banner_path  │              │
└──────────────┘              │
       │                      │
       │ 1:N                  │
       ▼                      ▼
┌──────────────────────────────────────────┐
│                 products                  │
├──────────────────────────────────────────┤
│ id           PK                          │
│ user_id      FK → users.id              │
│ category_id  FK → categories.id         │
│ name, image_path, amount_wan, rate_*    │
│ period_*, repay_method, regions         │
│ intro, status, reject_reason            │
│ reviewed_by, reviewed_at                │
└──────────────────────────────────────────┘

       │ 1:N (posts 旧版)
       ▼
┌──────────────────────────────────────────┐
│                 posts                     │
├──────────────────────────────────────────┤
│ id           PK                          │
│ user_id      FK → users.id              │
│ category_id  FK → categories.id         │
│ title, content, contact, contact_phone  │
│ status, reject_reason, reviewed_*       │
└──────────────────┬───────────────────────┘
                   │ 1:N
                   ▼
┌──────────────────────────────────────────┐
│              attachments                  │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│               sms_logs                    │
└──────────────────────────────────────────┘
```

### 身份与分类映射

业务身份（`users.identity`）决定店铺/产品可选分类，配置于 `internal/identity/identity.go`：

| 身份 | 可选一级分类 |
|------|-------------|
| demander（需求方） | 新能源项目、企业类项目、其他类 |
| supplier（设备供应商） | 新能源项目、企业类项目、电站出售方、其他类 |
| funder（资金方） | 租赁公司、企业类项目、电站收购方、其他类 |

## 8. 部署架构

```
┌─────────────────────────────────────────────┐
│              生产环境服务器                    │
│  ┌───────────────────────────────────────┐  │
│  │           Nginx (端口 80/443)          │  │
│  │  - SSL 终结                            │  │
│  │  - 静态文件缓存                         │  │
│  │  - 反向代理到 Go 服务                   │  │
│  └───────────────┬───────────────────────┘  │
│                  │                           │
│  ┌───────────────▼───────────────────────┐  │
│  │      Go 应用 (端口 8080)               │  │
│  │  - 单一可执行文件                       │  │
│  │  - systemd 管理                       │  │
│  └───────────────┬───────────────────────┘  │
│                  │                           │
│  ┌───────────────▼───────────────────────┐  │
│  │         PostgreSQL (端口 5432)         │  │
│  └───────────────────────────────────────┘  │
│                                              │
│  ┌───────────────────────────────────────┐  │
│  │         /opt/corp-site/uploads/        │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### 单机部署特性
- Go 编译为单一二进制文件，内嵌模板和静态资源(通过 `embed` 包)
- 仅依赖 PostgreSQL 数据库，无需 Redis/MQ 等中间件
- systemd 管理进程，支持自动重启

## 9. 项目目录结构

```
corp-site/
├── cmd/
│   └── server/
│       └── main.go                  # 入口文件
├── internal/
│   ├── config/
│   │   └── config.go                # 配置结构体 + viper 加载
│   ├── database/
│   │   ├── postgres.go              # GORM 连接单例
│   │   ├── migrate.go               # AutoMigrate + 初始化数据
│   │   └── categories.go            # 分类树 Seed + 旧分类迁移
│   ├── identity/
│   │   └── identity.go              # 业务身份枚举 + 分类映射
│   ├── data/
│   │   └── regions.go               # 省份列表
│   ├── model/
│   │   ├── user.go                  # User GORM 模型
│   │   ├── category.go              # Category GORM 模型(两级树)
│   │   ├── shop.go                  # Shop GORM 模型
│   │   ├── product.go               # Product GORM 模型
│   │   ├── post.go                  # Post GORM 模型
│   │   ├── attachment.go            # Attachment GORM 模型
│   │   └── sms_log.go               # SmsLog GORM 模型
│   ├── handler/
│   │   ├── auth.go                  # 认证 Handler
│   │   ├── post.go                  # 信息发布 Handler
│   │   ├── user_center.go           # 用户中心 Handler
│   │   ├── admin.go                 # 管理端 Handler
│   │   ├── export.go                # 信息/用户 Excel 导出
│   │   ├── stats.go                 # 分类维度统计
│   │   ├── upload_serve.go          # 上传文件鉴权访问
│   │   └── render.go                # 页面渲染辅助
│   ├── middleware/
│   │   ├── jwt.go                   # JWT 校验中间件
│   │   ├── role.go                  # 角色校验中间件
│   │   └── ratelimit.go             # 频率限制中间件
│   ├── service/
│   │   ├── auth.go                  # 认证业务逻辑
│   │   ├── post.go                  # 信息业务逻辑
│   │   ├── admin.go                 # 管理业务逻辑
│   │   └── export.go                # 导出业务逻辑
│   └── sms/
│       └── tencent.go               # 腾讯云短信封装
├── web/
│   ├── templates/
│   │   ├── layout/
│   │   │   ├── base.html            # 用户端公共布局(含分类导航)
│   │   │   ├── user.html            # 用户中心布局(左侧导航)
│   │   │   └── admin.html           # 管理端公共布局
│   │   ├── public/
│   │   │   ├── index.html           # 首页
│   │   │   └── post_detail.html     # 信息详情页
│   │   ├── user/
│   │   │   ├── login.html           # 用户登录
│   │   │   ├── register.html        # 用户注册
│   │   │   ├── center_home.html     # 用户中心首页
│   │   │   ├── shop.html            # 店铺信息
│   │   │   ├── product_create.html  # 添加产品
│   │   │   ├── product_list.html    # 产品列表
│   │   │   ├── profile.html         # 基本信息
│   │   │   ├── dashboard.html       # 我的发布(旧版)
│   │   │   └── post_create.html     # 发布信息(旧版)
│   │   └── admin/
│   │       ├── login.html           # 管理员登录
│   │       ├── dashboard.html       # 管理首页
│   │       ├── review.html          # 信息待审核
│   │       ├── product_review.html  # 产品待审核
│   │       ├── posts.html           # 全部信息管理
│   │       └── export.html          # 导出筛选
│   └── static/
│       └── css/
│           └── tailwind.min.css     # Tailwind 编译产物
├── uploads/                          # 上传文件存储目录
├── docs/
│   ├── architecture.md              # 本文档
│   ├── features.md                  # 功能点文档
│   └── user-center.md               # 用户中心与产品管理设计
├── config.yaml                      # 配置文件
├── go.mod
├── go.sum
└── Makefile                         # 构建脚本
```

## 10. 安全设计

| 安全措施 | 实现方式 |
|---------|---------|
| 密码存储 | bcrypt 哈希，cost=10 |
| 接口认证 | JWT Bearer Token，含过期时间 |
| 角色隔离 | 用户端和管理端路由分离，中间件校验 role |
| SQL 注入防护 | GORM 参数化查询 |
| XSS 防护 | Go 模板自动 HTML 转义 |
| 短信防刷 | 同一手机号 N分钟内限制发送次数 |
| 文件上传安全 | 类型白名单 + 大小限制 + UUID 文件名 + 路径遍历防护 |
| **文件访问控制** | `/uploads/*` 鉴权访问；管理员全权；普通用户仅本人资源 |
| **公开页隐私** | 联系人/电话/发布者匿名（`MaskName`/`MaskPhone`）；公开页不展示附件 |
| CSRF | 表单提交使用 CSRF Token |

### 10.1 管理端分类统计

`LoadCategoryStats()`（`internal/handler/stats.go`）按一级、二级分类分别统计 **posts** 与 **products** 数量，在管理首页表格展示。

### 10.2 上传文件访问规则

| 资源类型 | 允许访问 |
|---------|---------|
| 信息附件 | 发布者本人、管理员 |
| 企业认证照 | 用户本人、管理员 |
| 店铺 Banner | 店铺所属用户、管理员 |
| 产品图片 | 产品所属用户、管理员 |
| 未关联 post 的上传 | 任意已登录用户（发布流程中的临时文件） |
| 未登录 | 全部拒绝 |

## 11. 性能优化

- 首页列表使用分页查询(LIMIT/OFFSET)
- 审核通过信息列表使用条件索引(`WHERE status = 'approved'`)
- 静态资源(CSS/JS)设置强缓存(Nginx)
- 模板文件编译期间通过 embed 内嵌，减少 IO
