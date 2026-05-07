<template>
  <div>
    <v-card elevation="2" rounded="lg">
      <v-card-title class="pa-6 pb-2">
        <v-icon color="primary" class="mr-2">mdi-{{ isEdit ? 'pencil' : 'plus' }}</v-icon>
        {{ isEdit ? '编辑文档' : '新建文档' }}
      </v-card-title>

      <v-divider class="my-2" />

      <v-card-text class="pa-6">
        <v-form ref="formRef" @submit.prevent="handleSubmit">
          <v-text-field
            v-model="form.title"
            label="标题"
            prepend-inner-icon="mdi-format-title"
            variant="outlined"
            :rules="[v => !!v || '请输入文档标题']"
            density="comfortable"
            class="mb-3"
          />

          <v-row>
            <v-col cols="12" sm="6">
              <v-select
                v-model="form.categoryID"
                :items="[{ title: '不选择分类', value: null }, ...categories.map(c => ({ title: c.name, value: c.id }))]"
                item-title="title"
                item-value="value"
                label="分类"
                prepend-inner-icon="mdi-folder"
                variant="outlined"
                density="comfortable"
                clearable
              />
            </v-col>
            <v-col cols="12" sm="6">
              <v-combobox
                v-model="form.tags"
                :items="[]"
                label="标签（回车添加）"
                prepend-inner-icon="mdi-tag"
                variant="outlined"
                density="comfortable"
                multiple
                chips
                closable-chips
                clearable
              />
            </v-col>
          </v-row>

          <v-textarea
            v-model="form.summary"
            label="摘要"
            prepend-inner-icon="mdi-text-short"
            variant="outlined"
            rows="2"
            density="comfortable"
            class="mb-3"
          />

          <v-textarea
            v-model="form.content"
            label="内容（支持 Markdown）"
            prepend-inner-icon="mdi-language-markdown"
            variant="outlined"
            rows="16"
            density="comfortable"
            class="mb-3"
          />

          <div class="d-flex justify-end ga-3">
            <v-btn variant="outlined" @click="$router.back()">取消</v-btn>
            <v-btn color="primary" variant="elevated" :loading="loading" type="submit" rounded="lg">
              保存
            </v-btn>
          </div>
        </v-form>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getDocument, createDocument, updateDocument, getCategories } from '@/api'
import { showSnackbar } from '@/api/request'

const route = useRoute()
const router = useRouter()
const formRef = ref(null)
const loading = ref(false)
const categories = ref([])

const isEdit = computed(() => !!route.params.id)

const form = reactive({
  title: '',
  categoryID: null,
  tags: [],
  summary: '',
  content: ''
})

const loadDocument = async () => {
  if (!route.params.id) return
  try {
    const res = await getDocument(route.params.id)
    const doc = res.data
    form.title = doc.title
    form.categoryID = doc.categoryID || null
    form.tags = doc.tags?.map(t => t.name) || []
    form.summary = doc.summary || ''
    form.content = doc.content || ''
  } catch (error) {
    // handled
  }
}

const loadCategories = async () => {
  try {
    const res = await getCategories()
    categories.value = res.data || []
  } catch (error) {
    // handled
  }
}

const handleSubmit = async () => {
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    const data = {
      title: form.title,
      categoryID: form.categoryID || undefined,
      tags: form.tags,
      summary: form.summary,
      content: form.content
    }

    if (isEdit.value) {
      await updateDocument(route.params.id, data)
      showSnackbar('更新成功')
    } else {
      await createDocument(data)
      showSnackbar('创建成功')
    }
    router.push('/documents')
  } catch (error) {
    // handled
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDocument()
  loadCategories()
})
</script>
