# 通义千问 SSE 流式：累计文本 vs 增量 Delta

## 一个 bug 牵出的问题

前端流式对话，通义千问返回的内容重复：

```
你好！你好！有什么有什么我可以帮助你的吗？
```

排查发现：通义千问每条 SSE 事件的 `output.text` 是**完整文本**，不是增量。

```json
// event 1
{"output":{"text":"你好"}}

// event 2
{"output":{"text":"你好！"}}

// event 3
{"output":{"text":"你好！有什么"}}
```

而 OpenAI 系列返回的是增量 delta：

```json
// event 1
{"choices":[{"delta":{"content":"你好"}}]}

// event 2
{"choices":[{"delta":{"content":"！"}}]}

// event 3
{"choices":[{"delta":{"content":"有什么"}}]}
```

如果后端把通义千问的每条 `output.text` 当作 delta 直接转发给前端，前端逐条拼接，自然重复。

## 为什么会有两种设计？

通义千问的 `output.text` 默认是累计的，但提供了 `incremental_output=true` 参数。

加了参数后，文档说是增量输出，但**实际行为是：首条返回完整文本，后续返回增量**。

更坑的是，这个参数的行为在 SDK 和 HTTP API 之间不一致。为了可靠性，我们选择**不依赖这个参数**，而是在后端自行计算增量。

## 增量计算的核心逻辑

```go
var lastText string

for scanner.Scan() {
    line := scanner.Text()
    if !strings.HasPrefix(line, "data:") {
        continue
    }

    data := strings.TrimPrefix(line, "data:")
    var result tongyiStreamEvent
    json.Unmarshal([]byte(data), &result)

    // 核心计算：只发送新增部分
    if result.Output.Text != "" && result.Output.Text != lastText {
        increment := strings.TrimPrefix(result.Output.Text, lastText)
        if increment != "" {
            callback(increment)
        }
        lastText = result.Output.Text
    }
}
```

`strings.TrimPrefix(result.Output.Text, lastText)` —— 用已发送长度截取新增部分。

这是 O(1) 的字符串操作，没有正则、没有分词，直接基于前缀匹配。

## 为什么不用下标切片？

直觉可能会写：

```go
increment := result.Output.Text[len(lastText):]
```

这更高效，但有风险：`output.text` 不保证严格递增。如果某条事件的文本被截断或修正（如敏感词过滤），下标切片会越界或产生乱码（UTF-8 多字节字符中间截断）。

`strings.TrimPrefix` 更安全：如果前缀不匹配，返回原字符串，不会 panic。代价是多一次字符串比较，在 SSE 场景下完全可忽略。

## 为什么不直接让前端去重？

假设前端收到累计文本，自己计算增量：

```javascript
let lastText = ''
event => {
    const delta = event.text.slice(lastText.length)
    lastText = event.text
    fullAnswer += delta
}
```

看起来也行，但问题在于：

1. **网络不可靠**：SSE 事件可能丢失或乱序，前端用下标切片会错位
2. **前端无法感知错误**：如果某条事件的 `text` 被修正（过滤后变短了），前端不知道该回退多少
3. **后端是单一责任点**：增量计算放在后端，前端只管累加，职责清晰

后端计算增量、前端只做 `fullAnswer += chunk`，这是最不容易出问题的架构。

## 回到最初：为什么会踩这个坑？

后端最初按 OpenAI 格式解析通义千问的 SSE：

```go
// 错误的结构体定义
type tongyiStreamEvent struct {
    Choices []struct {
        Delta struct {
            Content string `json:"content"`
        } `json:"delta"`
    } `json:"choices"`
}
```

通义千问根本没有 `choices` 字段，解析后是零值，callback 永远不被调用，前端只收到 `event:done`。

**根因**：想当然地认为所有 LLM API 都兼容 OpenAI 格式。实际上只有 DeepSeek、智谱等明确声明兼容，通义千问和 Ollama 各有各的格式。

## 格式差异一览

| Provider | 流式字段 | 增量/累计 | Usage 位置 |
|----------|---------|----------|-----------|
| 通义千问 | `output.text` | **累计** | 最后一条 `usage` |
| OpenAI | `choices[0].delta.content` | 增量 | 需 `stream_options` |
| DeepSeek | `choices[0].delta.content` | 增量 | 兼容 OpenAI |
| 智谱 | `choices[0].delta.content` | 增量 | 兼容 OpenAI |
| Ollama | `message.content` | 增量 | `done:true` 时 |

通义千问是唯一返回累计文本的。这意味着你的流式抽象层必须考虑这种差异，而不是假设所有 Provider 都是 delta。

## 抽象层设计

```go
type Provider interface {
    ChatStream(systemPrompt string, messages []Message, callback func(string)) (*ChatResponse, error)
}
```

`callback(string)` 的契约是：**传入的字符串是增量 delta**。

无论 Provider 内部是累计还是增量，对外统一为增量。这个归一化在 Provider 实现内部完成，调用方无需关心。

通义千问在内部做 `TrimPrefix` 转换，OpenAI 直接透传，Ollama 直接透传。调用方只看到 delta。

## 总结

一个"流式输出重复"的 bug，根因是 LLM API 格式差异。修复路径：

1. **SSE 无内容** → 结构体字段对不上，改用 `output.text`
2. **内容重复** → 累计 vs 增量，后端计算 delta
3. **用 TrimPrefix 不用下标切片** → 防御性编程，避免乱序和截断风险
4. **归一化在 Provider 内部** → 对外统一 delta 语义

核心教训：**不要假设 LLM API 都兼容 OpenAI 格式**。在 Provider 抽象层内部做好格式归一化，比在外部打补丁更可靠。
