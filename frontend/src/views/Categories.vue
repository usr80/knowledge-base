<template>
  <div>
    <v-card elevation="2" rounded="lg">
      <!-- 标题栏 -->
      <v-card-title class="d-flex align-center pa-6 pb-0">
        <v-icon color="primary" class="mr-2">mdi-folder-multiple</v-icon>
        <span class="text-h6 font-weight-bold">分类管理</span>
        <v-spacer />
        <v-btn color="primary" prepend-icon="mdi-plus" rounded="lg" @click="handleAdd">
          新建分类
        </v-btn>
      </v-card-title>

      <v-divider class="my-2" />

      <!-- 分类列表 -->
      <v-card-text>
        <v-table hover class="elevation-0">
          <thead>
            <tr>
              <th>分类名称</th>
              <th>图标</th>
              <th>排序</th>
              <th>创建时间</th>
              <th style="width: 120px;">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="5" class="text-center py-8">
                <v-progress-circular indeterminate color="primary" size="32" />
              </td>
            </tr>
            <tr v-else-if="categories.length === 0">
              <td colspan="5" class="text-center text-grey py-8">暂无分类</td>
            </tr>
            <tr v-for="cat in categories" :key="cat.id" v-else>
              <td>
                <v-icon v-if="cat.icon" class="mr-2">{{ cat.icon }}</v-icon>
                <span class="font-weight-medium">{{ cat.name }}</span>
              </td>
              <td>{{ cat.icon || '-' }}</td>
              <td class="text-center">{{ cat.sortOrder }}</td>
              <td>{{ formatDate(cat.createdAt) }}</td>
              <td>
                <v-btn icon variant="text" size="small" @click="handleEdit(cat)">
                  <v-icon>mdi-pencil</v-icon>
                  <v-tooltip activator="parent">编辑</v-tooltip>
                </v-btn>
                <v-btn icon variant="text" size="small" color="error" @click="handleDelete(cat.id)">
                  <v-icon>mdi-delete</v-icon>
                  <v-tooltip activator="parent">删除</v-tooltip>
                </v-btn>
              </td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
    </v-card>

    <!-- 新建/编辑对话框 -->
    <v-dialog v-model="dialogVisible" max-width="480">
      <v-card rounded="lg">
        <v-card-title class="pa-5 pb-2">
          <v-icon class="mr-2" color="primary">{{ editingId ? 'mdi-pencil' : 'mdi-plus' }}</v-icon>
          {{ editingId ? '编辑分类' : '新建分类' }}
        </v-card-title>

        <v-card-text class="pa-5 pt-0">
          <v-form ref="formRef" @submit.prevent="handleSubmit">
            <v-text-field
              v-model="form.name"
              label="分类名称"
              prepend-inner-icon="mdi-label"
              variant="outlined"
              :rules="[v => !!v || '请输入分类名称']"
              density="comfortable"
              class="mb-3"
            />
            <v-text-field
              v-model="form.icon"
              label="图标（如 mdi-folder，可选）"
              prepend-inner-icon="mdi-icon"
              variant="outlined"
              density="comfortable"
              class="mb-3"
            />
            <v-select
              v-model="form.parentID"
              :items="[{ title: '无（顶级分类）', value: null }, ...parentCategories.map(c => ({ title: c.name, value: c.id }))]"
              item-title="title"
              item-value="value"
              label="父分类"
              prepend-inner-icon="mdi-folder-account"
              variant="outlined"
              density="comfortable"
              clearable
            />
          </v-form>
        </v-card-text>

        <v-card-actions class="pa-5 pt-0">
          <v-spacer />
          <v-btn variant="text" @click="dialogVisible = false">取消</v-btn>
          <v-btn color="primary" variant="elevated" :loading="submitLoading" @click="handleSubmit" rounded="lg">
            确定
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- 删除确认 -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>确认删除</v-card-title>
        <v-card-text>确定要删除该分类吗？子分类也将被删除。</v-card-text>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { getCategories, createCategory, updateCategory, deleteCategory } from '@/api'
import { showSnackbar } from '@/api/request'

const loading = ref(false)
const submitLoading = ref(false)
const categories = ref([])
const dialogVisible = ref(false)
const deleteDialog = ref(false)
const deleteTarget = ref(null)
const editingId = ref(null)
const formRef = ref(null)

const form = reactive({
  name: '',
  icon: '',
  parentID: null
})

// 排除当前编辑项作为父分类候选
const parentCategories = computed(() =>
  categories.value.filter(c => c.id !== editingId.value)
)

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const loadCategories = async () => {
  loading.value = true
  try {
    const res = await getCategories()
    categories.value = res.data || []
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  editingId.value = null
  form.name = ''
  form.icon = ''
  form.parentID = null
  dialogVisible.value = true
}

const handleEdit = (row) => {
  editingId.value = row.id
  form.name = row.name
  form.icon = row.icon || ''
  form.parentID = row.parentID || null
  dialogVisible.value = true
}

const handleDelete = (id) => {
  deleteTarget.value = id
  deleteDialog.value = true
}

const confirmDelete = async () => {
  try {
    await deleteCategory(deleteTarget.value)
    showSnackbar('删除成功')
    deleteDialog.value = false
    loadCategories()
  } catch (error) {
    // handled
  }
}

const handleSubmit = async () => {
  const { valid } = await formRef.value.validate()
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      name: form.name,
      icon: form.icon,
      parentID: form.parentID || undefined
    }

    if (editingId.value) {
      await updateCategory(editingId.value, data)
      showSnackbar('更新成功')
    } else {
      await createCategory(data)
      showSnackbar('创建成功')
    }
    dialogVisible.value = false
    loadCategories()
  } catch (error) {
    // handled
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => loadCategories())
</script>

<style scoped>
:deep(.v-table tbody tr:hover) {
  background-color: rgba(25, 118, 210, 0.04) !important;
}
</style>
