<script setup lang="ts">
// ChannelList 频道栏（对齐 Main.html：分组折叠、文字/语音频道、人数、锁标识）。
import { computed, ref } from 'vue'
import {
  ChevronDown, Eye, Hash, Lock, Mic, Volume2, Headphones, HeadphoneOff,
  MicOff, Settings,
} from 'lucide-vue-next'
import { useServersStore } from '../stores/data'
import { useVoiceStore } from '../stores/voice'
import { useAuthStore } from '../stores/auth'

const servers = useServersStore()
const voice = useVoiceStore()
const auth = useAuthStore()

const collapsed = ref<Record<string, boolean>>({})

const groups = computed(() => {
  const map = new Map<string, typeof servers.currentServer.channels>()
  for (const ch of servers.currentServer.channels) {
    const list = map.get(ch.group) ?? []
    list.push(ch)
    map.set(ch.group, list)
  }
  return [...map.entries()]
})

function toggleGroup(name: string) {
  collapsed.value[name] = !collapsed.value[name]
}

async function pick(id: string, type: string) {
  servers.selectChannel(id)
  if (type === 'voice') {
    await voice.join(id)
  }
}
</script>

<template>
  <aside class="channels">
    <div class="server-head">
      <span class="name">{{ servers.currentServer.name }}</span>
      <ChevronDown :size="16" />
    </div>

    <div class="chan-scroll">
      <div v-for="[name, list] in groups" :key="name" class="cat"
           :class="{ closed: collapsed[name] }">
        <div class="cat-head" @click="toggleGroup(name)">
          <ChevronDown :size="13" />{{ name }}
        </div>
        <div class="cat-items">
          <div v-for="ch in list" :key="ch.id" class="chan"
               :class="{ active: ch.id === servers.currentChannelId, 'in-voice': ch.id === voice.channelId }"
               @click="pick(ch.id, ch.type)">
            <component :is="ch.type === 'text' ? Hash : (ch.locked ? Lock : (ch.id === 'ch-audience' ? Eye : Volume2))"
                       :size="16" />
            <span class="cname">{{ ch.name }}</span>
            <span v-if="ch.members" class="count"><Mic :size="12" />{{ ch.members }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户面板（双关态：状态点绿→红） -->
    <div class="me">
      <span class="me-dot" :class="{ dual: voice.dualOff }" />
      <div class="me-info">
        <div class="n">{{ auth.user?.nickname ?? '未登录' }}</div>
        <div class="t">#{{ auth.user?.tag ?? '0000' }} · 在线</div>
      </div>
      <div class="me-acts">
        <button class="icon-btn" :class="{ danger: voice.muted }"
                :title="voice.muted ? '取消静音' : '静音'"
                @click="voice.setMute(!voice.muted)">
          <MicOff v-if="voice.muted" :size="17" />
          <Mic v-else :size="17" />
        </button>
        <button class="icon-btn" :class="{ danger: voice.deafened }"
                :title="voice.deafened ? '取消耳聋' : '耳聋'"
                @click="voice.setDeafen(!voice.deafened)">
          <HeadphoneOff v-if="voice.deafened" :size="17" />
          <Headphones v-else :size="17" />
        </button>
        <button class="icon-btn" title="设置"><Settings :size="17" /></button>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.channels {
  width: 248px; flex: none;
  display: flex; flex-direction: column;
  background: var(--surface);
  border-right: 1px solid var(--line);
  min-height: 0;
}
.server-head {
  height: 52px; flex: none;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 16px; border-bottom: 1px solid var(--line);
  cursor: pointer; transition: background 0.15s;
}
.server-head:hover { background: var(--brand-50); }
.server-head .name { font-size: 15px; font-weight: 700; letter-spacing: 0.2px; }
.server-head :deep(svg) { color: var(--ink-400); }
.chan-scroll { flex: 1; overflow-y: auto; padding: 10px 10px 12px; }
.cat { margin-bottom: 14px; }
.cat-head {
  display: flex; align-items: center; gap: 3px;
  padding: 4px 6px; font-size: 11.5px; font-weight: 600;
  color: var(--ink-400); letter-spacing: 0.6px;
  cursor: pointer; user-select: none; border-radius: 7px;
  transition: color 0.15s;
}
.cat-head:hover { color: var(--ink-900); }
.cat-head :deep(svg) { transition: transform 0.25s var(--ease); }
.cat.closed .cat-head :deep(svg) { transform: rotate(-90deg); }
.cat.closed .cat-items { display: none; }
.chan {
  display: flex; align-items: center; gap: 8px;
  height: 34px; padding: 0 9px; margin: 1px 0;
  border-radius: 8px; color: var(--ink-600);
  cursor: pointer; transition: background 0.15s, color 0.15s;
}
.chan :deep(svg) { color: var(--ink-400); flex: none; transition: color 0.15s; }
.chan .cname {
  flex: 1; font-size: 13.5px; font-weight: 500;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.chan .count {
  font-size: 11.5px; color: var(--ink-400);
  display: flex; align-items: center; gap: 3px;
}
.chan:hover { background: var(--brand-50); color: var(--ink-900); }
.chan.active { background: rgba(79, 164, 243, 0.12); color: var(--brand-700); }
.chan.active :deep(svg) { color: var(--brand-600); }
.me {
  margin: 0 10px 10px; padding: 10px; border-radius: 12px;
  background: var(--brand-50);
  display: flex; align-items: center; gap: 10px;
}
.me-dot {
  position: relative; width: 9px; height: 9px; flex: none; margin-left: 2px;
}
.me-dot::before {
  content: ""; position: absolute; inset: 0;
  border-radius: 50%; background: var(--green);
  transition: background 0.2s;
}
.me-dot::after {
  content: ""; position: absolute; inset: -3px;
  border-radius: 50%; border: 1.5px solid var(--green);
  opacity: 0.4; animation: pulse 2.4s var(--ease) infinite;
  transition: border-color 0.2s;
}
/* 双关态：状态点绿→红 */
.me-dot.dual::before { background: var(--red); }
.me-dot.dual::after { border-color: var(--red); }
@keyframes pulse {
  0%, 100% { transform: scale(0.8); opacity: 0.45; }
  50% { transform: scale(1.15); opacity: 0.1; }
}
.me-info { flex: 1; min-width: 0; }
.me-info .n {
  font-size: 13.5px; font-weight: 600;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.me-info .t { font-size: 11px; color: var(--ink-400); }
.me-acts { display: flex; gap: 2px; }
.icon-btn {
  width: 32px; height: 32px; flex: none;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; color: var(--ink-600);
  transition: background 0.15s, color 0.15s, transform 0.1s var(--ease);
}
.icon-btn:hover { background: rgba(79, 164, 243, 0.12); color: var(--brand-700); }
.icon-btn:active { transform: scale(0.92); }
.icon-btn.danger { background: rgba(232, 91, 91, 0.12); color: var(--red); }
.icon-btn.danger:hover { background: rgba(232, 91, 91, 0.2); }
</style>
