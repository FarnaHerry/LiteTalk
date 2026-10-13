import { defineStore } from 'pinia'
import { ref } from 'vue'
import { goApp } from '../api/bridge'
import type { UserProfile } from '../api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserProfile | null>(null)
  const loggedIn = ref(false)

  async function login(email: string, password: string) {
    const r = await goApp().Login(email, password)
    user.value = r.user
    loggedIn.value = true
  }

  async function register(nickname: string, email: string, password: string, code: string) {
    const r = await goApp().Register(nickname, email, password, code)
    user.value = r.user
    loggedIn.value = true
  }

  async function sendEmailCode(email: string, purpose: 'register' | 'reset') {
    await goApp().SendEmailCode(email, purpose)
  }

  async function resetPassword(email: string, code: string, newPassword: string) {
    await goApp().ResetPassword(email, code, newPassword)
  }

  async function logout() {
    await goApp().Logout()
    user.value = null
    loggedIn.value = false
  }

  async function fetchMe() {
    user.value = await goApp().GetMe()
    loggedIn.value = true
  }

  return { user, loggedIn, login, register, sendEmailCode, resetPassword, logout, fetchMe }
})
