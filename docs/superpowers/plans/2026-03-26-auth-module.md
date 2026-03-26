# 授权模块实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现用户认证授权模块，支持注册登录、第三方登录、单设备登录控制

**Architecture:** DDD 分层架构，Domain 层独立无依赖，通过仓储接口实现依赖倒置，DO/PO 分离支持多数据库

**Tech Stack:** Go 1.22+, Gin, GORM, Redis, JWT (golang-jwt/jwt), bcrypt

---

## 文件结构映射

### 领域层 (Domain Layer)
| 文件 | 职责 |
|------|------|
| `internal/auth/domain/valueobject/email.go` | 邮箱值对象，验证格式 |
| `internal/auth/domain/valueobject/password.go` | 密码值对象，加密验证 |
| `internal/auth/domain/valueobject/phone.go` | 手机号值对象，国际号码支持 |
| `internal/auth/domain/valueobject/oauth_provider.go` | OAuth 提供商枚举 |
| `internal/auth/domain/entity/user.go` | 用户聚合根 |
| `internal/auth/domain/entity/third_party_bind.go` | 第三方绑定实体 |
| `internal/auth/domain/repository/user_repository.go` | 用户仓储接口 |
| `internal/auth/domain/repository/third_party_bind_repository.go` | 第三方绑定仓储接口 |
| `internal/auth/domain/service/token_service.go` | Token 服务接口 |
| `internal/auth/domain/service/oauth_strategy.go` | OAuth 策略接口 |

### 基础设施层 (Infrastructure Layer)
| 文件 | 职责 |
|------|------|
| `internal/auth/infrastructure/persistence/dao/user_dao.go` | 用户 DAO 接口 |
| `internal/auth/infrastructure/persistence/dao/third_party_bind_dao.go` | 第三方绑定 DAO 接口 |
| `internal/auth/infrastructure/persistence/mysql/user_po.go` | 用户 MySQL PO |
| `internal/auth/infrastructure/persistence/mysql/user_dao_impl.go` | 用户 MySQL DAO 实现 |
| `internal/auth/infrastructure/persistence/mysql/third_party_bind_po.go` | 第三方绑定 MySQL PO |
| `internal/auth/infrastructure/persistence/mysql/third_party_bind_dao_impl.go` | 第三方绑定 MySQL DAO 实现 |
| `internal/auth/infrastructure/persistence/converter/user_converter.go` | 用户 DO/PO 转换器 |
| `internal/auth/infrastructure/persistence/converter/third_party_bind_converter.go` | 第三方绑定转换器 |
| `internal/auth/infrastructure/persistence/repository/user_repository_impl.go` | 用户仓储实现 |
| `internal/auth/infrastructure/persistence/repository/third_party_bind_repository_impl.go` | 第三方绑定仓储实现 |
| `internal/auth/infrastructure/token/jwt_token_service.go` | JWT Token 服务实现 |
| `internal/auth/infrastructure/oauth/github_oauth.go` | GitHub OAuth 实现 |
| `internal/auth/infrastructure/oauth/wechat_oauth.go` | 微信 OAuth 实现 |

### 应用层 (Application Layer)
| 文件 | 职责 |
|------|------|
| `internal/auth/application/dto/auth_dto.go` | 认证相关 DTO |
| `internal/auth/application/dto/user_dto.go` | 用户相关 DTO |
| `internal/auth/application/assembler/user_assembler.go` | 用户 DTO 组装器 |
| `internal/auth/application/service/auth_service.go` | 认证应用服务 |
| `internal/auth/application/service/user_service.go` | 用户应用服务 |

### 接口层 (Interfaces Layer)
| 文件 | 职责 |
|------|------|
| `internal/auth/interfaces/http/response/response.go` | 统一响应 |
| `internal/auth/interfaces/http/middleware/auth_middleware.go` | 认证中间件 |
| `internal/auth/interfaces/http/handler/auth_handler.go` | 认证处理器 |
| `internal/auth/interfaces/http/handler/user_handler.go` | 用户处理器 |
| `internal/auth/interfaces/http/router.go` | 路由注册 |

### 配置和数据库
| 文件 | 职责 |
|------|------|
| `configs/auth.yaml` | 认证模块配置 |
| `scripts/sql/auth_tables.sql` | 数据库表结构 |

---

## Task 1: 值对象 - Email

**Files:**
- Create: `internal/auth/domain/valueobject/email.go`
- Create: `internal/auth/domain/valueobject/email_test.go`

- [ ] **Step 1: 编写 Email 值对象测试**

```go
// internal/auth/domain/valueobject/email_test.go
package valueobject_test

import (
    "testing"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewEmail(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"有效邮箱", "test@example.com", false},
        {"无效邮箱-无@", "testexample.com", true},
        {"无效邮箱-无域名", "test@", true},
        {"空字符串", "", false}, // 可选字段
        {"带子域名", "user@mail.example.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            email, err := valueobject.NewEmail(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && email != nil && email.Value() != tt.input {
                t.Errorf("Email.Value() = %v, want %v", email.Value(), tt.input)
            }
        })
    }
}

func TestEmailMask(t *testing.T) {
    email, _ := valueobject.NewEmail("test@example.com")
    if email.Mask() != "t***@example.com" {
        t.Errorf("Email.Mask() = %v, want t***@example.com", email.Mask())
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
cd /Users/wangyan/Code/AI/price-watch
go test ./internal/auth/domain/valueobject/... -v
```
Expected: FAIL (包不存在)

