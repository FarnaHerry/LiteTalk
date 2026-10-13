// useWindowClose 关窗（问题 4d 修订：点击即关，不做任何延迟等待）。
// 关闭行为（托盘/退出）由 Go 侧 onBeforeClose 兜底链决定，前端只负责立即执行。
import { goApp } from '../api/bridge'

export async function closeWithHint(): Promise<void> {
  await goApp().WindowClose()
}
