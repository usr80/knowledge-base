package services

import (
	"fmt"
	"log"
	"strings"

	"knowledge-base/config"
	"knowledge-base/models"

	"gorm.io/gorm"
)

// RAGService RAG 检索增强生成服务
type RAGService struct {
	db           *gorm.DB
	embeddingSvc *EmbeddingService
	llmSvc       *LLMService
	usageSvc     *UsageService
	searchSvc    *SearchService
}

func NewRAGService() *RAGService {
	return &RAGService{
		db:           config.DB,
		embeddingSvc: NewEmbeddingService(),
		llmSvc:       NewLLMService(),
		usageSvc:     NewUsageService(),
		searchSvc:    GetSearchService(),
	}
}

// IndexDocument 为文档创建向量索引
func (s *RAGService) IndexDocument(documentID, userID uint) error {
	log.Printf("[IndexDocument] docID=%d, userID=%d", documentID, userID)

	// 获取文档
	var doc models.Document
	if err := s.db.Where("id = ? AND user_id = ?", documentID, userID).First(&doc).Error; err != nil {
		return fmt.Errorf("文档不存在: %w", err)
	}

	// 删除旧的切片
	if err := s.db.Where("document_id = ?", documentID).Delete(&models.DocumentChunk{}).Error; err != nil {
		return fmt.Errorf("删除旧切片失败: %w", err)
	}

	// 切片文档
	chunks := s.embeddingSvc.SplitDocument(doc.Content)
	if len(chunks) == 0 {
		return fmt.Errorf("文档内容为空，无法创建索引")
	}

	// 获取向量嵌入
	embeddings, err := s.embeddingSvc.GetEmbeddings(chunks)
	if err != nil {
		return fmt.Errorf("获取嵌入向量失败: %w", err)
	}

	// 保存切片到 MySQL（仅内容，向量数据存 Meilisearch）
	for i, chunk := range chunks {
		docChunk := models.DocumentChunk{
			DocumentID: documentID,
			UserID:     userID,
			ChunkIndex: i,
			Content:    chunk,
		}
		if err := s.db.Create(&docChunk).Error; err != nil {
			log.Printf("保存切片 %d 失败: %v", i, err)
		}
	}

	// 索引到 Meilisearch（向量检索）
	if err := s.searchSvc.IndexChunks(documentID, userID, chunks, embeddings); err != nil {
		return fmt.Errorf("Meilisearch 索引失败: %w", err)
	}

	log.Printf("文档 %d 索引创建完成: %d 个切片", documentID, len(chunks))
	return nil
}

// SearchSimilarChunks 检索相似文档切片（Meilisearch 向量搜索）
func (s *RAGService) SearchSimilarChunks(query string, userID uint, docIDs []uint, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = config.LoadConfig().AI.TopK
	}

	// 获取查询向量
	queryVec, err := s.embeddingSvc.GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("获取查询向量失败: %w", err)
	}

	// Meilisearch 向量搜索
	chunkResults, err := s.searchSvc.VectorSearch(queryVec, userID, docIDs, topK)
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %w", err)
	}

	// 转换为 SearchResult 并过滤低分
	cfg := config.LoadConfig()
	results := make([]SearchResult, 0, len(chunkResults))
	for _, cr := range chunkResults {
		if cr.Score < cfg.AI.MinScore {
			continue
		}
		results = append(results, SearchResult{
			ChunkID:    cr.ID,
			DocumentID: cr.DocumentID,
			Content:    cr.Content,
			Score:      cr.Score,
		})
	}

	// 打印命中文档信息
	if len(results) > 0 {
		docIDsSet := make(map[uint]bool)
		for _, r := range results {
			docIDsSet[r.DocumentID] = true
		}
		docIDList := make([]uint, 0, len(docIDsSet))
		for id := range docIDsSet {
			docIDList = append(docIDList, id)
		}

		var docs []models.Document
		s.db.Where("id IN ?", docIDList).Find(&docs)
		docMap := make(map[uint]models.Document)
		for _, d := range docs {
			docMap[d.ID] = d
		}

		log.Printf("[RAG] 命中 %d 个切片，来自 %d 个文档:", len(results), len(docIDsSet))
		for _, r := range results {
			if doc, ok := docMap[r.DocumentID]; ok {
				log.Printf("[RAG]   - 文档: %s (ID=%d), 切片#%d, 相似度=%.4f", doc.Title, r.DocumentID, r.ChunkID%1000, r.Score)
			}
		}
	} else {
		log.Printf("[RAG] 未命中任何文档（query=%q, minScore=%.2f）", query, cfg.AI.MinScore)
	}

	return results, nil
}

