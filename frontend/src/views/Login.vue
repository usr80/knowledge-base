<template>
  <v-app>
    <v-main>
      <div class="login-wrapper">
        <v-container>
          <v-row justify="center" align="center" style="height: 100vh;">
            <v-col cols="12" sm="8" md="5" lg="4">
              <v-card elevation="12" rounded="xl" class="login-card">
                <!-- 顶部装饰 -->
                <div class="card-header">
                  <v-icon size="56" color="white">mdi-book-open-page-variant</v-icon>
                  <h2 class="text-h5 text-white mt-3 font-weight-bold">知识库</h2>
                  <p class="text-white text-body-2 mt-1" style="opacity: 0.8;">登录你的账号</p>
                </div>

                <v-card-text class="pa-8">
                  <v-form ref="formRef" @submit.prevent="handleLogin">
                    <v-text-field
                      v-model="form.username"
                      label="用户名"
                      prepend-inner-icon="mdi-account"
                      variant="outlined"
                      :rules="[v => !!v || '请输入用户名']"
                      density="comfortable"
                      class="mb-2"
                    />
                    <v-text-field
                      v-model="form.password"
                      label="密码"
                      prepend-inner-icon="mdi-lock"
                      variant="outlined"
                      :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                      :type="showPassword ? 'text' : 'password'"
                      @click:append-inner="showPassword = !showPassword"
                      :rules="[v => !!v || '请输入密码']"
                      density="comfortable"
                      class="mb-4"
                    />
                    <v-btn
                      type="submit"
                      color="primary"
                      size="large"
                      block
                      :loading="loading"
                      rounded="lg"
                    >
                      登录
                    </v-btn>
                  </v-form>
                </v-card-text>

                <v-card-actions class="justify-center pb-6">
                  <span class="text-body-2 text-grey">没有账号？</span>
                  <router-link to="/register" class="text-primary text-body-2 font-weight-bold ml-1" style="text-decoration: none;">
                    立即注册
                  </router-link>
                </v-card-actions>
              </v-card>
            </v-col>
          </v-row>
        </v-container>
      </div>
    </v-main>

    <v-snackbar v-model="snackbar.show" :color="snackbar.color" :timeout="3000" location="top">
      {{ snackbar.text }}
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { showSnackbar } from '@/api/request'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref(null)
const loading = ref(false)
const showPassword = ref(false)

const form = reactive({ username: '', password: '' })

const snackbar = reactive({ show: false, text: '', color: 'success' })
const notify = (text, color = 'success') => {
  snackbar.text = text
  snackbar.color = color
  snackbar.show = true
}

const handleLogin = async () => {
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    await userStore.login(form)
    notify('登录成功')
    router.push('/')
  } catch (error) {
    // request.js 已处理提示
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrapper {
  height: 100vh;
  background: linear-gradient(135deg, #1565C0 0%, #0D47A1 50%, #1A237E 100%);
}

.login-card {
  overflow: hidden;
}

.card-header {
  background: linear-gradient(135deg, #1976D2, #1565C0);
  text-align: center;
  padding: 32px 24px 24px;
}
</style>
