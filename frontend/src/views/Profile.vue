<template>
  <div>
    <v-row justify="center">
      <v-col cols="12" md="8" lg="6">
        <!-- 个人信息卡片 -->
        <v-card elevation="2" rounded="lg" class="mb-6">
          <v-card-title class="d-flex align-center pa-6 pb-2">
            <v-icon color="primary" class="mr-2">mdi-account-circle</v-icon>
            <span class="text-h6 font-weight-bold">个人信息</span>
          </v-card-title>

          <v-divider class="my-2" />

          <v-card-text class="pa-6">
            <v-form ref="formRef" @submit.prevent="handleUpdate">
              <v-text-field
                v-model="form.username"
                label="用户名"
                prepend-inner-icon="mdi-account"
                variant="outlined"
                disabled
                density="comfortable"
                class="mb-3"
                hint="用户名不可修改"
              />
              <v-text-field
                v-model="form.nickname"
                label="昵称"
                prepend-inner-icon="mdi-account-edit"
                variant="outlined"
                density="comfortable"
                class="mb-3"
              />
              <v-text-field
                v-model="form.email"
                label="邮箱"
                prepend-inner-icon="mdi-email"
                variant="outlined"
                :rules="[v => !v || /.+@.+\..+/.test(v) || '邮箱格式不正确']"
                density="comfortable"
                class="mb-4"
              />
              <div class="d-flex justify-end">
                <v-btn color="primary" variant="elevated" :loading="loading" type="submit" rounded="lg">
                  保存修改
                </v-btn>
              </div>
            </v-form>
          </v-card-text>
        </v-card>

        <!-- 修改密码卡片 -->
        <v-card elevation="2" rounded="lg">
          <v-card-title class="d-flex align-center pa-6 pb-2">
            <v-icon color="warning" class="mr-2">mdi-lock-reset</v-icon>
            <span class="text-h6 font-weight-bold">修改密码</span>
          </v-card-title>

          <v-divider class="my-2" />

          <v-card-text class="pa-6">
            <v-form ref="passwordFormRef" @submit.prevent="handleChangePassword">
              <v-text-field
                v-model="passwordForm.oldPassword"
                label="原密码"
                prepend-inner-icon="mdi-lock"
                variant="outlined"
                :append-inner-icon="showOld ? 'mdi-eye' : 'mdi-eye-off'"
                :type="showOld ? 'text' : 'password'"
                @click:append-inner="showOld = !showOld"
                :rules="[v => !!v || '请输入原密码']"
                density="comfortable"
                class="mb-3"
              />
              <v-text-field
                v-model="passwordForm.newPassword"
                label="新密码"
                prepend-inner-icon="mdi-lock-plus"
                variant="outlined"
                :append-inner-icon="showNew ? 'mdi-eye' : 'mdi-eye-off'"
                :type="showNew ? 'text' : 'password'"
                @click:append-inner="showNew = !showNew"
                :rules="[
                  v => !!v || '请输入新密码',
                  v => (v && v.length >= 6) || '密码至少 6 个字符'
                ]"
                density="comfortable"
                class="mb-3"
              />
              <v-text-field
                v-model="passwordForm.confirmPassword"
                label="确认新密码"
                prepend-inner-icon="mdi-lock-check"
                variant="outlined"
                :append-inner-icon="showConfirm ? 'mdi-eye' : 'mdi-eye-off'"
                :type="showConfirm ? 'text' : 'password'"
                @click:append-inner="showConfirm = !showConfirm"
                :rules="[
                  v => !!v || '请确认新密码',
                  v => v === passwordForm.newPassword || '两次密码不一致'
                ]"
                density="comfortable"
                class="mb-4"
              />
              <div class="d-flex justify-end">
                <v-btn color="warning" variant="elevated" :loading="passwordLoading" type="submit" rounded="lg">
                  修改密码
                </v-btn>
              </div>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { updateProfile, changePassword } from '@/api'
import { showSnackbar } from '@/api/request'

const userStore = useUserStore()
const formRef = ref(null)
const passwordFormRef = ref(null)
const loading = ref(false)
const passwordLoading = ref(false)

const showOld = ref(false)
const showNew = ref(false)
const showConfirm = ref(false)

const form = reactive({
  username: '',
  nickname: '',
  email: ''
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const loadProfile = () => {
  const user = userStore.userInfo
  if (user) {
    form.username = user.username
    form.nickname = user.nickname || ''
    form.email = user.email || ''
  }
}

const handleUpdate = async () => {
  const { valid } = await formRef.value.validate()
  if (!valid) return

  loading.value = true
  try {
    await updateProfile({
      nickname: form.nickname,
      email: form.email
    })
    showSnackbar('更新成功')
    await userStore.getProfileInfo()
  } catch (error) {
    // handled
  } finally {
    loading.value = false
  }
}

const handleChangePassword = async () => {
  const { valid } = await passwordFormRef.value.validate()
  if (!valid) return

  passwordLoading.value = true
  try {
    await changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword
    })
    showSnackbar('密码修改成功')
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (error) {
    // handled
  } finally {
    passwordLoading.value = false
  }
}

onMounted(() => {
  loadProfile()
  // 如果没有用户信息，尝试加载
  if (!userStore.userInfo && userStore.token) {
    userStore.getProfileInfo().then(loadProfile).catch(() => {})
  }
})
</script>
