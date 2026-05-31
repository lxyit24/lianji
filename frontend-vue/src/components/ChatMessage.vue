<template>
  <div class="message" :class="message.role">
    <div class="avatar">
      {{ message.role === 'user' ? '👤' : '🤖' }}
    </div>
    <div class="content">
      <div class="role-name">
        {{ message.role === 'user' ? '你' : 'Lianji AI' }}
      </div>
      <div class="text" v-html="renderedContent"></div>
      <div v-if="message.streaming" class="cursor-blink">▌</div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  message: { type: Object, required: true }
})

const renderedContent = computed(() => {
  let text = props.message.content || ''
  // 简单的 markdown 渲染
  text = text.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
  text = text.replace(/`{3}(\w*)\n?([\s\S]*?)`{3}/g, '<pre><code>$2</code></pre>')
  text = text.replace(/`([^`]+)`/g, '<code>$1</code>')
  text = text.replace(/\n/g, '<br>')
  return text
})
</script>

<style scoped>
.message {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.avatar {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-size: 16px;
  flex-shrink: 0;
  background: #27272a;
}

.message.user .avatar {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
}

.content {
  flex: 1;
  min-width: 0;
}

.role-name {
  font-size: 12px;
  font-weight: 600;
  color: #71717a;
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.text {
  font-size: 14px;
  line-height: 1.6;
  color: #d4d4d8;
}

.text :deep(pre) {
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
  overflow-x: auto;
  font-size: 13px;
}

.text :deep(code) {
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  font-size: 13px;
  background: #27272a;
  padding: 1px 5px;
  border-radius: 4px;
  color: #e4e4e7;
}

.text :deep(pre code) {
  background: transparent;
  padding: 0;
}

.text :deep(strong) {
  color: #fafafa;
}

.cursor-blink {
  display: inline;
  color: #6366f1;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}
</style>
