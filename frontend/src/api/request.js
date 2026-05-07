/**
 * request.js - Axios 请求封装（Vuetify 版，无 Element Plus 依赖）
 */
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import router from '@/router'

// 开发模式用 /api（Vite 代理），生产模式用相对路径
const baseURL = import.meta.env.DEV ? '/api' : './api'

const request = axios.create({
  baseURL,
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器 — 不再依赖 ElMessage，改用 window.__snackbar
request.interceptors.response.use(
  response => {
    // 如果请求期望 blob，直接返回原始 data（可能是 Blob 或 ArrayBuffer）
    if (response.config.responseType === 'blob') {
      return response.data
    }
    return response.data
  },
  error => {
    if (error.response) {
      const { status, data } = error.response
      // 处理 blob 请求的错误：尝试从 Blob 中解析 JSON
      if (error.config?.responseType === 'blob' && data instanceof Blob) {
        data.text().then(text => {
          try {
            const json = JSON.parse(text)
            showSnackbar(json.error || '请求失败', 'error')
          } catch {
            showSnackbar('请求失败', 'error')
          }
        })
      } else if (status === 401) {
        const userStore = useUserStore()
        userStore.logout()
        showSnackbar('登录已过期，请重新登录', 'error')
        router.push({ name: 'Login' })
      } else {
        showSnackbar(data?.error || '请求失败', 'error')
      }
    } else {
      showSnackbar('网络错误', 'error')
    }
    return Promise.reject(error)
  }
)

/** 全局 snackbar 触发器，Layout 中监听 */
export function showSnackbar(text, color = 'success') {
  window.dispatchEvent(new CustomEvent('snackbar', { detail: { text, color } }))
}

export default request
