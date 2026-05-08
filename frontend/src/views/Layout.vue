<template>
  <v-app>
    <!-- 侧边栏 -->
    <v-navigation-drawer
      v-model="drawer"
      :rail="rail"
      permanent
      color="primary"
      theme="dark"
    >
      <!-- Logo -->
      <v-list-item
        prepend-icon="mdi-book-open-page-variant"
        :title="rail ? '' : '知识库'"
        nav
        class="pa-3"
      >
        <template v-slot:append>
          <v-btn
            variant="text"
            :icon="rail ? 'mdi-chevron-right' : 'mdi-chevron-left'"
            @click="rail = !rail"
          />
        </template>
      </v-list-item>

      <v-divider />

      <!-- 导航菜单 -->
      <v-list density="compact" nav>
        <v-list-item
          prepend-icon="mdi-file-document-multiple"
          title="全部文档"
          value="documents"
          to="/documents"
          :active="activeMenu === 'documents'"
        />
        <v-list-item
          prepend-icon="mdi-folder-multiple"
          title="分类管理"
          value="categories"
          to="/categories"
          :active="activeMenu === 'categories'"
        />
        <v-list-item
          prepend-icon="mdi-robot-happy"
          title="AI 对话"
          value="chat"
          to="/chat"
          :active="activeMenu === 'chat'"
        />
        <v-list-item
          prepend-icon="mdi-account-circle"
          title="个人中心"
          value="profile"
          to="/profile"
          :active="activeMenu === 'profile'"
        />
      </v-list>
    </v-navigation-drawer>

    <!-- 顶栏 -->
    <v-app-bar elevation="1" color="surface">
      <v-app-bar-title>
        <v-btn
          color="primary"
          variant="elevated"
          prepend-icon="mdi-plus"
          @click="goToEdit"
          rounded="lg"
        >
          新建文档
        </v-btn>
      </v-app-bar-title>

      <v-spacer />

      <!-- 用户头像 & 菜单 -->
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn v-bind="props" variant="text" class="text-none">
            <v-avatar color="primary" size="36" class="mr-2">
              <span class="text-white text-body-2">{{ avatarLetter }}</span>
            </v-avatar>
            <span class="d-none d-sm-inline">{{ displayName }}</span>
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </template>
        <v-list density="compact" min-width="160">
          <v-list-item prepend-icon="mdi-account-circle" title="个人中心" to="/profile" />
          <v-divider />
          <v-list-item prepend-icon="mdi-logout" title="退出登录" @click="handleLogout" />
        </v-list>
      </v-menu>
    </v-app-bar>

    <!-- 主内容 -->
    <v-main>
      <v-container fluid class="pa-6" style="max-width: 1400px; margin: 0 auto; min-height: calc(100vh - 64px);">
        <router-view />
      </v-container>
    </v-main>

    <!-- 全局 Snackbar -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="3000"
      location="top"
    >
      {{ snackbar.text }}
      <template v-slot:actions>
        <v-btn variant="text" @click="snackbar.show = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { showSnackbar } from '@/api/request'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const drawer = ref(true)
const rail = ref(false)

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/categories')) return 'categories'
  if (path.startsWith('/chat')) return 'chat'
  if (path.startsWith('/profile')) return 'profile'
  return 'documents'
})

const displayName = computed(() => {
  return userStore.userInfo?.nickname || userStore.userInfo?.username || '用户'
})

const avatarLetter = computed(() => {
  const name = displayName.value
  return name ? name.charAt(0).toUpperCase() : 'U'
})

const goToEdit = () => {
  router.push('/document/edit')
}

const handleLogout = () => {
  userStore.logout()
  showSnackbar('已退出登录')
  router.push('/login')
}

// 全局 Snackbar
const snackbar = reactive({ show: false, text: '', color: 'success' })

const onSnackbar = (e) => {
  snackbar.text = e.detail.text
  snackbar.color = e.detail.color || 'success'
  snackbar.show = true
}

onMounted(() => {
  window.addEventListener('snackbar', onSnackbar)
  // 加载用户信息
  if (userStore.token && !userStore.userInfo) {
    userStore.getProfileInfo().catch(() => {})
  }
})

onUnmounted(() => {
  window.removeEventListener('snackbar', onSnackbar)
})
</script>

<style scoped>
/* 侧边栏样式微调 */
:deep(.v-navigation-drawer) {
  transition: width 0.2s ease;
}

:deep(.v-list-item--active) {
  background-color: rgba(255, 255, 255, 0.12) !important;
}
</style>
