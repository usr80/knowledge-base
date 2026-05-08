package services

import (
	"knowledge-base/config"
	"knowledge-base/services/llm"
)

// LLMService 大语言模型服务（使用全局 Provider Manager）
type LLMService struct{}

// chatMessage 聊天消息（兼容旧接口）
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewLLMService 创建 LLM 服务（初始化全局 Manager）
func NewLLMService() *LLMService {
	cfg := config.LoadConfig()

	// 构建提供商配置
	providerConfigs := make(map[string]llm.ProviderConfig)

	// 通义千问
	if cfg.AI.DashScopeAPIKey != "" {
		providerConfigs["tongyi"] = llm.ProviderConfig{
			APIKey:      cfg.AI.DashScopeAPIKey,
			Model:       cfg.AI.ChatModel,
			MaxTokens:   cfg.AI.MaxTokens,
			Temperature: cfg.AI.Temperature,
		}
	}

	// OpenAI
	if cfg.AI.OpenAIAPIKey != "" {
		providerConfigs["openai"] = llm.ProviderConfig{
			APIKey:      cfg.AI.OpenAIAPIKey,
			BaseURL:     cfg.AI.OpenAIBaseURL,
			Model:       cfg.AI.OpenAIModel,
			MaxTokens:   cfg.AI.MaxTokens,
			Temperature: cfg.AI.Temperature,
		}
	}

	// DeepSeek
	if cfg.AI.DeepSeekAPIKey != "" {
		providerConfigs["deepseek"] = llm.ProviderConfig{
			APIKey:      cfg.AI.DeepSeekAPIKey,
			Model:       cfg.AI.DeepSeekModel,
			MaxTokens:   cfg.AI.MaxTokens,
			Temperature: cfg.AI.Temperature,
		}
	}

	// 智谱
	if cfg.AI.ZhipuAPIKey != "" {
		providerConfigs["zhipu"] = llm.ProviderConfig{
			APIKey:      cfg.AI.ZhipuAPIKey,
			Model:       cfg.AI.ZhipuModel,
			MaxTokens:   cfg.AI.MaxTokens,
			Temperature: cfg.AI.Temperature,
		}
	}

	// Ollama（本地，不需要 API Key）
	if cfg.AI.OllamaBaseURL != "" || cfg.AI.OllamaModel != "" {
		providerConfigs["ollama"] = llm.ProviderConfig{
			BaseURL:     cfg.AI.OllamaBaseURL,
			Model:       cfg.AI.OllamaModel,
			MaxTokens:   cfg.AI.MaxTokens,
			Temperature: cfg.AI.Temperature,
		}
	}

	// 初始化全局管理器（sync.Once 保证只初始化一次）
	llm.InitManager(providerConfigs, cfg.AI.DefaultProvider)

	return &LLMService{}
}

// Chat 单轮对话（使用全局当前提供商）
func (s *LLMService) Chat(systemPrompt string, messages []chatMessage) (string, error) {
	converted := make([]llm.Message, len(messages))
	for i, m := range messages {
		converted[i] = llm.Message{Role: m.Role, Content: m.Content}
	}
	return llm.GetManager().Chat(systemPrompt, converted)
}

// ChatStream 流式对话（使用全局当前提供商）
func (s *LLMService) ChatStream(systemPrompt string, messages []chatMessage, callback func(string)) error {
	converted := make([]llm.Message, len(messages))
	for i, m := range messages {
		converted[i] = llm.Message{Role: m.Role, Content: m.Content}
	}
	return llm.GetManager().ChatStream(systemPrompt, converted, callback)
}

// ListProviders 列出可用的提供商
func (s *LLMService) ListProviders() []string {
	return llm.GetManager().List()
}

// ListModels 列出指定提供商的模型
func (s *LLMService) ListModels(provider string) ([]string, error) {
	return llm.GetManager().ListModels(provider)
}

// AllModels 列出所有提供商的所有模型
func (s *LLMService) AllModels() map[string][]string {
	return llm.GetManager().AllModels()
}

// SelectModel 切换当前使用的模型（全局生效）
func (s *LLMService) SelectModel(provider, model string) error {
	return llm.GetManager().SelectModel(provider, model)
}

// CurrentProvider 获取当前提供商
func (s *LLMService) CurrentProvider() string {
	return llm.GetManager().CurrentProvider()
}

// CurrentModel 获取当前模型
func (s *LLMService) CurrentModel() string {
	return llm.GetManager().CurrentModel()
}
