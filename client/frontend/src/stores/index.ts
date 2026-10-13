import { defineStore } from 'pinia'
import { ref } from 'vue'

// 示例 store：应用设置。
// 后续里程碑将拆分为
// auth / servers / channels / voice / friends / chat / settings。
export const useSettingsStore = defineStore('settings', () => {
  // 「最小化时休眠界面（省内存）」，默认关闭
  const hibernateOnMinimize = ref(false)

  function setHibernateOnMinimize(enabled: boolean) {
    hibernateOnMinimize.value = enabled
  }

  return { hibernateOnMinimize, setHibernateOnMinimize }
})
