package main

import (
	"fmt"
	"log"

	"claude-relay-core/internal/api/routes"
	"claude-relay-core/internal/config"
	"claude-relay-core/internal/oauth"
	"claude-relay-core/internal/proxy"

	"github.com/gin-gonic/gin"
)

// 适配器类型，解决循环依赖
type oauthClientAdapter struct {
	client *oauth.Client
}

func (a *oauthClientAdapter) RefreshAccessToken(refreshToken string, proxyConfig *proxy.ProxyConfig) (*proxy.OAuthData, error) {
	// 转换ProxyConfig类型
	var oauthProxyConfig *oauth.ProxyConfig
	if proxyConfig != nil {
		oauthProxyConfig = &oauth.ProxyConfig{
			Type:     proxyConfig.Type,
			Host:     proxyConfig.Host,
			Port:     proxyConfig.Port,
			Username: proxyConfig.Username,
			Password: proxyConfig.Password,
		}
	}

	// 调用oauth client
	oauthData, err := a.client.RefreshAccessToken(refreshToken, oauthProxyConfig)
	if err != nil {
		return nil, err
	}

	// 转换返回类型
	var relayProxyConfig *proxy.ProxyConfig
	if oauthData.ProxyConfig != nil {
		relayProxyConfig = &proxy.ProxyConfig{
			Type:     oauthData.ProxyConfig.Type,
			Host:     oauthData.ProxyConfig.Host,
			Port:     oauthData.ProxyConfig.Port,
			Username: oauthData.ProxyConfig.Username,
			Password: oauthData.ProxyConfig.Password,
		}
	}

	return &proxy.OAuthData{
		AccessToken:  oauthData.AccessToken,
		RefreshToken: oauthData.RefreshToken,
		ExpiresAt:    oauthData.ExpiresAt,
		Scopes:       oauthData.Scopes,
		ProxyConfig:  relayProxyConfig,
	}, nil
}

type storageAdapter struct {
	storage *oauth.Storage
}

func (a *storageAdapter) LoadOAuthData(accountName string) (*proxy.OAuthData, error) {
	oauthData, err := a.storage.LoadOAuthData(accountName)
	if err != nil {
		return nil, err
	}

	// 转换类型
	var relayProxyConfig *proxy.ProxyConfig
	if oauthData.ProxyConfig != nil {
		relayProxyConfig = &proxy.ProxyConfig{
			Type:     oauthData.ProxyConfig.Type,
			Host:     oauthData.ProxyConfig.Host,
			Port:     oauthData.ProxyConfig.Port,
			Username: oauthData.ProxyConfig.Username,
			Password: oauthData.ProxyConfig.Password,
		}
	}

	return &proxy.OAuthData{
		AccessToken:  oauthData.AccessToken,
		RefreshToken: oauthData.RefreshToken,
		ExpiresAt:    oauthData.ExpiresAt,
		Scopes:       oauthData.Scopes,
		ProxyConfig:  relayProxyConfig,
	}, nil
}

func (a *storageAdapter) SaveOAuthData(accountName string, data *proxy.OAuthData) error {
	// 转换类型
	var oauthProxyConfig *oauth.ProxyConfig
	if data.ProxyConfig != nil {
		oauthProxyConfig = &oauth.ProxyConfig{
			Type:     data.ProxyConfig.Type,
			Host:     data.ProxyConfig.Host,
			Port:     data.ProxyConfig.Port,
			Username: data.ProxyConfig.Username,
			Password: data.ProxyConfig.Password,
		}
	}

	oauthData := &oauth.OAuthData{
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		ExpiresAt:    data.ExpiresAt,
		Scopes:       data.Scopes,
		ProxyConfig:  oauthProxyConfig,
	}

	return a.storage.SaveOAuthData(accountName, oauthData)
}

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 创建OAuth客户端和存储
	oauthClient := oauth.NewClient(cfg)
	storage := oauth.NewStorage("./data")

	// 创建转发服务
	relayService := proxy.NewRelayService(cfg, &oauthClientAdapter{oauthClient}, &storageAdapter{storage})

	// 创建Gin路由器
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.ReleaseMode) // 设置为发布模式，减少日志输出
	}
	
	router := gin.Default()

	// 设置路由
	routes.SetupRoutes(router, cfg, oauthClient, storage, relayService)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("🚀 Claude Relay Service 启动成功\n")
	fmt.Printf("🌐 服务地址: http://%s\n", addr)
	fmt.Printf("🔗 代理端点: http://%s/api/v1/messages\n", addr)
	fmt.Printf("⚙️  OAuth管理: http://%s/oauth\n", addr)
	
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ 启动服务器失败: %v", err)
	}
}

