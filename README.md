# 金筹设备租赁 · 信息收集平台 (Corp Site)

> 类似猪八戒的信息收集与发布平台，支持身份驱动的用户中心、店铺与金融产品管理、管理员审核、Excel 报表导出，适配手机浏览器。

## 项目背景

构建一个企业级信息收集平台，满足以下核心场景：

- **身份注册**：用户注册时选择业务身份（需求方 / 设备供应商 / 资金方），身份不可修改
- **用户中心**：登录后管理店铺、发布金融产品，须完成企业认证后方可添加产品
- **分类导航**：两级分类树（6 个一级 + 悬停下拉二级），首页分段导航栏
- **信息发布（旧版）**：保留 posts 流程，与 products 并行
- **审核机制**：信息/产品均需管理员审核通过后才能对外展示
- **数据导出**：管理员可按条件筛选并导出 Excel 报表
- **双端分离**：用户端和管理端独立登录，权限隔离
- **移动适配**：支持手机浏览器访问，响应式布局

> 用户中心详细设计：[docs/user-center.md](./docs/user-center.md)

## 技术栈

| 层面 | 技术 |
|------|------|
| 语言 | Go 1.24 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | PostgreSQL 16 |
| 认证 | JWT (access + refresh token) |
| 密码加密 | bcrypt |
| Excel 导出 | excelize |
| 短信服务 | 腾讯云 SMS (支持 Mock 模式) |
| 前端 | Go Template SSR + Tailwind CSS (CDN) |
| 部署 | Docker Compose |

## 功能清单

### 用户端

- [x] 手机号 + 短信验证码注册（真实姓名、企业名称、业务身份）
- [x] 手机号 + 密码登录，跳转用户中心
- [x] 首页分类导航 + 关键词搜索 + 无限滚动
- [x] 信息详情（公开页联系人/发布者匿名，附件不展示）
- [x] **用户中心**：首页统计、店铺信息、添加/管理产品、基本信息
- [x] **企业认证**：上传企业证明照片，即时完成认证
- [x] **公开页隐私**：联系人匿名、附件不可下载、上传文件鉴权访问
- [x] 发布信息（旧版 posts：分类/标题/内容/联系人/附件上传）
- [x] 我的发布（旧版，查看待审核/已通过/已驳回）
- [ ] 已审核产品展示到首页（待迭代）

### 管理端

- [x] 独立管理员登录页
- [x] 仪表盘（用户/信息/产品汇总 + **一级/二级分类统计**）
- [x] 信息审核（通过/驳回，驳回需填写原因）
- [x] **产品审核**
- [x] 信息管理（按分类/状态筛选，关键词搜索）
- [x] Excel 信息导出（数据明细 + 统计汇总）
- [x] **Excel 用户导出**（用户管理页，含店铺信息）
- [x] 用户管理（搜索/启用/禁用）

### 公共

- [x] 短信频率限制（每分钟/每小时）
- [x] 附件上传（类型白名单 + 大小限制 + 鉴权访问）
- [x] 公开页隐私（联系人/电话匿名，照片与附件不对他人开放）
- [x] 移动端响应式布局
- [x] 默认分类和默认管理员自动初始化

## 项目结构

```
corp-site/
├── cmd/server/main.go              # 入口
├── internal/
│   ├── config/config.go            # 配置 + 环境变量覆盖
│   ├── database/                   # GORM 连接 + 分类迁移
│   ├── identity/identity.go        # 业务身份 + 分类映射
│   ├── data/regions.go             # 省份列表
│   ├── model/                      # user, shop, product, category, post...
│   ├── handler/                    # auth, user_center, post, admin...
│   ├── middleware/jwt.go           # JWT 中间件
│   └── sms/sms.go                  # 短信封装
├── web/templates/                  # HTML 模板
│   ├── layout/                     # base, user, admin
│   ├── public/                     # 公开页面
│   ├── user/                       # 用户中心 + 旧版发布
│   └── admin/                      # 管理端页面
├── docs/
│   ├── architecture.md             # 架构文档
│   ├── features.md                 # 功能点文档
│   └── user-center.md              # 用户中心设计文档
├── config.yaml                     # 配置文件
├── Dockerfile
├── docker-compose.yml
├── .env                            # Docker 环境变量
├── Makefile
└── README.md
```

