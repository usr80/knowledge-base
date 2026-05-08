package llm

// DeepSeekProvider DeepSeek 提供商
// 使用 OpenAI 兼容 API 格式
type DeepSeekProvider struct {
	*OpenAIProvider
}

// NewDeepSeekProvider 创建 DeepSeek 提供商
func NewDeepSeekProvider(config ProviderConfig) *DeepSeekProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.deepseek.com/v1/chat/completions"
	}
	if config.Model == "" {
		config.Model = "deepseek-chat"
	}
	return &DeepSeekProvider{
		OpenAIProvider: NewOpenAIProvider(config),
	}
}

func (p *DeepSeekProvider) Name() string {
	return "deepseek"
}

func (p *DeepSeekProvider) Models() []string {
	return []string{
		"deepseek-chat",
		"deepseek-reasoner",
	}
}
