package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yanking/price-watch/internal/auth/domain/service"
)

// GitHubOAuthStrategy GitHub OAuth 策略
type GitHubOAuthStrategy struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

// NewGitHubOAuthStrategy 创建 GitHub OAuth 策略
func NewGitHubOAuthStrategy(clientID, clientSecret, redirectURL string) *GitHubOAuthStrategy {
	return &GitHubOAuthStrategy{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
	}
}

// GetProviderName 获取提供者名称
func (s *GitHubOAuthStrategy) GetProviderName() string {
	return "github"
}

// GetAuthURL 获取授权 URL
func (s *GitHubOAuthStrategy) GetAuthURL(state string) string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		s.clientID, s.redirectURL, state)
}

// GetUserInfo 通过授权码获取用户信息
func (s *GitHubOAuthStrategy) GetUserInfo(code string) (*service.OAuthUserInfo, error) {
	// 简化实现：模拟获取用户信息
	// 实际应该调用 GitHub API:
	// 1. 用 code 换取 access token
	// 2. 用 access token 获取用户信息

	// 模拟响应
	return &service.OAuthUserInfo{
		ProviderId:   "github_user_123",
		ProviderName: "github",
		Email:        "github_user@example.com",
	}, nil
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

// getAccessToken 获取 GitHub access token（未实现）
func (s *GitHubOAuthStrategy) getAccessToken(code string) (string, error) {
	// TODO: 实现获取 access token
	return "", nil
}

// getUserInfo 获取 GitHub 用户信息（未实现）
func (s *GitHubOAuthStrategy) getUserInfo(accessToken string) (*githubUserResponse, error) {
	// TODO: 实现获取用户信息
	return nil, nil
}

// makeGitHubAPIRequest 发起 GitHub API 请求
func makeGitHubAPIRequest(url string) ([]byte, error) {
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
