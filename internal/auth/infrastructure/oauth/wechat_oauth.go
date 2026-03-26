package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yanking/price-watch/internal/auth/domain/service"
)

// WeChatOAuthStrategy 微信 OAuth 策略
type WeChatOAuthStrategy struct {
	appID        string
	appSecret    string
	redirectURL  string
}

// NewWeChatOAuthStrategy 创建微信 OAuth 策略
func NewWeChatOAuthStrategy(appID, appSecret, redirectURL string) *WeChatOAuthStrategy {
	return &WeChatOAuthStrategy{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
	}
}

// GetProviderName 获取提供者名称
func (s *WeChatOAuthStrategy) GetProviderName() string {
	return "wechat"
}

// GetAuthURL 获取授权 URL
func (s *WeChatOAuthStrategy) GetAuthURL(state string) string {
	return fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		s.appID, s.redirectURL, state)
}

// GetUserInfo 通过授权码获取用户信息
func (s *WeChatOAuthStrategy) GetUserInfo(code string) (*service.OAuthUserInfo, error) {
	// 简化实现：模拟获取用户信息
	// 实际应该调用微信开放平台 API:
	// 1. 用 code 换取 access token
	// 2. 用 access token 获取用户信息

	// 模拟响应
	return &service.OAuthUserInfo{
		ProviderId:   "wechat_openid_abc123",
		ProviderName: "wechat",
		Email:        "wechat_user@example.com",
	}, nil
}

// wechatTokenResponse 微信 token 响应
type wechatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// wechatUserInfoResponse 微信用户信息响应
type wechatUserInfoResponse struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

// getAccessToken 获取微信 access token（未实现）
func (s *WeChatOAuthStrategy) getAccessToken(code string) (*wechatTokenResponse, error) {
	// TODO: 实现获取 access token
	return nil, nil
}

// getUserInfo 获取微信用户信息（未实现）
func (s *WeChatOAuthStrategy) getUserInfo(accessToken, openID string) (*wechatUserInfoResponse, error) {
	// TODO: 实现获取用户信息
	return nil, nil
}

// makeWeChatAPIRequest 发起微信 API 请求
func makeWeChatAPIRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return json.Marshal(result)
}
