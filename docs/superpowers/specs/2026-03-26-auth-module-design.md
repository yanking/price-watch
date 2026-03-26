# 授权模块设计文档

## 概述

本文档描述价格监控系统的用户认证授权模块设计，采用 DDD（领域驱动设计）架构，支持用户注册登录、第三方登录（微信、GitHub）、单设备登录控制等功能。

---

## 一、需求总结

| 项目 | 描述 |
|------|------|
| 用户类型 | 普通用户注册登录 |
| 认证方式 | JWT Token（7天过期） |
| 密码策略 | 8-20位，必须包含字母和数字 |
| 第三方登录 | 微信、GitHub（策略模式实现） |
| 邮箱验证 | 可选 |
| 找回密码 | 通过邮箱发送重置链接 |
| 单设备登录 | 新登录踢掉旧登录 |

---

## 二、架构设计

### 2.1 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    Interfaces Layer                      │
│              (HTTP Handlers, Middleware)                 │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   Application Layer                      │
│           (Services, DTOs, Assemblers)                   │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                     Domain Layer                         │
│      (Entities, Value Objects, Repositories, Services)   │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                     │
│      (Persistence, OAuth, Email, SMS Implementations)    │
└─────────────────────────────────────────────────────────┘
```

### 2.2 层间依赖原则

- 上层依赖下层，下层不依赖上层
- Domain 层不依赖任何外层
- 通过接口实现依赖倒置

---

## 三、领域模型

### 3.1 实体

#### User（用户聚合根）

```go
type User struct {
    id            int64         // 雪花ID，主键+对外
    username      string        // 用户名
    password      *Password     // 密码（值对象，可为空表示第三方登录）
    email         *Email        // 邮箱（值对象）
    emailVerified bool          // 邮箱是否验证
    areaCode      string        // 区号
    phone         string        // 手机号
    phoneVerified bool          // 手机号是否验证
    avatar        string        // 头像URL
    nickname      string        // 昵称
    status        UserStatus    // 用户状态
    createdAt     time.Time
    updatedAt     time.Time
}

// 行为方法
func (u *User) VerifyEmail() error
func (u *User) VerifyPhone() error
func (u *User) ChangePassword(old, new string)
func (u *User) ResetPassword(new string)
func (u *User) UpdateProfile(avatar, nickname string)
func (u *User) UpdatePhone(phone *Phone)
func (u *User) Activate()
func (u *User) Deactivate()
func (u *User) IsActive() bool
func (u *User) FullPhone() string
func (u *User) MaskedPhone() string
```

#### ThirdPartyBind（第三方绑定实体）

```go
type ThirdPartyBind struct {
    id           int64
    userId       int64
    provider     OAuthProvider  // 提供商（值对象）
    providerId   string         // 第三方用户ID
    providerName string         // 第三方用户名
    createdAt    time.Time
}
```

### 3.2 值对象

```go
type Email struct {
    value string
}

func NewEmail(value string) (*Email, error)  // 验证格式
func (e *Email) Value() string
func (e *Email) Mask() string

type Password struct {
    hash string
}

func NewPassword(plain string) (*Password, error)  // 验证规则并加密
func NewPasswordFromHash(hash string) *Password
func (p *Password) Verify(plain string) bool
func (p *Password) Hash() string

type Phone struct {
    areaCode string
    number   string
}

func NewPhone(areaCode, number string) (*Phone, error)
func (p *Phone) AreaCode() string
func (p *Phone) Number() string
func (p *Phone) Full() string      // +8613800138000
func (p *Phone) Mask() string      // 138****8000

type OAuthProvider int

const (
    OAuthProviderWeChat OAuthProvider = iota + 1
    OAuthProviderGitHub
)
```

### 3.3 模型关系

```
┌─────────────────────────────────────────────────────────────────┐
│                         User（聚合根）                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  *Email     │  │ *Password   │  │  *Phone     │  （值对象）   │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
         │ 1                           │ 1
         │                             │
         │ *                           │ *
         ↓                             ↓
┌─────────────────┐           ┌─────────────────────────┐
│   Token (Redis) │           │   ThirdPartyBind        │
│   版本号控制     │           │   第三方绑定实体          │
└─────────────────┘           └─────────────────────────┘
```

### 3.4 仓储接口

```go
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    FindById(ctx context.Context, id int64) (*User, error)
    FindByUsername(ctx context.Context, username string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindByPhone(ctx context.Context, areaCode, phone string) (*User, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error)
}

