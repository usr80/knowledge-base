package llm

import (
	"fmt"
	"sync"
)

// Manager LLM 提供商管理器
type Manager struct {
	providers       map[string]Provider
	defaultProvider string
	currentProvider string
	currentModel    string
	mu              sync.RWMutex
}

// NewManager 创建管理器
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// Register 注册提供商
func (m *Manager) Register(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// SetDefault 设置默认提供商
func (m *Manager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.providers[name]; !ok {
		return fmt.Errorf("提供商 %s 未注册", name)
	}
	m.defaultProvider = name
	if m.currentProvider == "" {
		m.currentProvider = name
	}
	return nil
}

// Get 获取提供商
func (m *Manager) Get(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if name == "" {
		name = m.currentProvider
	}
	if name == "" {
		name = m.defaultProvider
	}
	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("提供商 %s 未注册", name)
	}
	return provider, nil
}

// Default 获取默认提供商
func (m *Manager) Default() (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.defaultProvider == "" {
		return nil, fmt.Errorf("未设置默认提供商")
	}
	return m.providers[m.defaultProvider], nil
}

// List 列出所有提供商
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, 0, len(m.providers))
	for name := range m.providers {
		result = append(result, name)
	}
	return result
}

// ListModels 列出指定提供商的所有模型
func (m *Manager) ListModels(providerName string) ([]string, error) {
	provider, err := m.Get(providerName)
	if err != nil {
		return nil, err
	}
	return provider.Models(), nil
}

// AllModels 列出所有提供商的所有模型
func (m *Manager) AllModels() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string][]string)
	for name, provider := range m.providers {
		result[name] = provider.Models()
	}
	return result
}

// Chat 使用当前提供商进行对话
func (m *Manager) Chat(systemPrompt string, messages []Message) (string, error) {
	provider, err := m.Get("")
	if err != nil {
		return "", err
	}
	return provider.Chat(systemPrompt, messages)
}

// ChatWithUsage 使用当前提供商进行对话（带 token 使用量）
func (m *Manager) ChatWithUsage(systemPrompt string, messages []Message) (*ChatResponse, error) {
	provider, err := m.Get("")
	if err != nil {
		return nil, err
	}
	return provider.ChatWithUsage(systemPrompt, messages)
}

// ChatStream 使用当前提供商进行流式对话（返回 token 使用量）
func (m *Manager) ChatStream(systemPrompt string, messages []Message, callback func(string)) (*ChatResponse, error) {
	provider, err := m.Get("")
	if err != nil {
		return nil, err
	}
	return provider.ChatStream(systemPrompt, messages, callback)
}

// SelectModel 全局切换当前使用的提供商和模型
func (m *Manager) SelectModel(providerName, model string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if providerName != "" {
		p, ok := m.providers[providerName]
		if !ok {
			return fmt.Errorf("提供商 %s 未注册", providerName)
		}
		m.currentProvider = providerName
		if model != "" {
			p.SetModel(model)
		}
		m.currentModel = model
	} else if model != "" {
		// 只切换模型，保持当前提供商
		if m.currentProvider == "" {
			m.currentProvider = m.defaultProvider
		}
		if p, ok := m.providers[m.currentProvider]; ok {
			p.SetModel(model)
		}
		m.currentModel = model
	}

	return nil
}

// CurrentProvider 获取当前提供商名称
func (m *Manager) CurrentProvider() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentProvider
}

// CurrentModel 获取当前模型名称
func (m *Manager) CurrentModel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentModel
}

// 全局管理器实例
var GlobalManager *Manager
var managerOnce sync.Once

// InitManager 初始化全局管理器
func InitManager(configs map[string]ProviderConfig, defaultProvider string) *Manager {
	managerOnce.Do(func() {
		GlobalManager = NewManager()

		// 根据配置注册提供商
		for name, cfg := range configs {
			var provider Provider
			switch name {
			case "tongyi":
				provider = NewTongyiProvider(cfg)
			case "openai":
				provider = NewOpenAIProvider(cfg)
			case "deepseek":
				provider = NewDeepSeekProvider(cfg)
			case "zhipu":
				provider = NewZhipuProvider(cfg)
			case "ollama":
				provider = NewOllamaProvider(cfg)
			default:
				continue
			}
			GlobalManager.Register(provider)
		}

		// 设置默认提供商
		if defaultProvider != "" {
			GlobalManager.SetDefault(defaultProvider)
		}
	})

	return GlobalManager
}

// GetManager 获取全局管理器
func GetManager() *Manager {
	if GlobalManager == nil {
		GlobalManager = NewManager()
	}
	return GlobalManager
}
