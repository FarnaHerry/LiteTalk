<script setup lang="ts">
// SettingsView 设置页（§9.3：连接/账号/常规/音频设备/语音/快捷键/关于）。
// 「最小化时休眠界面」→ GetHibernateOnMinimize / SetHibernateOnMinimize（app.go 已有）。
// 「连接」→ GetServerConfig / SetServerConfig / TestServerConnection；重启后生效。
import { onMounted, ref } from 'vue'
import { X, Server, Wifi, RefreshCw } from 'lucide-vue-next'
import { goApp, type AudioDeviceDTO } from '../api/bridge'
import type { ConnConfigView } from '../api/types'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/data'
import { useHibernateStore } from '../stores/hibernate'
import { HOTKEYS } from '../composables/useHotkeys'
import Toast from './Toast.vue'

const auth = useAuthStore()
const settings = useSettingsStore()
const hibernateStore = useHibernateStore()
const version = ref('dev')
const hibernate = ref(false)
const caps = ref<AudioDeviceDTO[]>([])
const plays = ref<AudioDeviceDTO[]>([])

// 连接配置
const conn = ref<ConnConfigView>({ base_url: '', ws_url: '', source: 'default' })
const serverInput = ref('')
const testing = ref(false)
const testResult = ref('')
const testOk = ref(false)
const toast = ref<InstanceType<typeof Toast>>()

const SOURCE_LABEL: Record<string, string> = { env: '环境变量', file: '配置文件', default: '默认' }

const emit = defineEmits<{ (e: 'close'): void }>()

onMounted(async () => {
  version.value = await goApp().GetVersion()
  hibernate.value = await goApp().GetHibernateOnMinimize()
  const devs = await goApp().ListAudioDevices()
  caps.value = devs.capture
  plays.value = devs.playback
  conn.value = await goApp().GetServerConfig()
  serverInput.value = conn.value.base_url
})

async function toggleHibernate(v: boolean) {
  hibernate.value = v
  hibernateStore.setEnabled(v)
  await goApp().SetHibernateOnMinimize(v)
}

// 设备下拉即时应用：切换即热重启设备流（编解码/混音状态保留）。
async function applyDevice() {
  await goApp().SetAudioDevices(settings.captureDevice, settings.playbackDevice)
}

// 测试连接：对当前输入框地址（可未保存）做 healthz 探活。
async function testConn() {
  testing.value = true
  testResult.value = ''
  try {
    const r = await goApp().TestServerConnection(serverInput.value.trim())
    testOk.value = r.ok
    testResult.value = r.ok ? `${r.latency_ms}ms · ${r.version || 'ok'}` : (r.error || '连接失败')
  } finally {
    testing.value = false
  }
}

// 保存：持久化后重启生效（运行中不热切换信令）。
async function saveConn() {
  try {
    await goApp().SetServerConfig(serverInput.value.trim())
    conn.value = await goApp().GetServerConfig()
    toast.value?.show('已保存，重启后生效')
  } catch {
    toast.value?.show('地址无效', true)
  }
}
</script>

<template>
  <div class="settings">
    <div class="s-head">
      <b>设置</b>
      <button class="icon-btn" title="关闭" @click="emit('close')"><X :size="18" /></button>
    </div>

    <div class="s-body">
      <section class="s-sec">
        <h3>连接</h3>
        <div class="s-row conn-row">
          <Server :size="16" class="row-icon" />
          <input
            v-model="serverInput"
            class="conn-input"
            :readonly="conn.source === 'env'"
            placeholder="https://your-domain.com"
          >
          <span class="badge">{{ SOURCE_LABEL[conn.source] }}</span>
        </div>
        <div v-if="conn.source === 'env'" class="conn-hint">由环境变量指定</div>
        <div v-else class="conn-hint">当前：{{ conn.ws_url }}</div>
        <div class="s-row">
          <button class="btn" :disabled="testing" @click="testConn">
            <RefreshCw :size="14" class="btn-icon" :class="{ spin: testing }" />测试连接
          </button>
          <button class="btn primary" @click="saveConn">保存</button>
          <span v-if="testResult" class="test-result" :class="{ ok: testOk }">
            <Wifi :size="13" class="btn-icon" />{{ testResult }}
          </span>
        </div>
      </section>

      <section class="s-sec">
        <h3>账号</h3>
        <div class="s-row">
          <div><div class="k">昵称</div><div class="v">{{ auth.user?.nickname ?? '—' }}</div></div>
          <div><div class="k">邮箱</div><div class="v">{{ auth.user?.email ?? '—' }}</div></div>
        </div>
      </section>

      <section class="s-sec">
        <h3>常规</h3>
        <div class="s-row">
          <div class="k">最小化时休眠界面（省内存）</div>
          <label class="switch">
            <input type="checkbox" :checked="hibernate" @change="toggleHibernate(($event.target as HTMLInputElement).checked)">
            <i />
          </label>
        </div>
        <div class="s-row">
          <div class="k">关闭到托盘</div>
          <span class="v">始终开启</span>
        </div>
      </section>

      <section class="s-sec">
        <h3>音频设备</h3>
        <div class="s-row">
          <div class="k">麦克风</div>
          <select v-model="settings.captureDevice" @change="applyDevice">
            <option v-for="d in caps" :key="d.id" :value="d.id">{{ d.name }}</option>
          </select>
        </div>
        <div class="s-row">
          <div class="k">扬声器</div>
          <select v-model="settings.playbackDevice" @change="applyDevice">
            <option v-for="d in plays" :key="d.id" :value="d.id">{{ d.name }}</option>
          </select>
        </div>
      </section>

      <section class="s-sec">
        <h3>语音</h3>
        <div class="s-row">
          <div class="k">降噪</div>
          <label class="switch">
            <input v-model="settings.denoise" type="checkbox"><i />
          </label>
        </div>
        <div class="s-row">
          <div class="k">输入增益</div>
          <input v-model.number="settings.gain" type="range" min="0" max="2" step="0.1">
          <span class="v">{{ settings.gain.toFixed(1) }}</span>
        </div>
      </section>

      <section class="s-sec">
        <h3>快捷键</h3>
        <div v-for="h in HOTKEYS" :key="h.id" class="s-row">
          <div class="k">{{ h.label }}</div><span class="v">{{ h.combo }}</span>
        </div>
      </section>

      <section class="s-sec">
        <h3>关于</h3>
        <div class="s-row"><div class="k">版本</div><span class="v">{{ version }}</span></div>
      </section>
    </div>
    <Toast ref="toast" />
  </div>
