package proxy

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

// OAuthData OAuth数据结构（简化版，避免循环导入）
type OAuthData struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Scopes       []string     `json:"scopes"`
	ProxyConfig  *ProxyConfig `json:"proxy_config,omitempty"`
}

// Storage 存储接口
type Storage interface {
	LoadOAuthData(accountName string) (*OAuthData, error)
	SaveOAuthData(accountName string, data *OAuthData) error
}

// OAuthClient OAuth客户端接口
type OAuthClient interface {
	RefreshAccessToken(refreshToken string, proxyConfig *ProxyConfig) (*OAuthData, error)
}

// RelayService 请求转发服务
type RelayService struct {
	config      *config.Config
	oauthClient OAuthClient
	storage     Storage
}

// NewRelayService 创建转发服务
func NewRelayService(cfg *config.Config, oauthClient OAuthClient, storage Storage) *RelayService {
	return &RelayService{
		config:      cfg,
		oauthClient: oauthClient,
		storage:     storage,
	}
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

// RelayRequest 转发请求到Claude API
func (r *RelayService) RelayRequest(accountName string, requestBody []byte) (*http.Response, error) {
	// 1. 获取有效的OAuth token
	oauthData, err := r.getValidToken(accountName)
	if err != nil {
		return nil, fmt.Errorf("获取有效token失败: %w", err)
	}

	// 2. 创建HTTP客户端（支持代理）
	httpClient, err := r.createHTTPClient(oauthData.ProxyConfig)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP客户端失败: %w", err)
	}

	// 3. 构建Claude API请求
	req, err := r.buildClaudeRequest(requestBody, oauthData.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("构建Claude API请求失败: %w", err)
	}

	// 4. 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送Claude API请求失败: %w", err)
	}

	// 5. 检查响应状态
	if err := r.handleResponse(resp, accountName); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

// getValidToken 获取有效的OAuth token
func (r *RelayService) getValidToken(accountName string) (*OAuthData, error) {
	// 加载OAuth数据
	oauthData, err := r.storage.LoadOAuthData(accountName)
	if err != nil {
		return nil, fmt.Errorf("加载OAuth数据失败: %w", err)
	}

	// 检查token是否需要刷新
	if oauthData.NeedRefresh() {
		fmt.Printf("🔄 Token即将过期，正在刷新...\n")
		
		// 刷新token
		newOAuthData, err := r.oauthClient.RefreshAccessToken(oauthData.RefreshToken, oauthData.ProxyConfig)
		if err != nil {
			return nil, fmt.Errorf("刷新token失败: %w", err)
		}

		// 保存新的OAuth数据
		if err := r.storage.SaveOAuthData(accountName, newOAuthData); err != nil {
			return nil, fmt.Errorf("保存刷新后的OAuth数据失败: %w", err)
		}

		fmt.Printf("✅ Token刷新成功\n")
		return newOAuthData, nil
	}

	// 检查token是否有效
	if !oauthData.IsValid() {
		return nil, fmt.Errorf("token已过期且无法刷新")
	}

	return oauthData, nil
}

// createHTTPClient 创建HTTP客户端
func (r *RelayService) createHTTPClient(proxyConfig *ProxyConfig) (*http.Client, error) {
	var finalProxyConfig *ProxyConfig
	
	// 优先使用全局代理配置
	if r.config.Proxy.GlobalProxy != nil && r.config.Proxy.GlobalProxy.Enabled {
		finalProxyConfig = &ProxyConfig{
			Type:     r.config.Proxy.GlobalProxy.Type,
			Host:     r.config.Proxy.GlobalProxy.Host,
			Port:     r.config.Proxy.GlobalProxy.Port,
			Username: r.config.Proxy.GlobalProxy.Username,
			Password: r.config.Proxy.GlobalProxy.Password,
		}
	} else if proxyConfig != nil {
		// 如果没有全局代理，使用账户特定的代理
		finalProxyConfig = proxyConfig
	}
	
	transport, err := CreateProxyTransport(finalProxyConfig)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: transport,
		Timeout:   r.config.Claude.Timeout,
	}, nil
}

// buildClaudeRequest 构建Claude API请求
func (r *RelayService) buildClaudeRequest(requestBody []byte, accessToken string) (*http.Request, error) {
	// 创建请求
	req, err := http.NewRequest("POST", r.config.Claude.APIUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("anthropic-version", r.config.Claude.APIVersion)
	req.Header.Set("User-Agent", "claude-cli/1.0.53 (external, cli)")

	// 设置beta header
	if r.config.Claude.BetaHeader != "" {
		req.Header.Set("anthropic-beta", r.config.Claude.BetaHeader)
	}

	return req, nil
}

// handleResponse 处理响应
func (r *RelayService) handleResponse(resp *http.Response, accountName string) error {
	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 读取错误响应
		body, _ := io.ReadAll(resp.Body)
		
		// 检查是否为限流错误
		if r.isRateLimitError(string(body)) {
			fmt.Printf("🚫 检测到限流错误 (账户: %s)\n", accountName)
		}
		
		return fmt.Errorf("Claude API错误 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// isRateLimitError 检查是否为限流错误
func (r *RelayService) isRateLimitError(responseBody string) bool {
	lowerBody := strings.ToLower(responseBody)
	return strings.Contains(lowerBody, "rate limit") || 
		   strings.Contains(lowerBody, "exceed your account's rate limit")
}

// ProcessRequest 处理完整的请求流程
func (r *RelayService) ProcessRequest(accountName string, requestData interface{}) (interface{}, error) {
	// 序列化请求数据
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	fmt.Printf("📤 正在处理API请求 (账户: %s)\n", accountName)

	// 转发请求
	resp, err := r.RelayRequest(accountName, requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var responseData interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	fmt.Printf("✅ API请求处理完成\n")
	return responseData, nil
}