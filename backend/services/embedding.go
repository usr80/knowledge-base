package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"knowledge-base/config"
)

// EmbeddingService 向量嵌入服务
type EmbeddingService struct {
	apiKey     string
	model      string
	baseURL    string
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *embeddingParams `json:"parameters,omitempty"`
}

type embeddingParams struct {
	Truncate bool `json:"truncate,omitempty"`
}

type embeddingResponse struct {
	Output struct {
		Embeddings []struct {
			Embedding []float64 `json:"embedding"`
			TextIndex int       `json:"text_index"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

func NewEmbeddingService() *EmbeddingService {
	cfg := config.LoadConfig()
	return &EmbeddingService{
		apiKey:  cfg.AI.DashScopeAPIKey,
		model:   cfg.AI.EmbeddingModel,
		baseURL: "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding",
	}
}

// GetEmbedding 获取单个文本的向量
func (s *EmbeddingService) GetEmbedding(text string) ([]float64, error) {
	vecs, err := s.GetEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("未获取到嵌入向量")
	}
	return vecs[0], nil
}

// GetEmbeddings 批量获取文本向量
func (s *EmbeddingService) GetEmbeddings(texts []string) ([][]float64, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("请配置 DASHSCOPE_API_KEY")
	}

	// 限制单次请求文本数（通义千问限制 10 条）
	batchSize := 10
	var allEmbeddings [][]float64

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		reqBody := embeddingRequest{
			Model: s.model,
		}
		reqBody.Input.Texts = batch

		body, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("序列化请求失败: %w", err)
		}

		req, err := http.NewRequest("POST", s.baseURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+s.apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求嵌入 API 失败: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("嵌入 API 返回错误 (status=%d): %s", resp.StatusCode, string(respBody))
		}

		var result embeddingResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		if result.Code != "" && result.Code != "Success" {
			return nil, fmt.Errorf("嵌入 API 错误: [%s] %s", result.Code, result.Message)
		}

		// 按 text_index 排序
		ordered := make([][]float64, len(result.Output.Embeddings))
		for _, emb := range result.Output.Embeddings {
			ordered[emb.TextIndex] = emb.Embedding
		}

		allEmbeddings = append(allEmbeddings, ordered...)
	}

	return allEmbeddings, nil
}

// SplitDocument 将文档内容切片
func (s *EmbeddingService) SplitDocument(content string) []string {
	cfg := config.LoadConfig()
	chunkSize := cfg.AI.ChunkSize
	overlap := cfg.AI.ChunkOverlap

	if chunkSize <= 0 {
		chunkSize = 500
	}
	if overlap < 0 {
		overlap = 50
	}

	// 按段落优先切片
	paragraphs := strings.Split(content, "\n\n")
	var chunks []string
	var currentChunk strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果当前块 + 新段落不超过大小，则合并
		if currentChunk.Len()+len(para) <= chunkSize {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
		} else {
			// 保存当前块
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				// 重叠部分
				lastPart := getOverlapText(currentChunk.String(), overlap)
				currentChunk.Reset()
				currentChunk.WriteString(lastPart)
				if len(lastPart) > 0 {
					currentChunk.WriteString("\n\n")
				}
			}
			// 如果单段落超过 chunkSize，强制切割
			if len(para) > chunkSize {
				for len(para) > chunkSize {
					chunks = append(chunks, para[:chunkSize])
					para = para[chunkSize-overlap:]
				}
				if len(para) > 0 {
					currentChunk.WriteString(para)
				}
			} else {
				currentChunk.WriteString(para)
			}
		}
	}

	// 最后一块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	// 过滤空切片
	var result []string
	for _, c := range chunks {
		c = strings.TrimSpace(c)
		if c != "" {
			result = append(result, c)
		}
	}

	log.Printf("文档切片完成: 原文 %d 字符, 切为 %d 块", len(content), len(result))
	return result
}

// getOverlapText 获取文本末尾的重叠部分
func getOverlapText(text string, overlap int) string {
	runes := []rune(text)
	if len(runes) <= overlap {
		return text
	}
	return string(runes[len(runes)-overlap:])
}
