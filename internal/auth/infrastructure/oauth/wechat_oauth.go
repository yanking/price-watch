package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yanking/price-watch/internal/auth/domain/service"
)

// WeChatOAuthStrategy 微信 OAuth 策略
type WeChatOAuthStrategy struct {
	appID      string
	appSecret  string
	redirectURL string
	httpClient  *http.Client
}

// NewWeChatOAuthStrategy 创建微信 OAuth 策略
func NewWeChatOAuthStrategy(appID, appSecret, redirectURL string) *WeChatOAuthStrategy {
	return &WeChatOAuthStrategy{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
		httpClient:  &http.Client{},
	}
}

// GetProviderName 获取提供者名称
func (s *WeChatOAuthStrategy) GetProviderName() string {
	return "wechat"
}

// GetAuthURL 获取授权 URL
func (s *WeChatOAuthStrategy) GetAuthURL(state string) string {
	return fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		s.appID, url.QueryEscape(s.redirectURL), state)
}

// GetUserInfo 通过授权码获取用户信息
func (s *WeChatOAuthStrategy) GetUserInfo(code string) (*service.OAuthUserInfo, error) {
	// 1. 用 code 换取 access_token + openid
	tokenResp, err := s.getAccessToken(code)
	if err != nil {
		return nil, fmt.Errorf("获取微信 access_token 失败: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信返回错误: errcode=%d, errmsg=%s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// 2. 用 access_token + openid 获取用户信息
	userInfo, err := s.getUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		return nil, fmt.Errorf("获取微信用户信息失败: %w", err)
	}

	return &service.OAuthUserInfo{
		ProviderId:   userInfo.OpenID,
		ProviderName: userInfo.Nickname,
		Email:        "", // 微信不返回邮箱
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
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
}

// getAccessToken 用 code 换取 access_token + openid
func (s *WeChatOAuthStrategy) getAccessToken(code string) (*wechatTokenResponse, error) {
	reqURL := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(s.appID), url.QueryEscape(s.appSecret), url.QueryEscape(code))

	resp, err := s.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp wechatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &tokenResp, nil
}

// getUserInfo 用 access_token + openid 获取用户信息
func (s *WeChatOAuthStrategy) getUserInfo(accessToken, openID string) (*wechatUserInfoResponse, error) {
	reqURL := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		url.QueryEscape(accessToken), url.QueryEscape(openID))

	resp, err := s.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var userInfo wechatUserInfoResponse
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("微信返回错误: errcode=%d, errmsg=%s", userInfo.ErrCode, userInfo.ErrMsg)
	}

	return &userInfo, nil
}
