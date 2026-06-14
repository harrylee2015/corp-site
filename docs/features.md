# 功能点文档

## 1. 用户端功能

### 1.1 用户注册
| 功能点 | 详细说明 |
|-------|---------|
| 手机号输入 | 11位手机号，实时格式校验 |
| 短信验证码 | 调用腾讯云 SMS 发送6位数字验证码，有效期5分钟 |
| 倒计时 | 发送按钮60秒倒计时，防止重复点击 |
| 防刷限制 | 同一手机号每分钟限制发送1次，每小时限制5次 |
| 验证码校验 | 比对 sms_logs 表，校验未过期、未使用的验证码 |
| 密码设置 | 8-20位，须含字母+数字，bcrypt 加密存储 |
| 确认密码 | 须与登录密码一致 |
| 真实姓名 | 必填，2-20字符，写入 `real_name` |
| 企业名称 | 必填，2-100字符，写入 `company` |
| 用户身份 | 必填，三选一：需求方 / 设备供应商 / 资金方，写入 `identity`，**注册后不可修改** |
| 完成注册 | 写入 users 表，role=user，status=active，verify_status=none，自动登录并跳转 `/my` |

> 详细设计见 [user-center.md](./user-center.md)

### 1.2 用户登录
| 功能点 | 详细说明 |
|-------|---------|
| 登录方式 | 手机号 + 密码 |
| 身份校验 | 查询 users 表验证 phone + password_hash，仅允许 role=user |
| 账号状态 | 检查 status=active，禁用账号提示联系管理员 |
| Token 签发 | 返回 access_token(2h) + refresh_token(7天)，写入 Cookie |
| 登录成功 | 跳转到 `/my` 用户中心 |

### 1.3 首页(信息广场)
| 功能点 | 详细说明 |
|-------|---------|
| 分类导航 | 顶部导航栏「全部 + 6 个一级分类」，同一行分段等宽；有二级的一级悬停展开下拉 |
| 关键词搜索 | 搜索框，搜索标题和内容中匹配的关键词 |
| 列表展示 | 卡片式布局，每项显示：标题、分类标签、发布者（**匿名**）、发布时间 |
| 分页 | 底部无限滚动加载，每页12条 |
| 排序 | 默认按发布时间倒序 |
| 移动适配 | 小屏卡片占满宽度，大屏2-3列网格 |
| 状态过滤 | 仅展示 status=approved 的信息 |
| 品牌展示 | 站点品牌「金筹设备租赁」，顶部左侧客服热线 |
| 隐私 | 不展示附件数量；发布者姓名/手机号匿名化 |

### 1.4 信息详情
| 功能点 | 详细说明 |
|-------|---------|
| 标题 | 顶部大标题 |
| 分类标签 | 所属分类（一级 · 二级） |
| 内容 | 完整正文，保留换行 |
| 联系信息 | **公开访问时匿名**：联系人 `张**`、电话 `138****1234` |
| 完整信息 | 仅**发布者本人**或**管理员**可见真实联系人、电话及附件 |
| 附件 | 公开页**不展示**附件列表，不可下载 |
| 发布者信息 | 公开页发布者姓名匿名展示 |
| 返回 | 返回首页/上一页 |

### 1.5 用户中心

左侧竖向导航布局（`layout/user.html`），登录后默认入口。

#### 1.5.1 用户首页 (/my)
| 功能点 | 详细说明 |
|-------|---------|
| 统计卡片 | 全部产品数、已发布数、待审核数 |
| 状态提示 | 企业认证状态、店铺是否已完善 |
| 快捷入口 | 完善店铺、添加产品、产品列表 |

#### 1.5.2 店铺信息 (/my/shop)
| 功能点 | 详细说明 |
|-------|---------|
| 店铺名称 | 必填 |
| 可做区域 | 省份多选（`internal/data/regions.go`） |
| 可做类型 | 根据用户身份过滤的二级分类多选 |
| 联系信息 | 联系人、手机、电话、公司地址 |
| 公司介绍 | 文本域 |
| Banner | 图片上传 |
| 操作 | 重置 / 确认保存 → POST `/api/my/shop` |

#### 1.5.3 添加产品 (/my/products/new)
| 功能点 | 详细说明 |
|-------|---------|
| 前置条件 | 企业认证 `verify_status=approved` |
| 产品名称 | 必填 |
| 分类 | 二级分类单选（按身份过滤） |
| 产品图片 | 上传 |
| 额度 | 万元 |
| 利率 | 日利率 / 年利率 + 数值(%) |
| 还款期数 | 按月，如 12 = 1 年 12 期 |
| 还款方式 | 等额本息 / 等额本金 |
| 可做区域 | 省份多选 |
| 产品介绍 | 文本域 |
| 提交 | status=pending，等待管理员审核 |

