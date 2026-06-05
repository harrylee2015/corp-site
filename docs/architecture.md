# 系统架构文档

## 1. 项目概述

构建一个信息收集与发布平台(类似猪八戒)，支持用户发布各类商业信息，管理员审核后对外展示，并提供数据导出功能。

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
│  │  export  │  export  │  sms_log │                   │  │
│  │  upload  │  sms     │  attach  │                   │  │
│  └──────────┴──────────┴──────────┴──────────────────┘  │
│  ┌──────────────────────────────────────────────────┐   │
│  │              Template Engine (html/template)       │   │
│  │       SSR 渲染，内嵌 tailwind.min.css + 静态资源    │   │
│  └──────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────┤
│                    PostgreSQL 15+                        │
│  ┌────────┬────────┬──────┬──────┬───────────────┐     │
│  │ users  │ posts  │ cats │ sms  │ attachments   │     │
│  └────────┴────────┴──────┴──────┴───────────────┘     │
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
GET    /                          # 首页(已审核信息列表+分类筛选+搜索)
GET    /posts/:id                 # 信息详情页
GET    /login                     # 用户登录页
GET    /register                  # 用户注册页
POST   /api/sms/send              # 发送短信验证码
POST   /api/auth/register         # 用户注册
POST   /api/auth/login            # 用户登录

# ===== 用户端(需 JWT, role=user) =====
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
GET    /admin                     # 管理后台首页(统计)
GET    /admin/review              # 待审核列表
GET    /admin/posts               # 全部信息管理
GET    /admin/export              # 导出筛选页
GET    /admin/users               # 用户列表
POST   /api/admin/login           # 管理员登录
POST   /api/admin/posts/:id/review # 审核操作
POST   /api/admin/export          # 生成并下载 Excel
PUT    /api/admin/users/:id/status # 启用/禁用用户

# ===== 静态文件 =====
GET    /uploads/:filename         # 访问上传文件
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
│ phone        │       │ name         │
│ password_hash│       │ sort_order   │
│ role         │       │ created_at   │
│ nickname     │       └──────┬───────┘
│ company      │              │
│ status       │              │
│ created_at   │              │
│ updated_at   │              │
└──────┬───────┘              │
       │                      │
       │  ┌───────────────────┘
       │  │
       ▼  ▼
┌──────────────────────────────────────────┐
│                 posts                     │
├──────────────────────────────────────────┤
│ id           PK                          │
│ user_id      FK → users.id              │
│ category_id  FK → categories.id         │
│ title                                    │
│ content                                  │
│ contact                                  │
│ contact_phone                            │
│ status       (pending/approved/rejected) │
│ reject_reason                            │
│ reviewed_by  FK → users.id (nullable)    │
│ reviewed_at  (nullable)                  │
│ created_at                               │
│ updated_at                               │
└──────────────────┬───────────────────────┘
                   │
                   │ 1:N
                   ▼
┌──────────────────────────────────────────┐
│              attachments                  │
├──────────────────────────────────────────┤
│ id           PK                          │
│ post_id      FK → posts.id ON DELETE     │
│ file_name                                │
│ file_path                                │
│ file_size                                │
│ created_at                               │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│               sms_logs                    │
├──────────────────────────────────────────┤
│ id           PK                          │
│ phone                                    │
│ code                                     │
│ scene        (register/login/reset)      │
│ expired_at                               │
│ used                                     │
│ created_at                               │
└──────────────────────────────────────────┘
```

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
│   │   └── migrate.go               # AutoMigrate + 初始化数据
│   ├── model/
│   │   ├── user.go                  # User GORM 模型
│   │   ├── category.go              # Category GORM 模型
│   │   ├── post.go                  # Post GORM 模型
│   │   ├── attachment.go            # Attachment GORM 模型
│   │   └── sms_log.go               # SmsLog GORM 模型
│   ├── handler/
│   │   ├── auth.go                  # 认证 Handler
│   │   ├── post.go                  # 信息发布 Handler
│   │   ├── admin.go                 # 管理端 Handler
│   │   ├── export.go                # 导出 Handler
│   │   └── upload.go                # 上传 Handler
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
│   │   │   ├── base.html            # 用户端公共布局
│   │   │   └── admin.html           # 管理端公共布局
│   │   ├── public/
│   │   │   ├── index.html           # 首页
│   │   │   └── post_detail.html     # 信息详情页
│   │   ├── user/
│   │   │   ├── login.html           # 用户登录
│   │   │   ├── register.html        # 用户注册
│   │   │   ├── dashboard.html       # 我的发布
│   │   │   └── post_create.html     # 发布信息
│   │   └── admin/
│   │       ├── login.html           # 管理员登录
│   │       ├── dashboard.html       # 管理首页
│   │       ├── review.html          # 待审核列表
│   │       ├── posts.html           # 全部信息管理
│   │       └── export.html          # 导出筛选
│   └── static/
│       └── css/
│           └── tailwind.min.css     # Tailwind 编译产物
├── uploads/                          # 上传文件存储目录
├── docs/
│   ├── architecture.md              # 本文档
│   └── features.md                  # 功能点文档
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
| 文件上传安全 | 校验文件类型白名单 + 大小限制 + 随机文件名 |
| CSRF | 表单提交使用 CSRF Token(可选) |

## 11. 性能优化

- 首页列表使用分页查询(LIMIT/OFFSET)
- 审核通过信息列表使用条件索引(`WHERE status = 'approved'`)
- 静态资源(CSS/JS)设置强缓存(Nginx)
- 模板文件编译期间通过 embed 内嵌，减少 IO
