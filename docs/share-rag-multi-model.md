# RAG + 多模型架构设计

## 一、RAG 流程

```
文档 → 分片 → Embedding → 存储向量
                          ↓
问题 → Embedding → 余弦相似度检索 → TopK 片段 → 构建 Prompt → LLM 回答
```

### 1.1 文档索引

```go
func (s *RAGService) IndexDocument(userID, documentID uint) error {
    // 获取文档
    var doc models.Document
    s.db.Where("id = ? AND user_id = ?", documentID, userID).First(&doc)

    // 分片（ChunkSize=500, Overlap=50，保留上下文衔接）
    chunks := s.chunkSvc.Chunk(doc.Content, chunkSize, chunkOverlap)

    // 批量 Embedding（通义千问 text-embedding-v2，batch ≤ 10）
    embeddings, _ := s.embeddingSvc.GetEmbeddings(chunks)

    // 存储到 MySQL（Embedding 字段为 []float64 JSON）
    for i, chunk := range chunks {
        docChunk := models.DocumentChunk{
            DocumentID: documentID, UserID: userID,
            ChunkIndex: i, Content: chunk,
        }
        docChunk.SetEmbedding(embeddings[i])
        s.db.Create(&docChunk)
    }
    return nil
}
```

**分片策略：**
- ChunkSize = 500 字符（平衡检索精度和上下文完整性）
- Overlap = 50 字符（避免边界信息丢失）
- 按 `\n\n`（段落）优先切分，保证语义完整性

### 1.2 语义检索

```go
func (s *RAGService) SearchSimilarChunks(query string, userID uint, topK int) ([]SearchResult, error) {
    queryVec, _ := s.embeddingSvc.GetEmbedding(query)

    var chunks []models.DocumentChunk
    s.db.Where("user_id = ?", userID).Find(&chunks)

    results := []SearchResult{}
    for _, chunk := range chunks {
        chunkVec, _ := chunk.GetEmbedding()
        score := cosineSimilarity(queryVec, chunkVec)
        if score >= minScore {
            results = append(results, SearchResult{Content: chunk.Content, Score: score})
        }
    }
    sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
    if len(results) > topK {
        results = results[:topK]
    }
    return results, nil
}
```

**余弦相似度：**

```go
func cosineSimilarity(a, b []float64) float64 {
    var dot, normA, normB float64
    for i := range a {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB) + 1e-8)
}
```

向量维度：1536（text-embedding-v2）

### 1.3 Prompt 构建

```go
func (s *RAGService) Ask(userID uint, question string, docIDs []uint) (string, error) {
    results, _ := s.SearchSimilarChunks(question, userID, docIDs, topK)

    var messages []chatMessage
    if len(results) > 0 {
        // RAG 模式：拼接检索到的文档片段作为上下文
        var ctx strings.Builder
        for i, r := range results {
            ctx.WriteString(fmt.Sprintf("【片段 %d】\n%s\n\n", i+1, r.Content))
        }
        messages = []chatMessage{{Role: "user", Content: ctx.String() + "问题：" + question}}
    } else {
        // 通用对话：无检索结果，直接用 LLM 回答
        messages = []chatMessage{{Role: "user", Content: question}}
    }

    return s.llmSvc.Chat(systemPrompt, messages)
}
```

**System Prompt 设计：**
```
你是一个智能助手，可以基于知识库内容和自身知识回答用户问题。

要求：
1. 优先使用知识库中的信息回答
2. 如果知识库中没有相关信息，可以基于自身知识给出回答
3. 回答要简洁准确，必要时引用文档编号
```

---

## 二、多模型 Provider 架构

### 2.1 设计思路

**问题：** 各家 API 格式不同，硬编码无法扩展。

```
┌─────────────────────────────────────────────────┐
│                 LLMService (无状态)              │
│  Chat() / ChatStream() / SelectModel()          │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│            Manager (全局单例)                    │
│  providers: map[string]Provider                  │
│  currentProvider: string                        │
│  currentModel: string                           │
│  mu: sync.RWMutex                               │
└────────────────────┬────────────────────────────┘
                     │
        ┌────────────┼────────────┬────────────┐
        │            │            │            │
        ▼            ▼            ▼            ▼
   TongyiProvider  OpenAI    DeepSeek    OllamaProvider
   (input.messages) (messages) (messages)  (/api/chat)
```