- [ ] **Step 3: 实现 Email 值对象**

```go
// internal/auth/domain/valueobject/email.go
package valueobject

import (
    "errors"
    "regexp"
    "strings"
)

type Email struct {
    value string
}

func NewEmail(value string) (*Email, error) {
    if value == "" {
        return nil, nil
    }

    value = strings.TrimSpace(value)

    // 验证邮箱格式
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    matched, _ := regexp.MatchString(pattern, value)
    if !matched {
        return nil, errors.New("邮箱格式不正确")
    }

    return &Email{value: value}, nil
}

func (e *Email) Value() string {
    return e.value
}

func (e *Email) Mask() string {
    if e.value == "" {
        return ""
    }
    at := strings.Index(e.value, "@")
    if at <= 1 {
        return e.value
    }
    return string(e.value[0]) + "***" + e.value[at:]
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
cd /Users/wangyan/Code/AI/price-watch
go test ./internal/auth/domain/valueobject/... -v
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/domain/valueobject/email.go internal/auth/domain/valueobject/email_test.go
git commit -m "feat(auth): 添加 Email 值对象

- 支持邮箱格式验证
- 支持脱敏显示

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 值对象 - Password

**Files:**
- Create: `internal/auth/domain/valueobject/password.go`
- Create: `internal/auth/domain/valueobject/password_test.go`

- [ ] **Step 1: 编写 Password 值对象测试**

```go
// internal/auth/domain/valueobject/password_test.go
package valueobject_test

import (
    "testing"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewPassword(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"有效密码", "Test1234", false},
        {"太短", "Test12", true},
        {"太长", "Test12345678901234567890", true},
        {"无数字", "Testtest", true},
        {"无字母", "12345678", true},
        {"空字符串", "", false}, // 可选字段，第三方登录无密码
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pwd, err := valueobject.NewPassword(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewPassword() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && pwd != nil && pwd.Hash() == "" {
                t.Error("Password.Hash() should not be empty")
            }
        })
    }
}

func TestPasswordVerify(t *testing.T) {
    pwd, _ := valueobject.NewPassword("Test1234")

    if !pwd.Verify("Test1234") {
        t.Error("Verify() should return true for correct password")
    }
    if pwd.Verify("Wrong123") {
        t.Error("Verify() should return false for wrong password")
    }
}

func TestNewPasswordFromHash(t *testing.T) {
    pwd1, _ := valueobject.NewPassword("Test1234")
    pwd2 := valueobject.NewPasswordFromHash(pwd1.Hash())

    if !pwd2.Verify("Test1234") {
        t.Error("NewPasswordFromHash should preserve verification")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/auth/domain/valueobject/... -v -run TestPassword
```
Expected: FAIL

- [ ] **Step 3: 实现 Password 值对象**

```go
// internal/auth/domain/valueobject/password.go
package valueobject

import (
    "errors"
    "regexp"
    "golang.org/x/crypto/bcrypt"
)

type Password struct {
    hash string
}

func NewPassword(plain string) (*Password, error) {
    if plain == "" {
        return nil, nil
    }

    // 验证长度
    if len(plain) < 8 {
        return nil, errors.New("密码长度不能少于8位")
    }
    if len(plain) > 20 {
        return nil, errors.New("密码长度不能超过20位")
    }

    // 验证必须包含字母
    hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, plain)
    if !hasLetter {
        return nil, errors.New("密码必须包含字母")
    }

    // 验证必须包含数字
    hasDigit, _ := regexp.MatchString(`[0-9]`, plain)
    if !hasDigit {
        return nil, errors.New("密码必须包含数字")
    }

    // 加密
    hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.New("密码加密失败")
    }

    return &Password{hash: string(hash)}, nil
}

func NewPasswordFromHash(hash string) *Password {
    return &Password{hash: hash}
}

func (p *Password) Hash() string {
    return p.hash
}

func (p *Password) Verify(plain string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plain))
    return err == nil
}
```

- [ ] **Step 4: 安装依赖并运行测试**

```bash
go get golang.org/x/crypto/bcrypt
go test ./internal/auth/domain/valueobject/... -v -run TestPassword
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/domain/valueobject/password.go internal/auth/domain/valueobject/password_test.go go.mod go.sum
git commit -m "feat(auth): 添加 Password 值对象

- 支持8-20位密码验证
- 必须包含字母和数字
- 使用 bcrypt 加密

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 值对象 - Phone

**Files:**
- Create: `internal/auth/domain/valueobject/phone.go`
- Create: `internal/auth/domain/valueobject/phone_test.go`

