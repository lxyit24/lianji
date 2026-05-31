<template>
  <div class="chat-panel">
    <!-- 空状态 -->
    <div v-if="messages.length === 0" class="empty-state">
      <div class="empty-icon">🏗️</div>
      <h2>Lianji AI 建站</h2>
      <p>描述你想要的网站，AI 帮你一键生成</p>
      <div class="examples">
        <button
          v-for="ex in examples"
          :key="ex"
          class="example-btn"
          @click="$emit('send', ex)"
        >{{ ex }}</button>
      </div>
    </div>

    <!-- 消息列表 -->
    <div v-else class="messages" ref="messagesEl">
      <ChatMessage
        v-for="(msg, i) in messages"
        :key="i"
        :message="msg"
      />
      <!-- 加载指示器 -->
      <div v-if="loading" class="typing-indicator">
        <span></span><span></span><span></span>
      </div>
    </div>

    <!-- 输入框 -->
    <div class="input-area">
      <div class="input-wrapper">
        <textarea
          ref="inputEl"
          v-model="input"
          class="chat-input"
          placeholder="描述你想要的网站..."
          rows="1"
          @keydown.enter.exact.prevent="handleSend"
          @input="autoResize"
          :disabled="loading"
        ></textarea>
        <button
          v-if="!loading"
          class="send-btn"
          @click="handleSend"
          :disabled="!input.trim()"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 2L11 13"/><path d="M22 2l-7 20-4-9-9-4 20-7z"/>
          </svg>
        </button>
        <button v-else class="stop-btn" @click="$emit('stop')">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
            <rect x="6" y="6" width="12" height="12" rx="2"/>
          </svg>
        </button>
      </div>
      <p class="input-hint">Enter 发送 · 描述越详细效果越好</p>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, watch } from 'vue'
import ChatMessage from './ChatMessage.vue'

const props = defineProps({
  messages: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false }
})

const emit = defineEmits(['send', 'stop'])

const input = ref('')
const inputEl = ref(null)
const messagesEl = ref(null)

const examples = [
  '创建一个现代化的企业官网，包含首页、关于我们、产品展示、联系我们',
  '制作一个个人博客，支持文章列表、标签分类、评论功能',
  '设计一个在线商城主页，有轮播图、商品分类、促销活动',
  '做一个落地页，推广一款 AI 产品，包含功能特性、价格、CTA'
]

function autoResize() {
  const el = inputEl.value
  if (el) {
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 120) + 'px'
  }
}

function handleSend() {
  const text = input.value.trim()
  if (!text || props.loading) return
  emit('send', text)
  input.value = ''
  nextTick(autoResize)
}

// 自动滚动到底部
watch(() => props.messages.length, () => {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
})
</script>

<style scoped>
.chat-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #0a0a0b;
}

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-state h2 {
  font-size: 22px;
  font-weight: 700;
  color: #fafafa;
  margin-bottom: 8px;
}

.empty-state p {
  font-size: 14px;
  color: #71717a;
  margin-bottom: 24px;
}

.examples {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 400px;
  width: 100%;
}

.example-btn {
  padding: 10px 16px;
  background: #18181b;
  border: 1px solid #27272a;
  border-radius: 10px;
  color: #a1a1aa;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}

.example-btn:hover {
  background: #27272a;
  color: #e4e4e7;
  border-color: #6366f1;
}

/* 消息列表 */
.messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

/* 输入区域 */
.input-area {
  padding: 12px 16px;
  background: #18181b;
  border-top: 1px solid #27272a;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: #27272a;
  border: 1px solid #3f3f46;
  border-radius: 12px;
  padding: 8px 12px;
  transition: border-color 0.15s;
}

.input-wrapper:focus-within {
  border-color: #6366f1;
}

.chat-input {
  flex: 1;
  background: transparent;
  border: none;
  color: #e4e4e7;
  font-size: 14px;
  font-family: inherit;
  resize: none;
  outline: none;
  line-height: 1.5;
  max-height: 120px;
}

.chat-input::placeholder {
  color: #52525b;
}

.send-btn, .stop-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}

.send-btn {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}

.send-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #5558e6, #7c3aed);
}

.send-btn:disabled {
  background: #3f3f46;
  color: #52525b;
  cursor: not-allowed;
}

.stop-btn {
  background: #ef4444;
  color: white;
}

.input-hint {
  margin-top: 6px;
  font-size: 11px;
  color: #52525b;
  text-align: center;
}

/* 加载动画 */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 8px 16px;
}

.typing-indicator span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #6366f1;
  animation: bounce 1.4s infinite ease-in-out both;
}

.typing-indicator span:nth-child(1) { animation-delay: -0.32s; }
.typing-indicator span:nth-child(2) { animation-delay: -0.16s; }
.typing-indicator span:nth-child(3) { animation-delay: 0s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); opacity: 0.3; }
  40% { transform: scale(1); opacity: 1; }
}
</style>