**关键点：**
- LLMService 无状态，每次调用通过 Manager 获取当前 Provider
- Manager 是全局单例（`sync.Once`），模型切换全局生效
- 新增提供商只需实现 Provider 接口，注册到 Manager

### 2.2 Provider 接口

```go
type Provider interface {
    Name() string
    Chat(systemPrompt string, messages []Message) (string, error)
    ChatStream(systemPrompt string, messages []Message, callback func(string)) error
    Models() []string
    SetModel(model string)
}
```

**5 个方法，覆盖所有场景：**
- `Name()` - 提供商标识（tongyi/openai/deepseek/zhipu/ollama）
- `Chat()` - 非流式对话
- `ChatStream()` - 流式对话（SSE）
- `Models()` - 返回支持的模型列表
- `SetModel()` - 切换模型

### 2.3 全局 Manager

```go
type Manager struct {
    providers       map[string]Provider
    defaultProvider string
    currentProvider string  // 当前活跃提供商
    currentModel    string
    mu              sync.RWMutex
}

var GlobalManager *Manager
var managerOnce sync.Once

func InitManager(configs map[string]ProviderConfig, defaultProvider string) *Manager {
    managerOnce.Do(func() {
        GlobalManager = NewManager()
        for name, cfg := range configs {
            switch name {
            case "tongyi":   GlobalManager.Register(NewTongyiProvider(cfg))
            case "openai":   GlobalManager.Register(NewOpenAIProvider(cfg))
            case "deepseek": GlobalManager.Register(NewDeepSeekProvider(cfg))
            case "zhipu":    GlobalManager.Register(NewZhipuProvider(cfg))
            case "ollama":   GlobalManager.Register(NewOllamaProvider(cfg))
            }
        }
        GlobalManager.SetDefault(defaultProvider)
    })
    return GlobalManager
}

// 全局切换模型（所有服务共享）
func (m *Manager) SelectModel(providerName, model string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.currentProvider = providerName
    if p, ok := m.providers[providerName]; ok && model != "" {
        p.SetModel(model)
    }
    m.currentModel = model
    return nil
}
```

**为什么全局单例：** ChatController 和 RAGService 各持有 LLMService 实例，但底层共享同一个 Manager。切换模型只需调一次 `Manager.SelectModel()`，全局生效。

### 2.4 各平台 API 差异

| | 请求格式 | 响应路径 | 流式方式 |
|---|---|---|---|
| 通义千问 | `{"model","input":{messages},"parameters"}` | `output.text` | SSE + `X-DashScope-SSE: enable` |
| OpenAI | `{"model","messages","max_tokens"}` | `choices[0].message.content` | SSE `data: [DONE]` |
| DeepSeek | 同 OpenAI | 同 OpenAI | 同 OpenAI |
| 智谱 | 同 OpenAI | 同 OpenAI | 同 OpenAI |
| Ollama | `{"model","messages","stream":true}` | `message.content` | NDJSON `{"done":true}` |

**通义千问非流式响应：**
```json
{
  "output": {
    "text": "回答内容",
    "finish_reason": "stop"
  },
  "usage": { "input_tokens": 100, "output_tokens": 50 }
}
```

**OpenAI 兼容格式响应：**
```json
{
  "choices": [{
    "message": { "role": "assistant", "content": "回答内容" },
    "finish_reason": "stop"
  }],
  "usage": { "prompt_tokens": 100, "completion_tokens": 50 }
}
```

### 2.5 Provider 实现示例（通义千问）

```go
type TongyiProvider struct {
    config ProviderConfig
}

func (p *TongyiProvider) Chat(systemPrompt string, messages []Message) (string, error) {
    fullMessages := buildMessages(systemPrompt, messages)

    reqBody := tongyiRequest{
        Model: p.config.Model,
        Input: tongyiInput{Messages: fullMessages},
        Parameters: &tongyiParams{
            MaxTokens:   p.config.MaxTokens,
            Temperature: p.config.Temperature,
        },
    }

    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", p.config.BaseURL, bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()

    var result tongyiResponse
    json.NewDecoder(resp.Body).Decode(&result)

    return result.Output.Text, nil
}

func (p *TongyiProvider) ChatStream(systemPrompt string, messages []Message, callback func(string)) error {
    fullMessages := buildMessages(systemPrompt, messages)

    reqBody := tongyiRequest{
        Model: p.config.Model,
        Input: tongyiInput{Messages: fullMessages},
    }

    body, _ := json.Marshal(reqBody)
    streamURL := p.config.BaseURL + "?incremental_output=true"

    req, _ := http.NewRequest("POST", streamURL, bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-DashScope-SSE", "enable")

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if !strings.HasPrefix(line, "data:") { continue }

        data := strings.TrimPrefix(line, "data:")
        if data == "[DONE]" { continue }

        var event tongyiStreamEvent
        json.Unmarshal([]byte(data), &event)

        if len(event.Output.Choices) > 0 {
            content := event.Output.Choices[0].Delta.Content
            if content != "" { callback(content) }
        }
    }
    return scanner.Err()
}
```

