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
