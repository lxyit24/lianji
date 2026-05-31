import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const isGenerating = ref(false)

  function addMessage(msg) {
    messages.value.push({
      role: msg.role,
      content: msg.content || '',
      streaming: msg.streaming || false,
      timestamp: Date.now()
    })
  }

  function updateLastMessage(content, streaming = true) {
    if (messages.value.length > 0) {
      const last = messages.value[messages.value.length - 1]
      last.content = content
      last.streaming = streaming
    }
  }

  function clearMessages() {
    messages.value = []
  }

  return {
    messages,
    isGenerating,
    addMessage,
    updateLastMessage,
    clearMessages
  }
})
