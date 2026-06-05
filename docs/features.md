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
| 昵称 | 可选，2-20个字符 |
| 公司名称 | 可选 |
| 完成注册 | 写入 users 表，role=user，status=active，跳转登录页 |

### 1.2 用户登录
| 功能点 | 详细说明 |
|-------|---------|
| 登录方式 | 手机号 + 密码 |
| 身份校验 | 查询 users 表验证 phone + password_hash，仅允许 role=user |
| 账号状态 | 检查 status=active，禁用账号提示联系管理员 |
| Token 签发 | 返回 access_token(2h) + refresh_token(7天)，写入 Cookie |
| 登录成功 | 跳转到首页或个人中心 |

### 1.3 首页(信息广场)
| 功能点 | 详细说明 |
|-------|---------|
| 分类筛选 | 顶部 Tab 栏展示所有分类，"全部" + 各分类名称 |
| 关键词搜索 | 搜索框，搜索标题和内容中匹配的关键词 |
| 列表展示 | 卡片式布局，每项显示：标题、分类标签、发布者、发布时间、联系信息 |
| 分页 | 底部页码导航，每页20条 |
| 排序 | 默认按发布时间倒序 |
| 移动适配 | 小屏卡片占满宽度，大屏2-3列网格 |
| 状态过滤 | 仅展示 status=approved 的信息 |

### 1.4 信息详情
| 功能点 | 详细说明 |
|-------|---------|
| 标题 | 顶部大标题 |
| 分类标签 | 所属分类 |
| 内容 | 完整正文，保留换行 |
| 联系信息 | 联系人 + 联系电话 |
| 附件列表 | 文件名 + 文件大小 + 下载链接 |
| 发布者信息 | 昵称 + 发布时间 |
| 返回 | 返回首页/上一页 |

### 1.5 我的发布(个人中心)
| 功能点 | 详细说明 |
|-------|---------|
| 发布列表 | 展示当前用户所有发布信息 |
| 状态标识 | 待审核(黄色) / 已通过(绿色) / 已驳回(红色) |
| 驳回原因 | 已驳回信息展示驳回理由 |
| 删除操作 | 仅 pending 状态可删除 |
| 新建发布 | "发布信息"按钮，跳转发布表单 |

### 1.6 发布信息
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

### 1.7 退出登录
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
| 统计数据 | 总用户数、总信息数、待审核数、今日新增 |
| 快捷入口 | 审核入口、导出入口、用户管理入口 |
| 导航栏 | 侧边导航：首页 / 审核管理 / 信息管理 / 导出报表 / 用户管理 |

### 2.3 审核管理
| 功能点 | 详细说明 |
|-------|---------|
| 待审核列表 | 展示所有 status=pending 的信息 |
| 列表字段 | 标题、分类、发布人、发布时间 |
| 详情查看 | 点击查看信息完整内容 + 附件 |
| 通过操作 | 设置 status=approved，记录 reviewed_by + reviewed_at |
| 驳回操作 | 弹出驳回原因输入框(必填)，设置 status=rejected，记录 reject_reason |
| 审核日志 | 记录审核人、审核时间、审核结果 |

### 2.4 信息管理
| 功能点 | 详细说明 |
|-------|---------|
| 全部信息列表 | 展示所有状态的信息(待审核/已通过/已驳回) |
| 筛选 | 按分类、按状态、按发布时间范围筛选 |
| 搜索 | 关键词搜索标题 |
| 分页 | 每页20条 |
| 操作 | 查看详情、删除(硬删除，含关联附件) |

### 2.5 导出报表
| 功能点 | 详细说明 |
|-------|---------|
| 筛选条件 | 分类(多选)、状态(多选)、发布时间范围、关键词 |
| 预览 | 点击"预览"展示符合条件的记录数和前20条数据 |
| 导出按钮 | 点击"导出Excel"触发下载 |
| Excel内容 | Sheet1 数据明细：标题、分类、内容、联系人、电话、发布人、状态、发布时间 |
|  | Sheet2 统计汇总：按分类统计数量、按状态统计数量 |
| 文件名 | 导出_YYYYMMDD_HHmmss.xlsx |

