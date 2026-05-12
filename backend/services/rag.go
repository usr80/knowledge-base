package services

import (
	"fmt"
	"log"
	"math"
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

	// 删除旧的切片（MySQL 中的记录仍保留，用于备份；Meilisearch 中的索引由 IndexChunks 自动清理）
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

	// 保存切片到 MySQL（保留作为备份，不再存储 embedding 字段）
	for i, chunk := range chunks {
		docChunk := models.DocumentChunk{
			DocumentID: documentID,
			UserID:     userID,
			ChunkIndex: i,
			Content:    chunk,
		}
		// 仍然保存 embedding 到 MySQL 作为备份
		if err := docChunk.SetEmbedding(embeddings[i]); err != nil {
			log.Printf("保存切片 %d 向量失败: %v", i, err)
		}
		if err := s.db.Create(&docChunk).Error; err != nil {
			log.Printf("保存切片 %d 失败: %v", i, err)
		}
	}

	// 索引到 Meilisearch（向量检索用）
	if err := s.searchSvc.IndexChunks(documentID, userID, chunks, embeddings); err != nil {
		log.Printf("警告: Meilisearch 切片索引失败: %v（MySQL 备份已保存）", err)
	}

	log.Printf("文档 %d 索引创建完成: %d 个切片", documentID, len(chunks))
	return nil
}

// SearchSimilarChunks 检索相似文档切片（优先使用 Meilisearch 向量搜索）
func (s *RAGService) SearchSimilarChunks(query string, userID uint, docIDs []uint, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = config.LoadConfig().AI.TopK
	}

	// 获取查询向量
	queryVec, err := s.embeddingSvc.GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("获取查询向量失败: %w", err)
	}

	// 优先使用 Meilisearch 向量搜索
	chunkResults, err := s.searchSvc.VectorSearch(queryVec, userID, docIDs, topK)
	if err == nil && len(chunkResults) > 0 {
		// 转换为 SearchResult
		results := make([]SearchResult, 0, len(chunkResults))
		for _, cr := range chunkResults {
			// 过滤低分结果
			cfg := config.LoadConfig()
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
		return results, nil
	}

	// Meilisearch 搜索失败，降级到 MySQL 余弦相似度
	if err != nil {
		log.Printf("Meilisearch 向量搜索失败，降级到 MySQL: %v", err)
	}
	return s.searchSimilarChunksMySQL(queryVec, userID, docIDs, topK)
}

// searchSimilarChunksMySQL MySQL 降级：手动计算余弦相似度
func (s *RAGService) searchSimilarChunksMySQL(queryVec []float64, userID uint, docIDs []uint, topK int) ([]SearchResult, error) {
	// 获取用户的所有文档切片
	dbQuery := s.db.Model(&models.DocumentChunk{}).Where("user_id = ?", userID)
	if len(docIDs) > 0 {
		dbQuery = dbQuery.Where("document_id IN ?", docIDs)
	}

	var chunks []models.DocumentChunk
	if err := dbQuery.Find(&chunks).Error; err != nil {
		return nil, fmt.Errorf("查询切片失败: %w", err)
	}

	if len(chunks) == 0 {
		return nil, nil
	}

	// 计算相似度并排序
	results := make([]SearchResult, 0)
	for _, chunk := range chunks {
		chunkVec, err := chunk.GetEmbedding()
		if err != nil || len(chunkVec) == 0 {
			continue
		}

		score := cosineSimilarity(queryVec, chunkVec)
		if score < config.LoadConfig().AI.MinScore {
			continue
		}

		results = append(results, SearchResult{
			ChunkID:    chunk.ID,
			DocumentID: chunk.DocumentID,
			Content:    chunk.Content,
			Score:      score,
		})
	}

	// 按相似度降序排序，取 topK
	sortResults(results)
	if len(results) > topK {
		results = results[:topK]
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

// AskStream 流式回答问题（记录 token 使用量）
func (s *RAGService) AskStream(userID uint, question string, docIDs []uint, sessionID string, callback func(string)) error {
	// 检索相关文档
	results, err := s.SearchSimilarChunks(question, userID, docIDs, 0)
	if err != nil {
		return fmt.Errorf("检索失败: %w", err)
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
		return fmt.Errorf("生成回答失败: %w", err)
	}

	// 记录 token 使用量
	if resp.InputTokens > 0 || resp.OutputTokens > 0 {
		provider := s.llmSvc.CurrentProvider()
		model := s.llmSvc.CurrentModel()
		cost := CalculateCost(provider, model, resp.InputTokens, resp.OutputTokens)
		go s.usageSvc.LogUsage(userID, provider, model, "chat_stream", resp.InputTokens, resp.OutputTokens, cost, sessionID, 0)
	}

	return nil
}

// SearchResult 搜索结果
type SearchResult struct {
	ChunkID    uint
	DocumentID uint
	Content    string
	Score      float64
}

// cosineSimilarity 余弦相似度（MySQL 降级时使用）
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// sortResults 按相似度降序排序
func sortResults(results []SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