#### 1.5.4 产品列表 (/my/products)
| 功能点 | 详细说明 |
|-------|---------|
| 列表字段 | 产品名称、分类、额度、利率、期数、状态、时间 |
| 状态筛选 | 待审核 / 已发布 / 已驳回 / 已下架 |
| 删除 | 待审核、已驳回状态可删除 |

#### 1.5.5 基本信息 (/my/profile)
| 功能点 | 详细说明 |
|-------|---------|
| 账号信息 | 手机、真实姓名、企业名称、身份（只读） |
| 企业认证 | 上传企业证明照片（jpg/png/gif/webp），上传后即时 `approved` |
| 修改密码 | 原密码 + 新密码 + 确认 → POST `/api/my/password` |

### 1.6 我的发布(旧版，保留)
| 功能点 | 详细说明 |
|-------|---------|
| 发布列表 | 展示当前用户所有发布信息 |
| 状态标识 | 待审核(黄色) / 已通过(绿色) / 已驳回(红色) |
| 驳回原因 | 已驳回信息展示驳回理由 |
| 删除操作 | 仅 pending 状态可删除 |
| 新建发布 | "发布信息"按钮，跳转发布表单 |

> 路径：`/my/posts`，与新版产品管理并行保留。

### 1.7 发布信息
| 功能点 | 详细说明 |
|-------|---------|
| 分类选择 | 下拉框选择信息分类 |
| 标题 | 必填，5-100字 |
| 内容 | 必填，textarea，支持多行文本 |
| 联系人 | 选填 |
| 联系电话 | 选填，11位手机号格式校验 |
| 附件上传 | 支持多文件上传(jpg/png/pdf/doc/docx/xls/xlsx)，单文件≤10MB |
| 上传进度 | 文件上传时展示进度提示 |
| 提交 | 数据写 posts 表，status=pending |
| 提示 | 提交成功后提示"信息已提交，请等待审核" |

### 1.8 退出登录
| 功能点 | 详细说明 |
|-------|---------|
| 登出 | 清除 Cookie 中的 JWT Token，跳转首页 |

---

## 2. 管理端功能

### 2.1 管理员登录
| 功能点 | 详细说明 |
|-------|---------|
| 登录页面 | 独立于用户端的登录页(/admin/login) |
| 登录方式 | 手机号 + 密码 |
| 身份校验 | 仅允许 role=admin 的用户登录 |
| 登录成功 | 跳转管理后台首页 |

### 2.2 管理后台首页
| 功能点 | 详细说明 |
|-------|---------|
| 汇总统计 | 用户总数、信息总数、信息待审、产品总数、产品待审、今日新增信息 |
| **分类统计** | 按一级分类汇总信息数/产品数，并展开各二级分类明细 |
| 快捷入口 | 信息审核、产品审核、信息管理、导出报表、用户管理 |
| 导航栏 | 首页 / 信息审核 / 产品审核 / 信息管理 / 导出报表 / 用户管理 |

### 2.3 信息审核（posts）
| 功能点 | 详细说明 |
|-------|---------|
| 待审核列表 | 展示所有 status=pending 的信息 |
| 列表字段 | 标题、分类、发布人、发布时间 |
| 详情查看 | 点击查看信息完整内容 + 附件 |
| 通过操作 | 设置 status=approved，记录 reviewed_by + reviewed_at |
| 驳回操作 | 弹出驳回原因输入框(必填)，设置 status=rejected，记录 reject_reason |
| 审核日志 | 记录审核人、审核时间、审核结果 |

### 2.4 产品审核（products）
| 功能点 | 详细说明 |
|-------|---------|
| 待审核列表 | GET `/admin/product-review`，展示 status=pending 的产品 |
| 列表字段 | 产品名称、分类、发布人、额度、利率、期数、提交时间 |
| 通过操作 | POST `/api/admin/products/:id/review`，status=approved |
| 驳回操作 | 填写驳回原因，status=rejected |

### 2.5 用户导出
| 功能点 | 详细说明 |
|-------|---------|
| 入口 | 用户管理页「导出用户 Excel」 |
| 接口 | GET `/api/admin/users/export`，支持 `keyword` 筛选 |
| Excel 字段 | 手机号、真实姓名、企业名称、身份、认证状态、账号状态、店铺名称、店铺联系人、店铺手机、注册时间 |

