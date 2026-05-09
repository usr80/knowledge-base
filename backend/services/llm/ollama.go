package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaProvider Ollama 本地模型提供商
type OllamaProvider struct {
	config ProviderConfig
}

// ollamaRequest Ollama 请求格式
type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ollamaResponse Ollama 响应格式
type ollamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool   `json:"done"`
	TotalDuration       int64 `json:"total_duration,omitempty"`
	EvalCount          int `json:"eval_count,omitempty"`
	PromptEvalCount    int `json:"prompt_eval_count,omitempty"`
	Error    string `json:"error,omitempty"`
}

// NewOllamaProvider 创建 Ollama 提供商
func NewOllamaProvider(config ProviderConfig) *OllamaProvider {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434/api/chat"
	}
	if config.Model == "" {
		config.Model = "llama3"
	}
	return &OllamaProvider{config: config}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) Models() []string {
	// 返回常用模型，实际可用模型取决于本地安装
	return []string{
		"llama3",
		"llama3.1",
		"llama3.2",
		"qwen2.5",
		"mistral",
		"deepseek-coder",
		"codellama",
		"gemma2",
	}
}

func (p *OllamaProvider) SetModel(model string) {
	p.config.Model = model
}

// Chat 单轮对话
func (p *OllamaProvider) Chat(systemPrompt string, messages []Message) (string, error) {
	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := ollamaRequest{
		Model:    p.config.Model,
		Messages: fullMessages,
		Stream:   false,
		Options: map[string]interface{}{
			"num_predict": p.config.MaxTokens,
			"temperature": p.config.Temperature,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 Ollama 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result ollamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("Ollama 错误: %s", result.Error)
	}

	if result.Message.Content != "" {
		return result.Message.Content, nil
	}

	return "", fmt.Errorf("未返回有效回答")
}

// ChatWithUsage 单轮对话（带 token 使用量）
func (p *OllamaProvider) ChatWithUsage(systemPrompt string, messages []Message) (*ChatResponse, error) {
	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := ollamaRequest{
		Model:    p.config.Model,
		Messages: fullMessages,
		Stream:   false,
		Options: map[string]interface{}{
			"num_predict": p.config.MaxTokens,
			"temperature": p.config.Temperature,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Ollama 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result ollamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("Ollama 错误: %s", result.Error)
	}

	if result.Message.Content == "" {
		return nil, fmt.Errorf("未返回有效回答")
	}

	return &ChatResponse{
		Content:      result.Message.Content,
		InputTokens:  result.PromptEvalCount,  // Ollama 使用 prompt_eval_count
		OutputTokens: result.EvalCount,         // Ollama 使用 eval_count
	}, nil
}

// ChatStream 流式对话（返回 token 使用量）
func (p *OllamaProvider) ChatStream(systemPrompt string, messages []Message, callback func(string)) (*ChatResponse, error) {
	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := ollamaRequest{
		Model:    p.config.Model,
		Messages: fullMessages,
		Stream:   true,
		Options: map[string]interface{}{
			"num_predict": p.config.MaxTokens,
			"temperature": p.config.Temperature,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Ollama 失败: %w", err)
	}
	defer resp.Body.Close()

	// Ollama 返回 JSON Lines 格式，每行一个 JSON 对象
	var inputTokens, outputTokens int

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var result ollamaResponse
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		if result.Error != "" {
			return nil, fmt.Errorf("Ollama 错误: %s", result.Error)
		}

		if result.Message.Content != "" {
			callback(result.Message.Content)
		}

		// 最后一条消息（done=true）包含 token 使用量
		if result.Done {
			inputTokens = result.PromptEvalCount
			outputTokens = result.EvalCount
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &ChatResponse{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}, nil
}
