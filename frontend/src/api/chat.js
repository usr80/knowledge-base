import request from './request'

// 聊天相关 API
export const chatAPI = {
  // 提问
  ask(data) {
    return request.post('/chat/ask', data)
  },

  // 流式提问
  askStream(data, onMessage, onDone, onError) {
    const userStore = require('@/stores/user').useUserStore()
    const baseURL = import.meta.env.DEV ? '/api' : './api'

    const eventSource = new EventSource(
      `${baseURL}/chat/ask/stream`,
      {
        headers: {
          'Authorization': `Bearer ${userStore.token}`,
          'Content-Type': 'application/json'
        }
      }
    )

    // 注意：EventSource 不支持 POST，这里用 fetch 替代
    fetch(`${baseURL}/chat/ask/stream`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${userStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    }).then(response => {
      const reader = response.body.getReader()
      const decoder = new TextDecoder()

      const read = () => {
        reader.read().then(({ done, value }) => {
          if (done) {
            onDone && onDone()
            return
          }

          const text = decoder.decode(value)
          const lines = text.split('\n')

          lines.forEach(line => {
            if (line.startsWith('data:')) {
              const data = line.slice(5).trim()
              if (data && data !== '[DONE]') {
                try {
                  const parsed = JSON.parse(data)
                  if (parsed.content) {
                    onMessage && onMessage(parsed.content)
                  }
                } catch (e) {
                  // 忽略解析错误
                }
              }
            }
          })

          read()
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
