/**
 * request.js - Axios 请求封装（Vuetify 版，无 Element Plus 依赖）
 */
import axios from 'axios'
import { useUserStore } from '@/stores/user'

const request = axios.create({
  baseURL: '/api',
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
  response => response.data,
  error => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        const userStore = useUserStore()
        userStore.logout()
        showSnackbar('登录已过期，请重新登录', 'error')
        window.location.href = '/login'
      } else {
        showSnackbar(data.error || '请求失败', 'error')
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
