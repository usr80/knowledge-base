<template>
  <v-container fluid class="chat-container pa-0">
    <v-row no-gutters class="fill-height">
      <!-- 左侧会话列表 -->
      <v-col cols="12" md="3" class="session-list-col">
        <v-card height="100%" flat class="rounded-0">
          <v-card-title class="d-flex align-center pa-4">
            <v-icon start>mdi-chat</v-icon>
            <span class="text-h6">AI 对话</span>
            <v-spacer />
            <v-btn
              icon="mdi-plus"
              variant="text"
              size="small"
              @click="newSession"
              title="新建会话"
            />
          </v-card-title>

          <v-divider />

          <!-- 模型选择 -->
          <div class="pa-4 pb-0">
            <v-select
              v-model="selectedProvider"
              :items="modelProviders"
              label="提供商"
              density="compact"
              variant="outlined"
              hide-details
              class="mb-2"
              @update:model-value="changeModel"
            >
              <template v-slot:selection="{ item }">
                <span class="text-capitalize">{{ item.value }}</span>
              </template>
              <template v-slot:item="{ item, props }">
                <v-list-item v-bind="props">
                  <v-list-item-title class="text-capitalize">{{ item.value }}</v-list-item-title>
                </v-list-item>
              </template>
            </v-select>

            <v-select
              v-model="selectedModel"
              :items="currentModels"
              label="模型"
              density="compact"
              variant="outlined"
              hide-details
              @update:model-value="changeModel"
            />
          </div>

          <v-divider class="mt-4" />

          <v-list density="compact" nav class="session-list">
            <v-list-item
              v-for="session in sessions"
              :key="session.sessionID"
              :active="currentSessionID === session.sessionID"
              @click="selectSession(session.sessionID)"
              class="session-item"
            >
              <v-list-item-title class="text-truncate">
                {{ session.title }}
              </v-list-item-title>
              <v-list-item-subtitle class="text-caption">
                {{ formatDate(session.updatedAt) }}
              </v-list-item-subtitle>
              <template v-slot:append>
                <v-btn
                  icon="mdi-delete"
                  variant="text"
                  size="x-small"
                  @click.stop="deleteSession(session.sessionID)"
                />
              </template>
            </v-list-item>

            <v-list-item v-if="sessions.length === 0" class="pa-4">
              <v-list-item-title class="text-center text-medium-emphasis">
                暂无对话记录
              </v-list-item-title>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>

      <!-- 右侧对话区域 -->
      <v-col cols="12" md="9" class="chat-area-col">
        <v-card height="100%" flat class="rounded-0 d-flex flex-column">
          <!-- 消息列表 -->
          <div class="message-list flex-grow-1 overflow-y-auto pa-4" ref="messageList">
            <div v-if="messages.length === 0" class="d-flex flex-column align-center justify-center h-100">
              <v-icon size="80" color="primary" class="mb-4">mdi-robot-happy</v-icon>
              <h2 class="text-h5 mb-2">AI 知识助手</h2>
              <p class="text-body-1 text-medium-emphasis text-center max-w-50">
                基于您的知识库内容，智能回答问题。<br>
                开始对话前，请先为文档创建索引。
              </p>
            </div>

            <div
              v-for="(msg, index) in messages"
              :key="index"
              :class="['message-item', 'mb-4', msg.role]"
            >
              <v-card
                :color="msg.role === 'user' ? 'primary' : undefined"
                :variant="msg.role === 'user' ? 'elevated' : 'outlined'"
                max-width="80%"
                :class="msg.role === 'user' ? 'ml-auto' : 'mr-auto assistant-card'"
              >
                <v-card-text class="message-content">
                  <div v-html="renderMarkdown(msg.content)"></div>
                </v-card-text>
              </v-card>
              <!-- 引用来源（只显示在最后一条助手消息下方） -->
              <div
                v-if="msg.role === 'assistant' && index === messages.length - 1 && currentReferences.length > 0 && !loading"
                class="references-area mt-1"
              >
                <div class="text-caption text-medium-emphasis mb-1">
                  <v-icon size="14" class="mr-1">mdi-bookmark</v-icon>引用来源
                </div>
                <v-chip
                  v-for="ref in currentReferences"
                  :key="ref.documentID"
                  size="small"
                  variant="tonal"
                  color="primary"
                  class="mr-1 mb-1"
                >
                  {{ ref.title }}
                  <span class="text-medium-emphasis ml-1">{{ (ref.score * 100).toFixed(0) }}%</span>
                </v-chip>
              </div>
            </div>

            <!-- 加载中 -->
            <div v-if="loading" class="d-flex align-center mb-4">
              <v-card variant="outlined" max-width="80%">
                <v-card-text>
                  <v-progress-circular indeterminate size="20" class="mr-2" />
                  <span class="text-medium-emphasis">思考中...</span>
                </v-card-text>
              </v-card>
            </div>
          </div>

          <!-- 输入区域 -->
          <v-divider />
          <div class="input-area pa-4">
            <v-form @submit.prevent="sendMessage">
              <v-textarea
                v-model="inputMessage"
                placeholder="输入您的问题..."
                auto-grow
                rows="1"
                max-rows="4"
                variant="outlined"
                hide-details
                :disabled="loading"
                @keydown.enter.exact.prevent="sendMessage"
                class="mb-2"
              />
              <div class="d-flex align-center">
                <v-chip
                  v-if="selectedDocuments.length > 0"
                  closable
                  color="primary"
                  variant="outlined"
                  class="mr-2"
                  @click:close="clearDocuments"
                >
                  已选 {{ selectedDocuments.length }} 篇文档
                </v-chip>
                <v-btn
                  icon="mdi-file-document-multiple"
                  variant="text"
                  size="small"
                  @click="showDocSelect = true"
                  title="选择检索文档"
                />
                <v-spacer />
                <v-btn
                  type="submit"
                  color="primary"
                  :disabled="!inputMessage.trim() || loading"
                  :loading="loading"
                  prepend-icon="mdi-send"
                >
                  发送
                </v-btn>
              </div>
            </v-form>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- 文档选择对话框 -->
    <v-dialog v-model="showDocSelect" max-width="600">
      <v-card>
        <v-card-title>选择检索文档</v-card-title>
        <v-card-text>
          <v-list density="compact" class="border rounded">
            <v-list-item
              v-for="doc in allDocuments"
              :key="doc.id"
              @click="toggleDocument(doc.id)"
            >
              <template v-slot:prepend>
                <v-checkbox
                  :model-value="selectedDocuments.includes(doc.id)"
                  hide-details
                  readonly
                />
              </template>
              <v-list-item-title>{{ doc.title }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="showDocSelect = false">取消</v-btn>
          <v-btn color="primary" @click="showDocSelect = false">确定</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, onMounted, nextTick, computed } from 'vue'