// Ask 基于文档回答问题（支持记录 token 使用量）
func (s *RAGService) Ask(userID uint, question string, docIDs []uint, sessionID string) (string, error) {
	// 检索相关文档
	results, err := s.SearchSimilarChunks(question, userID, docIDs, 0)
	if err != nil {
		return "", fmt.Errorf("检索失败: %w", err)
	}

	// 构建系统提示词
	systemPrompt := `你是一个智能助手，可以基于知识库内容和自身知识回答用户问题。

要求：
1. 优先使用知识库中的信息回答
2. 如果知识库中没有相关信息，可以基于自身知识给出回答，但要说明哪些内容来自知识库、哪些是通用知识
3. 回答要简洁准确，必要时引用文档编号
4. 可以综合多个文档片段的信息`

	var messages []chatMessage
	if len(results) > 0 {
		// 有检索结果：构建带上下文的提示
		var contextBuilder strings.Builder
		contextBuilder.WriteString("以下是知识库中的相关内容：\n\n")
		for i, r := range results {
			contextBuilder.WriteString(fmt.Sprintf("【文档片段 %d】\n%s\n\n", i+1, r.Content))
		}
		contextStr := contextBuilder.String()
		messages = []chatMessage{
			{
				Role:    "user",
				Content: contextStr + "\n用户问题：" + question,
			},
		}
	} else {
		// 无检索结果：通用对话模式
		messages = []chatMessage{
			{
				Role:    "user",
				Content: question,
			},
		}
	}

	// 调用 LLM 生成回答（带 token 使用量）
	resp, err := s.llmSvc.ChatWithUsage(systemPrompt, messages)
	if err != nil {
		return "", fmt.Errorf("生成回答失败: %w", err)
	}

	// 记录 token 使用量
	provider := s.llmSvc.CurrentProvider()
	model := s.llmSvc.CurrentModel()
	cost := CalculateCost(provider, model, resp.InputTokens, resp.OutputTokens)
	go s.usageSvc.LogUsage(userID, provider, model, "chat", resp.InputTokens, resp.OutputTokens, cost, sessionID, 0)

	// 添加引用来源（仅有检索结果时）
	if len(results) > 0 {
		// 获取文档信息用于引用
		docIDsSet := make(map[uint]bool)
		for _, r := range results {
			docIDsSet[r.DocumentID] = true
		}
		docIDList := make([]uint, 0)
		for id := range docIDsSet {
			docIDList = append(docIDList, id)
		}

		var docs []models.Document
		s.db.Where("id IN ?", docIDList).Find(&docs)
		docMap := make(map[uint]models.Document)
		for _, d := range docs {
			docMap[d.ID] = d
		}

		refBuilder := strings.Builder{}
		refBuilder.WriteString("\n\n---\n**参考来源：**\n")
		seen := make(map[uint]bool)
		for _, r := range results {
			if !seen[r.DocumentID] {
				if doc, ok := docMap[r.DocumentID]; ok {
					refBuilder.WriteString(fmt.Sprintf("- %s\n", doc.Title))
				}
				seen[r.DocumentID] = true
			}
		}
		return resp.Content + refBuilder.String(), nil
	}

	return resp.Content, nil
}

// AskStream 流式回答问题（返回搜索结果供调用方构造引用）
func (s *RAGService) AskStream(userID uint, question string, docIDs []uint, sessionID string, callback func(string)) ([]SearchResult, error) {
	// 检索相关文档
	results, err := s.SearchSimilarChunks(question, userID, docIDs, 0)
	if err != nil {
		return nil, fmt.Errorf("检索失败: %w", err)
	}

	systemPrompt := `你是一个智能助手，可以基于知识库内容和自身知识回答用户问题。

要求：
1. 优先使用知识库中的信息回答
2. 如果知识库中没有相关信息，可以基于自身知识给出回答，但要说明哪些内容来自知识库、哪些是通用知识
3. 回答要简洁准确，必要时引用文档编号
4. 可以综合多个文档片段的信息`

	var messages []chatMessage
	if len(results) > 0 {
		// 有检索结果：构建带上下文的提示
		var contextBuilder strings.Builder
		contextBuilder.WriteString("以下是知识库中的相关内容：\n\n")
		for i, r := range results {
			contextBuilder.WriteString(fmt.Sprintf("【文档片段 %d】\n%s\n\n", i+1, r.Content))
		}
		contextStr := contextBuilder.String()
		messages = []chatMessage{
			{
				Role:    "user",
				Content: contextStr + "\n用户问题：" + question,
			},
		}
	} else {
		// 无检索结果：通用对话模式
		messages = []chatMessage{
			{
				Role:    "user",
				Content: question,
			},
		}
	}

	// 调用 LLM 流式生成回答（返回 token 使用量）
	resp, err := s.llmSvc.ChatStream(systemPrompt, messages, callback)
	if err != nil {
		return nil, fmt.Errorf("生成回答失败: %w", err)
	}

	// 记录 token 使用量
	if resp.InputTokens > 0 || resp.OutputTokens > 0 {
		provider := s.llmSvc.CurrentProvider()
		model := s.llmSvc.CurrentModel()
		cost := CalculateCost(provider, model, resp.InputTokens, resp.OutputTokens)
		go s.usageSvc.LogUsage(userID, provider, model, "chat_stream", resp.InputTokens, resp.OutputTokens, cost, sessionID, 0)
	}

	// 返回搜索结果，供调用方构造引用
	return results, nil
}

// SearchResult 搜索结果
type SearchResult struct {
	ChunkID    uint
	DocumentID uint
	Content    string
	Score      float64
}

// Reference 引用来源
type Reference struct {
	DocumentID   uint    `json:"documentId"`
	DocumentName string  `json:"documentName"`
	Score        float64 `json:"score"`
}


