<template>
  <div>
    <v-card elevation="2" rounded="lg">
      <!-- 工具栏 -->
      <v-card-title class="d-flex align-center pa-4 pb-0">
        <v-icon color="primary" class="mr-2">mdi-file-document-multiple</v-icon>
        <span>全部文档</span>
        <v-spacer />
        <v-text-field
          v-model="keyword"
          prepend-inner-icon="mdi-magnify"
          placeholder="搜索文档..."
          variant="outlined"
          density="compact"
          hide-details
          clearable
          style="max-width: 280px;"
          @click:clear="handleSearch"
          @keyup.enter="handleSearch"
          class="mr-3"
        />
        <v-select
          v-model="categoryID"
          :items="[{ title: '全部分类', value: null }, ...categories.map(c => ({ title: c.name, value: c.id }))]"
          item-title="title"
          item-value="value"
          placeholder="分类筛选"
          variant="outlined"
          density="compact"
          hide-details
          style="max-width: 160px;"
          class="mr-2"
          @update:modelValue="handleSearch"
        />
        <v-btn color="primary" prepend-icon="mdi-plus" to="/document/edit" rounded="lg">
          新建
        </v-btn>
      </v-card-title>

      <v-divider class="my-2" />

      <!-- 文档列表 -->
      <v-card-text class="pt-2">
        <v-data-table-server
          :headers="headers"
          :items="documents"
          :items-length="total"
          :loading="loading"
          :page="page"
          :items-per-page="pageSize"
          item-value="id"
          @update:page="page = $event; loadDocuments()"
          @update:items-per-page="pageSize = $event; loadDocuments()"
          hover
          class="elevation-0"
        >
          <template v-slot:item.title="{ item }">
            <router-link :to="`/documents/${item.id}`" class="text-decoration-none font-weight-medium">
              {{ item.title }}
            </router-link>
          </template>
          <template v-slot:item.category="{ item }">
            <v-chip v-if="item.category" size="small" color="primary" variant="tonal">
              {{ item.category.name }}
            </v-chip>
            <span v-else class="text-grey">未分类</span>
          </template>
          <template v-slot:item.tags="{ item }">
            <v-chip
              v-for="tag in item.tags"
              :key="tag.id"
              size="x-small"
              color="info"
              variant="tonal"
              class="mr-1"
            >
              {{ tag.name }}
            </v-chip>
          </template>
          <template v-slot:item.viewCount="{ item }">
            <v-icon size="small" class="mr-1">mdi-eye</v-icon>
            {{ item.viewCount }}
          </template>
          <template v-slot:item.updatedAt="{ item }">
            {{ formatDate(item.updatedAt) }}
          </template>
          <template v-slot:item.actions="{ item }">
            <v-btn icon variant="text" size="small" :to="`/document/edit/${item.id}`">
              <v-icon>mdi-pencil</v-icon>
              <v-tooltip activator="parent">编辑</v-tooltip>
            </v-btn>
            <v-btn icon variant="text" size="small" color="error" @click="handleDelete(item.id)">
              <v-icon>mdi-delete</v-icon>
              <v-tooltip activator="parent">删除</v-tooltip>
            </v-btn>
          </template>
        </v-data-table-server>
      </v-card-text>
    </v-card>

    <!-- 删除确认 -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>确认删除</v-card-title>
        <v-card-text>确定要删除该文档吗？此操作不可撤销。</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">取消</v-btn>
          <v-btn color="error" variant="elevated" @click="confirmDelete">删除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getDocuments, getCategories, deleteDocument } from '@/api'
import { showSnackbar } from '@/api/request'

const loading = ref(false)
const documents = ref([])
const categories = ref([])
const keyword = ref('')
const categoryID = ref(null)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const deleteDialog = ref(false)
const deleteTarget = ref(null)

const headers = [
  { title: '标题', key: 'title', minWidth: 200 },
  { title: '分类', key: 'category', width: 120 },
  { title: '标签', key: 'tags', width: 180 },
  { title: '阅读', key: 'viewCount', width: 80, align: 'center' },
  { title: '更新时间', key: 'updatedAt', width: 160 },
  { title: '操作', key: 'actions', width: 100, align: 'center', sortable: false }
]

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const loadDocuments = async () => {
  loading.value = true
  try {
    const res = await getDocuments({
      page: page.value,
      pageSize: pageSize.value,
      categoryID: categoryID.value || undefined,
      keyword: keyword.value || undefined
    })
    documents.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const loadCategories = async () => {
  try {
    const res = await getCategories()
    categories.value = res.data || []
  } catch (error) {
    // ignore
  }
}

const handleSearch = () => {
  page.value = 1
  loadDocuments()
}

const handleDelete = (id) => {
  deleteTarget.value = id
  deleteDialog.value = true
}

const confirmDelete = async () => {
  try {
    await deleteDocument(deleteTarget.value)
    showSnackbar('删除成功')
    deleteDialog.value = false
    loadDocuments()
  } catch (error) {
    // handled
  }
}

onMounted(() => {
  loadDocuments()
  loadCategories()
})
</script>

<style scoped>
:deep(.v-data-table tbody tr:hover) {
  background-color: rgba(25, 118, 210, 0.04) !important;
}
</style>
