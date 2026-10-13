<script setup lang="ts">
// SettingsView 设置页（§9.3：账号/常规/音频设备/语音/快捷键/关于）。
// 「最小化时休眠界面」→ GetHibernateOnMinimize / SetHibernateOnMinimize（app.go 已有）。
import { onMounted, ref } from 'vue'
import { X } from 'lucide-vue-next'
import { goApp, type AudioDeviceDTO } from '../api/bridge'
import { useAuthStore } from '../stores/auth'
import { useSettingsStore } from '../stores/data'
import { useHibernateStore } from '../stores/hibernate'
import { HOTKEYS } from '../composables/useHotkeys'

const auth = useAuthStore()
const settings = useSettingsStore()
const hibernateStore = useHibernateStore()
const version = ref('dev')
const hibernate = ref(false)
const caps = ref<AudioDeviceDTO[]>([])
const plays = ref<AudioDeviceDTO[]>([])

const emit = defineEmits<{ (e: 'close'): void }>()

onMounted(async () => {
  version.value = await goApp().GetVersion()
  hibernate.value = await goApp().GetHibernateOnMinimize()
  const devs = await goApp().ListAudioDevices()
  caps.value = devs.capture
  plays.value = devs.playback
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
</script>

<template>
  <div class="settings">
    <div class="s-head">
      <b>设置</b>
      <button class="icon-btn" title="关闭" @click="emit('close')"><X :size="18" /></button>
    </div>

    <div class="s-body">
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
</style>
