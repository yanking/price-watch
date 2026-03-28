package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yanking/price-watch/internal/auth/domain/service"
)

// GitHubOAuthStrategy GitHub OAuth 策略
type GitHubOAuthStrategy struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

// NewGitHubOAuthStrategy 创建 GitHub OAuth 策略
func NewGitHubOAuthStrategy(clientID, clientSecret, redirectURL string) *GitHubOAuthStrategy {
	return &GitHubOAuthStrategy{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		httpClient:   &http.Client{},
	}
}

// GetProviderName 获取提供者名称
func (s *GitHubOAuthStrategy) GetProviderName() string {
	return "github"
}

// GetAuthURL 获取授权 URL
func (s *GitHubOAuthStrategy) GetAuthURL(state string) string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		s.clientID, url.QueryEscape(s.redirectURL), state)
}

// GetUserInfo 通过授权码获取用户信息
func (s *GitHubOAuthStrategy) GetUserInfo(code string) (*service.OAuthUserInfo, error) {
	// 1. 用 code 换取 access_token
	accessToken, err := s.getAccessToken(code)
	if err != nil {
		return nil, fmt.Errorf("获取 GitHub access_token 失败: %w", err)
	}

	// 2. 用 access_token 获取用户信息
	userInfo, err := s.getUserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("获取 GitHub 用户信息失败: %w", err)
	}

	// 3. 获取用户邮箱（GitHub API 可能不返回 primary email）
	email, err := s.getPrimaryEmail(accessToken)
	if err != nil {
		// 邮箱获取失败不影响登录，使用 userInfo 中的邮箱
		email = userInfo.Email
	}

	result := &service.OAuthUserInfo{
		ProviderId:   fmt.Sprintf("%d", userInfo.ID),
		ProviderName: userInfo.Login,
		Email:        email,
	}

	if email == "" && userInfo.Email != "" {
		result.Email = userInfo.Email
	}

	return result, nil
}

// githubTokenResponse GitHub token 响应
type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// githubUserResponse GitHub 用户信息响应
type githubUserResponse struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// githubEmail GitHub 邮箱信息
type githubEmail struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}

// getAccessToken 用 code 换取 access_token
func (s *GitHubOAuthStrategy) getAccessToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.redirectURL)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp githubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("GitHub 未返回 access_token")
	}

	return tokenResp.AccessToken, nil
}

// getUserInfo 用 access_token 获取用户信息
func (s *GitHubOAuthStrategy) getUserInfo(accessToken string) (*githubUserResponse, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var userInfo githubUserResponse
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &userInfo, nil
}

// getPrimaryEmail 获取用户的主要已验证邮箱
func (s *GitHubOAuthStrategy) getPrimaryEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 返回错误状态码 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var emails []githubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}

	// 没有 primary 邮箱，返回第一个
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", nil
}
