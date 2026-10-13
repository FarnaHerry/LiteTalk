// hibernate 休眠状态机：监听 sys:hibernate / sys:wake 事件，
// 进入休眠态时卸载业务组件、清理事件订阅与定时器；唤醒时恢复。
// 状态机迁移函数为纯函数，可单测。
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { eventsOn } from '../api/bridge'

// HibernateEvent 状态机事件。
export type HibernateEvent = 'hibernate' | 'wake' | 'toggle-setting'

// hibernateReduce 纯状态机：给定当前状态与事件，返回下一状态。
// 规则：hibernate → 休眠；wake → 唤醒；toggle-setting 不改变状态（仅设置项）。
export function hibernateReduce(state: boolean, ev: HibernateEvent): boolean {
  switch (ev) {
    case 'hibernate':
      return true
    case 'wake':
      return false
    default:
      return state
  }
}

export const useHibernateStore = defineStore('hibernate', () => {
  const hibernated = ref(false)
  const enabled = ref(false) // 设置项「最小化时休眠界面」

  // onStateChange 由 MainView 注入：休眠/唤醒时卸载/恢复业务组件。
  let onStateChange: ((hibernated: boolean) => void) | null = null

  function bind(fn: (hibernated: boolean) => void) {
    onStateChange = fn
  }

  function apply(ev: HibernateEvent) {
    const next = hibernateReduce(hibernated.value, ev)
    if (next === hibernated.value) return
    hibernated.value = next
    onStateChange?.(next)
  }

  // init 订阅 Go 侧窗口事件（sys:hibernate / sys:wake）。
  function init() {
    eventsOn('sys:hibernate', () => apply('hibernate'))
    eventsOn('sys:wake', () => apply('wake'))
    // 窗口重新获得焦点视为唤醒兜底（如从任务栏还原）。
    window.addEventListener('focus', () => apply('wake'))
  }

  function setEnabled(v: boolean) {
    enabled.value = v
    apply('toggle-setting')
  }

  return { hibernated, enabled, init, bind, apply, setEnabled }
})