</template>

<style scoped>
.settings {
  position: absolute; inset: 0; z-index: 25;
  background: var(--surface);
  display: flex; flex-direction: column;
}
.s-head {
  height: 56px; flex: none;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 22px; border-bottom: 1px solid var(--line);
}
.s-head b { font-size: 15px; }
.icon-btn {
  width: 32px; height: 32px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; color: var(--ink-400);
  transition: background 0.15s, color 0.15s;
}
.icon-btn:hover { background: var(--brand-50); color: var(--ink-900); }
.s-body { flex: 1; overflow-y: auto; padding: 22px; max-width: 560px; }
.s-sec { margin-bottom: 26px; }
.s-sec h3 {
  font-size: 12px; font-weight: 700; letter-spacing: 1.5px;
  color: var(--ink-400); margin-bottom: 12px;
}
.s-row {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 14px; border-radius: 10px;
  background: var(--brand-50); margin-bottom: 8px;
}
.k { font-size: 13.5px; font-weight: 500; color: var(--ink-600); flex: 1; }
.v { font-size: 13px; color: var(--ink-400); }
select {
  height: 32px; border-radius: 8px; border: 1.5px solid var(--brand-200);
  padding: 0 8px; font-size: 13px; color: var(--ink-900);
  background: var(--surface); outline: none;
}
input[type="range"] { flex: 0 0 140px; accent-color: var(--brand-500); }
.switch { position: relative; width: 36px; height: 20px; flex: none; }
.switch input { display: none; }
.switch i {
  position: absolute; inset: 0;
  border-radius: 10px; background: var(--brand-200);
  transition: background 0.2s;
}
.switch i::after {
  content: ""; position: absolute; left: 2px; top: 2px;
  width: 16px; height: 16px; border-radius: 50%;
  background: #fff; box-shadow: var(--shadow-1);
  transition: transform 0.25s var(--ease);
}
.switch input:checked + i { background: var(--brand-500); }
.switch input:checked + i::after { transform: translateX(16px); }

/* 连接分组 */
.conn-row { gap: 10px; }
.row-icon { color: var(--ink-400); flex: none; }
.conn-input {
  flex: 1; height: 32px; border-radius: 8px;
  border: 1.5px solid var(--brand-200); padding: 0 10px;
  font-size: 13px; color: var(--ink-900); background: var(--surface); outline: none;
  transition: border-color 0.2s var(--ease), box-shadow 0.2s var(--ease);
}
.conn-input:focus { border-color: var(--brand-500); box-shadow: 0 0 0 3px rgba(79, 164, 243, 0.15); }
.conn-input[readonly] { background: var(--brand-50); color: var(--ink-400); }
.badge {
  flex: none; font-size: 11px; font-weight: 600;
  padding: 2px 8px; border-radius: 99px;
  background: rgba(79, 164, 243, 0.1); color: var(--brand-600);
}
.conn-hint { font-size: 11.5px; color: var(--ink-400); margin: 2px 0 8px 26px; }
.btn {
  display: inline-flex; align-items: center; gap: 6px;
  height: 32px; padding: 0 14px; border-radius: 8px;
  font-size: 13px; font-weight: 600; color: var(--ink-600);
  background: var(--brand-50); transition: background 0.2s, color 0.2s, transform 0.1s var(--ease);
}
.btn:hover:not(:disabled) { background: rgba(79, 164, 243, 0.15); color: var(--brand-700); }
.btn:active:not(:disabled) { transform: scale(0.96); }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn.primary { background: var(--brand-500); color: #fff; }
.btn.primary:hover { background: var(--brand-600); color: #fff; }
.btn-icon { flex: none; }
.spin { animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.test-result {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 12px; color: var(--red); margin-left: auto;
}
.test-result.ok { color: var(--green); }
</style>
