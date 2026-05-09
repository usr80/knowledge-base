package models

import (
	"time"
)

// UsageLog Token 使用记录
type UsageLog struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userID"`
	Provider     string    `gorm:"size:50;not null" json:"provider"`     // tongyi / openai / deepseek / zhipu / ollama
	Model        string    `gorm:"size:100;not null" json:"model"`       // 具体模型名称
	RequestType  string    `gorm:"size:20;not null" json:"requestType"`  // embedding / chat / completion
	InputTokens  int       `gorm:"not null" json:"inputTokens"`          // 输入 tokens
	OutputTokens int       `gorm:"not null;default:0" json:"outputTokens"` // 输出 tokens
	Cost         float64   `gorm:"type:decimal(10,6);not null;default:0" json:"cost"` // 费用（美元）
	SessionID    string    `gorm:"index;size:64" json:"sessionID"`       // 关联会话
	DocumentID   uint      `gorm:"index" json:"documentID"`              // 关联文档（embedding）
	CreatedAt    time.Time `json:"createdAt"`
}

func (ul *UsageLog) TableName() string {
	return "usage_logs"
}

// UsageStats 用量统计汇总
type UsageStats struct {
	TotalRequests  int     `json:"totalRequests"`
	TotalInput     int64   `json:"totalInput"`
	TotalOutput    int64   `json:"totalOutput"`
	TotalCost      float64 `json:"totalCost"`
	ByProvider     []ProviderStats `json:"byProvider"`
	ByModel        []ModelStats `json:"byModel"`
	ByDate         []DateStats `json:"byDate"`
}

// ProviderStats 按提供商统计
type ProviderStats struct {
	Provider    string  `json:"provider"`
	Requests    int     `json:"requests"`
	InputTokens int64   `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
	Cost        float64 `json:"cost"`
}

// ModelStats 按模型统计
type ModelStats struct {
	Model       string  `json:"model"`
	Requests    int     `json:"requests"`
	InputTokens int64   `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
	Cost        float64 `json:"cost"`
}

// DateStats 按日期统计
type DateStats struct {
	Date        string  `json:"date"`
	Requests    int     `json:"requests"`
	InputTokens int64   `json:"inputTokens"`
	OutputTokens int64  `json:"outputTokens"`
	Cost        float64 `json:"cost"`
}
