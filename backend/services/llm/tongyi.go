package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// TongyiProvider 通义千问提供商
type TongyiProvider struct {
	config ProviderConfig
}

// tongyiRequest 通义千问请求格式
type tongyiRequest struct {
	Model      string           `json:"model"`
	Input      tongyiInput      `json:"input"`
	Parameters *tongyiParams    `json:"parameters,omitempty"`
}

type tongyiInput struct {
	Messages []Message `json:"messages"`
}

type tongyiParams struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// tongyiResponse 通义千问响应格式
type tongyiResponse struct {
	Output struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
		Choices []struct {
			FinishReason string  `json:"finish_reason"`
			Message      Message `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

// tongyiStreamEvent 流式事件（修复：使用 output.text 格式）
type tongyiStreamEvent struct {
	Output struct {
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"output"`
	Usage struct {
		TotalTokens  int `json:"total_tokens"`
		OutputTokens int `json:"output_tokens"`
		InputTokens  int `json:"input_tokens"`
	} `json:"usage"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// NewTongyiProvider 创建通义千问提供商
func NewTongyiProvider(config ProviderConfig) *TongyiProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	}
	if config.Model == "" {
		config.Model = "qwen-turbo"
	}
	return &TongyiProvider{config: config}
}

func (p *TongyiProvider) Name() string {
	return "tongyi"
}

func (p *TongyiProvider) Models() []string {
	return []string{
		"qwen-turbo",
		"qwen-turbo-latest",
		"qwen-plus",
		"qwen-plus-latest",
		"qwen-max",
		"qwen-max-latest",
		"qwen-long",
	}
}

func (p *TongyiProvider) SetModel(model string) {
	p.config.Model = model
}

// Chat 单轮对话
func (p *TongyiProvider) Chat(systemPrompt string, messages []Message) (string, error) {
	if p.config.APIKey == "" {
		return "", fmt.Errorf("请配置通义千问 API Key")
	}

	// 构建完整消息
	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := tongyiRequest{
		Model: p.config.Model,
		Input: tongyiInput{
			Messages: fullMessages,
		},
		Parameters: &tongyiParams{
			MaxTokens:   p.config.MaxTokens,
			Temperature: p.config.Temperature,
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	var result tongyiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != "" && result.Code != "Success" {
		return "", fmt.Errorf("API 错误: [%s] %s", result.Code, result.Message)
	}

	// 优先使用 output.text（非流式响应）
	if result.Output.Text != "" {
		return result.Output.Text, nil
	}

	// 兼容 choices 格式
	if len(result.Output.Choices) > 0 {
		return result.Output.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未返回有效回答")
}

// ChatWithUsage 单轮对话（带 token 使用量）
func (p *TongyiProvider) ChatWithUsage(systemPrompt string, messages []Message) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("请配置通义千问 API Key")
	}

	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := tongyiRequest{
		Model: p.config.Model,
		Input: tongyiInput{
			Messages: fullMessages,
		},
		Parameters: &tongyiParams{
			MaxTokens:   p.config.MaxTokens,
			Temperature: p.config.Temperature,
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
	}

	var result tongyiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != "" && result.Code != "Success" {
		return nil, fmt.Errorf("API 错误: [%s] %s", result.Code, result.Message)
	}

	chatResp := &ChatResponse{
		InputTokens:  result.Usage.InputTokens,
		OutputTokens: result.Usage.OutputTokens,
	}

	// 优先使用 output.text
	if result.Output.Text != "" {
		chatResp.Content = result.Output.Text
	} else if len(result.Output.Choices) > 0 {
		chatResp.Content = result.Output.Choices[0].Message.Content
	} else {
		return nil, fmt.Errorf("未返回有效回答")
	}

	return chatResp, nil
}

// ChatStream 流式对话（修复：正确解析 output.text 格式，计算增量输出，返回 token 使用量）
func (p *TongyiProvider) ChatStream(systemPrompt string, messages []Message, callback func(string)) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("请配置通义千问 API Key")
	}

	fullMessages := buildMessages(systemPrompt, messages)

	reqBody := tongyiRequest{
		Model: p.config.Model,
		Input: tongyiInput{
			Messages: fullMessages,
		},
		Parameters: &tongyiParams{
			MaxTokens:   p.config.MaxTokens,
			Temperature: p.config.Temperature,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 流式 API 使用增量输出
	streamURL := p.config.BaseURL + "?incremental_output=true"

	req, err := http.NewRequest("POST", streamURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-SSE", "enable")

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
	// 通义千问流式格式（incremental_output=true）：
	//   每条 SSE data 的 output.text 是**累计文本**（非增量 delta）
	//   需要计算增量，只发送新增部分
	var lastText string
	var inputTokens, outputTokens int

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// 通义千问 SSE 格式：
		// id:1
		// event:result
		// :HTTP_STATUS/200
		// data:{"output":{"text":"Hello","finish_reason":"null"},...}

		// 跳过非 data 行
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)
		if data == "" || data == "[DONE]" {
			continue
		}

		var result tongyiStreamEvent
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			log.Printf("解析 SSE 数据失败: %v, data: %s", err, data)
			continue
		}

		if result.Code != "" && result.Code != "Success" {
			return nil, fmt.Errorf("API 错误: [%s] %s", result.Code, result.Message)
		}

		// 计算增量：只发送新增部分
		if result.Output.Text != "" && result.Output.Text != lastText {
			increment := strings.TrimPrefix(result.Output.Text, lastText)
			if increment != "" {
				callback(increment)
			}
			lastText = result.Output.Text
		}

		// 捕获 token 使用量（通常在最后一条消息中）
		if result.Usage.InputTokens > 0 || result.Usage.OutputTokens > 0 {
			inputTokens = result.Usage.InputTokens
			outputTokens = result.Usage.OutputTokens
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &ChatResponse{
		Content:      lastText,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}, nil
}

// buildMessages 构建完整消息列表
func buildMessages(systemPrompt string, messages []Message) []Message {
	result := make([]Message, 0)
	if systemPrompt != "" {
		result = append(result, Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	result = append(result, messages...)
	return result
}