## 快速开始

### Docker Compose (推荐)

```bash
# 启动（首次自动初始化数据库）
docker compose up -d

# 访问
# 用户端：http://localhost:8080
# 管理端：http://localhost:8080/admin/login

# 停止
docker compose down

# 清除数据重新开始
docker compose down -v && docker compose up -d
```

### 本地开发

**前置条件**：Go 1.24+、PostgreSQL 16+

```bash
# 1. 创建数据库
createdb corp_site

# 2. 修改 config.yaml 中的数据库连接信息

# 3. 安装依赖
go mod tidy

# 4. 启动（首次自动建表 + 初始化数据）
go run ./cmd/server/

# 或编译后运行
make build && ./bin/server
```

### 默认账号

| 角色 | 手机号 | 密码 |
|------|--------|------|
| 管理员 | 13800000000 | Admin@123 |

### 默认分类（两级树）

一级：新能源项目、企业类项目、电站出售方、电站收购方、租赁公司、其他类。详见 [docs/user-center.md](./docs/user-center.md)。

## 配置说明

### config.yaml

```yaml
server:
  port: 8080          # 服务端口
  mode: debug         # debug | release

database:
  host: localhost
  port: 5432
  user: postgres
  password: ""
  dbname: corp_site
  sslmode: disable

jwt:
  secret: "your-secret" # 生产环境务必修改
  access_ttl: 2h
  refresh_ttl: 168h

sms:
  mock: true          # true=控制台打印验证码, false=调用腾讯云
  provider: tencent
  secret_id: ""
  secret_key: ""
  sdk_app_id: ""
  sign_name: ""
  template_id: ""

upload:
  path: ./uploads
  max_size: 10485760  # 10MB
  allowed_types: jpg,jpeg,png,pdf,doc,docx,xls,xlsx
```

### 环境变量 (Docker)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| DB_HOST | postgres | 数据库地址 |
| DB_PORT | 5432 | 数据库端口 |
| DB_USER | postgres | 数据库用户 |
| DB_PASSWORD | corp_site_2024 | 数据库密码 |
| DB_NAME | corp_site | 数据库名称 |
| SERVER_PORT | 8080 | 服务端口 |
| JWT_SECRET | config.yaml 中的值 | JWT 密钥 |

## 数据模型

![ER](./docs/er.svg)

| 表 | 说明 |
|----|------|
| users | 用户表（含身份、企业认证状态） |
| categories | 两级分类树 |
| shops | 店铺信息（每用户一条） |
| products | 金融产品（含审核状态） |
| posts | 信息发布表（旧版，含审核状态） |
| attachments | 附件表 |
| sms_logs | 短信验证码日志 |

## 状态流转

```
产品：pending → approved（首页展示待迭代）→ rejected

信息(posts)：pending → approved (首页展示) → rejected

企业认证：none → approved（上传照片即完成）

公开访问：联系人/电话/发布者匿名；附件与上传文件不对他人开放

用户：active ⇄ disabled (管理员操作)
```

## 短信配置

开发阶段 `sms.mock: true`，验证码在**控制台打印**并返回前端弹窗。

接入腾讯云短信：
1. 修改 `config.yaml` 中 `sms.mock: false`
2. 填入腾讯云控制台获取的 `secret_id`、`secret_key`、`sdk_app_id`、`sign_name`、`template_id`
3. 如需自定义模板参数，编辑 `internal/sms/sms.go` 中的 `TencentProvider.Send()`

## Makefile

```bash
make build    # 编译
make run      # 编译并运行
make clean    # 清除二进制
make tidy     # 整理依赖
make vet      # 静态检查
```

## License

MIT
