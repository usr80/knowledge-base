package llm

// ZhipuProvider 智谱 AI 提供商
// 使用 OpenAI 兼容 API 格式
type ZhipuProvider struct {
	*OpenAIProvider
}

// NewZhipuProvider 创建智谱提供商
func NewZhipuProvider(config ProviderConfig) *ZhipuProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	}
	if config.Model == "" {
		config.Model = "glm-4-flash"
	}
	return &ZhipuProvider{
		OpenAIProvider: NewOpenAIProvider(config),
	}
}

func (p *ZhipuProvider) Name() string {
	return "zhipu"
}

func (p *ZhipuProvider) Models() []string {
	return []string{
		"glm-4-plus",
		"glm-4-0520",
		"glm-4-air",
		"glm-4-airx",
		"glm-4-flash",
		"glm-4-long",
	}
}
