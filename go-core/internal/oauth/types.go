package oauth

import (
	"time"
)

// ProxyConfig 代理配置
type ProxyConfig struct {
	Type     string `json:"type"`     // "socks5", "http", "https"
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// OAuthData OAuth数据结构
type OAuthData struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Scopes       []string     `json:"scopes"`
	ProxyConfig  *ProxyConfig `json:"proxy_config,omitempty"`
}

// PKCEData PKCE流程数据
type PKCEData struct {
	CodeVerifier  string `json:"code_verifier"`
	CodeChallenge string `json:"code_challenge"`
	State         string `json:"state"`
	AuthURL       string `json:"auth_url"`
}

// TokenResponse token响应结构
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// IsValid 检查token是否有效
func (o *OAuthData) IsValid() bool {
	return o.AccessToken != "" && time.Now().Before(o.ExpiresAt.Add(-60*time.Second))
}

// NeedRefresh 检查是否需要刷新token
func (o *OAuthData) NeedRefresh() bool {
	// 提前60秒刷新
	return time.Now().After(o.ExpiresAt.Add(-60*time.Second))
}