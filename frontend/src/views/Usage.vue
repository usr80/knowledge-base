<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-4">用量统计</h1>
      </v-col>
    </v-row>

    <!-- 总计卡片 -->
    <v-row>
      <v-col cols="12" sm="6" md="3">
        <v-card>
          <v-card-text class="text-center">
            <div class="text-caption text-grey">总请求数</div>
            <div class="text-h4">{{ stats.totalRequests }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card>
          <v-card-text class="text-center">
            <div class="text-caption text-grey">输入 Tokens</div>
            <div class="text-h4">{{ formatNumber(stats.totalInput) }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card>
          <v-card-text class="text-center">
            <div class="text-caption text-grey">输出 Tokens</div>
            <div class="text-h4">{{ formatNumber(stats.totalOutput) }}</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card>
          <v-card-text class="text-center">
            <div class="text-caption text-grey">总费用（$）</div>
            <div class="text-h4">{{ stats.totalCost?.toFixed(4) || '0.0000' }}</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- 按提供商统计 -->
    <v-row class="mt-4">
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>按提供商统计</v-card-title>
          <v-card-text>
            <v-list v-if="stats.byProvider && stats.byProvider.length > 0">
              <v-list-item v-for="item in stats.byProvider" :key="item.provider">
                <template v-slot:prepend>
                  <v-avatar size="32" color="primary">
                    {{ item.provider.charAt(0).toUpperCase() }}
                  </v-avatar>
                </template>
                <v-list-item-title>{{ item.provider }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ item.requests }} 次 · {{ formatNumber(item.inputTokens + item.outputTokens) }} tokens
                </v-list-item-subtitle>
                <template v-slot:append>
                  <span class="text-body-2">${{ item.cost?.toFixed(4) || '0.0000' }}</span>
                </template>
              </v-list-item>
            </v-list>
            <v-alert v-else type="info" variant="tonal">暂无数据</v-alert>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- 按模型统计 -->
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>按模型统计</v-card-title>
          <v-card-text>
            <v-list v-if="stats.byModel && stats.byModel.length > 0">
              <v-list-item v-for="item in stats.byModel" :key="item.model">
                <v-list-item-title>{{ item.model }}</v-list-item-title>
                <v-list-item-subtitle>
                  {{ item.requests }} 次 · {{ formatNumber(item.inputTokens + item.outputTokens) }} tokens
                </v-list-item-subtitle>
                <template v-slot:append>
                  <span class="text-body-2">${{ item.cost?.toFixed(4) || '0.0000' }}</span>
                </template>
              </v-list-item>
            </v-list>
            <v-alert v-else type="info" variant="tonal">暂无数据</v-alert>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- 按日期统计 -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>最近 30 天用量趋势</v-card-title>
          <v-card-text>
            <v-simple-table v-if="stats.byDate && stats.byDate.length > 0">
              <thead>
                <tr>
                  <th>日期</th>
                  <th>请求数</th>
                  <th>输入 Tokens</th>
                  <th>输出 Tokens</th>
                  <th>费用</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in stats.byDate" :key="item.date">
                  <td>{{ item.date }}</td>
                  <td>{{ item.requests }}</td>
                  <td>{{ formatNumber(item.inputTokens) }}</td>
                  <td>{{ formatNumber(item.outputTokens) }}</td>
                  <td>${{ item.cost?.toFixed(4) || '0.0000' }}</td>
                </tr>
              </tbody>
            </v-simple-table>
            <v-alert v-else type="info" variant="tonal">暂无数据</v-alert>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- 最近使用记录 -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>最近使用记录</v-card-title>
          <v-card-text>
            <v-data-table
              :headers="logHeaders"
              :items="logs"
              :loading="loading"
              :items-per-page="10"
            >
              <template v-slot:item.createdAt="{ item }">
                {{ formatDateTime(item.createdAt) }}
              </template>
              <template v-slot:item.cost="{ item }">
                ${{ item.cost?.toFixed(6) || '0.000000' }}
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { chatAPI } from '@/api/chat'

const loading = ref(false)
const stats = ref({
  totalRequests: 0,
  totalInput: 0,
  totalOutput: 0,
  totalCost: 0,
  byProvider: [],
  byModel: [],
  byDate: []
})
const logs = ref([])

const logHeaders = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '提供商', key: 'provider' },
  { title: '模型', key: 'model' },
  { title: '类型', key: 'requestType' },
  { title: '输入 Tokens', key: 'inputTokens' },
  { title: '输出 Tokens', key: 'outputTokens' },
  { title: '费用', key: 'cost' },
  { title: '时间', key: 'createdAt' }
]

const formatNumber = (num) => {
  if (!num) return '0'
  return num.toLocaleString()
}

const formatDateTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

const loadStats = async () => {
  loading.value = true
  try {
    const res = await chatAPI.getUsageStats()
    stats.value = res || {}
  } catch (e) {
    console.error('加载统计数据失败:', e)
  } finally {
    loading.value = false
  }
}

const loadLogs = async () => {
  try {
    const res = await chatAPI.getUsageLogs({ limit: 50 })
    logs.value = res.logs || []
  } catch (e) {
    console.error('加载使用记录失败:', e)
  }
}

onMounted(() => {
  loadStats()
  loadLogs()
})
</script>
