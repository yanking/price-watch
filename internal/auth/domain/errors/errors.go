package errors

import "errors"

// 领域错误定义
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrInvalidToken      = errors.New("无效的令牌")
	ErrTokenExpired      = errors.New("令牌已过期")
	ErrOAuthFailed       = errors.New("OAuth 认证失败")
	ErrUsernameExists    = errors.New("用户名已存在")
	ErrEmailExists       = errors.New("邮箱已被注册")
	ErrEmailUsedByOther  = errors.New("邮箱已被其他用户使用")
	ErrUserSuspended     = errors.New("用户已被停用")
	ErrAccountPassword   = errors.New("账号或密码错误")
	ErrBindExists        = errors.New("第三方绑定已存在")
	ErrBindNotFound      = errors.New("第三方绑定不存在")
)