### 2.6 信息管理
| 功能点 | 详细说明 |
|-------|---------|
| 全部信息列表 | 展示所有状态的信息(待审核/已通过/已驳回) |
| 筛选 | 按分类、按状态、按发布时间范围筛选 |
| 搜索 | 关键词搜索标题 |
| 分页 | 每页20条 |
| 操作 | 查看详情、删除(硬删除，含关联附件) |

### 2.7 导出报表
| 功能点 | 详细说明 |
|-------|---------|
| 筛选条件 | 分类(多选)、状态(多选)、发布时间范围、关键词 |
| 预览 | 点击"预览"展示符合条件的记录数和前20条数据 |
| 导出按钮 | 点击"导出Excel"触发下载 |
| Excel内容 | Sheet1 数据明细：标题、分类、内容、联系人、电话、发布人、状态、发布时间 |
|  | Sheet2 统计汇总：按分类统计数量、按状态统计数量 |
| 文件名 | 导出_YYYYMMDD_HHmmss.xlsx |

### 2.8 用户管理
| 功能点 | 详细说明 |
|-------|---------|
| 用户列表 | 展示所有用户(管理员除外) |
| 列表字段 | 手机号、真实姓名、企业名称、身份、认证状态、注册时间、账号状态 |
| 搜索 | 按手机号、真实姓名、企业名称或昵称搜索 |
| 导出 | 导出当前筛选结果为 Excel（见 2.5） |
| 启用/禁用 | 管理员可禁用违规用户(status=disabled)，禁用后无法登录 |
| 分页 | 每页20条 |

---

## 3. 公共功能

### 3.1 短信验证码
| 功能点 | 详细说明 |
|-------|---------|
| 发送接口 | POST /api/sms/send，参数：phone + scene |
| 场景 | register(注册)、login(登录)、reset(重置密码，预留) |
| 验证码生成 | 6位随机数字 |
| 存储 | 写入 sms_logs 表，包含手机号、验证码、过期时间 |
| 腾讯云对接 | 调用腾讯云 SMS SDK 发送，支持模板变量 |
| 频率限制 | 单手机号每分钟1次，每小时5次 |
| Mock模式 | 开发环境下跳过真实发送，控制台打印验证码 |

### 3.2 附件上传与访问
| 功能点 | 详细说明 |
|-------|---------|
| 上传接口 | POST `/api/upload`（multipart/form-data），需 JWT 认证 |
| 文件类型 | jpg, jpeg, png, pdf, doc, docx, xls, xlsx |
| 文件大小 | 单文件 ≤ 10MB |
| 存储路径 | `/uploads/{YYYYMM}/{UUID}.{ext}` |
| 安全处理 | UUID 重命名 + 路径遍历防护 |
| **访问控制** | GET `/uploads/*` 非公开静态目录，需鉴权（见 3.4） |
| 删除 | 删除 post 时级联删除附件文件和数据库记录 |

### 3.3 公开页隐私与安全
| 功能点 | 详细说明 |
|-------|---------|
| 联系人匿名 | `MaskName`：保留首字，其余 `*`（如 张三 → 张*） |
| 电话匿名 | `MaskPhone`：138****1234 |
| 发布者匿名 | 首页列表、详情页对非本人/非管理员匿名展示 |
| 附件隐藏 | 公开信息详情不展示附件；API 列表不返回附件数量 |
| 文件访问 | 未登录拒绝；登录用户仅可访问**本人**资源（认证照、店铺 Banner、产品图、自己的信息附件）；管理员可访问全部 |
| 实现 | `internal/handler/upload_serve.go`、`MaskName`/`MaskPhone` |

### 3.4 移动端适配
| 功能点 | 详细说明 |
|-------|---------|
| 响应式布局 | Tailwind CSS 断点：sm(640) md(768) lg(1024) |
| 导航栏 | 小屏折叠为汉堡菜单，点击展开 |
| 表格 | 小屏转为卡片堆叠展示 |
| 表单 | 所有输入框 w-full，按钮全宽，表单间距加大 |
| 触摸友好 | 按钮最小 44px 高度，链接间距合理 |
| 字体 | 基准 16px，防止 iOS 缩放 |

---

## 4. 数据模型

### 4.1 用户 (users)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 主键 |
| phone | VARCHAR(11) | UNIQUE, NOT NULL | 手机号 |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt 密码哈希 |
| role | VARCHAR(10) | NOT NULL, DEFAULT 'user' | user / admin |
| nickname | VARCHAR(50) | - | 昵称（兼容旧数据） |
| real_name | VARCHAR(50) | - | 真实姓名 |
| company | VARCHAR(100) | - | 企业名称 |
| identity | VARCHAR(20) | - | demander / supplier / funder |
| verify_status | VARCHAR(15) | DEFAULT 'none' | none / approved（上传照片即 approved） |
| verify_doc_path | VARCHAR(500) | - | 企业证明材料路径 |
| status | VARCHAR(10) | NOT NULL, DEFAULT 'active' | active / disabled |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新时间 |

