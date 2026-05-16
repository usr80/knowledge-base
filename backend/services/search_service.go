package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"knowledge-base/config"
	"knowledge-base/models"

	"github.com/meilisearch/meilisearch-go"
)

// SearchService Meilisearch 搜索服务
type SearchService struct {
	client       meilisearch.ServiceManager
	docIndex     string // 文档全文索引名
	chunkIndex   string // 切片向量索引名
}

// ==================== 文档全文索引结构 ====================

// DocumentIndex 文档索引结构
type DocumentIndex struct {
	ID           uint     `json:"id"`
	UserID       uint     `json:"user_id"`
	CategoryID   *uint    `json:"category_id,omitempty"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Summary      string   `json:"summary,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CategoryName string   `json:"category_name,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// DocSearchResult 搜索结果
type DocSearchResult struct {
	ID           uint     `json:"id"`
	UserID       uint     `json:"user_id"`
	CategoryID   *uint    `json:"category_id,omitempty"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Summary      string   `json:"summary,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CategoryName string   `json:"category_name,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// DocSearchResponse 搜索响应
type DocSearchResponse struct {
	Documents []DocSearchResult `json:"list"`
	Total     int64             `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"pageSize"`
}

// ==================== 切片向量索引结构 ====================

// ChunkIndex 文档切片索引结构（用于 RAG 向量检索）
type ChunkIndex struct {
	ID         uint            `json:"id"`
	UserID     uint            `json:"user_id"`
	DocumentID uint            `json:"document_id"`
	ChunkIndex int             `json:"chunk_index"`
	Content    string          `json:"content"`
	Vectors    *VectorEmbedder `json:"_vectors,omitempty"`
}

// VectorEmbedder 向量嵌入器结构
type VectorEmbedder struct {
	Manual *ManualVector `json:"manual,omitempty"`
}

// ManualVector 手动向量（userProvided 模式）
type ManualVector struct {
	Embeddings [][]float32 `json:"embeddings"`
	Regenerate bool        `json:"regenerate"`
}

// ChunkSearchResult 切片搜索结果
type ChunkSearchResult struct {
	ID         uint    `json:"id"`
	UserID     uint    `json:"user_id"`
	DocumentID uint    `json:"document_id"`
	ChunkIndex int     `json:"chunk_index"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

var searchService *SearchService

// GetSearchService 获取搜索服务单例
func GetSearchService() *SearchService {
	if searchService == nil {
		cfg := config.LoadConfig()

		client := meilisearch.New(cfg.Search.Host, meilisearch.WithAPIKey(cfg.Search.APIKey))

		searchService = &SearchService{
			client:     client,
			docIndex:   cfg.Search.Index,
			chunkIndex: cfg.Search.Index + "_chunks",
		}

		// 初始化索引
		searchService.initDocIndex()
		searchService.initChunkIndex()
	}
	return searchService
}

// ==================== 文档全文索引 ====================

// initDocIndex 初始化文档全文索引配置
func (s *SearchService) initDocIndex() {
	index := s.client.Index(s.docIndex)

	// 创建索引（如果不存在）
	filterableAttrs := &[]interface{}{"user_id", "category_id", "tags"}
	_, err := index.UpdateFilterableAttributes(filterableAttrs)
	if err != nil {
		log.Printf("警告: 设置过滤属性失败: %v", err)
	}

	// 设置可排序属性
	sortableAttrs := &[]string{"created_at", "updated_at", "title"}
	_, err = index.UpdateSortableAttributes(sortableAttrs)
	if err != nil {
		log.Printf("警告: 设置排序属性失败: %v", err)
	}

	// 设置搜索字段（权重：标题 > 内容 > 摘要）
	searchableAttrs := &[]string{"title", "content", "summary", "tags"}
	_, err = index.UpdateSearchableAttributes(searchableAttrs)
	if err != nil {
		log.Printf("警告: 设置搜索字段失败: %v", err)
	}

	// 设置排序规则（默认按更新时间倒序）
	rankingRules := &[]string{
		"words",
		"typo",
		"proximity",
		"attribute",
		"sort",
		"exactness",
	}
	_, err = index.UpdateRankingRules(rankingRules)
	if err != nil {
		log.Printf("警告: 设置排序规则失败: %v", err)
	}

	log.Printf("Meilisearch 文档索引初始化完成: %s", s.docIndex)
}

// initChunkIndex 初始化切片向量索引配置
func (s *SearchService) initChunkIndex() {
	index := s.client.Index(s.chunkIndex)

	// 配置 embedder（userProvided 模式）
	cfg := config.LoadConfig()
	embeddingDims := cfg.AI.EmbeddingDimensions

	embedders := map[string]meilisearch.Embedder{
		"manual": {
			Source:     meilisearch.UserProvidedEmbedderSource,
			Dimensions: embeddingDims,
		},
	}
	task, err := index.UpdateEmbedders(embedders)
	if err != nil {
		log.Printf("警告: 设置 embedder 失败: %v", err)
	} else {
		log.Printf("设置 embedder 任务已提交: TaskUID=%d", task.TaskUID)
	}

	// 设置过滤属性
	filterableAttrs := &[]interface{}{"user_id", "document_id"}
	_, err = index.UpdateFilterableAttributes(filterableAttrs)
	if err != nil {
		log.Printf("警告: 设置切片过滤属性失败: %v", err)
	}

	log.Printf("Meilisearch 切片索引初始化完成: %s (dimensions=%d)", s.chunkIndex, embeddingDims)
}

// IndexDocument 索引单个文档
func (s *SearchService) IndexDocument(doc *models.Document) error {
	index := s.client.Index(s.docIndex)

	docIndex := s.toDocumentIndex(doc)

	primaryKey := "id"
	_, err := index.AddDocumentsWithContext(context.Background(), []DocumentIndex{docIndex}, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return fmt.Errorf("索引文档失败: %w", err)
	}

	return nil
}

// DeleteDocument 删除文档索引
func (s *SearchService) DeleteDocument(id uint) error {
	index := s.client.Index(s.docIndex)

	_, err := index.DeleteDocument(strconv.FormatUint(uint64(id), 10), nil)
	if err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}

	return nil
}

// Search 搜索文档
func (s *SearchService) Search(userID uint, keyword string, page, pageSize int, categoryID *uint) (*DocSearchResponse, error) {
	index := s.client.Index(s.docIndex)

	// 构建过滤条件
	filters := []string{fmt.Sprintf("user_id = %d", userID)}
	if categoryID != nil && *categoryID > 0 {
		filters = append(filters, fmt.Sprintf("category_id = %d", *categoryID))
	}

	// 构建过滤表达式
	filterExpr := ""
	if len(filters) == 1 {
		filterExpr = filters[0]
	} else {
		filterExpr = fmt.Sprintf("%s AND %s", filters[0], filters[1])
	}

	// 设置搜索参数
	searchParams := &meilisearch.SearchRequest{
		AttributesToHighlight: []string{"title", "content", "summary"},
		AttributesToCrop:     []string{"content"},
		CropLength:           200,
		Filter:               &filterExpr,
		Sort:                 []string{"updated_at:desc"},
		Limit:                int64(pageSize),
		Offset:               int64((page - 1) * pageSize),
	}

	// 执行搜索
	resp, err := index.SearchWithContext(context.Background(), keyword, searchParams)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	// 转换结果
	results := make([]DocSearchResult, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var result DocSearchResult
		if err := hit.DecodeInto(&result); err != nil {
			log.Printf("警告: 解析搜索结果失败: %v", err)
			result = s.parseSearchResult(hit)
		}
		results = append(results, result)
	}

	// 获取总数
	total := resp.EstimatedTotalHits
	if resp.TotalHits > 0 {
		total = resp.TotalHits
	}

	return &DocSearchResponse{
		Documents: results,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// RebuildIndex 重建文档全文索引（全量同步）
func (s *SearchService) RebuildIndex() error {
	index := s.client.Index(s.docIndex)

	// 获取所有文档
	var docs []models.Document
	if err := config.DB.Preload("Category").Preload("Tags").Find(&docs).Error; err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}

	// 转换为索引结构
	docIndexes := make([]DocumentIndex, 0, len(docs))
	for _, doc := range docs {
		docIndexes = append(docIndexes, s.toDocumentIndex(&doc))
	}

	// 批量索引
	primaryKey := "id"
	task, err := index.AddDocumentsWithContext(context.Background(), docIndexes, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return fmt.Errorf("重建索引失败: %w", err)
	}

	log.Printf("重建索引任务已提交，TaskUID: %d，文档数: %d", task.TaskUID, len(docIndexes))

	return nil
}

// ==================== 切片向量索引 ====================

// IndexChunks 索引文档的切片（含向量）
func (s *SearchService) IndexChunks(documentID, userID uint, chunks []string, embeddings [][]float64) error {
	index := s.client.Index(s.chunkIndex)

	if len(chunks) != len(embeddings) {
		return fmt.Errorf("切片数(%d)与向量数(%d)不匹配", len(chunks), len(embeddings))
	}

	// 先删除该文档的旧切片索引
	s.DeleteChunksByDocument(documentID)

	// 构建切片索引
	chunkIndexes := make([]ChunkIndex, 0, len(chunks))
	for i, chunk := range chunks {
		// float64 → float32
		vec32 := make([]float32, len(embeddings[i]))
		for j, v := range embeddings[i] {
			vec32[j] = float32(v)
		}

		chunkIdx := ChunkIndex{
			ID:         documentID*1000 + uint(i), // 用 documentID*1000+i 作为唯一 ID
			UserID:     userID,
			DocumentID: documentID,
			ChunkIndex: i,
			Content:    chunk,
			Vectors: &VectorEmbedder{
				Manual: &ManualVector{
					Embeddings: [][]float32{vec32},
					Regenerate: false,
				},
			},
		}
		chunkIndexes = append(chunkIndexes, chunkIdx)
	}

	// 批量索引
	primaryKey := "id"
	task, err := index.AddDocumentsWithContext(context.Background(), chunkIndexes, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	if err != nil {
		return fmt.Errorf("索引切片失败: %w", err)
	}

	log.Printf("文档 %d 切片索引已提交: TaskUID=%d, 切片数=%d", documentID, task.TaskUID, len(chunkIndexes))
	return nil
}

// DeleteChunksByDocument 删除文档的所有切片索引
func (s *SearchService) DeleteChunksByDocument(documentID uint) error {
	index := s.client.Index(s.chunkIndex)

	// 通过过滤删除
	filter := fmt.Sprintf("document_id = %d", documentID)
	_, err := index.DeleteDocumentsByFilter(filter, nil)
	if err != nil {
		log.Printf("警告: 删除文档 %d 切片索引失败: %v", documentID, err)
		return err
	}

	log.Printf("文档 %d 切片索引已删除", documentID)
	return nil
}

// VectorSearch 向量搜索（纯向量检索，用于 RAG）
func (s *SearchService) VectorSearch(queryVec []float64, userID uint, docIDs []uint, topK int) ([]ChunkSearchResult, error) {
	if topK <= 0 {
		topK = config.LoadConfig().AI.TopK
	}

	index := s.client.Index(s.chunkIndex)

	// float64 → float32
	vec32 := make([]float32, len(queryVec))
	for i, v := range queryVec {
		vec32[i] = float32(v)
	}

	// 构建过滤条件
	filters := []string{fmt.Sprintf("user_id = %d", userID)}
	if len(docIDs) > 0 {
		docFilter := "document_id IN ["
		for i, id := range docIDs {
			if i > 0 {
				docFilter += ", "
			}
			docFilter += strconv.FormatUint(uint64(id), 10)
		}
		docFilter += "]"
		filters = append(filters, docFilter)
	}

	filterExpr := ""
	if len(filters) == 1 {
		filterExpr = filters[0]
	} else {
		filterExpr = fmt.Sprintf("%s AND %s", filters[0], filters[1])
	}

	// 向量搜索参数
	searchParams := &meilisearch.SearchRequest{
		Vector:  vec32,
		Filter:  &filterExpr,
		Limit:   int64(topK),
		Hybrid: &meilisearch.SearchRequestHybrid{
			SemanticRatio: 1.0, // 纯向量搜索
			Embedder:      "manual",
		},
		RetrieveVectors:  false,
		ShowRankingScore: true, // 必须！否则 _rankingScore 不会返回
	}

	// 执行搜索
	resp, err := index.SearchWithContext(context.Background(), "", searchParams)
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %w", err)
	}

	// 解析结果
	results := make([]ChunkSearchResult, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var chunk ChunkSearchResult
		if err := hit.DecodeInto(&chunk); err != nil {
			log.Printf("警告: 解析向量搜索结果失败: %v", err)
			continue
		}

		// Meilisearch 向量搜索的 _rankingScore 字段
		if raw, ok := hit["_rankingScore"]; ok {
			var score float64
			if err := json.Unmarshal(raw, &score); err == nil {
				chunk.Score = score
			}
		}

		results = append(results, chunk)
	}

	return results, nil
}

// HybridSearch 混合搜索（关键词 + 向量，语义比例可调）
func (s *SearchService) HybridSearch(query string, queryVec []float64, userID uint, docIDs []uint, topK int, semanticRatio float64) ([]ChunkSearchResult, error) {
	if topK <= 0 {
		topK = config.LoadConfig().AI.TopK
	}
	if semanticRatio <= 0 {
		semanticRatio = 0.5
	}
	if semanticRatio > 1.0 {
		semanticRatio = 1.0
	}

	index := s.client.Index(s.chunkIndex)

	// float64 → float32
	vec32 := make([]float32, len(queryVec))
	for i, v := range queryVec {
		vec32[i] = float32(v)
	}

	// 构建过滤条件
	filters := []string{fmt.Sprintf("user_id = %d", userID)}
	if len(docIDs) > 0 {
		docFilter := "document_id IN ["
		for i, id := range docIDs {
			if i > 0 {
				docFilter += ", "
			}
			docFilter += strconv.FormatUint(uint64(id), 10)
		}
		docFilter += "]"
		filters = append(filters, docFilter)
	}

	filterExpr := ""
	if len(filters) == 1 {
		filterExpr = filters[0]
	} else {
		filterExpr = fmt.Sprintf("%s AND %s", filters[0], filters[1])
	}

	// 混合搜索参数
	searchParams := &meilisearch.SearchRequest{
		Query:  query,
		Vector: vec32,
		Filter: &filterExpr,
		Limit:  int64(topK),
		Hybrid: &meilisearch.SearchRequestHybrid{
			SemanticRatio: semanticRatio,
			Embedder:      "manual",
		},
		RetrieveVectors:  false,
		ShowRankingScore: true, // 必须！否则 _rankingScore 不会返回
	}

	resp, err := index.SearchWithContext(context.Background(), query, searchParams)
	if err != nil {
		return nil, fmt.Errorf("混合搜索失败: %w", err)
	}

	results := make([]ChunkSearchResult, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var chunk ChunkSearchResult
		if err := hit.DecodeInto(&chunk); err != nil {
			log.Printf("警告: 解析混合搜索结果失败: %v", err)
			continue
		}

		if raw, ok := hit["_rankingScore"]; ok {
			var score float64
			if err := json.Unmarshal(raw, &score); err == nil {
				chunk.Score = score
			}
		}

		results = append(results, chunk)
	}

	return results, nil
}

// RebuildChunkIndex 重建切片向量索引（全量同步）
func (s *SearchService) RebuildChunkIndex() error {
	embeddingSvc := NewEmbeddingService()

	// 获取所有文档
	var docs []models.Document
	if err := config.DB.Find(&docs).Error; err != nil {
		return fmt.Errorf("获取文档失败: %w", err)
	}

	totalChunks := 0
	for _, doc := range docs {
		// 切片
		chunks := embeddingSvc.SplitDocument(doc.Content)
		if len(chunks) == 0 {
			continue
		}

		// 获取向量
		embeddings, err := embeddingSvc.GetEmbeddings(chunks)
		if err != nil {
			log.Printf("文档 %d 获取向量失败: %v", doc.ID, err)
			continue
		}

		// 索引到 Meilisearch
		if err := s.IndexChunks(doc.ID, doc.UserID, chunks, embeddings); err != nil {
			log.Printf("文档 %d 索引切片失败: %v", doc.ID, err)
			continue
		}

		totalChunks += len(chunks)
	}

	log.Printf("切片索引重建完成: %d 个文档, %d 个切片", len(docs), totalChunks)
	return nil
}

// ==================== 辅助方法 ====================

// toDocumentIndex 将文档转换为索引结构
func (s *SearchService) toDocumentIndex(doc *models.Document) DocumentIndex {
	idx := DocumentIndex{
		ID:         doc.ID,
		UserID:     doc.UserID,
		CategoryID: doc.CategoryID,
		Title:      doc.Title,
		Content:    doc.Content,
		Summary:    doc.Summary,
		CreatedAt:  doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  doc.UpdatedAt.Format(time.RFC3339),
	}

	// 提取标签名称
	if len(doc.Tags) > 0 {
		tags := make([]string, 0, len(doc.Tags))
		for _, tag := range doc.Tags {
			tags = append(tags, tag.Name)
		}
		idx.Tags = tags
	}

	// 提取分类名称
	if doc.Category != nil {
		idx.CategoryName = doc.Category.Name
	}

	return idx
}

// parseSearchResult 降级解析搜索结果（手动从 json.RawMessage 提取）
func (s *SearchService) parseSearchResult(hit meilisearch.Hit) DocSearchResult {
	result := DocSearchResult{}

	if raw, ok := hit["id"]; ok {
		var id float64
		if err := json.Unmarshal(raw, &id); err == nil {
			result.ID = uint(id)
		}
	}
	if raw, ok := hit["user_id"]; ok {
		var uid float64
		if err := json.Unmarshal(raw, &uid); err == nil {
			result.UserID = uint(uid)
		}
	}
	if raw, ok := hit["category_id"]; ok {
		var cid float64
		if err := json.Unmarshal(raw, &cid); err == nil {
			v := uint(cid)
			result.CategoryID = &v
		}
	}
	if raw, ok := hit["title"]; ok {
		_ = json.Unmarshal(raw, &result.Title)
	}
	if raw, ok := hit["content"]; ok {
		_ = json.Unmarshal(raw, &result.Content)
	}
	if raw, ok := hit["summary"]; ok {
		_ = json.Unmarshal(raw, &result.Summary)
	}
	if raw, ok := hit["category_name"]; ok {
		_ = json.Unmarshal(raw, &result.CategoryName)
	}
	if raw, ok := hit["created_at"]; ok {
		_ = json.Unmarshal(raw, &result.CreatedAt)
	}
	if raw, ok := hit["updated_at"]; ok {
		_ = json.Unmarshal(raw, &result.UpdatedAt)
	}
	if raw, ok := hit["tags"]; ok {
		_ = json.Unmarshal(raw, &result.Tags)
	}

	return result
}

// Close 关闭连接
func (s *SearchService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