type ThirdPartyBindRepository interface {
    Save(ctx context.Context, bind *ThirdPartyBind) error
    Delete(ctx context.Context, userId int64, provider OAuthProvider) error
    FindByProvider(ctx context.Context, provider OAuthProvider, providerId string) (*ThirdPartyBind, error)
    FindByUserId(ctx context.Context, userId int64) ([]*ThirdPartyBind, error)
    ExistsByProvider(ctx context.Context, provider OAuthProvider, providerId string) (bool, error)
}
```

### 3.5 领域服务

```go
// TokenService - 令牌领域服务
type TokenService interface {
    GenerateToken(user *User) (token string, version int64, err error)
    ParseToken(token string) (userId int64, err error)
    IncrementVersion(userId int64) (int64, error)
    GetVersion(userId int64) (int64, error)
}

// OAuthStrategy - OAuth策略接口
type OAuthStrategy interface {
    GetProviderName() string
    GetAuthURL(state string) string
    GetUserInfo(code string) (*OAuthUserInfo, error)
}

type OAuthUserInfo struct {
    ProviderId   string
    ProviderName string
    Email        string
}
```

---

## 四、Token 设计

### 4.1 单设备登录方案

使用 JWT + Redis 版本号实现单设备登录控制：

```
Redis Key: user_token_version:{userId} → {version}

登录时：version + 1
验证时：比对 JWT 中的 version 与 Redis 中的 version
```

### 4.2 JWT Payload

```json
{
  "uid": 1234567890123456789,
  "ver": 5,
  "exp": 1700000000,
  "iat": 1699999900
}
```

### 4.3 验证流程

```
请求携带 Token
  → 解析 JWT 获取 uid、version
  → Redis GET user_token_version:{uid}
  → 比对 version
     ├─ 一致 → 通过
     └─ 不一致/不存在 → 401（已在其他设备登录）
```

---

## 五、数据库设计

### 5.1 users 表

```sql
CREATE TABLE users (
    id             BIGINT UNSIGNED NOT NULL PRIMARY KEY COMMENT '雪花ID',
    username       VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password_hash  VARCHAR(255) DEFAULT NULL COMMENT '密码哈希',
    email          VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
    email_verified TINYINT(1) DEFAULT 0 COMMENT '邮箱是否验证',
    area_code      VARCHAR(10) DEFAULT NULL COMMENT '区号',
    phone          VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    phone_verified TINYINT(1) DEFAULT 0 COMMENT '手机号是否验证',
    avatar         VARCHAR(500) DEFAULT NULL COMMENT '头像URL',
    nickname       VARCHAR(50) DEFAULT NULL COMMENT '昵称',
    status         TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常 2停用',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_username (username),
    INDEX idx_email (email),
    UNIQUE INDEX idx_area_phone (area_code, phone),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

### 5.2 third_party_binds 表

```sql
CREATE TABLE third_party_binds (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id       BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    provider      TINYINT NOT NULL COMMENT '提供商：1微信 2GitHub',
    provider_id   VARCHAR(100) NOT NULL COMMENT '第三方用户ID',
    provider_name VARCHAR(100) DEFAULT NULL COMMENT '第三方用户名',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_provider_user (provider, provider_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='第三方账号绑定表';
```

### 5.3 Redis 存储

```
# 用户 Token 版本号
user_token_version:{userId} → {version}  TTL: 7天

# 密码重置 Token
password_reset:{token} → {userId}  TTL: 15分钟

# 邮箱验证 Token
email_verify:{token} → {userId}  TTL: 24小时

# OAuth State
oauth_state:{state} → {provider}  TTL: 10分钟

# 手机验证码
phone_verify_code:{userId} → {code}  TTL: 5分钟
```

---

## 六、API 设计

### 6.1 认证相关

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/v1/users` | 注册用户 | 否 |
| POST | `/v1/auth/login` | 登录 | 否 |
| POST | `/v1/auth/logout` | 登出 | 是 |
| GET | `/v1/auth/oauth/{provider}/url` | 获取第三方授权URL | 否 |
| POST | `/v1/auth/oauth/{provider}` | 第三方登录 | 否 |
| POST | `/v1/auth/password/forgot` | 忘记密码 | 否 |
| POST | `/v1/auth/password/reset` | 重置密码 | 否 |

### 6.2 用户相关

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/v1/users/{userId}` | 获取用户信息 | 是 |
| PATCH | `/v1/users/{userId}` | 更新用户信息 | 是 |
| POST | `/v1/users/{userId}/password` | 修改密码 | 是 |
| POST | `/v1/users/{userId}/phone-verification` | 发送手机验证码 | 是 |
| POST | `/v1/users/{userId}/phone-verification/confirm` | 确认手机验证 | 是 |
| POST | `/v1/users/{userId}/email-verification` | 发送邮箱验证 | 是 |
| POST | `/v1/users/{userId}/email-verification/confirm` | 确认邮箱验证 | 否 |
| GET | `/v1/users/{userId}/oauth-bindings` | 获取第三方绑定列表 | 是 |
| DELETE | `/v1/users/{userId}/oauth-bindings/{provider}` | 解绑第三方账号 | 是 |

### 6.3 请求/响应示例

**注册**
```json
POST /v1/users

{
  "username": "testuser",
  "password": "Test1234",
  "email": "test@example.com",
  "phone": {
    "areaCode": "86",
    "number": "13800138000"
  }
}

→ 201 Created
{
  "id": "1234567890123456789",
  "username": "testuser",
  "email": "test@example.com",
  "emailVerified": false,
  "areaCode": "86",
  "phone": "138****8000",
  "phoneVerified": false,
  "createdAt": "2026-03-26T10:00:00Z"
}
```

**登录**
```json
POST /v1/auth/login

{
  "account": "testuser",
  "password": "Test1234"
}

→ 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 604800,
  "user": {
    "id": "1234567890123456789",
    "username": "testuser"
  }
}
```

**第三方登录**
```json
GET /v1/auth/oauth/github/url

→ 200 OK
{
  "url": "https://github.com/login/oauth/authorize?...",
  "state": "xyz789"
}

POST /v1/auth/oauth/github

{
  "code": "abc123",
  "state": "xyz789"
}

→ 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 604800,
  "isNewUser": true,
  "user": {
    "id": "1234567890123456789",
    "username": "github_octocat"
  }
}
```

### 6.4 HTTP 状态码

| 状态码 | 场景 |
|--------|------|
| 200 OK | 成功（GET、PATCH） |
| 201 Created | 创建成功（POST） |
| 204 No Content | 删除成功（DELETE） |
| 202 Accepted | 请求已接受，异步处理 |
| 400 Bad Request | 参数错误 |
| 401 Unauthorized | 未认证/Token 失效 |
| 403 Forbidden | 无权限 |
| 404 Not Found | 资源不存在 |
| 409 Conflict | 资源冲突 |
| 422 Unprocessable Entity | 验证失败 |
| 500 Internal Server Error | 服务器错误 |

### 6.5 错误响应格式

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "参数验证失败",
    "details": [
      {
        "field": "password",
        "message": "密码必须包含字母和数字"
      }
    ]
  }
}
```

---

## 七、关键流程

### 7.1 注册流程

```
Client → Handler → Service → Repo
         │          │          │
         │ 验证参数   │          │
         │─────────►│          │
         │          │ 检查用户名 │
         │          │─────────►│
         │          │◄─────────│
         │          │ 创建实体   │
         │          │ (生成雪花ID│
         │          │  密码加密) │
         │          │─────────►│
         │          │◄─────────│
         │◄─────────│          │
         │ 返回DTO   │          │
