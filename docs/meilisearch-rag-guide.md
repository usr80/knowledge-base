# Meilisearch：轻量搜索引擎的优雅选择，以及它在 RAG 中的应用

> 从 MySQL LIKE 到全文搜索，再到向量检索——一个搜索引擎如何同时搞定关键词搜索和 AI 语义检索？

## 为什么需要 Meilisearch

大多数个人项目和小型团队的搜索方案，都经历过这样的阶段：

```sql
SELECT * FROM documents WHERE title LIKE '%关键词%' OR content LIKE '%关键词%';
```

MySQL LIKE 够用吗？数据量小的时候够用。但它有三个硬伤：

1. **没有中文分词**——搜「机器学习」找不到「机器 学习方法」
2. **全表扫描**——数据量一大就慢
3. **没有相关性排序**——结果顺序全靠运气

选 Elasticsearch？它当然是行业标准，但对个人项目来说太重了：

| 对比项 | Elasticsearch | Meilisearch |
|--------|--------------|-------------|
| 内存占用 | 2GB+ | ~100MB |
| 安装方式 | JDK + 集群配置 | 单二进制文件 |
| 中文分词 | 需装插件（IK） | 内置支持 |
| API 风格 | 复杂的 JSON DSL | 简洁的 RESTful |
| 上手成本 | 高 | 低 |
| 适用场景 | 日志分析、大型项目 | 个人/小型项目（<10万文档） |

**结论**：如果你的文档量在 10 万以内，Meilisearch 是更务实的选择。不是 Elasticsearch 不好，是杀鸡用牛刀。

## Meilisearch 核心特性

### 1. 开箱即用的中文分词

不需要安装任何插件，Meilisearch 内置中文分词能力。索引时自动拆词，搜索时自动匹配。

### 2. Typo Tolerance（容错搜索）

搜「meilisearch」能匹配「meilisearc」「meilserch」，基于 Damerau-Levenshtein 距离算法自动纠错。

### 3. 前缀搜索

输入「数据」即可匹配「数据库」「数据结构」「数据分析」，实时响应。

### 4. 过滤与排序

支持字段过滤（如按用户 ID、分类 ID 筛选）和多字段排序，满足业务查询需求。

### 5. 向量搜索（v1.3+）

这是本文的重点——Meilisearch 不仅做关键词搜索，还支持向量检索和混合搜索，让它在 RAG 场景中也能胜任。

## 快速上手

### 安装

```bash
# macOS
brew install meilisearch

# Linux
curl -L https://install.meilisearch.com | sh

# Windows（scoop）
scoop install meilisearch

# Docker
docker run -d -p 7700:7700 \
  -v $(pwd)/meili_data:/meili_data \
  getmeili/meilisearch:v1.3 \
  --master-key=your-master-key
```

### 启动

```bash
meilisearch --master-key=your-master-key
```

打开 `http://localhost:7700`，你会看到一个搜索预览界面——对，它自带 Web UI。

### 基本操作

```bash
# 创建索引 + 添加文档（一步到位）
curl -X POST 'http://localhost:7700/indexes/articles/documents' \
  -H 'Authorization: Bearer your-master-key' \
  -H 'Content-Type: application/json' \
  --data '[{
    "id": 1,
    "title": "Go 并发编程",
    "content": "goroutine 和 channel 是 Go 并发的核心...",
    "tags": ["go", "并发"]
  }]'

# 搜索
curl 'http://localhost:7700/indexes/articles/search?q=并发&filter=user_id=1'
```

就是这么简单。添加文档的那一刻，索引就已经建好了。

## Go 集成实战

在我们的知识库项目中，使用 `meilisearch-go` SDK 集成 Meilisearch。以下是核心代码。

### 安装 SDK

```bash
go get github.com/meilisearch/meilisearch-go
```

### 连接与初始化

```go
package services

import (
    "log"
    "github.com/meilisearch/meilisearch-go"
)

type SearchService struct {
    client     meilisearch.ServiceManager
    docIndex   string
    chunkIndex string
}

func GetSearchService() *SearchService {
    cfg := config.LoadConfig()

    client := meilisearch.New(
        cfg.Search.Host,
        meilisearch.WithAPIKey(cfg.Search.APIKey),
    )

    svc := &SearchService{
        client:     client,
        docIndex:   cfg.Search.Index,
        chunkIndex: cfg.Search.Index + "_chunks",
    }

    svc.initDocIndex()
    svc.initChunkIndex()
    return svc
}
```

