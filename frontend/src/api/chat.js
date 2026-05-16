import request from './request'
import { useUserStore } from '@/stores/user'

// 聊天相关 API
export const chatAPI = {
  // 提问
  ask(data) {
    return request.post('/chat/ask', data)
  },

  // 流式提问
  askStream(data, onMessage, onDone, onError, onReferences) {
    const userStore = useUserStore()
    const baseURL = import.meta.env.DEV ? '/api' : './api'

    fetch(`${baseURL}/chat/ask/stream`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${userStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    }).then(response => {
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let sessionID = ''
      let buffer = ''  // 缓冲区：处理不完整的 SSE 行

      const read = () => {
        reader.read().then(({ done, value }) => {
          if (done) {
            onDone && onDone(sessionID)
            return
          }

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          // 最后一行可能不完整，保留在缓冲区
          buffer = lines.pop() || ''

          for (const line of lines) {
            if (!line.startsWith('data:')) continue
            const data = line.slice(5).trim()
            if (!data) continue

            try {
              const parsed = JSON.parse(data)
              if (parsed.type === 'references' && parsed.references) {
                onReferences && onReferences(parsed.references)
              } else {
                if (parsed.content) {
                  onMessage && onMessage(parsed.content)
                }
              }
              if (parsed.sessionID) {
                sessionID = parsed.sessionID
              }
            } catch (e) {
              // 忽略解析错误
            }
          }

          read()
        }).catch(err => {
          onError && onError(err)
        })
      }

      read()
    }).catch(err => {
      onError && onError(err)
    })
  },

  // 获取会话列表
  getSessions() {
    return request.get('/chat/sessions')
  },

  // 获取会话详情
  getSession(sessionID) {
    return request.get(`/chat/sessions/${sessionID}`)
  },

  // 删除会话
  deleteSession(sessionID) {
    return request.delete(`/chat/sessions/${sessionID}`)
  },

  // 创建文档索引
  indexDocument(documentID) {
    return request.post(`/documents/${documentID}/index`)
  },

  // 获取可用模型列表
  getModels() {
    return request.get('/models')
  },

  // 选择模型
  selectModel(data) {
    return request.post('/models/select', data)
  },

  // 获取用量统计
  getUsageStats(params) {
    return request.get('/chat/usage/stats', { params })
  },

  // 获取用量记录
  getUsageLogs(params) {
    return request.get('/chat/usage/logs', { params })
  }
}

// 文档相关 API（补充）
export const documentAPI = {
  getList(params) {
    return request.get('/documents', { params }).then(res => ({
      documents: res.documents || [],
      total: res.total || 0
    }))
  },

  get(id) {
    return request.get(`/documents/${id}`)
  },

  create(data) {
    return request.post('/documents', data)
  },

  update(id, data) {
    return request.put(`/documents/${id}`, data)
  },

  delete(id) {
    return request.delete(`/documents/${id}`)
  },

  import(formData) {
    return request.post('/documents/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  exportMarkdown(id) {
    return request.get(`/documents/${id}/export/markdown`, { responseType: 'blob' })
  },

  // 创建索引
  createIndex(id) {
    return request.post(`/documents/${id}/index`)
  }
}