- [ ] **Step 1: 编写 Phone 值对象测试**

```go
// internal/auth/domain/valueobject/phone_test.go
package valueobject_test

import (
    "testing"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewPhone(t *testing.T) {
    tests := []struct {
        name     string
        areaCode string
        number   string
        wantErr  bool
    }{
        {"中国手机号", "86", "13800138000", false},
        {"美国手机号", "1", "2125551234", false},
        {"无效区号", "0", "13800138000", true},
        {"中国无效号码", "86", "12345", true},
        {"空值", "", "", false},
        {"只有区号", "86", "", true},
        {"只有号码", "", "13800138000", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            phone, err := valueobject.NewPhone(tt.areaCode, tt.number)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewPhone() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && phone != nil {
                if phone.AreaCode() != tt.areaCode {
                    t.Errorf("AreaCode() = %v, want %v", phone.AreaCode(), tt.areaCode)
                }
                if phone.Number() != tt.number {
                    t.Errorf("Number() = %v, want %v", phone.Number(), tt.number)
                }
            }
        })
    }
}

func TestPhoneFull(t *testing.T) {
    phone, _ := valueobject.NewPhone("86", "13800138000")
    if phone.Full() != "+8613800138000" {
        t.Errorf("Full() = %v, want +8613800138000", phone.Full())
    }
}

func TestPhoneMask(t *testing.T) {
    phone, _ := valueobject.NewPhone("86", "13800138000")
    if phone.Mask() != "138****8000" {
        t.Errorf("Mask() = %v, want 138****8000", phone.Mask())
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/auth/domain/valueobject/... -v -run TestPhone
```
Expected: FAIL

- [ ] **Step 3: 实现 Phone 值对象**

```go
// internal/auth/domain/valueobject/phone.go
package valueobject

import (
    "errors"
    "regexp"
    "strings"
)

type Phone struct {
    areaCode string
    number   string
}

func NewPhone(areaCode, number string) (*Phone, error) {
    if areaCode == "" && number == "" {
        return nil, nil
    }

    if areaCode == "" {
        return nil, errors.New("区号不能为空")
    }
    if number == "" {
        return nil, errors.New("手机号不能为空")
    }

    // 清理格式
    areaCode = strings.TrimPrefix(areaCode, "+")
    areaCode = strings.TrimSpace(areaCode)
    number = strings.ReplaceAll(number, " ", "")
    number = strings.ReplaceAll(number, "-", "")

    // 验证区号：1-4位数字，不以0开头
    if matched, _ := regexp.MatchString(`^[1-9]\d{0,3}$`, areaCode); !matched {
        return nil, errors.New("区号格式不正确")
    }

    // 根据区号验证号码
    switch areaCode {
    case "86": // 中国
        if matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, number); !matched {
            return nil, errors.New("手机号格式不正确")
        }
    case "1": // 美国/加拿大
        if matched, _ := regexp.MatchString(`^\d{10}$`, number); !matched {
            return nil, errors.New("手机号格式不正确")
        }
    default:
        // 其他国家：6-15位数字
        if matched, _ := regexp.MatchString(`^\d{6,15}$`, number); !matched {
            return nil, errors.New("手机号格式不正确")
        }
    }

    return &Phone{areaCode: areaCode, number: number}, nil
}

func (p *Phone) AreaCode() string {
    return p.areaCode
}

func (p *Phone) Number() string {
    return p.number
}

func (p *Phone) Full() string {
    if p.areaCode == "" || p.number == "" {
        return ""
    }
    return "+" + p.areaCode + p.number
}

func (p *Phone) Mask() string {
    if p.number == "" {
        return ""
    }
    if len(p.number) > 7 {
        return p.number[:3] + "****" + p.number[len(p.number)-4:]
    }
    return p.number
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/auth/domain/valueobject/... -v -run TestPhone
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/domain/valueobject/phone.go internal/auth/domain/valueobject/phone_test.go
git commit -m "feat(auth): 添加 Phone 值对象

- 支持国际号码格式
- 区号和号码分开存储
- 支持脱敏显示

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 值对象 - OAuthProvider

**Files:**
- Create: `internal/auth/domain/valueobject/oauth_provider.go`
- Create: `internal/auth/domain/valueobject/oauth_provider_test.go`

- [ ] **Step 1: 编写 OAuthProvider 测试**

```go
// internal/auth/domain/valueobject/oauth_provider_test.go
package valueobject_test