---

## 三、关键设计决策

### 3.1 为什么用 MySQL 存向量

**当前方案：**
```go
type DocumentChunk struct {
    ID         uint      `gorm:"primaryKey"`
    DocumentID uint      `gorm:"index"`
    UserID     uint      `gorm:"index"`
    ChunkIndex int
    Content    string
    Embedding  []float64 `gorm:"type:JSON"`  // 1536 维向量
    CreatedAt  time.Time
}
```

**原因：**
- 快速验证，零依赖
- 文档量 < 1 万条时，全表扫描 + 内存余弦相似度性能可接受
- 后续迁移到 Milvus/Qdrant 只需替换检索层

**演进路径：**
| 阶段 | 存储 | 检索 | 规模 |
|------|------|------|------|
| 当前 | MySQL JSON | 全表扫描 + 内存余弦 | < 1 万 |
| 下一步 | Milvus / Qdrant | ANN 索引（HNSW） | 10 万+ |
| 混合 | 向量库 + ES | 向量 + BM25 | 生产级 |

### 3.2 为什么全局单例 Manager

**问题：** 如果 Manager 不是单例，每次 `NewLLMService()` 创建新 Manager，模型切换不会全局生效。

**方案：**
```go
var managerOnce sync.Once

func InitManager(...) *Manager {
    managerOnce.Do(func() {
        GlobalManager = NewManager()
        // 注册 providers
    })
    return GlobalManager
}
```

**效果：** 无论 ChatController 还是 RAGService，调用 `NewLLMService()` 时，底层共享同一个 GlobalManager。

### 3.3 Embedding Batch Size 限制

**通义千问 API 限制：** 单次请求最多 10 条文本

**处理：**
```go
const batchSize = 10

func (s *EmbeddingService) GetEmbeddings(texts []string) ([][]float64, error) {
    var results [][]float64
    for i := 0; i < len(texts); i += batchSize {
        end := min(i+batchSize, len(texts))
        batch := texts[i:end]
        embeddings, _ := s.callEmbeddingAPI(batch)
        results = append(results, embeddings...)
    }
    return results, nil
}
```

---

## 四、数据流图

```
┌──────────────────────────────────────────────────────────────────┐
│                        前端 Chat.vue                              │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────┐     │
│  │ 选择 Provider│  │ 选择 Model  │  │ POST /api/chat/ask    │     │
│  │ (tongyi)    │  │ (qwen-turbo)│  │ Body: {question}      │     │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬───────────┘     │
│         │                │                     │                  │
│         └────────────────┼─────────────────────┘                  │
│                          ▼                                        │
│               POST /api/models/select                              │
│               Body: {provider, model}                             │
└──────────────────────────┬───────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                     ChatController.Ask                           │
│  1. 获取 userID from context                                      │
│  2. ragSvc.Ask(userID, question, docIDs)                         │
└──────────────────────────┬───────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                       RAGService.Ask                             │
│  1. embeddingSvc.GetEmbedding(question) → queryVec               │
│  2. SearchSimilarChunks(queryVec) → TopK 片段                     │
│  3. 构建 Prompt: context + question                              │
│  4. llmSvc.Chat(systemPrompt, messages)                          │
└──────────────────────────┬───────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                   LLMService.Chat                                │
│  manager.Get() → provider (当前活跃)                              │
│  provider.Chat(systemPrompt, messages)                          │
└──────────────────────────┬───────────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│                  TongyiProvider.Chat                             │
│  POST https://dashscope.aliyuncs.com/api/v1/.../generation       │
│  Headers: Authorization: Bearer {API_KEY}                        │
│  Body: {"model":"qwen-turbo","input":{"messages":[...]}}        │
└──────────────────────────────────────────────────────────────────┘
```