import { chatAPI, documentAPI } from '@/api/chat'
import { marked } from 'marked'

const sessions = ref([])
const currentSessionID = ref('')
const messages = ref([])
const inputMessage = ref('')
const loading = ref(false)
const messageList = ref(null)
const showDocSelect = ref(false)
const allDocuments = ref([])
const selectedDocuments = ref([])
const currentReferences = ref([])

// 模型选择
const modelProviders = ref([])
const modelList = ref({})
const selectedProvider = ref('tongyi')
const selectedModel = ref('qwen-turbo')

// 加载可用模型
const loadModels = async () => {
  try {
    const res = await chatAPI.getModels()
    modelProviders.value = res.providers || []
    modelList.value = res.models || {}
  } catch (err) {
    console.error('加载模型列表失败:', err)
  }
}

// 切换模型
const changeModel = async () => {
  try {
    await chatAPI.selectModel({
      provider: selectedProvider.value,
      model: selectedModel.value
    })
  } catch (err) {
    console.error('切换模型失败:', err)
  }
}

// 当前提供商的模型列表
const currentModels = computed(() => {
  return modelList.value[selectedProvider.value] || []
})

// 加载会话列表
const loadSessions = async () => {
  try {
    const res = await chatAPI.getSessions()
    sessions.value = res.sessions || []
  } catch (err) {
    console.error('加载会话失败:', err)
  }
}

// 加载文档列表
const loadDocuments = async () => {
  try {
    const res = await documentAPI.getList()
    allDocuments.value = res.documents || []
  } catch (err) {
    console.error('加载文档失败:', err)
  }
}

