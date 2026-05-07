<template>
  <div>
    <v-card elevation="2" rounded="lg" :loading="loading">
      <!-- 标题栏 -->
      <v-card-title class="d-flex align-center pa-6">
        <v-btn icon variant="text" @click="$router.back()" class="mr-2">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
        <h2 class="text-h6 font-weight-bold flex-grow-1">{{ document.title }}</h2>
        <v-btn color="primary" prepend-icon="mdi-pencil" :to="`/document/edit/${route.params.id}`" rounded="lg">
          编辑
        </v-btn>
      </v-card-title>

      <v-divider />

      <v-card-text class="pa-6">
        <!-- 元信息 -->
        <div class="d-flex align-center mb-6">
          <v-chip v-if="document.category" size="small" color="primary" variant="tonal" prepend-icon="mdi-folder">
            {{ document.category.name }}
          </v-chip>
          <span v-else class="text-grey text-body-2 mr-3">未分类</span>
          <v-chip
            v-for="tag in document.tags"
            :key="tag.id"
            size="small"
            color="info"
            variant="tonal"
            class="ml-2"
          >
            {{ tag.name }}
          </v-chip>
          <v-spacer />
          <span class="text-body-2 text-grey">
            <v-icon size="small" class="mr-1">mdi-clock-outline</v-icon>
            更新于 {{ formatDate(document.updatedAt) }}
          </span>
        </div>

        <v-divider class="mb-6" />

        <!-- 内容 -->
        <div class="doc-content" v-html="renderedContent"></div>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getDocument } from '@/api'

const route = useRoute()
const loading = ref(false)
const document = ref({})

const renderedContent = computed(() => {
  const content = document.value.content || ''
  return content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/`(.+?)`/g, '<code>$1</code>')
    .replace(/^\- (.+)$/gm, '<li>$1</li>')
    .replace(/^> (.+)$/gm, '<blockquote>$1</blockquote>')
    .replace(/\n/g, '<br>')
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const loadDocument = async () => {
  loading.value = true
  try {
    const res = await getDocument(route.params.id)
    document.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(() => loadDocument())
</script>

<style scoped>
.doc-content {
  line-height: 1.8;
  font-size: 16px;
}
.doc-content :deep(h1) { margin-top: 24px; margin-bottom: 16px; font-size: 1.8em; }
.doc-content :deep(h2) { margin-top: 20px; margin-bottom: 12px; font-size: 1.4em; }
.doc-content :deep(h3) { margin-top: 16px; margin-bottom: 8px; font-size: 1.2em; }
.doc-content :deep(code) {
  background-color: #E8EAF6;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.9em;
}
.doc-content :deep(blockquote) {
  border-left: 4px solid #1976D2;
  padding-left: 16px;
  color: #666;
  margin: 16px 0;
}
</style>
