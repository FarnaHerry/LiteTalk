<script setup lang="ts">
// RegisterView 注册页（对齐 Auth.html：昵称/邮箱/验证码倒计时/密码强度）。
import { computed, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Check, Eye, EyeOff, LockKeyhole, Mail, ShieldCheck, UserRound } from 'lucide-vue-next'
import AuthLayout from './AuthLayout.vue'
import Toast from '../../components/Toast.vue'
import { useAuthStore } from '../../stores/auth'
import './auth-forms.css'

const router = useRouter()
const auth = useAuthStore()
const toast = ref<InstanceType<typeof Toast>>()

const form = reactive({ nickname: '', email: '', code: '', password: '' })
const invalid = reactive({ name: false, email: false, code: false, password: false })
const showPwd = ref(false)
const loading = ref(false)

const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const cd = ref(0)
let cdTimer: ReturnType<typeof setInterval> | null = null

const strengthClass = computed(() => {
  const v = form.password
  if (!v) return ''
  let s = 0
  if (v.length >= 8) s++
  if (/\d/.test(v) && /[a-zA-Z]/.test(v)) s++
  if (v.length >= 12 && /[^a-zA-Z0-9]/.test(v)) s++
  return s ? `s${s}` : ''
})

function validate(): boolean {
  invalid.name = form.nickname.trim().length < 2 || form.nickname.trim().length > 16
  invalid.email = !emailRe.test(form.email.trim())
  invalid.code = !/^\d{6}$/.test(form.code)
  invalid.password = form.password.length < 8
  return !(invalid.name || invalid.email || invalid.code || invalid.password)
}

async function sendCode() {
  invalid.email = !emailRe.test(form.email.trim())
  if (invalid.email) return
  try {
    await auth.sendEmailCode(form.email.trim(), 'register')
    toast.value?.show('验证码已发送')
    cd.value = 60
    cdTimer = setInterval(() => {
      cd.value--
      if (cd.value <= 0 && cdTimer) { clearInterval(cdTimer); cdTimer = null }
    }, 1000)
  } catch (e) {
    toast.value?.show(String(e), true)
  }
}

async function submit() {
  if (!validate()) return
  loading.value = true
  try {
    await auth.register(form.nickname.trim(), form.email.trim(), form.password, form.code)
    toast.value?.show('注册成功')
    router.push('/')
  } catch (e) {
    toast.value?.show(String(e), true)
  } finally {
    loading.value = false
  }
}

onUnmounted(() => { if (cdTimer) clearInterval(cdTimer) })
</script>

<template>
  <AuthLayout>
    <div class="seg reg">
      <div class="seg-thumb" />
      <button @click="router.push('/login')">登录</button>
      <button class="on">注册</button>
    </div>

    <form @submit.prevent="submit" novalidate>
      <div class="field" :class="{ invalid: invalid.name }">
        <label>昵称</label>
        <div class="input-wrap">
          <UserRound :size="17" />
          <input v-model="form.nickname" type="text" placeholder="2–16 个字符"
                 maxlength="16" @input="invalid.name = false">
        </div>
        <p class="field-err">昵称需为 2–16 个字符</p>
      </div>

      <div class="field" :class="{ invalid: invalid.email }">
        <label>邮箱</label>
        <div class="code-row">
          <div class="input-wrap">
            <Mail :size="17" />
            <input v-model="form.email" type="email" placeholder="you@example.com"
                   autocomplete="email" @input="invalid.email = false">
          </div>
          <button type="button" class="code-btn" :disabled="cd > 0" @click="sendCode">
            {{ cd > 0 ? `${cd}s` : '发送验证码' }}
          </button>
        </div>
        <p class="field-err">请输入有效的邮箱地址</p>
      </div>

      <div class="field" :class="{ invalid: invalid.code }">
        <label>邮箱验证码</label>
        <div class="input-wrap">
          <ShieldCheck :size="17" />
          <input v-model="form.code" type="text" placeholder="6 位数字"
                 maxlength="6" inputmode="numeric" @input="invalid.code = false">
        </div>
        <p class="field-err">请输入 6 位验证码</p>
      </div>

      <div class="field" :class="{ invalid: invalid.password }">
        <label>设置密码</label>
        <div class="input-wrap">
          <LockKeyhole :size="17" />
          <input v-model="form.password" :type="showPwd ? 'text' : 'password'"
                 placeholder="至少 8 位，建议混合字母与数字"
                 autocomplete="new-password" @input="invalid.password = false">
          <button type="button" class="eye" tabindex="-1" @click="showPwd = !showPwd">
            <EyeOff v-if="showPwd" :size="16" />
            <Eye v-else :size="16" />
          </button>
        </div>
        <div class="strength" :class="strengthClass"><i /><i /><i /></div>
        <p class="field-err">密码至少 8 位</p>
      </div>

      <button type="submit" class="btn-primary" :class="{ loading }"
              :disabled="loading" style="margin-top: 6px">
        <span class="spinner" />
        <span class="btn-text">创建账号</span>
      </button>
    </form>

    <Toast ref="toast" />
  </AuthLayout>
</template>
