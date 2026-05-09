package llm

import "io"

// Provider LLM 提供商接口
type Provider interface {
	// Name 返回提供商名称
	Name() string

	// Chat 单轮对话（非流式）
	Chat(systemPrompt string, messages []Message) (string, error)

	// ChatWithUsage 单轮对话（带 token 使用量）
	ChatWithUsage(systemPrompt string, messages []Message) (*ChatResponse, error)

	// ChatStream 流式对话（返回 token 使用量）
	ChatStream(systemPrompt string, messages []Message, callback func(string)) (*ChatResponse, error)

	// Models 返回支持的模型列表
	Models() []string

	// SetModel 设置使用的模型
	SetModel(model string)
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 对话响应（带 token 使用量）
type ChatResponse struct {
	Content      string `json:"content"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name      string  // 提供商名称
	APIKey    string  // API 密钥
	BaseURL   string  // API 地址（可选，用于自定义端点）
	Model     string  // 默认模型
	MaxTokens int     // 最大输出 token
	Temperature float64 // 温度参数
}

// StreamResponse 流式响应
type StreamResponse struct {
	Content      string
	Done         bool
	Error        error
	InputTokens  int
	OutputTokens int
}

// StreamWriter 流式写入器接口
type StreamWriter interface {
	io.Writer
	Flush() error
}
