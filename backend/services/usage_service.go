package services

import (
	"time"

	"knowledge-base/config"
	"knowledge-base/models"
)

// UsageService Token 使用量服务
type UsageService struct{}

// NewUsageService 创建使用量服务
func NewUsageService() *UsageService {
	return &UsageService{}
}

// LogUsage 记录一次使用量
func (s *UsageService) LogUsage(userID uint, provider, model, requestType string,
	inputTokens, outputTokens int, cost float64, sessionID string, documentID uint) error {

	log := models.UsageLog{
		UserID:       userID,
		Provider:     provider,
		Model:        model,
		RequestType:  requestType,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Cost:         cost,
		SessionID:    sessionID,
		DocumentID:   documentID,
		CreatedAt:    time.Now(),
	}

	return config.DB.Create(&log).Error
}

// GetStats 获取用户的使用统计
func (s *UsageService) GetStats(userID uint, startDate, endDate string) (*models.UsageStats, error) {
	stats := &models.UsageStats{}

	query := config.DB.Model(&models.UsageLog{}).Where("user_id = ?", userID)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	// 总计
	var totalRequests int64
	var totalInput, totalOutput int64
	var totalCost float64

	query.Count(&totalRequests)
	query.Select("COALESCE(SUM(input_tokens), 0)").Scan(&totalInput)
	query.Select("COALESCE(SUM(output_tokens), 0)").Scan(&totalOutput)
	query.Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)

	stats.TotalRequests = int(totalRequests)
	stats.TotalInput = totalInput
	stats.TotalOutput = totalOutput
	stats.TotalCost = totalCost

	// 按提供商统计
	config.DB.Model(&models.UsageLog{}).
		Select("provider, COUNT(*) as requests, COALESCE(SUM(input_tokens), 0) as input_tokens, COALESCE(SUM(output_tokens), 0) as output_tokens, COALESCE(SUM(cost), 0) as cost").
		Where("user_id = ?", userID).
		Group("provider").
		Scan(&stats.ByProvider)

	// 按模型统计
	config.DB.Model(&models.UsageLog{}).
		Select("model, COUNT(*) as requests, COALESCE(SUM(input_tokens), 0) as input_tokens, COALESCE(SUM(output_tokens), 0) as output_tokens, COALESCE(SUM(cost), 0) as cost").
		Where("user_id = ?", userID).
		Group("model").
		Scan(&stats.ByModel)

	// 按日期统计（最近 30 天）
	dateQuery := config.DB.Model(&models.UsageLog{}).
		Select("DATE(created_at) as date, COUNT(*) as requests, COALESCE(SUM(input_tokens), 0) as input_tokens, COALESCE(SUM(output_tokens), 0) as output_tokens, COALESCE(SUM(cost), 0) as cost").
		Where("user_id = ?", userID).
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)").
		Group("DATE(created_at)").
		Order("date DESC")

	if startDate != "" {
		dateQuery = dateQuery.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		dateQuery = dateQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}

	dateQuery.Scan(&stats.ByDate)

	return stats, nil
}

// GetRecentLogs 获取最近的使用记录
func (s *UsageService) GetRecentLogs(userID uint, limit int) ([]models.UsageLog, error) {
	var logs []models.UsageLog
	err := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// CalculateCost 计算 token 成本（美元）
// 价格参考（2024年1月，按千tokens计价）：
// - GPT-4: 输入 $0.03, 输出 $0.06
// - GPT-3.5-turbo: 输入 $0.0015, 输出 $0.002
// - DeepSeek: 输入 $0.0001, 输出 $0.0002
// - 通义千问: 输入 $0.0008, 输出 $0.002
// - 智谱 GLM-4: 输入 $0.014, 输出 $0.014
func CalculateCost(provider, model string, inputTokens, outputTokens int) float64 {
	// 每千 tokens 的价格（美元）
	priceMap := map[string]struct{ input, output float64 }{
		"openai-gpt4":      {0.03, 0.06},
		"openai-gpt35":     {0.0015, 0.002},
		"deepseek":         {0.0001, 0.0002},
		"tongyi":           {0.0008, 0.002},
		"zhipu":            {0.014, 0.014},
	}

	key := provider
	if provider == "openai" {
		if model == "gpt-4" || model == "gpt-4-turbo" || model == "gpt-4o" {
			key = "openai-gpt4"
		} else {
			key = "openai-gpt35"
		}
	}

	prices, ok := priceMap[key]
	if !ok {
		// 默认价格（保守估计）
		prices = struct{ input, output float64 }{0.002, 0.002}
	}

	inputCost := float64(inputTokens) / 1000 * prices.input
	outputCost := float64(outputTokens) / 1000 * prices.output

	return inputCost + outputCost
}
