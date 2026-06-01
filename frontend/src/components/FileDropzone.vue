<template>
  <div
    class="dropzone"
    :class="{ active: isDragging, filled: file }"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
    @drop.prevent="onDrop"
    @click="triggerInput"
  >
    <input
      ref="inputRef"
      type="file"
      :accept="accept"
      hidden
      @change="onChange"
    />
    <div v-if="file" class="file-info">
      <span class="file-icon">📄</span>
      <div class="file-details">
        <span class="file-name">{{ file.name }}</span>
        <span class="file-size">{{ formatSize(file.size) }}</span>
      </div>
      <button class="clear-btn" @click.stop="$emit('clear')">✕</button>
    </div>
    <div v-else class="placeholder">
      <span class="upload-icon">⬆</span>
      <span class="label">{{ label }}</span>
      <span class="hint">{{ hint }}</span>
      <span class="hint">点击或拖拽文件到此处</span>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  label: String,
  hint: String,
  accept: String,
  file: File,
})

const emit = defineEmits(['select', 'clear'])

const inputRef = ref(null)
const isDragging = ref(false)

function triggerInput() {
  inputRef.value?.click()
}

function onChange(e) {
  const f = e.target.files?.[0]
  if (f) emit('select', f)
  e.target.value = ''
}

function onDrop(e) {
  isDragging.value = false
  const f = e.dataTransfer.files?.[0]
  if (f) emit('select', f)
}

function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}
</script>

<style scoped>
.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  padding: 32px 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dropzone:hover, .dropzone.active {
  border-color: var(--primary);
  background: var(--primary-light);
}

.dropzone.filled {
  border-style: solid;
  border-color: var(--success);
  background: var(--success-light);
}

.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.upload-icon {
  font-size: 32px;
  margin-bottom: 4px;
}

.label {
  font-weight: 600;
  font-size: 15px;
}

.hint {
  font-size: 12px;
  color: var(--text-secondary);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.file-icon {
  font-size: 28px;
}

.file-details {
  flex: 1;
  text-align: left;
}

.file-name {
  display: block;
  font-weight: 600;
  font-size: 14px;
  word-break: break-all;
}

.file-size {
  font-size: 12px;
  color: var(--text-secondary);
}

.clear-btn {
  background: none;
  border: none;
  font-size: 16px;
  color: var(--text-secondary);
  padding: 4px 8px;
  border-radius: 4px;
}

.clear-btn:hover {
  background: rgba(0,0,0,0.08);
  color: var(--danger);
}
</style>