```

### 7.2 登录流程

```
Client → Handler → Service → Repo → Redis
         │          │          │       │
         │          │ 查询用户  │       │
         │          │─────────►│       │
         │          │◄─────────│       │
         │          │ 验证密码  │       │
         │          │          │       │
         │          │ 版本号+1  │       │
         │          │─────────────────►│
         │          │◄─────────────────│
         │          │ 生成JWT  │       │
         │◄─────────│          │       │
```

### 7.3 第三方登录流程

```
Client → Handler → Service → Strategy → OAuth Provider
         │          │          │              │
         │ 获取URL   │          │              │
         │─────────►│─────────►│              │
         │◄─────────│◄─────────│              │
         │          │          │              │
         │ 用户授权...          │              │
         │          │          │              │
         │ 回调(code)│          │              │
         │─────────►│          │              │
         │          │ 验证state │              │
         │          │─────────►│              │
         │          │          │ 获取用户信息  │
         │          │          │─────────────►│
         │          │          │◄─────────────│
         │          │◄─────────│              │
         │          │ 查找/创建 │              │
         │          │ 用户      │              │
         │◄─────────│          │              │
```

---

## 八、通知服务设计

### 8.1 多服务商配置

支持配置多个 SMS/Email 服务商，实现负载均衡和故障转移。

```yaml
sms:
  providers:
    - name: aliyun1
      type: aliyun
      weight: 5
      enabled: true
      config:
        access_key_id: ""
        access_key_secret: ""
        sign_name: ""
        template_code: ""

    - name: twilio
      type: twilio
      weight: 2
      enabled: true
      config:
        account_sid: ""
        auth_token: ""
        from_number: ""

  strategy: weighted_round_robin  # round_robin / weighted_round_robin / failover
  retry_times: 2
  retry_interval: 100  # ms
  code_expire: 300     # seconds

email:
  providers:
    - name: smtp1
      type: smtp
      weight: 5
      enabled: true
      config:
        host: "smtp.example.com"
        port: 587
        username: ""
        password: ""
        from: ""

    - name: sendgrid
      type: sendgrid
      weight: 2
      enabled: true
      config:
        api_key: ""
        from: ""

  strategy: weighted_round_robin
  retry_times: 2
  retry_interval: 100
