package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 应用程序配置
type Config struct {
	Server ServerConfig `json:"server"`
	OAuth  OAuthConfig  `json:"oauth"`
	Claude ClaudeConfig `json:"claude"`
	Proxy  ProxyConfig  `json:"proxy"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// OAuthConfig OAuth配置
type OAuthConfig struct {
	ClientID     string `json:"client_id"`
	AuthorizeURL string `json:"authorize_url"`
	TokenURL     string `json:"token_url"`
	RedirectURI  string `json:"redirect_uri"`
	Scopes       string `json:"scopes"`
}

// ClaudeConfig Claude API配置
type ClaudeConfig struct {
	APIUrl     string `json:"api_url"`
	APIVersion string `json:"api_version"`
	BetaHeader string `json:"beta_header"`
	Timeout    time.Duration `json:"timeout"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
	
	// 全局代理设置 - 用于所有Claude Code服务器请求
	GlobalProxy *GlobalProxyConfig `json:"global_proxy,omitempty"`
}

// GlobalProxyConfig 全局代理配置
type GlobalProxyConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`     // socks5, http, https
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Load 加载配置
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port: getEnvInt("PORT", 3000),
			Host: getEnvString("HOST", "0.0.0.0"),
		},
		OAuth: OAuthConfig{
			ClientID:     "9d1c250a-e61b-44d9-88ed-5944d1962f5e", // Claude Code固定ClientID
			AuthorizeURL: "https://claude.ai/oauth/authorize",
			TokenURL:     "https://console.anthropic.com/v1/oauth/token",
			RedirectURI:  "https://console.anthropic.com/oauth/code/callback",
			Scopes:       "org:create_api_key user:profile user:inference",
		},
		Claude: ClaudeConfig{
			APIUrl:     getEnvString("CLAUDE_API_URL", "https://api.anthropic.com/v1/messages"),
			APIVersion: getEnvString("CLAUDE_API_VERSION", "2023-06-01"),
			BetaHeader: getEnvString("CLAUDE_BETA_HEADER", "claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,fine-grained-tool-streaming-2025-05-14"),
			Timeout:    getEnvDuration("CLAUDE_TIMEOUT", 30*time.Second),
		},
		Proxy: ProxyConfig{
			Timeout:     getEnvDuration("PROXY_TIMEOUT", 30*time.Second),
			MaxRetries:  getEnvInt("PROXY_MAX_RETRIES", 3),
			GlobalProxy: loadGlobalProxyConfig(),
		},
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", c.Server.Port)
	}

	if c.OAuth.ClientID == "" {
		return fmt.Errorf("OAuth ClientID 不能为空")
	}

	if c.Claude.APIUrl == "" {
		return fmt.Errorf("Claude API URL 不能为空")
	}

	return nil
}

// 辅助函数
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// loadGlobalProxyConfig 加载全局代理配置
func loadGlobalProxyConfig() *GlobalProxyConfig {
	// 检查是否启用全局代理
	enabled := getEnvString("GLOBAL_PROXY_ENABLED", "false") == "true"
	if !enabled {
		return nil
	}

	// 加载代理配置
	proxyType := getEnvString("GLOBAL_PROXY_TYPE", "")
	if proxyType == "" {
		return nil
	}

	host := getEnvString("GLOBAL_PROXY_HOST", "")
	if host == "" {
		return nil
	}

	port := getEnvInt("GLOBAL_PROXY_PORT", 0)
	if port <= 0 {
		return nil
	}

	return &GlobalProxyConfig{
		Enabled:  true,
		Type:     proxyType,
		Host:     host,
		Port:     port,
		Username: getEnvString("GLOBAL_PROXY_USERNAME", ""),
		Password: getEnvString("GLOBAL_PROXY_PASSWORD", ""),
	}
}