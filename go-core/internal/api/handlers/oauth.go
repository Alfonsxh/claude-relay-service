package handlers

import (
	"net/http"

	"claude-relay-core/internal/config"
	"claude-relay-core/internal/oauth"

	"github.com/gin-gonic/gin"
)

// OAuthHandler OAuth处理器
type OAuthHandler struct {
	config      *config.Config
	oauthClient *oauth.Client
	storage     *oauth.Storage
}

// NewOAuthHandler 创建OAuth处理器
func NewOAuthHandler(cfg *config.Config, oauthClient *oauth.Client, storage *oauth.Storage) *OAuthHandler {
	return &OAuthHandler{
		config:      cfg,
		oauthClient: oauthClient,
		storage:     storage,
	}
}

// GenerateAuthURL 生成OAuth授权URL
func (h *OAuthHandler) GenerateAuthURL(c *gin.Context) {
	var req struct {
		ProxyConfig *oauth.ProxyConfig `json:"proxy_config,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 生成OAuth参数
	pkceData, err := oauth.GenerateOAuthParams(h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成OAuth参数失败"})
		return
	}

	// 保存PKCE数据
	if err := h.storage.SavePKCEData(pkceData.State, pkceData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存PKCE数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url":       pkceData.AuthURL,
		"state":          pkceData.State,
		"code_challenge": pkceData.CodeChallenge,
	})
}

// ExchangeToken 交换授权码获取token
func (h *OAuthHandler) ExchangeToken(c *gin.Context) {
	var req struct {
		AuthorizationCode string               `json:"authorization_code"`
		State             string               `json:"state"`
		AccountName       string               `json:"account_name"`
		ProxyConfig       *oauth.ProxyConfig   `json:"proxy_config,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 加载PKCE数据
	pkceData, err := h.storage.LoadPKCEData(req.State)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的state参数"})
		return
	}

	// 交换token
	oauthData, err := h.oauthClient.ExchangeCodeForToken(
		req.AuthorizationCode,
		pkceData.CodeVerifier,
		req.State,
		req.ProxyConfig,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token交换失败: " + err.Error()})
		return
	}

	// 保存OAuth数据
	if err := h.storage.SaveOAuthData(req.AccountName, oauthData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存OAuth数据失败"})
		return
	}

	// 清理PKCE临时数据
	h.storage.DeletePKCEData(req.State)

	c.JSON(http.StatusOK, gin.H{
		"message":    "OAuth认证成功",
		"account":    req.AccountName,
		"expires_at": oauthData.ExpiresAt,
		"scopes":     oauthData.Scopes,
	})
}

// ListAccounts 列出账户
func (h *OAuthHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.storage.ListAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取账户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// GetAccountStatus 检查账户状态
func (h *OAuthHandler) GetAccountStatus(c *gin.Context) {
	accountName := c.Param("name")
	
	oauthData, err := h.storage.LoadOAuthData(accountName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "账户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account":      accountName,
		"is_valid":     oauthData.IsValid(),
		"need_refresh": oauthData.NeedRefresh(),
		"expires_at":   oauthData.ExpiresAt,
		"scopes":       oauthData.Scopes,
	})
}