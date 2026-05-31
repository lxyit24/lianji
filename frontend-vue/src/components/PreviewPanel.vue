<template>
  <div class="preview-panel">
    <!-- 工具栏 -->
    <div class="preview-toolbar">
      <div class="toolbar-left">
        <button class="tb-btn" @click="refreshPreview">🔄 刷新</button>
        <button class="tb-btn" @click="openInNewTab">🔗 新窗口</button>
      </div>
      <div class="toolbar-right">
        <span class="device-label">预览尺寸:</span>
        <button
          v-for="d in devices"
          :key="d.name"
          class="tb-btn"
          :class="{ active: device === d.name }"
          @click="device = d.name"
          :title="d.label"
        >{{ d.icon }}</button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading && !hasContent" class="preview-loading">
      <div class="loading-spinner"></div>
      <p>AI 正在生成网站...</p>
    </div>

    <!-- 空状态 -->
    <div v-if="!hasContent && !loading" class="preview-empty">
      <div class="empty-icon">🖼️</div>
      <p>网站预览将在这里显示</p>
    </div>

    <!-- 预览 iframe -->
    <div v-show="hasContent" class="preview-frame-wrapper" :class="device">
      <iframe
        ref="iframeEl"
        class="preview-iframe"
        :srcdoc="combinedHTML"
        sandbox="allow-scripts allow-same-origin"
        @load="onIframeLoad"
      ></iframe>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  html: { type: String, default: '' },
  css: { type: String, default: '' },
  loading: { type: Boolean, default: false }
})

const device = ref('desktop')
const iframeEl = ref(null)

const devices = [
  { name: 'mobile', label: '手机', icon: '📱' },
  { name: 'tablet', label: '平板', icon: '📋' },
  { name: 'desktop', label: '桌面', icon: '🖥️' }
]

const hasContent = computed(() => props.html.length > 0)

const combinedHTML = computed(() => {
  if (!props.html) return '<html><body style="background:#fff;color:#333;display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif">等待生成...</body></html>'
  
  let html = props.html
  // 在 </head> 前插入 CSS
  if (props.css) {
    if (html.includes('</head>')) {
      html = html.replace('</head>', `<style>${props.css}</style></head>`)
    } else if (html.includes('<body')) {
      html = html.replace('<body', `<head><style>${props.css}</style></head><body`)
    } else {
      html = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><style>${props.css}</style></head><body>${html}</body></html>`
    }
  }
  // 确保有 viewport meta
  if (!html.includes('viewport')) {
    html = html.replace('<head>', '<head><meta name="viewport" content="width=device-width, initial-scale=1.0">')
  }
  return html
})

function refreshPreview() {
  if (iframeEl.value) {
    iframeEl.value.srcdoc = combinedHTML.value
  }
}

function openInNewTab() {
  const win = window.open('', '_blank')
  if (win) {
    win.document.write(combinedHTML.value)
    win.document.close()
  }
}

function onIframeLoad() {
  // iframe loaded
}

// 当代码更新时自动刷新
watch([() => props.html, () => props.css], () => {
  if (hasContent.value) {
    refreshPreview()
  }
})
</script>

<style scoped>
.preview-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1a1a1e;
}

.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #18181b;
  border-bottom: 1px solid #27272a;
  flex-shrink: 0;
}

.toolbar-left, .toolbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tb-btn {
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

.tb-btn:hover {
  background: #27272a;
  color: #e4e4e7;
}

.tb-btn.active {
  background: #27272a;
  color: #fafafa;
  border-color: #6366f1;
}

.device-label {
  font-size: 11px;
  color: #52525b;
}

/* 加载 */
.preview-loading, .preview-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #52525b;
  gap: 16px;
}

.empty-icon {
  font-size: 48px;
  opacity: 0.5;
}

.loading-spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #27272a;
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 预览框架 */
.preview-frame-wrapper {
  flex: 1;
  overflow: hidden;
  display: flex;
  justify-content: center;
  background: #27272a;
}

.preview-frame-wrapper.desktop {
  padding: 16px;
}

.preview-frame-wrapper.tablet {
  padding: 16px;
}

.preview-frame-wrapper.tablet .preview-iframe {
  max-width: 768px;
}

.preview-frame-wrapper.mobile {
  padding: 16px;
}

.preview-frame-wrapper.mobile .preview-iframe {
  max-width: 375px;
}

.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.4);
}
</style>
