// useWindowClose 关窗提示（问题 4d）：首次关窗先提示「已最小化到托盘」，
// 避免用户误以为已退出；localStorage 记一次，之后直接关。
// 提示后短暂延迟再真正关闭，保证 Toast 可见。
import { goApp } from '../api/bridge'

const HINT_KEY = 'litetalk.tray-hint-shown'

export interface CloseHintHost {
  show(msg: string, err?: boolean): void
}

// closeWithHint 执行关窗：托盘模式且首次 → 提示并稍延迟后关；其余直接关。
// 关闭行为（托盘/退出）由 Go 侧 onBeforeClose 兜底链决定，前端只管提示。
export async function closeWithHint(toast?: CloseHintHost): Promise<void> {
  const trayMode = await goApp().GetCloseToTray().catch(() => true)
  const firstTime = !localStorage.getItem(HINT_KEY)
  if (trayMode && firstTime && toast) {
    localStorage.setItem(HINT_KEY, '1')
    toast.show('已最小化到托盘')
    await new Promise((r) => setTimeout(r, 900))
  }
  await goApp().WindowClose()
}
