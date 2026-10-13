import { defineStore } from 'pinia'
import { ref } from 'vue'
import { eventsOn, goApp } from '../api/bridge'
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

  return { messages, send, bindEvents, clear }
})
