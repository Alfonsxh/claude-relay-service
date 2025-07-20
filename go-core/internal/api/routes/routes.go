package routes

import (
	"net/http"

	"claude-relay-core/internal/api/handlers"
	"claude-relay-core/internal/config"
	"claude-relay-core/internal/oauth"
	"claude-relay-core/internal/proxy"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine, cfg *config.Config, oauthClient *oauth.Client, storage *oauth.Storage, relayService *proxy.RelayService) {
	// 创建处理器
	oauthHandler := handlers.NewOAuthHandler(cfg, oauthClient, storage)
	relayHandler := handlers.NewRelayHandler(relayService)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "claude-relay-service",
			"version": "1.0.0-mvp",
		})
	})

	// OAuth管理路由组
	setupOAuthRoutes(router, oauthHandler)

	// API转发路由组
	setupAPIRoutes(router, relayHandler)

	// 根路径信息
	setupRootRoute(router)
}

// setupOAuthRoutes 设置OAuth相关路由
func setupOAuthRoutes(router *gin.Engine, handler *handlers.OAuthHandler) {
	oauthGroup := router.Group("/oauth")
	{
		// 生成OAuth授权URL
		oauthGroup.POST("/auth-url", handler.GenerateAuthURL)

		// 交换授权码获取token
		oauthGroup.POST("/token", handler.ExchangeToken)

		// 列出账户
		oauthGroup.GET("/accounts", handler.ListAccounts)

		// 检查账户状态
		oauthGroup.GET("/accounts/:name/status", handler.GetAccountStatus)
	}
}

// setupAPIRoutes 设置API转发相关路由
func setupAPIRoutes(router *gin.Engine, handler *handlers.RelayHandler) {
	apiGroup := router.Group("/api/v1")
	{
		// Claude API消息转发
		apiGroup.POST("/messages", handler.ProcessMessages)

		// 简单的模型列表兼容接口
		apiGroup.GET("/models", handler.GetModels)
	}
}

// setupRootRoute 设置根路径路由
func setupRootRoute(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "Claude Relay Service MVP",
			"version":     "1.0.0-mvp",
			"description": "Claude Code OAuth认证 + 请求转发服务",
			"endpoints": gin.H{
				"health":         "GET /health",
				"oauth_auth_url": "POST /oauth/auth-url",
				"oauth_token":    "POST /oauth/token", 
				"oauth_accounts": "GET /oauth/accounts",
				"api_messages":   "POST /api/v1/messages",
				"api_models":     "GET /api/v1/models",
			},
			"usage": gin.H{
				"1": "首先调用 POST /oauth/auth-url 生成授权URL",
				"2": "访问授权URL完成Claude Code认证",
				"3": "使用授权码调用 POST /oauth/token 完成认证",
				"4": "使用 POST /api/v1/messages?account=账户名 转发请求",
			},
		})
	})
}