<template>
  <div class="code-panel">
    <!-- 标签切换 -->
    <div class="code-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="code-tab"
        :class="{ active: activeTab === tab.id }"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
        <span class="tab-lang">{{ tab.lang }}</span>
      </button>
      <div class="tab-actions">
        <button class="copy-btn" @click="copyCode" :title="copied ? '已复制!' : '复制代码'">
          {{ copied ? '✅ 已复制' : '📋 复制' }}
        </button>
      </div>
    </div>

    <!-- 代码编辑器 -->
    <div class="code-editor" ref="editorEl"></div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { EditorView, basicSetup } from 'codemirror'
import { html } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { oneDark } from '@codemirror/theme-one-dark'

const props = defineProps({
  html: { type: String, default: '' },
  css: { type: String, default: '' }
})

const activeTab = ref('html')
const copied = ref(false)
const editorEl = ref(null)
let editorView = null

const tabs = [
  { id: 'html', label: 'HTML', lang: '.html' },
  { id: 'css', label: 'CSS', lang: '.css' }
]

const currentCode = computed(() => {
  return activeTab.value === 'html' ? props.html : props.css
})

function createEditor() {
  if (!editorEl.value) return
  
  if (editorView) {
    editorView.destroy()
  }

  const lang = activeTab.value === 'html' ? html() : css()
  
  editorView = new EditorView({
    doc: currentCode.value || '',
    extensions: [
      basicSetup,
      lang,
      oneDark,
      EditorView.editable.of(false),
      EditorView.theme({
        '&': { height: '100%' },
        '.cm-scroller': { overflow: 'auto' },
        '.cm-content': { fontFamily: "'Fira Code', 'Cascadia Code', monospace", fontSize: '13px' },
        '.cm-gutters': { background: '#18181b', borderRight: '1px solid #27272a', color: '#52525b' }
      })
    ],
    parent: editorEl.value
  })
}

watch(activeTab, () => {
  nextTick(createEditor)
})

watch([() => props.html, () => props.css], () => {
  nextTick(createEditor)
})

onMounted(() => {
  nextTick(createEditor)
})

onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy()
  }
})

async function copyCode() {
  try {
    await navigator.clipboard.writeText(currentCode.value)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch {
    // fallback
    const textarea = document.createElement('textarea')
    textarea.value = currentCode.value
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  }
}
</script>

<style scoped>
.code-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #0a0a0b;
}

.code-tabs {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 8px 12px;
  background: #18181b;
  border-bottom: 1px solid #27272a;
}

.code-tab {
  padding: 6px 14px;
  border: none;
  background: transparent;
  color: #71717a;
  font-size: 13px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
  font-family: inherit;
  display: flex;
  align-items: center;
  gap: 6px;
}

.code-tab:hover {
  background: #27272a;
  color: #a1a1aa;
}

.code-tab.active {
  background: #27272a;
  color: #fafafa;
}

.tab-lang {
  font-size: 10px;
  color: #52525b;
  text-transform: uppercase;
}

.tab-actions {
  margin-left: auto;
}

.copy-btn {
  padding: 4px 10px;
  border: 1px solid #3f3f46;
  border-radius: 6px;
  background: transparent;
  color: #a1a1aa;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}

.copy-btn:hover {
  background: #27272a;
  color: #e4e4e7;
}

.code-editor {
  flex: 1;
  overflow: hidden;
}
</style>
