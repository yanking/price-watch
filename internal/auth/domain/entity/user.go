package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

// 类型别名，方便在 entity 包中使用
type Email = valueobject.Email
type Password = valueobject.Password

// 导出值对象构造函数
var NewEmail = valueobject.NewEmail
var NewPassword = valueobject.NewPassword
var NewPasswordFromHash = valueobject.NewPasswordFromHash

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User 用户聚合根
type User struct {
	id            uint64
	username      string
	password      *Password
	email         *Email
	areaCode      string
	phone         string
	emailVerified bool
	phoneVerified bool
	status        UserStatus
	oauthProvider string
	oauthID       string
	createdAt     time.Time
	updatedAt     time.Time
}

// NewUser 创建新用户
func NewUser(
	username string,
	password *Password,
	email *Email,
) (*User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("用户名不能为空")
	}

	now := time.Now()

	return &User{
		username:      username,
		password:      password,
		email:         email,
		status:        UserStatusActive,
		emailVerified: false,
		phoneVerified: false,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// NewUserFromData 从数据库数据重建用户
func NewUserFromData(
	id uint64,
	username string,
	passwordHash string,
	email string,
	areaCode string,
	phone string,
	emailVerified bool,
	phoneVerified bool,
	status string,
	oauthProvider string,
	oauthID string,
	createdAt time.Time,
	updatedAt time.Time,
) (*User, error) {
	pwd := NewPasswordFromHash(passwordHash)

	var emailVO *Email
	if email != "" {
		e, err := NewEmail(email)
		if err != nil {
			return nil, fmt.Errorf("恢复邮箱失败: %w", err)
		}
		emailVO = e
	}

	return &User{
		id:            id,
		username:      username,
		password:      pwd,
		email:         emailVO,
		areaCode:      areaCode,
		phone:         phone,
		emailVerified: emailVerified,
		phoneVerified: phoneVerified,
		status:        UserStatus(status),
		oauthProvider: oauthProvider,
		oauthID:       oauthID,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

// Getters

func (u *User) ID() uint64 {
	return u.id
}

func (u *User) Username() string {
	return u.username
}

func (u *User) Password() *Password {
	return u.password
}

func (u *User) Email() *Email {
	return u.email
}

func (u *User) AreaCode() string {
	return u.areaCode
}

func (u *User) Phone() string {
	return u.phone
}

func (u *User) EmailVerified() bool {
	return u.emailVerified
}

func (u *User) PhoneVerified() bool {
	return u.phoneVerified
}

func (u *User) Status() UserStatus {
	return u.status
}

func (u *User) OAuthProvider() string {
	return u.oauthProvider
}

func (u *User) OAuthID() string {
	return u.oauthID
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Setters (用于从数据库重建时设置内部状态)

func (u *User) SetID(id uint64) {
	u.id = id
}

func (u *User) SetEmail(email *Email) {
	u.email = email
	u.updatedAt = time.Now()
}

func (u *User) SetPassword(password *Password) {
	u.password = password
	u.updatedAt = time.Now()
}

func (u *User) SetAreaCode(areaCode string) {
	u.areaCode = areaCode
	u.updatedAt = time.Now()
}

func (u *User) SetPhone(phone string) {
	u.phone = phone
	u.updatedAt = time.Now()
}

func (u *User) SetOAuthProvider(provider string, oauthID string) {
	u.oauthProvider = provider
	u.oauthID = oauthID
	u.updatedAt = time.Now()
}

func (u *User) SetStatus(status UserStatus) {
	u.status = status
	u.updatedAt = time.Now()
}

func (u *User) SetCreatedAt(createdAt time.Time) {
	u.createdAt = createdAt
}

func (u *User) SetUpdatedAt(updatedAt time.Time) {
	u.updatedAt = updatedAt
}

// 业务方法

// VerifyEmail 验证邮箱
func (u *User) VerifyEmail() {
	u.emailVerified = true
	u.updatedAt = time.Now()
}

// VerifyPhone 验证手机号
func (u *User) VerifyPhone() {
	u.phoneVerified = true
	u.updatedAt = time.Now()
}

// ChangePassword 修改密码
func (u *User) ChangePassword(oldPassword string, newPassword *Password) error {
	if u.password == nil {
		return errors.New("未设置密码")
	}

	if !u.password.Verify(oldPassword) {
		return errors.New("原密码错误")
	}

	u.password = newPassword
	u.updatedAt = time.Now()
	return nil
}

// ResetPassword 重置密码（管理员操作或找回密码）
func (u *User) ResetPassword(newPassword *Password) {
	u.password = newPassword
	u.updatedAt = time.Now()
}

// UpdateProfile 更新用户资料
func (u *User) UpdateProfile(username string) error {
	if strings.TrimSpace(username) == "" {
		return errors.New("用户名不能为空")
	}

	u.username = username
	u.updatedAt = time.Now()
	return nil
}

// UpdatePhone 更新手机号
func (u *User) UpdatePhone(areaCode string, phone string) {
	u.areaCode = areaCode
	u.phone = phone
	u.phoneVerified = false // 更新手机号后需要重新验证
	u.updatedAt = time.Now()
}

// Activate 激活用户
func (u *User) Activate() {
	u.status = UserStatusActive
	u.updatedAt = time.Now()
}

// Deactive 停用用户
func (u *User) Deactive() {
	u.status = UserStatusInactive
	u.updatedAt = time.Now()
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.status == UserStatusActive
}

// FullPhone 返回完整手机号（带区号）
func (u *User) FullPhone() string {
	if u.areaCode == "" || u.phone == "" {
		return ""
	}
	return fmt.Sprintf("+%s%s", u.areaCode, u.phone)
}

// MaskedPhone 返回脱敏手机号
func (u *User) MaskedPhone() string {
	if u.phone == "" {
		return ""
	}

	// 简单脱敏：保留前3位和后4位
	runes := []rune(u.phone)
	length := len(runes)

	if length <= 7 {
		// 如果太短，只显示前3位
		return string(runes[:3]) + "****"
	}

	return string(runes[:3]) + "****" + string(runes[length-4:])
}
