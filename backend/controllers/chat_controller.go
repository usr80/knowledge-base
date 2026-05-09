package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"knowledge-base/config"
	"knowledge-base/models"
	"knowledge-base/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatController struct {
	ragSvc *services.RAGService
}

func NewChatController() *ChatController {
	return &ChatController{
		ragSvc: services.NewRAGService(),
	}
}

// AskRequest 提问请求
type AskRequest struct {
	Question    string `json:"question" binding:"required"`
	DocumentIDs []uint `json:"documentIDs"` // 可选：限定检索文档范围
	SessionID   string `json:"sessionID"`   // 可选：会话 ID（用于多轮对话）
}

// AskResponse 提问响应
type AskResponse struct {
	SessionID string `json:"sessionID"`
	Answer    string `json:"answer"`
}

// Ask 提问接口
func (c *ChatController) Ask(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req AskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 生成或使用会话 ID
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// 保存用户消息
	userMsg := models.ChatMessage{
		UserID:    userID,
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Question,
	}
	config.DB.Create(&userMsg)

	// 更新或创建会话
	var session models.ChatSession
	result := config.DB.Where("session_id = ?", sessionID).First(&session)
	if result.Error != nil {
		// 新会话
		title := req.Question
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		session = models.ChatSession{
			UserID:    userID,
			SessionID: sessionID,
			Title:     title,
		}
		config.DB.Create(&session)
	} else {
		// 更新时间
		config.DB.Model(&session).Update("updated_at", time.Now())
	}

	// 调用 RAG 服务
	answer, err := c.ragSvc.Ask(userID, req.Question, req.DocumentIDs, sessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "回答生成失败: " + err.Error()})
		return
	}

	// 保存助手消息
	assistantMsg := models.ChatMessage{
		UserID:    userID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   answer,
	}
	config.DB.Create(&assistantMsg)

	ctx.JSON(http.StatusOK, AskResponse{
		SessionID: sessionID,
		Answer:    answer,
	})
}

// AskStream 流式提问接口
func (c *ChatController) AskStream(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req AskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 生成或使用会话 ID
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// 保存用户消息
	userMsg := models.ChatMessage{
		UserID:    userID,
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Question,
	}
	config.DB.Create(&userMsg)

	// 更新或创建会话
	var session models.ChatSession
	result := config.DB.Where("session_id = ?", sessionID).First(&session)
	if result.Error != nil {
		title := req.Question
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		session = models.ChatSession{
			UserID:    userID,
			SessionID: sessionID,
			Title:     title,
		}
		config.DB.Create(&session)
	} else {
		config.DB.Model(&session).Update("updated_at", time.Now())
	}

	// 设置 SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	var answerBuilder strings.Builder

	err := c.ragSvc.AskStream(userID, req.Question, req.DocumentIDs, sessionID, func(chunk string) {
		answerBuilder.WriteString(chunk)
		// 发送 SSE 事件
		ctx.SSEvent("message", gin.H{"content": chunk})
		ctx.Writer.Flush()
	})

	if err != nil {
		ctx.SSEvent("error", gin.H{"error": err.Error()})
		ctx.Writer.Flush()
		return
	}

	// 保存完整回答
	assistantMsg := models.ChatMessage{
		UserID:    userID,
		SessionID: sessionID,
		Role:      "assistant",
		Content:   answerBuilder.String(),
	}
	config.DB.Create(&assistantMsg)

	// 发送结束事件
	ctx.SSEvent("done", gin.H{"sessionID": sessionID})
	ctx.Writer.Flush()
}

// IndexDocument 创建文档索引
func (c *ChatController) IndexDocument(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	docIDStr := ctx.Param("id")
	docID, err := strconv.ParseUint(docIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档 ID"})
		return
	}

	log.Printf("[IndexDocument] userID=%d, docID=%d", userID, docID)

	// 调用 RAG 服务创建索引
	err = c.ragSvc.IndexDocument(uint(docID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "文档索引创建成功",
		"documentID": docID,
	})
}

// ListSessions 获取会话列表
func (c *ChatController) ListSessions(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var sessions []models.ChatSession
	config.DB.Where("user_id = ?", userID).Order("updated_at DESC").Limit(50).Find(&sessions)

	ctx.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// GetSession 获取会话详情（消息列表）
func (c *ChatController) GetSession(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	sessionID := ctx.Param("id")

	var session models.ChatSession
	result := config.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&session)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	var messages []models.ChatMessage
	config.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages)

	ctx.JSON(http.StatusOK, gin.H{
		"session":  session,
		"messages": messages,
	})
}

// DeleteSession 删除会话
func (c *ChatController) DeleteSession(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	sessionID := ctx.Param("id")

	// 删除消息
	config.DB.Where("session_id = ? AND user_id IN (SELECT id FROM chat_sessions WHERE user_id = ?)", sessionID, userID).Delete(&models.ChatMessage{})
	
	// 删除会话
	result := config.DB.Where("session_id = ? AND user_id = ?", sessionID, userID).Delete(&models.ChatSession{})
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListModels 获取可用的模型列表
func (c *ChatController) ListModels(ctx *gin.Context) {
	llmSvc := services.NewLLMService()
	models := llmSvc.AllModels()
	providers := llmSvc.ListProviders()
	
	ctx.JSON(http.StatusOK, gin.H{
		"providers":        providers,
		"models":           models,
		"currentProvider":  llmSvc.CurrentProvider(),
		"currentModel":     llmSvc.CurrentModel(),
	})
}

// SelectModelRequest 选择模型请求
type SelectModelRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// SelectModel 选择使用的模型
func (c *ChatController) SelectModel(ctx *gin.Context) {
	var req SelectModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	llmSvc := services.NewLLMService()
	if err := llmSvc.SelectModel(req.Provider, req.Model); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "模型切换成功",
		"provider": req.Provider,
		"model":    req.Model,
	})
}

// GetUsageStats 获取用量统计
func (c *ChatController) GetUsageStats(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")

	usageSvc := services.NewUsageService()
	stats, err := usageSvc.GetStats(userID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// GetUsageLogs 获取最近使用记录
func (c *ChatController) GetUsageLogs(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	limit := 50
	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	usageSvc := services.NewUsageService()
	logs, err := usageSvc.GetRecentLogs(userID, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取记录失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"logs": logs})
}
