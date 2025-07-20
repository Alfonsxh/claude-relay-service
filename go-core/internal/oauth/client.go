package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"claude-relay-core/internal/config"
)

// Client OAuth客户端
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// NewClient 创建OAuth客户端
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Proxy.Timeout,
		},
	}
}

// ExchangeCodeForToken 交换授权码获取token
func (c *Client) ExchangeCodeForToken(authorizationCode, codeVerifier, state string, proxyConfig *ProxyConfig) (*OAuthData, error) {
	// 清理授权码，移除URL片段
	cleanedCode := strings.Split(authorizationCode, "#")[0]
	cleanedCode = strings.Split(cleanedCode, "&")[0]

	// 构建请求参数
	params := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     c.config.OAuth.ClientID,
		"code":          cleanedCode,
		"redirect_uri":  c.config.OAuth.RedirectURI,
		"code_verifier": codeVerifier,
		"state":         state,
	}

	// 发送token交换请求
	tokenResp, err := c.sendTokenRequest(params, proxyConfig)
	if err != nil {
		return nil, fmt.Errorf("token交换失败: %w", err)
	}

	// 构建OAuth数据
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	scopes := strings.Fields(tokenResp.Scope)

	oauthData := &OAuthData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt,
		Scopes:       scopes,
		ProxyConfig:  proxyConfig,
	}

	return oauthData, nil
}

// RefreshAccessToken 刷新访问token
func (c *Client) RefreshAccessToken(refreshToken string, proxyConfig *ProxyConfig) (*OAuthData, error) {
	// 构建刷新请求参数
	params := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     c.config.OAuth.ClientID,
	}

	// 发送token刷新请求
	tokenResp, err := c.sendTokenRequest(params, proxyConfig)
	if err != nil {
		return nil, fmt.Errorf("token刷新失败: %w", err)
	}

	// 构建OAuth数据
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	scopes := strings.Fields(tokenResp.Scope)

	oauthData := &OAuthData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt,
		Scopes:       scopes,
		ProxyConfig:  proxyConfig,
	}

	return oauthData, nil
}

// sendTokenRequest 发送token请求的通用方法
func (c *Client) sendTokenRequest(params map[string]string, proxyConfig *ProxyConfig) (*TokenResponse, error) {
	// 构建请求体
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.config.OAuth.TokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://claude.ai/")
	req.Header.Set("Origin", "https://claude.ai")

	// 配置代理（优先使用全局代理，然后是请求特定代理）
	httpClient := c.httpClient
	var finalProxyConfig *ProxyConfig
	
	// 优先使用全局代理配置
	if c.config.Proxy.GlobalProxy != nil && c.config.Proxy.GlobalProxy.Enabled {
		finalProxyConfig = &ProxyConfig{
			Type:     c.config.Proxy.GlobalProxy.Type,
			Host:     c.config.Proxy.GlobalProxy.Host,
			Port:     c.config.Proxy.GlobalProxy.Port,
			Username: c.config.Proxy.GlobalProxy.Username,
			Password: c.config.Proxy.GlobalProxy.Password,
		}
	} else if proxyConfig != nil {
		// 如果没有全局代理，使用请求特定的代理
		finalProxyConfig = proxyConfig
	}
	
	if finalProxyConfig != nil {
		transport, err := createProxyTransport(finalProxyConfig)
		if err != nil {
			return nil, fmt.Errorf("创建代理传输失败: %w", err)
		}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   c.config.Proxy.Timeout,
		}
	}

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析token响应失败: %w", err)
	}

	return &tokenResp, nil
}