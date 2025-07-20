package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

// ProxyConfig 代理配置（重复定义，保持模块独立性）
type ProxyConfig struct {
	Type     string `json:"type"`     // "socks5", "http", "https"
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateProxyTransport 创建代理传输
func CreateProxyTransport(proxyConfig *ProxyConfig) (http.RoundTripper, error) {
	if proxyConfig == nil {
		return http.DefaultTransport, nil
	}

	switch proxyConfig.Type {
	case "socks5":
		return createSOCKS5Transport(proxyConfig)
	case "http", "https":
		return createHTTPTransport(proxyConfig)
	default:
		return nil, fmt.Errorf("不支持的代理类型: %s", proxyConfig.Type)
	}
}

// createSOCKS5Transport 创建SOCKS5代理传输
func createSOCKS5Transport(proxyConfig *ProxyConfig) (http.RoundTripper, error) {
	// 构建SOCKS5地址
	addr := fmt.Sprintf("%s:%d", proxyConfig.Host, proxyConfig.Port)

	var auth *proxy.Auth
	if proxyConfig.Username != "" && proxyConfig.Password != "" {
		auth = &proxy.Auth{
			User:     proxyConfig.Username,
			Password: proxyConfig.Password,
		}
	}

	// 创建SOCKS5代理拨号器
	dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("创建SOCKS5代理失败: %w", err)
	}

	// 创建自定义传输
	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	return transport, nil
}

// createHTTPTransport 创建HTTP代理传输
func createHTTPTransport(proxyConfig *ProxyConfig) (http.RoundTripper, error) {
	// 构建代理URL
	var proxyURL *url.URL
	var err error

	if proxyConfig.Username != "" && proxyConfig.Password != "" {
		// 带认证的代理URL
		proxyURL, err = url.Parse(fmt.Sprintf("%s://%s:%s@%s:%d",
			proxyConfig.Type,
			url.QueryEscape(proxyConfig.Username),
			url.QueryEscape(proxyConfig.Password),
			proxyConfig.Host,
			proxyConfig.Port))
	} else {
		// 不带认证的代理URL
		proxyURL, err = url.Parse(fmt.Sprintf("%s://%s:%d",
			proxyConfig.Type,
			proxyConfig.Host,
			proxyConfig.Port))
	}

	if err != nil {
		return nil, fmt.Errorf("解析代理URL失败: %w", err)
	}

	// 创建自定义传输
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return transport, nil
}

// CreateProxyClient 创建代理HTTP客户端
func CreateProxyClient(proxyConfig *ProxyConfig) (*http.Client, error) {
	transport, err := CreateProxyTransport(proxyConfig)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: transport,
	}, nil
}

// TestProxyConnection 测试代理连接
func TestProxyConnection(proxyConfig *ProxyConfig) error {
	if proxyConfig == nil {
		return nil
	}

	// 简单的连接测试
	addr := fmt.Sprintf("%s:%d", proxyConfig.Host, proxyConfig.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("代理连接测试失败: %w", err)
	}
	defer conn.Close()

	return nil
}