package config

import "time"

// Config 授权模块配置
type Config struct {
	JWT   JWTConfig   `mapstructure:"jwt"`
	OAuth OAuthConfig `mapstructure:"oauth"`
}

// JWTConfig JWT 令牌配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`      // JWT 签名密钥
	ExpireDays int           `mapstructure:"expire_days"` // 过期天数
}

// ExpireDuration 返回过期时间时长
func (c *JWTConfig) ExpireDuration() time.Duration {
	return time.Duration(c.ExpireDays) * 24 * time.Hour
}

// OAuthConfig OAuth 认证配置
type OAuthConfig struct {
	GitHub OAuthProviderConfig `mapstructure:"github"`
	Wechat OAuthProviderConfig `mapstructure:"wechat"`
}

// OAuthProviderConfig OAuth 提供者配置
type OAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`     // 客户端 ID
	ClientSecret string `mapstructure:"client_secret"` // 客户端密钥
	RedirectURL  string `mapstructure:"redirect_url"`  // 回调地址
}
