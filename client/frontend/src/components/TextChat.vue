<script setup lang="ts">
// TextChat 文字频道视图（对齐 Main.html：消息列表 + 输入发送）。
import { ref } from 'vue'
import { Hash, Send } from 'lucide-vue-next'
import { useChatStore } from '../stores/chat'
import { useServersStore } from '../stores/data'

const chat = useChatStore()
const servers = useServersStore()
const text = ref('')

async function send() {
  const v = text.value.trim()
  if (!v) return
  try {
    await chat.send(servers.currentChannelId, v)
    // 本地回显（服务端广播 chat.message 到达后由 store 去重；M3 服务端就绪后移除）。
    chat.messages.push({
      id: String(Date.now()),
      channel_id: servers.currentChannelId,
      user_id: 'me',
      nickname: '我',
      content: v,
      ts: String(Date.now()),
    })
    text.value = ''
  } catch { /* 发送失败静默，服务端 error 消息另行提示 */ }
}
</script>

<template>
  <div class="chat">
    <div class="chat-head">
      <Hash :size="17" />
      <b>{{ servers.currentChannel?.name ?? '闲聊广场' }}</b>
    </div>

    <div class="chat-list">
      <div v-for="m in chat.messages" :key="m.id" class="msg">
        <div class="m-head">
          <span class="m-name">{{ m.nickname }}</span>
          <span class="m-time">{{ m.ts }}</span>
        </div>
        <p class="m-text">{{ m.content }}</p>
      </div>
    </div>

    <div class="composer">
      <div class="c-input">
        <input v-model="text"
               :placeholder="`发送到 #${servers.currentChannel?.name ?? ''}`"
               maxlength="2000"
               @keydown.enter.prevent="send">
      </div>
      <button class="send-btn" title="发送" @click="send"><Send :size="17" /></button>
    </div>
  </div>
</template>

<style scoped>
.chat {
  flex: 1; min-height: 0;
  display: flex; flex-direction: column;
  background: var(--brand-50);
}
.chat-head {
  flex: none; padding: 14px 22px 0;
  display: flex; align-items: center; gap: 8px;
  color: var(--ink-600);
}
.chat-head :deep(svg) { color: var(--ink-400); }
.chat-head b { font-size: 15px; color: var(--ink-900); }
.chat-list {
  flex: 1; overflow-y: auto;
  padding: 14px 22px;
  display: flex; flex-direction: column; gap: 14px;
}
.msg .m-head {
  display: flex; align-items: baseline; gap: 8px;
  margin-bottom: 3px;
}
.msg .m-name { font-size: 13.5px; font-weight: 600; }
.msg .m-time { font-size: 11px; color: var(--ink-300); }
.msg .m-text {
  font-size: 14px; color: var(--ink-600);
  line-height: 1.65; text-wrap: pretty;
}
.composer { flex: none; padding: 12px 22px 18px; display: flex; gap: 10px; }
.composer .c-input {
  flex: 1; display: flex; align-items: center;
  background: var(--surface);
  border: 1.5px solid var(--brand-200); border-radius: 12px;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.composer .c-input:focus-within {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px rgba(79, 164, 243, 0.15);
}
.composer input {
  flex: 1; height: 42px; padding: 0 14px;
  border: none; outline: none; font-size: 14px;
  background: transparent; color: var(--ink-900);
}
.composer input::placeholder { color: var(--ink-300); }
.send-btn {
  width: 42px; height: 42px; flex: none;
  border-radius: 12px; background: var(--brand-500); color: #fff;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.2s, transform 0.1s var(--ease);
  box-shadow: 0 4px 12px rgba(79, 164, 243, 0.3);
}
.send-btn:hover { background: var(--brand-600); }
.send-btn:active { transform: scale(0.93); }
</style>
