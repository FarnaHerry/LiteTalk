import { createRouter, createWebHashHistory } from 'vue-router'
import MainView from '../views/MainView.vue'
import LoginView from '../views/auth/LoginView.vue'
import RegisterView from '../views/auth/RegisterView.vue'
import ForgotView from '../views/auth/ForgotView.vue'
import { useAuthStore } from '../stores/auth'

// Wails 嵌入场景使用 hash 路由，避免刷新 404。
const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', name: 'main', component: MainView },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/register', name: 'register', component: RegisterView },
    { path: '/forgot', name: 'forgot', component: ForgotView },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// 鉴权守卫：未登录只能访问登录/注册/找回；已登录访问鉴权页时回主界面。
router.beforeEach((to) => {
  const auth = useAuthStore()
  const authPages = ['login', 'register', 'forgot']
  if (!auth.loggedIn && !authPages.includes(to.name as string)) {
    return { name: 'login' }
  }
  if (auth.loggedIn && authPages.includes(to.name as string)) {
    return { name: 'main' }
  }
})

export default router
