package handlers

import (
	"net/http"

	"claude-relay-core/internal/proxy"

	"github.com/gin-gonic/gin"
)

// RelayHandler API转发处理器
type RelayHandler struct {
	relayService *proxy.RelayService
}

// NewRelayHandler 创建API转发处理器
func NewRelayHandler(relayService *proxy.RelayService) *RelayHandler {
	return &RelayHandler{
		relayService: relayService,
	}
}

// ProcessMessages Claude API消息转发
func (h *RelayHandler) ProcessMessages(c *gin.Context) {
	// 获取账户名（从查询参数或头部）
	accountName := c.Query("account")
	if accountName == "" {
		accountName = c.GetHeader("X-Account-Name")
	}
	if accountName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少账户名参数 (account 或 X-Account-Name)"})
		return
	}

	// 解析请求体
	var requestData interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON请求体"})
		return
	}

	// 处理请求
	responseData, err := h.relayService.ProcessRequest(accountName, requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, responseData)
}

// GetModels 获取模型列表（兼容性接口）
func (h *RelayHandler) GetModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []gin.H{
			{
				"id":      "claude-3-5-sonnet-20241022",
				"object":  "model",
				"created": 1234567890,
				"owned_by": "anthropic",
			},
			{
				"id":      "claude-3-5-haiku-20241022",
				"object":  "model",
				"created": 1234567890,
				"owned_by": "anthropic",
			},
		},
	})
}