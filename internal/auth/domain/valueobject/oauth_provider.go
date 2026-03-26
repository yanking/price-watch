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