### 文档索引结构

```go
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
```

### 索引配置

```go
func (s *SearchService) initDocIndex() {
    index := s.client.Index(s.docIndex)

    // 可过滤字段
    filterable := &[]interface{}{"user_id", "category_id", "tags"}
    index.UpdateFilterableAttributes(filterable)

    // 可排序字段
    sortable := &[]string{"created_at", "updated_at", "title"}
    index.UpdateSortableAttributes(sortable)

    // 搜索字段（权重从高到低：标题 > 内容 > 摘要 > 标签）
    searchable := &[]string{"title", "content", "summary", "tags"}
    index.UpdateSearchableAttributes(searchable)

    // 排序规则
    ranking := &[]string{
        "words", "typo", "proximity",
        "attribute", "sort", "updated_at:desc",
    }
    index.UpdateRankingRules(ranking)
}
```

### 搜索

```go
func (s *SearchService) Search(keyword string, userID uint, page, pageSize int) (*DocSearchResponse, error) {
    index := s.client.Index(s.docIndex)

    filter := fmt.Sprintf("user_id = %d", userID)

    result, err := index.Search(keyword, &meilisearch.SearchRequest{
        Filter: filter,
        HitsPerPage: pageSize,
        Page: page,
    })
    if err != nil {
        return nil, err
    }

    // 解析结果...
    return response, nil
}
```

### 文档同步

文档增删改时自动同步索引：

```go
// 创建文档后
searchSvc.IndexDocument(doc)

// 更新文档后
searchSvc.IndexDocument(doc)

// 删除文档后
searchSvc.DeleteDocument(docID)
```

### 前端智能切换

前端根据是否有搜索关键词，智能选择数据源：

```javascript
async loadDocuments() {
    if (this.searchKeyword) {
        // 有关键词 → Meilisearch 全文搜索（中文分词、相关性排序）
        const res = await searchDocuments({
            keyword: this.searchKeyword,
            page: this.currentPage,
            categoryID: this.selectedCategoryID,
        })
        this.documents = res.list
        this.total = res.total
    } else {
        // 无关键词 → MySQL 列表（完整浏览）
        const res = await getDocuments({
            page: this.currentPage,
            categoryID: this.selectedCategoryID,
        })
        this.documents = res.list
        this.total = res.total
    }
}
```

到这里，全文搜索部分就完成了。但 Meilisearch 的能力不止于此——下面进入正题。

---

## 向量搜索：Meilisearch 在 RAG 中的应用

### 什么是 RAG

RAG（Retrieval-Augmented Generation，检索增强生成）的核心思路：

```
用户提问 → 检索相关文档 → 将文档作为上下文 → LLM 生成回答
```

关键步骤是「检索相关文档」。传统方案用关键词匹配，但关键词匹配有局限——用户问「如何优化并发性能」，相关文档可能写的是「goroutine 调度器原理」，关键词完全不同，但语义相关。

向量检索解决的就是这个问题：**按语义相似度检索，而非关键词匹配**。

### 向量检索原理

1. 用 Embedding 模型将文本转为高维向量（如 1024 维）
2. 相似语义的文本在向量空间中距离更近
3. 查询时计算向量距离，返回最相似的文档

```
"并发优化" → [0.12, -0.34, 0.56, ...]  ─┐
                                           ├─ 余弦相似度高 → 匹配！
"goroutine 调度" → [0.11, -0.31, 0.54, ...] ─┘

"数据库索引" → [0.87, 0.22, -0.15, ...]  ── 余弦相似度低 → 不匹配
```

### 为什么选择 Meilisearch 做向量检索

之前我们的 RAG 架构：

```
文档 → 切片 → Embedding → 存 MySQL（JSON blob）
                              ↓
              查询时 Go 代码手动算余弦相似度
```

问题很明显：
- MySQL 没有 ANN（近似最近邻）索引，全表计算相似度
- 搜索走 Meilisearch，RAG 走 MySQL，两套索引重复存储
- 同步逻辑复杂，维护成本高

统一到 Meilisearch 后：

```
文档 → 切片 → Embedding → 存 Meilisearch（_vectors 字段）
                              ↓
              ├─ 关键词搜索：全文检索
              └─ RAG 检索：向量检索 / 混合检索
```

**一个索引，两种检索方式，架构大幅简化。**

### userProvided 模式

Meilisearch 支持多种 Embedder 模式：OpenAI、HuggingFace、Ollama 等。但我们选择了 `userProvided`——手动提供向量。