### 2.6 用户管理
| 功能点 | 详细说明 |
|-------|---------|
| 用户列表 | 展示所有用户(管理员除外) |
| 列表字段 | 手机号、昵称、公司、注册时间、状态 |
| 搜索 | 按手机号或昵称搜索 |
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

### 3.2 附件上传
| 功能点 | 详细说明 |
|-------|---------|
| 上传接口 | POST /api/upload(multipart/form-data)，需 JWT 认证 |
| 文件类型 | jpg, jpeg, png, pdf, doc, docx, xls, xlsx |
| 文件大小 | 单文件 ≤ 10MB |
| 存储路径 | /uploads/{YYYYMM}/{UUID}_{原文件名} |
| 安全处理 | 随机 UUID 重命名防止路径遍历 |
| 返回 | 文件ID、文件名、文件路径、文件大小 |
| 删除 | 删除 post 时级联删除附件文件和数据库记录 |

### 3.3 移动端适配
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
| nickname | VARCHAR(50) | - | 昵称 |
| company | VARCHAR(100) | - | 公司名称 |
| status | VARCHAR(10) | NOT NULL, DEFAULT 'active' | active / disabled |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL | 更新时间 |

### 4.2 分类 (categories)
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| name | VARCHAR(50) | UNIQUE, NOT NULL | 分类名称 |
| sort_order | INT | DEFAULT 0 | 排序权重 |
| created_at | TIMESTAMPTZ | NOT NULL | 创建时间 |

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

---

## 5. 状态流转

### 5.1 信息状态机
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

### 5.2 用户状态机
```
┌────────┐     管理员禁用     ┌───────────┐
│ active │ ─────────────────→ │ disabled  │
└──────┬─┘                    └─────┬─────┘
       │                            │
       │      管理员启用             │
       └────────────────────────────┘
```

### 5.3 注册流程
```
输入手机号 → 获取验证码 → 输入验证码 → 设置密码 → 填写信息 → 注册成功
```

### 5.4 发布-审核-展示流程
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
| GET | /my/posts | 用户 | 我的发布列表 |
| GET | /my/posts/new | 用户 | 发布信息表单 |
| POST | /api/posts | 用户 | 创建信息 |
| GET | /api/posts/:id | 用户 | 信息详情(我的) |
| DELETE | /api/posts/:id | 用户 | 删除信息 |
| POST | /api/upload | 用户 | 上传附件 |
| GET | /logout | - | 退出 |
| GET | /admin/login | 无 | 管理登录页 |
| POST | /api/admin/login | 无 | 管理员登录 |
| GET | /admin | 管理员 | 管理首页 |
| GET | /admin/review | 管理员 | 待审核列表 |
| GET | /admin/posts | 管理员 | 全部信息 |
| GET | /admin/export | 管理员 | 导出筛选页 |
| GET | /admin/users | 管理员 | 用户列表 |
| POST | /api/admin/posts/:id/review | 管理员 | 审核操作 |
| POST | /api/admin/export | 管理员 | 导出 Excel |
| PUT | /api/admin/users/:id/status | 管理员 | 用户状态 |

---

## 7. 初始化数据

### 7.1 默认分类
| ID | 名称 | 排序 |
|----|------|------|
| 1 | 新能源 | 1 |
| 2 | 融资 | 2 |
| 3 | 租赁 | 3 |
| 4 | 技术合作 | 4 |
| 5 | 项目转让 | 5 |
| 6 | 其他 | 99 |

### 7.2 默认管理员
| 字段 | 值 |
|------|-----|
| 手机号 | 13800000000 |
| 密码 | Admin@123 |
| 角色 | admin |
| 昵称 | 超级管理员 |