import (
    "testing"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestOAuthProviderString(t *testing.T) {
    tests := []struct {
        provider valueobject.OAuthProvider
        want     string
    }{
        {valueobject.OAuthProviderWeChat, "wechat"},
        {valueobject.OAuthProviderGitHub, "github"},
    }

    for _, tt := range tests {
        t.Run(tt.want, func(t *testing.T) {
            if got := tt.provider.String(); got != tt.want {
                t.Errorf("String() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestParseOAuthProvider(t *testing.T) {
    tests := []struct {
        input   string
        want    valueobject.OAuthProvider
        wantErr bool
    }{
        {"wechat", valueobject.OAuthProviderWeChat, false},
        {"github", valueobject.OAuthProviderGitHub, false},
        {"unknown", 0, true},
        {"", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got, err := valueobject.ParseOAuthProvider(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseOAuthProvider() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("ParseOAuthProvider() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/auth/domain/valueobject/... -v -run TestOAuth
```
Expected: FAIL

- [ ] **Step 3: 实现 OAuthProvider**

```go
// internal/auth/domain/valueobject/oauth_provider.go
package valueobject

import "errors"

type OAuthProvider int

const (
    OAuthProviderWeChat OAuthProvider = iota + 1
    OAuthProviderGitHub
)

func (p OAuthProvider) String() string {
    switch p {
    case OAuthProviderWeChat:
        return "wechat"
    case OAuthProviderGitHub:
        return "github"
    default:
        return "unknown"
    }
}

func ParseOAuthProvider(s string) (OAuthProvider, error) {
    switch s {
    case "wechat":
        return OAuthProviderWeChat, nil
    case "github":
        return OAuthProviderGitHub, nil
    default:
        return 0, errors.New("不支持的OAuth提供商: " + s)
    }
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/auth/domain/valueobject/... -v -run TestOAuth
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/domain/valueobject/oauth_provider.go internal/auth/domain/valueobject/oauth_provider_test.go
git commit -m "feat(auth): 添加 OAuthProvider 值对象

- 支持微信和 GitHub 提供商
- 支持字符串解析

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 实体 - User

**Files:**
- Create: `internal/auth/domain/entity/user.go`
- Create: `internal/auth/domain/entity/user_test.go`

- [ ] **Step 1: 编写 User 实体测试**

```go
// internal/auth/domain/entity/user_test.go
package entity_test

import (
    "testing"
    "time"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewUser(t *testing.T) {
    pwd, _ := valueobject.NewPassword("Test1234")
    email, _ := valueobject.NewEmail("test@example.com")

    user, err := entity.NewUser("testuser", pwd, email, nil)
    if err != nil {
        t.Fatalf("NewUser() error = %v", err)
    }

    if user.Username() != "testuser" {
        t.Errorf("Username() = %v, want testuser", user.Username())
    }
    if user.Password() == nil {
        t.Error("Password() should not be nil")
    }
    if user.Email() == nil {
        t.Error("Email() should not be nil")
    }
    if !user.IsActive() {
        t.Error("IsActive() should be true for new user")
    }
}

func TestNewUser_EmptyUsername(t *testing.T) {
    _, err := entity.NewUser("", nil, nil, nil)
    if err == nil {
        t.Error("NewUser() should return error for empty username")
    }
}

func TestUserVerifyEmail(t *testing.T) {
    user, _ := entity.NewUser("testuser", nil, nil, nil)

    if user.EmailVerified() {
        t.Error("EmailVerified() should be false initially")
    }

    user.VerifyEmail()
    if !user.EmailVerified() {
        t.Error("EmailVerified() should be true after VerifyEmail()")
    }
}

func TestUserChangePassword(t *testing.T) {
    oldPwd, _ := valueobject.NewPassword("Test1234")
    user, _ := entity.NewUser("testuser", oldPwd, nil, nil)

    newPwd, _ := valueobject.NewPassword("NewTest12")
    err := user.ChangePassword("Test1234", newPwd)
    if err != nil {
        t.Fatalf("ChangePassword() error = %v", err)
    }

    if !user.Password().Verify("NewTest12") {
        t.Error("Password should be updated")
    }
}

func TestUserChangePassword_WrongOld(t *testing.T) {
    oldPwd, _ := valueobject.NewPassword("Test1234")
    user, _ := entity.NewUser("testuser", oldPwd, nil, nil)

    newPwd, _ := valueobject.NewPassword("NewTest12")
    err := user.ChangePassword("Wrong123", newPwd)
    if err == nil {
        t.Error("ChangePassword() should return error for wrong old password")
    }
}

func TestUserMaskedPhone(t *testing.T) {
    user, _ := entity.NewUser("testuser", nil, nil, nil)
    user.SetAreaCode("86")
    user.SetPhone("13800138000")

    if user.MaskedPhone() != "138****8000" {
        t.Errorf("MaskedPhone() = %v, want 138****8000", user.MaskedPhone())
    }
}

func TestUserFullPhone(t *testing.T) {
    user, _ := entity.NewUser("testuser", nil, nil, nil)
    user.SetAreaCode("86")
    user.SetPhone("13800138000")

    if user.FullPhone() != "+8613800138000" {
        t.Errorf("FullPhone() = %v, want +8613800138000", user.FullPhone())
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/auth/domain/entity/... -v
```
Expected: FAIL

- [ ] **Step 3: 实现 User 实体**

```go
// internal/auth/domain/entity/user.go
package entity

import (
    "errors"
    "time"

    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

type UserStatus int

const (
    UserStatusActive   UserStatus = 1
    UserStatusInactive UserStatus = 2
)

type User struct {
    id            int64
    username      string
    password      *valueobject.Password
    email         *valueobject.Email
    emailVerified bool
    areaCode      string
    phone         string
    phoneVerified bool
    avatar        string
    nickname      string
    status        UserStatus
    createdAt     time.Time
    updatedAt     time.Time
}

func NewUser(username string, password *valueobject.Password, email *valueobject.Email, phone *valueobject.Phone) (*User, error) {
    if username == "" {
        return nil, errors.New("用户名不能为空")
    }

    now := time.Now()
    user := &User{
        username:      username,
        password:      password,
        email:         email,
        emailVerified: false,
        status:        UserStatusActive,
        createdAt:     now,
        updatedAt:     now,
    }

    if phone != nil {
        user.areaCode = phone.AreaCode()
        user.phone = phone.Number()
        user.phoneVerified = false
    }

    return user, nil
}

// NewUserFromData 从数据库重建用户
func NewUserFromData(id int64, username string, emailVerified, phoneVerified bool, status int8, createdAt, updatedAt time.Time) *User {
    return &User{
        id:            id,
        username:      username,
        emailVerified: emailVerified,
        phoneVerified: phoneVerified,
        status:        UserStatus(status),
        createdAt:     createdAt,
        updatedAt:     updatedAt,
    }
}

// Getters
func (u *User) Id() int64                { return u.id }
func (u *User) Username() string         { return u.username }
func (u *User) Password() *valueobject.Password { return u.password }
func (u *User) Email() *valueobject.Email { return u.email }
func (u *User) EmailVerified() bool      { return u.emailVerified }
func (u *User) AreaCode() string         { return u.areaCode }
func (u *User) Phone() string            { return u.phone }
func (u *User) PhoneVerified() bool      { return u.phoneVerified }
func (u *User) Avatar() string           { return u.avatar }
func (u *User) Nickname() string         { return u.nickname }
func (u *User) Status() UserStatus       { return u.status }
func (u *User) CreatedAt() time.Time     { return u.createdAt }
func (u *User) UpdatedAt() time.Time     { return u.updatedAt }

// Setters (用于从数据库重建)
func (u *User) SetId(id int64)                  { u.id = id }
func (u *User) SetPassword(p *valueobject.Password) { u.password = p }
func (u *User) SetEmail(e *valueobject.Email)   { u.email = e }
func (u *User) SetEmailVerified(v bool)         { u.emailVerified = v }
func (u *User) SetAreaCode(s string)            { u.areaCode = s }
func (u *User) SetPhone(s string)               { u.phone = s }
func (u *User) SetPhoneVerified(v bool)         { u.phoneVerified = v }
func (u *User) SetAvatar(s string)              { u.avatar = s }
func (u *User) SetNickname(s string)            { u.nickname = s }
func (u *User) SetCreatedAt(t time.Time)        { u.createdAt = t }
func (u *User) SetUpdatedAt(t time.Time)        { u.updatedAt = t }

// 业务方法
func (u *User) VerifyEmail() {
    u.emailVerified = true
    u.updatedAt = time.Now()
}

func (u *User) VerifyPhone() {
    u.phoneVerified = true
    u.updatedAt = time.Now()
}

func (u *User) ChangePassword(oldPassword string, newPassword *valueobject.Password) error {
    if u.password == nil {
        return errors.New("用户未设置密码")
    }
    if !u.password.Verify(oldPassword) {
        return errors.New("原密码错误")
    }
    u.password = newPassword
    u.updatedAt = time.Now()
    return nil
}

func (u *User) ResetPassword(newPassword *valueobject.Password) {
    u.password = newPassword
    u.updatedAt = time.Now()
}

func (u *User) UpdateProfile(avatar, nickname string) {
    if avatar != "" {
        u.avatar = avatar
    }
    if nickname != "" {
        u.nickname = nickname
    }
    u.updatedAt = time.Now()
}

func (u *User) UpdatePhone(phone *valueobject.Phone) {
    u.areaCode = phone.AreaCode()
    u.phone = phone.Number()
    u.phoneVerified = false
    u.updatedAt = time.Now()
}

func (u *User) Activate() {
    u.status = UserStatusActive
    u.updatedAt = time.Now()
}

func (u *User) Deactivate() {
    u.status = UserStatusInactive
    u.updatedAt = time.Now()
}

func (u *User) IsActive() bool {
    return u.status == UserStatusActive
}

func (u *User) FullPhone() string {
    if u.areaCode == "" || u.phone == "" {
        return ""
    }
    return "+" + u.areaCode + u.phone
}

func (u *User) MaskedPhone() string {
    if u.phone == "" {
        return ""
    }
    if len(u.phone) > 7 {
        return u.phone[:3] + "****" + u.phone[len(u.phone)-4:]
    }
    return u.phone
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/auth/domain/entity/... -v
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/domain/entity/user.go internal/auth/domain/entity/user_test.go
git commit -m "feat(auth): 添加 User 聚合根

- 支持密码修改、重置
- 支持邮箱/手机验证
- 支持资料更新
- 支持状态管理

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 实体 - ThirdPartyBind

**Files:**
- Create: `internal/auth/domain/entity/third_party_bind.go`
- Create: `internal/auth/domain/entity/third_party_bind_test.go`

- [ ] **Step 1: 编写 ThirdPartyBind 实体测试**

```go
// internal/auth/domain/entity/third_party_bind_test.go
package entity_test

import (
    "testing"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewThirdPartyBind(t *testing.T) {
    bind := entity.NewThirdPartyBind(1, valueobject.OAuthProviderGitHub, "12345", "octocat")

    if bind.UserId() != 1 {
        t.Errorf("UserId() = %v, want 1", bind.UserId())
    }
    if bind.Provider() != valueobject.OAuthProviderGitHub {
        t.Errorf("Provider() = %v, want github", bind.Provider())
    }
    if bind.ProviderId() != "12345" {
        t.Errorf("ProviderId() = %v, want 12345", bind.ProviderId())
    }
    if bind.ProviderName() != "octocat" {
        t.Errorf("ProviderName() = %v, want octocat", bind.ProviderName())
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/auth/domain/entity/... -v -run TestThirdParty
```
Expected: FAIL

- [ ] **Step 3: 实现 ThirdPartyBind 实体**

```go
// internal/auth/domain/entity/third_party_bind.go
package entity

import (
    "time"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

type ThirdPartyBind struct {
    id           int64
    userId       int64
    provider     valueobject.OAuthProvider
    providerId   string
    providerName string
    createdAt    time.Time
}

func NewThirdPartyBind(userId int64, provider valueobject.OAuthProvider, providerId, providerName string) *ThirdPartyBind {
    return &ThirdPartyBind{
        userId:       userId,
        provider:     provider,
        providerId:   providerId,
        providerName: providerName,
        createdAt:    time.Now(),
    }
}

// Getters
func (b *ThirdPartyBind) Id() int64                          { return b.id }
func (b *ThirdPartyBind) UserId() int64                      { return b.userId }
func (b *ThirdPartyBind) Provider() valueobject.OAuthProvider { return b.provider }
func (b *ThirdPartyBind) ProviderId() string                 { return b.providerId }
func (b *ThirdPartyBind) ProviderName() string               { return b.providerName }
func (b *ThirdPartyBind) CreatedAt() time.Time               { return b.createdAt }

// Setters
func (b *ThirdPartyBind) SetId(id int64)                     { b.id = id }
func (b *ThirdPartyBind) SetProviderName(name string)        { b.providerName = name }
func (b *ThirdPartyBind) SetCreatedAt(t time.Time)           { b.createdAt = t }
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/auth/domain/entity/... -v -run TestThirdParty
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/domain/entity/third_party_bind.go internal/auth/domain/entity/third_party_bind_test.go
git commit -m "feat(auth): 添加 ThirdPartyBind 实体

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 仓储接口定义

**Files:**
- Create: `internal/auth/domain/repository/user_repository.go`
- Create: `internal/auth/domain/repository/third_party_bind_repository.go`

- [ ] **Step 1: 创建仓储接口**

```go
// internal/auth/domain/repository/user_repository.go
package repository

import (
    "context"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
)

type UserRepository interface {
    Save(ctx context.Context, user *entity.User) error
    Update(ctx context.Context, user *entity.User) error
    FindById(ctx context.Context, id int64) (*entity.User, error)
    FindByUsername(ctx context.Context, username string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindByPhone(ctx context.Context, areaCode, phone string) (*entity.User, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error)
}
```

```go
// internal/auth/domain/repository/third_party_bind_repository.go
package repository

import (
    "context"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

type ThirdPartyBindRepository interface {
    Save(ctx context.Context, bind *entity.ThirdPartyBind) error
    Delete(ctx context.Context, userId int64, provider valueobject.OAuthProvider) error
    FindByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (*entity.ThirdPartyBind, error)
    FindByUserId(ctx context.Context, userId int64) ([]*entity.ThirdPartyBind, error)
    ExistsByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (bool, error)
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./internal/auth/domain/...
```
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/auth/domain/repository/
git commit -m "feat(auth): 添加仓储接口定义

- UserRepository 用户仓储接口
- ThirdPartyBindRepository 第三方绑定仓储接口

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 领域服务接口

**Files:**
- Create: `internal/auth/domain/service/token_service.go`
- Create: `internal/auth/domain/service/oauth_strategy.go`

- [ ] **Step 1: 创建领域服务接口**

```go
// internal/auth/domain/service/token_service.go
package service

import "github.com/yanking/price-watch/internal/auth/domain/entity"

type TokenService interface {
    GenerateToken(user *entity.User) (token string, version int64, err error)
    ParseToken(token string) (userId int64, err error)
    IncrementVersion(userId int64) (int64, error)
    GetVersion(userId int64) (int64, error)
}
```

```go
// internal/auth/domain/service/oauth_strategy.go
package service

type OAuthUserInfo struct {
    ProviderId   string
    ProviderName string
    Email        string
}

type OAuthStrategy interface {
    GetProviderName() string
    GetAuthURL(state string) string
    GetUserInfo(code string) (*OAuthUserInfo, error)
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./internal/auth/domain/...
```
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/auth/domain/service/
git commit -m "feat(auth): 添加领域服务接口

- TokenService 令牌服务接口
- OAuthStrategy OAuth策略接口

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 数据库表结构

**Files:**
- Create: `scripts/sql/auth_tables.sql`

- [ ] **Step 1: 创建数据库表 SQL**

```sql
-- scripts/sql/auth_tables.sql

-- 用户表
CREATE TABLE IF NOT EXISTS users (
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

-- 第三方绑定表
CREATE TABLE IF NOT EXISTS third_party_binds (
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

- [ ] **Step 2: 提交**

```bash
git add scripts/sql/auth_tables.sql
git commit -m "feat(auth): 添加用户表和第三方绑定表 SQL

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 10: DAO 接口

**Files:**
- Create: `internal/auth/infrastructure/persistence/dao/user_dao.go`
- Create: `internal/auth/infrastructure/persistence/dao/third_party_bind_dao.go`

- [ ] **Step 1: 创建 DAO 接口**

```go
// internal/auth/infrastructure/persistence/dao/user_dao.go
package dao

import (
    "context"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
)

type UserDAO interface {
    Insert(ctx context.Context, user *entity.User) error
    Update(ctx context.Context, user *entity.User) error
    FindById(ctx context.Context, id int64) (*entity.User, error)
    FindByUsername(ctx context.Context, username string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindByPhone(ctx context.Context, areaCode, phone string) (*entity.User, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error)
}
```

```go
// internal/auth/infrastructure/persistence/dao/third_party_bind_dao.go
package dao

import (
    "context"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

type ThirdPartyBindDAO interface {
    Insert(ctx context.Context, bind *entity.ThirdPartyBind) error
    Delete(ctx context.Context, userId int64, provider valueobject.OAuthProvider) error
    FindByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (*entity.ThirdPartyBind, error)
    FindByUserId(ctx context.Context, userId int64) ([]*entity.ThirdPartyBind, error)
    ExistsByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (bool, error)
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./internal/auth/infrastructure/persistence/dao/...
```
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add internal/auth/infrastructure/persistence/dao/
git commit -m "feat(auth): 添加 DAO 接口定义

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 11: MySQL PO 和转换器

**Files:**
- Create: `internal/auth/infrastructure/persistence/mysql/user_po.go`
- Create: `internal/auth/infrastructure/persistence/mysql/third_party_bind_po.go`
- Create: `internal/auth/infrastructure/persistence/converter/user_converter.go`
- Create: `internal/auth/infrastructure/persistence/converter/user_converter_test.go`

- [ ] **Step 1: 创建 MySQL PO**

```go
// internal/auth/infrastructure/persistence/mysql/user_po.go
package mysql

import "time"

type UserPO struct {
    Id            int64     `gorm:"column:id;primaryKey"`
    Username      string    `gorm:"column:username;type:varchar(50);uniqueIndex"`
    PasswordHash  *string   `gorm:"column:password_hash;type:varchar(255)"`
    Email         *string   `gorm:"column:email;type:varchar(100)"`
    EmailVerified bool      `gorm:"column:email_verified"`
    AreaCode      *string   `gorm:"column:area_code;type:varchar(10)"`
    Phone         *string   `gorm:"column:phone;type:varchar(20)"`
    PhoneVerified bool      `gorm:"column:phone_verified"`
    Avatar        *string   `gorm:"column:avatar;type:varchar(500)"`
    Nickname      *string   `gorm:"column:nickname;type:varchar(50)"`
    Status        int8      `gorm:"column:status"`
    CreatedAt     time.Time `gorm:"column:created_at"`
    UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (UserPO) TableName() string {
    return "users"
}

// third_party_bind_po.go
type ThirdPartyBindPO struct {
    Id           int64     `gorm:"column:id;primaryKey;autoIncrement"`
    UserId       int64     `gorm:"column:user_id;not null;index"`
    Provider     int8      `gorm:"column:provider;type:tinyint;not null"`
    ProviderId   string    `gorm:"column:provider_id;type:varchar(100);not null"`
    ProviderName *string   `gorm:"column:provider_name;type:varchar(100)"`
    CreatedAt    time.Time `gorm:"column:created_at"`
}

func (ThirdPartyBindPO) TableName() string {
    return "third_party_binds"
}
```

- [ ] **Step 2: 创建转换器**

```go
// internal/auth/infrastructure/persistence/converter/user_converter.go
package converter

import (
    "github.com/yanking/price-watch/internal/auth/domain/entity"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
    "github.com/yanking/price-watch/internal/auth/infrastructure/persistence/mysql"
)

func UserToPO(user *entity.User) *mysql.UserPO {
    po := &mysql.UserPO{
        Id:            user.Id(),
        Username:      user.Username(),
        EmailVerified: user.EmailVerified(),
        PhoneVerified: user.PhoneVerified(),
        Status:        int8(user.Status()),
        CreatedAt:     user.CreatedAt(),
        UpdatedAt:     user.UpdatedAt(),
    }

    if user.Password() != nil {
        hash := user.Password().Hash()
        po.PasswordHash = &hash
    }
    if user.Email() != nil {
        email := user.Email().Value()
        po.Email = &email
    }
    if user.AreaCode() != "" {
        po.AreaCode = &user.AreaCode
    }
    if user.Phone() != "" {
        po.Phone = &user.Phone
    }
    if user.Avatar() != "" {
        po.Avatar = &user.Avatar
    }
    if user.Nickname() != "" {
        po.Nickname = &user.Nickname
    }

    return po
}

func POToUser(po *mysql.UserPO) *entity.User {
    user := entity.NewUserFromData(
        po.Id,
        po.Username,
        po.EmailVerified,
        po.PhoneVerified,
        po.Status,
        po.CreatedAt,
        po.UpdatedAt,
    )

    if po.PasswordHash != nil {
        user.SetPassword(valueobject.NewPasswordFromHash(*po.PasswordHash))
    }
    if po.Email != nil {
        email, _ := valueobject.NewEmail(*po.Email)
        user.SetEmail(email)
    }
    if po.AreaCode != nil {
        user.SetAreaCode(*po.AreaCode)
    }
    if po.Phone != nil {
        user.SetPhone(*po.Phone)
    }
    if po.Avatar != nil {
        user.SetAvatar(*po.Avatar)
    }
    if po.Nickname != nil {
        user.SetNickname(*po.Nickname)
    }

    return user
}
```

- [ ] **Step 3: 编写转换器测试**

```go
// internal/auth/infrastructure/persistence/converter/user_converter_test.go
package converter_test

import (
    "testing"
    "time"
    "github.com/yanking/price-watch/internal/auth/domain/entity"
    "github.com/yanking/price-watch/internal/auth/domain/valueobject"
    "github.com/yanking/price-watch/internal/auth/infrastructure/persistence/converter"
)

func TestUserConverter(t *testing.T) {
    pwd, _ := valueobject.NewPassword("Test1234")
    email, _ := valueobject.NewEmail("test@example.com")
    phone, _ := valueobject.NewPhone("86", "13800138000")

    user, _ := entity.NewUser("testuser", pwd, email, phone)
    user.SetId(1234567890123456789)
    user.SetAvatar("https://example.com/avatar.jpg")
    user.SetNickname("测试用户")

    // DO -> PO
    po := converter.UserToPO(user)
    if po.Username != "testuser" {
        t.Errorf("Username = %v, want testuser", po.Username)
    }
    if po.PasswordHash == nil {
        t.Error("PasswordHash should not be nil")
    }
    if po.Email == nil || *po.Email != "test@example.com" {
        t.Error("Email conversion failed")
    }
    if po.AreaCode == nil || *po.AreaCode != "86" {
        t.Error("AreaCode conversion failed")
    }

    // PO -> DO
    user2 := converter.POToUser(po)
    if user2.Username() != user.Username() {
        t.Error("Username not preserved")
    }
    if !user2.Password().Verify("Test1234") {
        t.Error("Password verification failed after conversion")
    }
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/auth/infrastructure/persistence/converter/... -v
```
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/auth/infrastructure/persistence/mysql/ internal/auth/infrastructure/persistence/converter/
git commit -m "feat(auth): 添加 MySQL PO 和转换器

- UserPO 用户持久化对象
- ThirdPartyBindPO 第三方绑定持久化对象
- UserConverter DO/PO 转换器

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## 后续任务概述

由于计划篇幅较长，以下是后续任务的概述：

### Task 12-15: 基础设施层实现
- MySQL DAO 实现（UserDAO、ThirdPartyBindDAO）
- 仓储实现（委托给 DAO）
- JWT Token 服务实现
- OAuth 策略实现（GitHub、微信）

### Task 16-18: 应用层实现
- DTO 定义
- Assembler 实现
- Auth/User Service 实现

### Task 19-22: 接口层实现
- 统一响应格式
- 认证中间件
- Handler 实现
- 路由注册

### Task 23-25: 配置和初始化
- 配置文件
- 模块初始化
- 集成测试

---

## 依赖安装清单

```bash
# JWT
go get github.com/golang-jwt/jwt/v5

# 密码加密
go get golang.org/x/crypto/bcrypt

# Web 框架
go get github.com/gin-gonic/gin

# 验证
go get github.com/go-playground/validator/v10
```

---

## 测试策略

1. **单元测试**：每个值对象、实体、转换器都有对应测试
2. **集成测试**：DAO 层使用 SQLite 内存数据库测试
3. **API 测试**：使用 httptest 测试 HTTP 接口

---

## 提交规范

每个 Task 完成后单独提交，commit message 格式：
```
feat(auth): 简短描述

详细说明（可选）

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```