为什么？因为国内常用通义千问、智谱等 Embedding API，Meilisearch 内置不支持。userProvided 模式让我们自由选择 Embedding 模型，只需把计算好的向量交给 Meilisearch 存储和检索。

### 代码实现

#### 切片索引结构

```go
// ChunkIndex 文档切片索引结构（用于 RAG 向量检索）
type ChunkIndex struct {
    ID         uint            `json:"id"`
    UserID     uint            `json:"user_id"`
    DocumentID uint            `json:"document_id"`
    ChunkIndex int             `json:"chunk_index"`
    Content    string          `json:"content"`
    Vectors    *VectorEmbedder `json:"_vectors,omitempty"`  // Meilisearch 向量字段
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
```

`_vectors` 字段是 Meilisearch 约定的特殊字段名。`manual` 是我们配置的 Embedder 名称。数据格式：

```json
{
    "id": 1,
    "user_id": 1,
    "document_id": 42,
    "chunk_index": 0,
    "content": "goroutine 和 channel 是 Go 并发的核心...",
    "_vectors": {
        "manual": {
            "embeddings": [[0.12, -0.34, 0.56, ...]],
            "regenerate": false
        }
    }
}
```

#### 配置 Embedder

```go
func (s *SearchService) initChunkIndex() {
    index := s.client.Index(s.chunkIndex)

    // 配置向量嵌入器
    embedders := map[string]meilisearch.Embedder{
        "manual": {
            Source:     meilisearch.UserProvidedEmbedderSource,
            Dimensions: 1024, // 通义千问 text-embedding-v3 输出维度
        },
    }
    index.UpdateEmbedders(embedders)

    // 设置可过滤字段
    filterable := &[]interface{}{"user_id", "document_id"}
    index.UpdateFilterableAttributes(filterable)
}
```

`Dimensions` 必须与你的 Embedding 模型输出维度一致。通义千问 `text-embedding-v3` 输出 1024 维，所以这里设 1024。

#### 索引切片

```go
func (s *SearchService) IndexChunks(documentID, userID uint, chunks []string, embeddings [][]float64) error {
    var chunkIndices []ChunkIndex

    for i, chunk := range chunks {
        // float64 → float32（Meilisearch 要求 float32）
        vec := make([]float32, len(embeddings[i]))
        for j, v := range embeddings[i] {
            vec[j] = float32(v)
        }

        chunkIndices = append(chunkIndices, ChunkIndex{
            ID:         uint(i + 1),
            UserID:     userID,
            DocumentID: documentID,
            ChunkIndex: i,
            Content:    chunk,
            Vectors: &VectorEmbedder{
                Manual: &ManualVector{
                    Embeddings: [][]float32{vec},
                    Regenerate: false,
                },
            },
        })
    }

    // 先清理该文档的旧索引
    s.DeleteChunksByDocument(documentID)

    // 批量写入
    _, err := s.client.Index(s.chunkIndex).AddDocumentsWithContext(
        context.Background(),
        &meilisearch.DocumentOptions{PrimaryKey: "id"},
        chunkIndices,
    )
    return err
}
```

#### 向量搜索

```go
func (s *SearchService) VectorSearch(queryVec []float64, userID uint, docIDs []uint, limit int) ([]ChunkSearchResult, error) {
    // float64 → float32
    vec := make([]float32, len(queryVec))
    for i, v := range queryVec {
        vec[i] = float32(v)
    }

    // 构建过滤条件
    filter := fmt.Sprintf("user_id = %d", userID)

    result, err := s.client.Index(s.chunkIndex).Search("", &meilisearch.SearchRequest{
        Vector: vec,
        Filter: filter,
        Limit:  int64(limit),
        Hybrid: &meilisearch.SearchRequestHybrid{
            SemanticRatio: 1.0,  // 纯向量搜索
            Embedder:      "manual",
        },
    })
    if err != nil {
        return nil, err
    }

    // 解析结果...
    return results, nil
}
```

`SemanticRatio` 控制搜索模式：
- `1.0`：纯向量搜索（语义匹配）
- `0.0`：纯关键词搜索
- `0.5`：混合搜索（各占一半权重）

#### 混合搜索