### 4.2 分类 (categories)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| parent_id | INT | FK → categories.id, NULL | NULL=一级导航；有值=二级 |
| name | VARCHAR(50) | NOT NULL | 分类名称（parent_id + name 联合唯一） |
| sort_order | INT | DEFAULT 0 | 排序权重 |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |

一级分类：新能源项目、企业类项目、电站出售方、电站收购方、租赁公司、其他类。启动时自动从旧 6 分类迁移。

### 4.3 信息 (posts)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 主键 |
| user_id | UUID | FK → users.id, NOT NULL | 发布人 |
| category_id | INT | FK → categories.id, NOT NULL | 分类 |
| title | VARCHAR(200) | NOT NULL | 标题 |
| content | TEXT | NOT NULL | 内容 |
| contact | VARCHAR(100) | - | 联系人 |
| contact_phone | VARCHAR(11) | - | 联系电话 |
| status | VARCHAR(15) | NOT NULL, DEFAULT 'pending' | pending/approved/rejected |
| reject_reason | TEXT | - | 驳回原因 |
| reviewed_by | UUID | FK → users.id | 审核人 |
| reviewed_at | TIMESTAMPTZ | - | 审核时间 |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新时间 |

### 4.4 附件 (attachments)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 主键 |
| post_id | UUID | FK → posts.id, ON DELETE CASCADE | 所属信息 |
| file_name | VARCHAR(255) | NOT NULL | 原始文件名 |
| file_path | VARCHAR(500) | NOT NULL | 存储路径 |
| file_size | BIGINT | NOT NULL | 文件大小(字节) |
| created_at | TIMESTAMPTZ | NOT NULL | 上传时间 |

### 4.5 短信日志 (sms_logs)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 主键 |
| phone | VARCHAR(11) | NOT NULL | 手机号 |
| code | VARCHAR(6) | NOT NULL | 验证码 |
| scene | VARCHAR(20) | NOT NULL | 场景标识 |
| expired_at | TIMESTAMPTZ | NOT NULL | 过期时间 |
| used | BOOLEAN | NOT NULL, DEFAULT FALSE | 是否已使用 |
| created_at | TIMESTAMPTZ | NOT NULL | 发送时间 |

### 4.6 店铺 (shops)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 主键 |
| user_id | UUID | UNIQUE, FK → users.id | 所属用户 |
| shop_name | VARCHAR(100) | - | 店铺名称 |
| regions | TEXT | - | JSON 省份数组 |
| category_ids | TEXT | - | JSON 分类 ID 数组 |
| contact, phone, tel, address | | | 联系信息 |
| intro | TEXT | - | 公司介绍 |
| banner_path | VARCHAR(500) | - | Banner 图路径 |

### 4.7 产品 (products)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK | 主键 |
| user_id | UUID | FK → users.id | 发布人 |
| category_id | INT | FK → categories.id | 叶子分类 |
| name | VARCHAR(200) | NOT NULL | 产品名称 |
| image_path | VARCHAR(500) | - | 产品图 |
| amount_wan | FLOAT | - | 额度（万元） |
| rate_type | VARCHAR(10) | - | daily / yearly |
| rate_percent | FLOAT | - | 利率 % |
| period_count | INT | - | 期数 |
| period_unit | VARCHAR(10) | DEFAULT 'month' | 期数单位 |
| repay_method | VARCHAR(30) | - | 还款方式 |
| regions | TEXT | - | JSON 省份数组 |
| intro | TEXT | - | 产品介绍 |
| status | VARCHAR(15) | DEFAULT 'pending' | pending/approved/rejected/delisted |
| reject_reason | TEXT | - | 驳回原因 |
| reviewed_by | UUID | FK → users.id | 审核人 |
| reviewed_at | TIMESTAMPTZ | - | 审核时间 |
| created_at, updated_at | TIMESTAMPTZ | | 时间戳 |

---

## 5. 状态流转

### 5.1 信息状态机（posts）
```
┌─────────┐     用户发布      ┌──────────┐     管理员审核      ┌──────────┐
│  (草稿)  │ ───────────────→ │  pending │ ─────────────────→ │ approved │
└─────────┘                   └────┬─────┘                    └──────────┘
                                   │
                                   │ 管理员驳回
                                   ▼
                              ┌──────────┐
                              │ rejected │
                              └──────────┘
```

