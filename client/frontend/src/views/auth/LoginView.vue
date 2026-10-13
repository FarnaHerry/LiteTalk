<script setup lang="ts">
// LoginView 登录页（对齐 Auth.html：分段切换登录/注册、记住我、忘记密码）。
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Eye, EyeOff, LockKeyhole, Mail, Check } from 'lucide-vue-next'
import AuthLayout from './AuthLayout.vue'
import Toast from '../../components/Toast.vue'
import { useAuthStore } from '../../stores/auth'
import './auth-forms.css'

const router = useRouter()
const auth = useAuthStore()
const toast = ref<InstanceType<typeof Toast>>()

const form = reactive({ email: '', password: '', remember: true })
const invalid = reactive({ email: false, password: false })
const showPwd = ref(false)
const loading = ref(false)

const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function validate(): boolean {
  invalid.email = !emailRe.test(form.email.trim())
  invalid.password = form.password.length < 8
  return !invalid.email && !invalid.password
}

async function submit() {
  if (!validate()) return
  loading.value = true
  try {
    await auth.login(form.email.trim(), form.password)
    toast.value?.show('登录成功')
    router.push('/')
  } catch (e) {
    toast.value?.show(String(e), true)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthLayout>
    <div class="seg">
      <div class="seg-thumb" />
      <button class="on">登录</button>
      <button @click="router.push('/register')">注册</button>
    </div>

    <form @submit.prevent="submit" novalidate>
      <div class="field" :class="{ invalid: invalid.email }">
        <label>邮箱</label>
        <div class="input-wrap">
          <Mail :size="17" />
          <input v-model="form.email" type="email" placeholder="you@example.com"
                 autocomplete="email" @input="invalid.email = false">
        </div>
        <p class="field-err">请输入有效的邮箱地址</p>
      </div>

      <div class="field" :class="{ invalid: invalid.password }">
        <label>密码</label>
        <div class="input-wrap">
          <LockKeyhole :size="17" />
          <input v-model="form.password" :type="showPwd ? 'text' : 'password'"
                 placeholder="请输入密码" autocomplete="current-password"
                 @input="invalid.password = false">
          <button type="button" class="eye" tabindex="-1" @click="showPwd = !showPwd">
            <EyeOff v-if="showPwd" :size="16" />
            <Eye v-else :size="16" />
          </button>
        </div>
        <p class="field-err">密码至少 8 位</p>
      </div>

      <div class="opt-row">
        <label class="check">
          <input v-model="form.remember" type="checkbox">
          <span class="box"><Check :size="11" /></span>记住我
        </label>
        <button type="button" class="link" @click="router.push('/forgot')">忘记密码？</button>
      </div>

      <button type="submit" class="btn-primary" :class="{ loading }" :disabled="loading">
        <span class="spinner" />
        <span class="btn-text">登录</span>
      </button>
    </form>

    <Toast ref="toast" />
  </AuthLayout>
</template>