```go
func (s *SearchService) HybridSearch(queryVec []float64, keyword string, userID uint, limit int, ratio float64) ([]ChunkSearchResult, error) {
    vec := make([]float32, len(queryVec))
    for i, v := range queryVec {
        vec[i] = float32(v)
    }

    result, err := s.client.Index(s.chunkIndex).Search(keyword, &meilisearch.SearchRequest{
        Vector: vec,
        Filter: fmt.Sprintf("user_id = %d", userID),
        Limit:  int64(limit),
        Hybrid: &meilisearch.SearchRequestHybrid{
            SemanticRatio: ratio,  // 0.5 = 关键词和向量各半
            Embedder:      "manual",
        },
    })
    if err != nil {
        return nil, err
    }

    // 解析结果...
    return results, nil
}
```

混合搜索是 Meilisearch 的杀手级特性——一次查询同时考虑关键词匹配和语义相似度，取两者最优结果。

### RAG 流程整合

```go
func (s *RAGService) SearchSimilarChunks(query string, userID uint, docIDs []uint, topK int) ([]SearchResult, error) {
    // 1. 获取查询向量
    queryVec, err := s.embeddingSvc.GetEmbedding(query)
    if err != nil {
        return nil, err
    }

    // 2. 优先使用 Meilisearch 向量搜索
    chunkResults, err := s.searchSvc.VectorSearch(queryVec, userID, docIDs, topK)
    if err == nil && len(chunkResults) > 0 {
        return chunkResults, nil
    }

    // 3. 降级到 MySQL 余弦相似度
    log.Printf("Meilisearch 向量搜索失败，降级到 MySQL: %v", err)
    return s.searchByMySQL(queryVec, userID, docIDs, topK)
}
```

降级策略很重要——Meilisearch 挂了，RAG 还能工作。MySQL 中的 embedding 备份此时派上用场。

---

## 完整架构

最终架构非常清晰：

```
用户请求
  ├─ 关键词搜索 → Meilisearch 全文检索 → 结果
  └─ AI 对话
        ├─ 用户提问 → Embedding API → 查询向量
        ├─ Meilisearch 向量搜索 → 相似文档切片
        ├─ （降级）MySQL 余弦相似度
        ├─ 拼接 Prompt：问题 + 上下文
        └─ LLM 生成回答 → 流式输出
```

一个 Meilisearch 实例，同时服务两种截然不同的搜索需求。

## 踩坑记录

### 1. meilisearch-go v0.36 API 不兼容

SDK v0.36 与文档示例有多处不兼容：

| 问题 | 解决 |
|------|------|
| `SearchResult` 重复定义 | 用 `hit.DecodeInto()` 解码 |
| `Total` 字段不存在 | 改用 `EstimatedTotalHits` |
| `UpdateFilterableAttributes` 需 `*[]interface{}` | 不能传 `*[]string` |
| `AddDocumentsWithContext` 第三参数是 `*DocumentOptions` | 传 `&DocumentOptions{PrimaryKey: "id"}` |
| `UpdateEmbedders` 参数是 `map[string]Embedder` | 传值而非指针 |
| Hit 类型是 `map[string]interface{}` | 手动映射到目标结构体 |

### 2. float64 vs float32

通义千问 Embedding API 返回 `float64`，Meilisearch SDK 要求 `float32`。必须手动转换，不能直接传。

### 3. _vectors 字段格式

`_vectors` 的格式是 `{embedder_name: {embeddings: [[vec]], regenerate: false}}`，不是简单的向量数组。`embeddings` 是二维数组——因为一个文档可以有多个向量（虽然我们每个切片只有一个）。

### 4. Settings 更新是异步的

`UpdateFilterableAttributes`、`UpdateEmbedders` 等操作是异步的，返回 `TaskInfo`。如果紧接着搜索，可能配置还没生效。生产环境应该等待 task 完成。

## 总结

Meilisearch 在个人/小型项目中是 Elasticsearch 的优雅替代品：

- **轻量**：单二进制，~100MB 内存，5 分钟上手
- **中文友好**：内置分词，无需插件
- **API 简洁**：RESTful，SDK 支持多语言
- **向量搜索**：userProvided 模式让你自由选择 Embedding 模型
- **混合搜索**：关键词 + 语义一次搞定

在 RAG 场景中，将全文索引和向量索引统一到 Meilisearch，大幅简化了架构——一个搜索服务同时支撑关键词搜索和语义检索，维护成本降低，代码更清晰。

如果你的项目也在 10 万文档以内，试试 Meilisearch，你会发现搜索这件事可以这么简单。

---

*本文基于知识库项目的实际开发经验，项目使用 Go + Gin + Vue 3，Meilisearch 同时服务全文搜索和 RAG 向量检索。*
