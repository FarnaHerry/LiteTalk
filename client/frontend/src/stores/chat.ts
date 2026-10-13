import { defineStore } from 'pinia'
import { ref } from 'vue'
import { eventsOff, eventsOn, goApp } from '../api/bridge'
import type { ChatMessage } from '../api/types'

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])

  async function send(channelId: string, content: string) {
    await goApp().SendChat(channelId, content)
  }

  function bindEvents() {
    eventsOn('chat:message', (m) => {
      messages.value.push(m as ChatMessage)
    })
  }

  function clear() {
    messages.value = []
  }

  // unbindEvents 休眠态清理：解除事件订阅。
  function unbindEvents() {
    eventsOff('chat:message')
  }

  return { messages, send, bindEvents, unbindEvents, clear }
})