// 选择会话
const selectSession = async (sessionID) => {
  currentSessionID.value = sessionID
  try {
    const res = await chatAPI.getSession(sessionID)
    messages.value = res.messages || []
    scrollToBottom()
  } catch (err) {
    console.error('加载会话消息失败:', err)
  }
}

// 新建会话
const newSession = () => {
  currentSessionID.value = ''
  messages.value = []
}

// 发送消息（流式输出）
const sendMessage = async () => {
  if (!inputMessage.value.trim() || loading.value) return

  const question = inputMessage.value.trim()
  inputMessage.value = ''

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: question
  })

  // 添加空的助手消息（用于流式填充）
  const assistantMsgIndex = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: ''
  })
  scrollToBottom()

  loading.value = true
  let fullAnswer = ''
  currentReferences.value = [] // 清空上次引用

  try {
    await chatAPI.askStream(
      {
        question,
        sessionID: currentSessionID.value,
        documentIDs: selectedDocuments.value.length > 0 ? selectedDocuments.value : null
      },
      // onMessage: 接收到流式内容
      (chunk) => {
        fullAnswer += chunk
        messages.value[assistantMsgIndex].content = fullAnswer
        scrollToBottom()
      },
      // onDone: 流式结束
      (sessionID) => {
        loading.value = false
        if (!currentSessionID.value && sessionID) {
          currentSessionID.value = sessionID
          loadSessions()
        }
      },
      // onError: 错误处理
      (err) => {
        loading.value = false
        messages.value[assistantMsgIndex].content = '抱歉，回答生成失败：' + (err.message || '网络错误')
      },
      // onReferences: 接收引用来源
      (refs) => {
        currentReferences.value = refs
      }
    )
  } catch (err) {
    loading.value = false
    messages.value[assistantMsgIndex].content = '抱歉，回答生成失败：' + (err.message || '未知错误')
  }
}

// 删除会话
const deleteSession = async (sessionID) => {
  try {
    await chatAPI.deleteSession(sessionID)
    sessions.value = sessions.value.filter(s => s.sessionID !== sessionID)
    if (currentSessionID.value === sessionID) {
      newSession()
    }
  } catch (err) {
    console.error('删除会话失败:', err)
  }
}

// 文档选择
const toggleDocument = (docID) => {
  const idx = selectedDocuments.value.indexOf(docID)
  if (idx >= 0) {
    selectedDocuments.value.splice(idx, 1)
  } else {
    selectedDocuments.value.push(docID)
  }
}

const clearDocuments = () => {
  selectedDocuments.value = []
}

// 渲染 Markdown
const renderMarkdown = (text) => {
  return marked.parse(text || '')
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + ' 分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + ' 小时前'
  return date.toLocaleDateString()
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messageList.value) {
      messageList.value.scrollTop = messageList.value.scrollHeight
    }
  })
}

onMounted(() => {
  loadSessions()
  loadDocuments()
  loadModels()
})
</script>

<style scoped>
.chat-container {
  height: calc(100vh - 64px);
  min-height: 500px;
}

.session-list-col {
  border-right: 1px solid rgba(0, 0, 0, 0.12);
}

.session-list {
  height: calc(100% - 64px);
  overflow-y: auto;
}

.session-item {
  cursor: pointer;
}

.session-item:hover {
  background-color: rgba(0, 0, 0, 0.04);
}

.message-list {
  height: calc(100% - 120px);
}

.message-item.user .v-card {
  margin-left: auto;
}

.message-item.assistant .v-card {
  margin-right: auto;
}

.message-item.assistant .message-content {
  color: rgba(0, 0, 0, 0.87);
}

.assistant-card {
  background-color: #f5f5f5 !important;
}

.assistant-card .message-content {
  color: rgba(0, 0, 0, 0.87) !important;
}

.message-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.input-area {
  background-color: rgba(0, 0, 0, 0.02);
}

.max-w-50 {
  max-width: 50%;
}

@media (max-width: 960px) {
  .session-list-col {
    display: none;
  }

  .max-w-50 {
    max-width: 90%;
  }
}
.references-area {
  max-width: 80%;
  margin-left: auto;
  padding-left: 8px;
}

.message-item.user .references-area {
  margin-left: 0;
  margin-right: auto;
  padding-left: 0;
}

.message-item.assistant .references-area {
  margin-left: 0;
  margin-right: auto;
}
</style>
