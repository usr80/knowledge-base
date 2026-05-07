<template>
  <v-app>
    <v-main>
      <div class="register-wrapper">
        <v-container>
          <v-row justify="center" align="center" style="height: 100vh;">
            <v-col cols="12" sm="8" md="5" lg="4">
              <v-card elevation="12" rounded="xl" class="register-card">
                <div class="card-header">
                  <v-icon size="56" color="white">mdi-account-plus</v-icon>
                  <h2 class="text-h5 text-white mt-3 font-weight-bold">注册账号</h2>
                </div>

                <v-card-text class="pa-8">
                  <v-form ref="formRef" @submit.prevent="handleRegister">
                    <v-text-field
                      v-model="form.username"
                      label="用户名"
                      prepend-inner-icon="mdi-account"
                      variant="outlined"
                      :rules="[
                        v => !!v || '请输入用户名',
                        v => (v && v.length >= 3) || '用户名至少 3 个字符'
                      ]"
                      density="comfortable"
                      class="mb-2"
                    />
                    <v-text-field
                      v-model="form.email"
                      label="邮箱（可选）"
                      prepend-inner-icon="mdi-email"
                      variant="outlined"
                      :rules="[v => !v || /.+@.+\..+/.test(v) || '邮箱格式不正确']"
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
                      :rules="[
                        v => !!v || '请输入密码',
                        v => (v && v.length >= 6) || '密码至少 6 个字符'
                      ]"
                      density="comfortable"
                      class="mb-2"
                    />
                    <v-text-field
                      v-model="form.confirmPassword"
                      label="确认密码"
                      prepend-inner-icon="mdi-lock-check"
                      variant="outlined"
                      :append-inner-icon="showConfirm ? 'mdi-eye' : 'mdi-eye-off'"
                      :type="showConfirm ? 'text' : 'password'"
                      @click:append-inner="showConfirm = !showConfirm"
                      :rules="[
                        v => !!v || '请确认密码',
                        v => v === form.password || '两次密码不一致'
                      ]"
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
                      注册
                    </v-btn>
                  </v-form>
                </v-card-text>

                <v-card-actions class="justify-center pb-6">
                  <span class="text-body-2 text-grey">已有账号？</span>
                  <router-link to="/login" class="text-primary text-body-2 font-weight-bold ml-1" style="text-decoration: none;">
                    立即登录
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
import { register } from '@/api'
import { showSnackbar } from '@/api/request'

const router = useRouter()
const formRef = ref(null)
const loading = ref(false)
const showPassword = ref(false)
const showConfirm = ref(false)

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const snackbar = reactive({ show: false, text: '', color: 'success' })
const notify = (text, color = 'success') => {
  snackbar.text = text
  snackbar.color = color
  snackbar.show = true
}

const handleRegister = async () => {
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    await register({
      username: form.username,
      email: form.email,
      password: form.password
    })
    notify('注册成功，请登录')
    router.push('/login')
  } catch (error) {
    // request.js 已处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-wrapper {
  height: 100vh;
  background: linear-gradient(135deg, #00897B 0%, #00695C 50%, #004D40 100%);
}

.register-card {
  overflow: hidden;
}

.card-header {
  background: linear-gradient(135deg, #26A69A, #00897B);
  text-align: center;
  padding: 32px 24px 24px;
}
</style>
