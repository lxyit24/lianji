<template>
  <div class="app-container">
    <!-- 顶部导航 -->
    <header class="topbar">
      <div class="topbar-left">
        <span class="logo">🏗️ Lianji</span>
        <span class="badge">AI 建站</span>
      </div>
      <div class="topbar-right">
        <button
          class="btn btn-sm"
          :class="{ active: viewMode === 'chat' }"
          @click="viewMode = 'chat'"
        >对话</button>
        <button
          class="btn btn-sm"
          :class="{ active: viewMode === 'code' }"
          @click="viewMode = 'code'"
          :disabled="!currentCode"
        >代码</button>
        <button
          class="btn btn-sm btn-primary"
          @click="deploySite"
          :disabled="!currentCode"
        >🚀 发布</button>
      </div>
    </header>

    <!-- 主体区域 -->
    <div class="main-area">
      <Splitpanes class="default-theme">
        <!-- 左侧面板：对话或代码 -->
        <Pane :size="viewMode === 'code' ? 35 : 45" :min-size="30">
          <ChatPanel
            v-if="viewMode === 'chat'"
            :messages="chatStore.messages"
            :loading="chatStore.isGenerating"
            @send="handleSend"
            @stop="handleStop"
          />
          <CodePanel
            v-else
            :html="currentCode?.html_code || ''"
            :css="currentCode?.css_code || ''"
          />
        </Pane>

        <!-- 右侧面板：实时预览 -->
        <Pane :size="viewMode === 'code' ? 65 : 55" :min-size="35">
          <PreviewPanel
            :html="currentCode?.html_code || ''"
            :css="currentCode?.css_code || ''"
            :loading="chatStore.isGenerating"
          />
        </Pane>
      </Splitpanes>
    </div>

    <!-- 底部状态栏 -->
    <footer class="statusbar">
      <span>{{ statusText }}</span>
      <span>Lianji v0.2.0</span>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'
import { useChatStore } from './stores/chat.js'
import ChatPanel from './components/ChatPanel.vue'
import PreviewPanel from './components/PreviewPanel.vue'
import CodePanel from './components/CodePanel.vue'

const chatStore = useChatStore()
const viewMode = ref('chat')
const currentCode = ref(null)

const statusText = computed(() => {
  if (chatStore.isGenerating) return '🤖 AI 正在生成...'
  if (currentCode.value) return '✅ 网站已生成'
  return '💡 输入描述开始创建网站'
})

async function handleSend(prompt) {
  chatStore.addMessage({ role: 'user', content: prompt })
  chatStore.isGenerating = true
  chatStore.addMessage({ role: 'assistant', content: '', streaming: true })

  let fullContent = ''

  try {
    const response = await fetch('/api/v1/generate/stream', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prompt,
        site_type: '通用网站',
        history: chatStore.messages.slice(0, -2).map(m => ({
          role: m.role,
          content: m.content
        }))
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6))
            
            if (data.type === 'chunk') {
              fullContent += data.content
              chatStore.updateLastMessage(fullContent)
            } else if (data.type === 'code') {
              currentCode.value = {
                html_code: data.html_code || '',
                css_code: data.css_code || ''
              }
            } else if (data.type === 'done') {
              if (data.html_code || data.css_code) {
                currentCode.value = {
                  html_code: data.html_code || '',
                  css_code: data.css_code || ''
                }
              }
            } else if (data.type === 'error') {
              chatStore.updateLastMessage('❌ 生成失败: ' + data.message)
            }
          } catch (e) {
            // skip malformed lines
          }
        }
      }
    }

    chatStore.updateLastMessage(fullContent, false)
    chatStore.isGenerating = false
    viewMode.value = 'code'

  } catch (err) {
    chatStore.updateLastMessage('❌ 网络错误: ' + err.message, false)
    chatStore.isGenerating = false
  }
}

function handleStop() {
  chatStore.isGenerating = false
}

async function deploySite() {
  if (!currentCode.value) return
  
  const domain = prompt('请输入要部署的域名（如 mysite.lianji.win）:')
  if (!domain) return

  try {
    const res = await fetch('/api/v1/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        prompt: chatStore.messages[0]?.content || 'AI 生成的网站',
        site_type: '通用网站'
      })
    })
    const data = await res.json()
    alert('部署请求已发送！状态: ' + JSON.stringify(data))
  } catch (err) {
    alert('部署失败: ' + err.message)
  }
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  background: #0a0a0b;
  color: #e4e4e7;
  overflow: hidden;
  height: 100vh;
}

.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

/* 顶部导航 */
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  height: 48px;
  background: #18181b;
  border-bottom: 1px solid #27272a;
  flex-shrink: 0;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo {
  font-size: 16px;
  font-weight: 700;
  color: #fafafa;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  font-weight: 500;
}

.topbar-right {
  display: flex;
  gap: 8px;
}

/* 按钮 */
.btn {
  padding: 6px 14px;
  border: 1px solid #3f3f46;
  border-radius: 8px;
  background: transparent;
  color: #a1a1aa;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}

.btn:hover:not(:disabled) {
  background: #27272a;
  color: #e4e4e7;
}

.btn.active {
  background: #27272a;
  color: #fafafa;
  border-color: #52525b;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: transparent;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #5558e6, #7c3aed);
}

.btn-sm {
  padding: 5px 12px;
  font-size: 12px;
}

/* 主体 */
.main-area {
  flex: 1;
  overflow: hidden;
}

/* Splitpanes 覆盖 */
.splitpanes.default-theme .splitpanes__pane {
  background: #0a0a0b;
}

.splitpanes.default-theme .splitpanes__splitter {
  background: #27272a;
  border: none;
  width: 4px;
  transition: background 0.15s;
}

.splitpanes.default-theme .splitpanes__splitter:hover {
  background: #6366f1;
}

/* 状态栏 */
.statusbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  height: 28px;
  background: #18181b;
  border-top: 1px solid #27272a;
  font-size: 11px;
  color: #71717a;
  flex-shrink: 0;
}

/* 滚动条 */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #3f3f46;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #52525b;
}
</style>