```

### 8.2 负载均衡策略

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| `round_robin` | 简单轮询，依次使用 | 各服务商能力相当 |
| `weighted_round_robin` | 加权轮询，按权重分配 | 服务商能力不同 |
| `failover` | 主备模式，优先用第一个 | 有明确主备关系 |

### 8.3 故障处理

```
发送请求
    ↓
选择服务商（排除故障中的）
    ↓
发送 → 成功 → 返回
    ↓
   失败
    ↓
标记故障（进入30秒冷却期）
    ↓
重试（选择其他服务商）
    ↓
达到重试次数 → 返回错误
```

---

## 九、项目目录结构

```
internal/
└── auth/
    ├── domain/
    │   ├── entity/
    │   │   ├── user.go
    │   │   └── third_party_bind.go
    │   ├── valueobject/
    │   │   ├── email.go
    │   │   ├── password.go
    │   │   ├── phone.go
    │   │   └── oauth_provider.go
    │   ├── repository/
    │   │   ├── user_repository.go
    │   │   └── third_party_bind_repository.go
    │   └── service/
    │       ├── token_service.go
    │       └── oauth_strategy.go
    ├── application/
    │   ├── service/
    │   │   ├── auth_service.go
    │   │   └── user_service.go
    │   ├── assembler/
    │   │   └── user_assembler.go
    │   └── dto/
    │       ├── auth_dto.go
    │       └── user_dto.go
    ├── infrastructure/
    │   ├── persistence/
    │   │   ├── user_repository_impl.go
    │   │   ├── third_party_bind_repository_impl.go
    │   │   └── models.go
    │   ├── oauth/
    │   │   ├── github_oauth.go
    │   │   └── wechat_oauth.go
    │   ├── token/
    │   │   └── jwt_token_service.go
    │   └── cache/
    │       └── token_cache.go
    └── interfaces/
        └── http/
            ├── handler/
            │   ├── auth_handler.go
            │   └── user_handler.go
            ├── middleware/
            │   └── auth_middleware.go
            ├── response/
            │   └── response.go
            └── router.go

infrastructure/
├── config/
│   └── notification_config.go
└── notification/
    ├── load_balancer.go
    ├── sms_service.go
    ├── email_service.go
    ├── provider_aliyun_sms.go
    ├── provider_twilio_sms.go
    ├── provider_smtp.go
    └── provider_sendgrid.go

configs/
└── auth.yaml
```

---

## 十、配置设计

```yaml
# configs/auth.yaml
auth:
  jwt:
    secret: "your-jwt-secret-key"
    expire: 168  # 7天（小时）

  password:
    min_length: 8
    max_length: 20
    require_letter: true
    require_digit: true

  oauth:
    wechat:
      app_id: ""
      app_secret: ""
      redirect_uri: ""

    github:
      client_id: ""
      client_secret: ""
      redirect_uri: ""

  sms:
    providers:
      - name: aliyun1
        type: aliyun
        weight: 5
        enabled: true
        config:
          access_key_id: ""
          access_key_secret: ""
          sign_name: ""
          template_code: ""

    strategy: weighted_round_robin
    retry_times: 2
    retry_interval: 100
    code_expire: 300

  email:
    providers:
      - name: smtp1
        type: smtp
        weight: 5
        enabled: true
        config:
          host: ""
          port: 587
          username: ""
          password: ""
          from: ""

    strategy: weighted_round_robin
    retry_times: 2
    retry_interval: 100
```

---

## 十一、技术选型

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| Web 框架 | Gin | 高性能 HTTP 框架 |
| ORM | GORM | Go 常用 ORM |
| 缓存 | Redis | Token 版本控制、验证码存储 |
| JWT | golang-jwt/jwt | JWT 生成和解析 |
| 密码加密 | bcrypt | 密码哈希 |
| ID 生成 | 雪花算法 | 分布式唯一 ID |
| 配置管理 | Viper | 配置文件解析 |
| 依赖注入 | Wire（可选） | 编译时依赖注入 |

---

## 十二、安全考虑

1. **密码安全**：使用 bcrypt 加密，不存储明文
2. **Token 安全**：JWT 签名验证，版本号控制失效
3. **单设备登录**：新登录踢掉旧登录
4. **敏感信息脱敏**：手机号、邮箱脱敏显示
5. **防暴力破解**：登录失败次数限制（可扩展）
6. **HTTPS**：生产环境强制 HTTPS

---

## 十三、扩展性

1. **新增 OAuth 提供商**：实现 `OAuthStrategy` 接口
2. **新增通知渠道**：实现 `Provider` 接口
3. **新增用户字段**：修改实体和数据库表
4. **多因素认证**：可在认证流程中扩展
