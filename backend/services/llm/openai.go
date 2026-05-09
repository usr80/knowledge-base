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

// OpenAIProvider OpenAI 及兼容 API 提供商
// 支持官方 OpenAI、Azure OpenAI、以及所有兼容 OpenAI API 格式的服务
type OpenAIProvider struct {
	config ProviderConfig
}

// openaiRequest OpenAI 请求格式
type openaiRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// openaiResponse OpenAI 响应格式
type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message,omitempty"`
		Delta        Message `json:"delta,omitempty"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(config ProviderConfig) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1/chat/completions"
	}
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}
	return &OpenAIProvider{config: config}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Models() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

func (p *OpenAIProvider) SetModel(model string) {
	p.config.Model = model
}

// Chat 单轮对话
func (p *OpenAIProvider) Chat(systemPrompt string, messages []Message) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("请配置 OpenAI API Key")
	}

	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := openaiRequest{
		Model:       p.config.Model,
		Messages:    fullMessages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 API 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result openaiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API 错误: [%s] %s", result.Error.Code, result.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未返回有效回答")
}

// ChatWithUsage 单轮对话（带 token 使用量）
func (p *OpenAIProvider) ChatWithUsage(systemPrompt string, messages []Message) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("请配置 OpenAI API Key")
	}

	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := openaiRequest{
		Model:       p.config.Model,
		Messages:    fullMessages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 API 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result openaiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API 错误: [%s] %s", result.Error.Code, result.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("未返回有效回答")
	}

	return &ChatResponse{
		Content:      result.Choices[0].Message.Content,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
	}, nil
}

// ChatStream 流式对话（返回 token 使用量）
func (p *OpenAIProvider) ChatStream(systemPrompt string, messages []Message, callback func(string)) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("请配置 OpenAI API Key")
	}

	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := openaiRequest{
		Model:       p.config.Model,
		Messages:    fullMessages,
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Stream:      true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	// 解析 SSE 流
	var inputTokens, outputTokens int

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)
		if data == "" || data == "[DONE]" {
			continue
		}

		var result openaiResponse
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			continue
		}

		if result.Error != nil {
			return nil, fmt.Errorf("API 错误: [%s] %s", result.Error.Code, result.Error.Message)
		}

		if len(result.Choices) > 0 {
			content := result.Choices[0].Delta.Content
			if content != "" {
				callback(content)
			}
		}

		// 捕获 token 使用量（在最后一条消息或 usage 专属消息中）
		if result.Usage.PromptTokens > 0 || result.Usage.CompletionTokens > 0 {
			inputTokens = result.Usage.PromptTokens
			outputTokens = result.Usage.CompletionTokens
		}

		if len(result.Choices) > 0 && result.Choices[0].FinishReason == "stop" {
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