### 5.2 产品状态机（products）
```
pending → approved（管理员通过，首页展示待后续迭代）
       → rejected（用户可见原因，可删除重发）
approved → delisted（下架，预留）
```

### 5.3 企业认证状态机
```
none → approved（用户上传企业证明照片，即时完成，无需管理员审核）
```

### 5.4 用户账号状态机
```
┌────────┐     管理员禁用     ┌───────────┐
│ active │ ─────────────────→ │ disabled  │
└──────┬─┘                    └─────┬─────┘
       │                            │
       │      管理员启用             │
       └────────────────────────────┘
```

### 5.5 注册流程
```
输入手机号 → 获取验证码 → 输入验证码 → 设置密码 → 填写真实姓名/企业名称/选身份 → 注册成功 → 跳转用户中心
```

### 5.6 产品发布-审核流程
```
注册 → 上传企业证明照片（即时认证）→ 完善店铺 → 添加产品(pending) → 管理员产品审核 → approved
```

### 5.7 信息发布-审核流程（旧版 posts）
```
用户登录 → 填写信息表单 → 上传附件 → 提交(pending)
→ 管理员查看待审核列表 → 查看详情
→ 通过(approved) → 首页对外展示
→ 驳回(rejected) → 用户查看驳回原因 → 可重新发布
```

---

## 6. 接口列表汇总

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | / | 无 | 首页 |
| GET | /posts/:id | 无 | 信息详情 |
| GET | /login | 无 | 用户登录页 |
| GET | /register | 无 | 用户注册页 |
| POST | /api/sms/send | 无 | 发送短信验证码 |
| POST | /api/auth/register | 无 | 用户注册 |
| POST | /api/auth/login | 无 | 用户登录 |
| GET | /my | 用户 | 用户中心首页 |
| GET | /my/shop | 用户 | 店铺信息页 |
| POST | /api/my/shop | 用户 | 保存店铺 |
| GET | /my/products/new | 用户 | 添加产品页 |
| GET | /my/products | 用户 | 产品列表 |
| POST | /api/my/products | 用户 | 创建产品 |
| POST | /api/my/products/:id/delete | 用户 | 删除产品 |
| GET | /my/profile | 用户 | 基本信息页 |
| POST | /api/my/password | 用户 | 修改密码 |
| POST | /api/my/verify | 用户 | 上传企业证明照片（即时认证） |
| GET | /my/posts | 用户 | 我的发布列表（旧版） |
| GET | /my/posts/new | 用户 | 发布信息表单 |
| POST | /api/posts | 用户 | 创建信息 |
| GET | /api/posts/:id | 用户 | 信息详情(我的) |
| DELETE | /api/posts/:id | 用户 | 删除信息 |
| POST | /api/upload | 用户 | 上传附件 |
| GET | /logout | - | 退出 |
| GET | /admin/login | 无 | 管理登录页 |
| POST | /api/admin/login | 无 | 管理员登录 |
| GET | /admin | 管理员 | 管理首页 |
| GET | /admin/review | 管理员 | 信息待审核列表 |
| GET | /admin/product-review | 管理员 | 产品待审核列表 |
| GET | /admin/posts | 管理员 | 全部信息 |
| GET | /admin/export | 管理员 | 导出筛选页 |
| GET | /admin/users | 管理员 | 用户列表 |
| POST | /api/admin/posts/:id/review | 管理员 | 信息审核 |
| POST | /api/admin/products/:id/review | 管理员 | 产品审核 |
| GET | /api/admin/users/export | 管理员 | 导出用户 Excel |
| POST | /api/admin/export | 管理员 | 导出信息 Excel |
| PUT | /api/admin/users/:id/status | 管理员 | 用户状态 |
| GET | /uploads/*filepath | 可选 JWT | 鉴权访问上传文件 |

---

## 7. 初始化数据

### 7.1 默认分类（两级树）

| 一级 | 二级 |
|------|------|
| 新能源项目 | 光伏发电、储能电站、风力发电、垃圾发电、水利发电、充电桩、光储充一体化项目 |
| 企业类项目 | 央国企设备租赁、上市公司设备租赁、中小微企业设备租赁 |
| 电站出售方 | 同新能源项目 7 项 |
| 电站收购方 | 同新能源项目 7 项 |
| 租赁公司 | 金租、商租、外资 |
| 其他类 | （无二级，本身为叶子） |

旧版 6 分类（新能源、融资、租赁等）在启动时自动迁移至新树。

### 7.2 默认管理员
| 字段 | 值 |
|------|-----|
| 手机号 | 13800000000 |
| 密码 | Admin@123 |
| 角色 | admin |
| 昵称 | 超级管理员 |
