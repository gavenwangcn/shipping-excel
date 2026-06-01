<template>
  <div class="container">
    <div class="page-header">
      <h1>上传 Excel 文件</h1>
      <p>上传包含「数据」工作表的源文件，以及包含「INVOICE」工作表的模板文件，系统将按 HS CODE 自动生成报关 Excel。</p>
    </div>

    <div class="card upload-card">
      <div class="upload-grid">
        <FileDropzone
          label="源数据 Excel"
          hint="需包含「数据」工作表（含 HS CODE 等字段）"
          accept=".xlsx,.xlsm,.xls"
          :file="sourceFile"
          @select="sourceFile = $event"
          @clear="sourceFile = null"
        />
        <FileDropzone
          label="模板 Excel"
          hint="需包含「INVOICE」工作表作为报关单模板"
          accept=".xlsx,.xlsm,.xls"
          :file="templateFile"
          @select="templateFile = $event"
          @clear="templateFile = null"
        />
      </div>

      <div v-if="error" class="alert alert-error">{{ error }}</div>

      <div class="actions">
        <button
          class="btn btn-primary submit-btn"
          :disabled="!canSubmit || uploading"
          @click="handleSubmit"
        >
          <span v-if="uploading" class="spinner"></span>
          {{ uploading ? '上传中...' : '开始生成报关 Excel' }}
        </button>
      </div>
    </div>

    <div v-if="recentJobs.length" class="card recent-card">
      <h2>最近任务</h2>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>任务 ID</th>
              <th>状态</th>
              <th>进度</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="job in recentJobs" :key="job.id">
              <td class="mono">{{ job.id.slice(0, 8) }}...</td>
              <td><StatusBadge :status="job.status" /></td>
              <td>{{ job.progress }} / {{ job.total || '—' }}</td>
              <td>{{ formatTime(job.created_at) }}</td>
              <td>
                <router-link :to="`/jobs/${job.id}`" class="btn btn-outline btn-sm">查看</router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { uploadFiles, listJobs } from '../api'
import FileDropzone from '../components/FileDropzone.vue'
import StatusBadge from '../components/StatusBadge.vue'

const router = useRouter()
const sourceFile = ref(null)
const templateFile = ref(null)
const uploading = ref(false)
const error = ref('')
const recentJobs = ref([])

const canSubmit = computed(() => sourceFile.value && templateFile.value)

async function handleSubmit() {
  if (!canSubmit.value) return
  uploading.value = true
  error.value = ''
  try {
    const result = await uploadFiles(sourceFile.value, templateFile.value)
    router.push(`/jobs/${result.id}`)
  } catch (e) {
    error.value = e.response?.data?.error || e.message || '上传失败'
  } finally {
    uploading.value = false
  }
}

function formatTime(t) {
  if (!t) return '—'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(async () => {
  try {
    recentJobs.value = await listJobs()
  } catch {
    /* ignore */
  }
})
</script>

<style scoped>
.page-header {
  margin-bottom: 28px;
}

.page-header h1 {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 8px;
}

.page-header p {
  color: var(--text-secondary);
  font-size: 15px;
}

.upload-card {
  margin-bottom: 32px;
}

.upload-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 28px;
}

@media (max-width: 640px) {
  .upload-grid {
    grid-template-columns: 1fr;
  }
}

.actions {
  display: flex;
  justify-content: center;
}

.submit-btn {
  min-width: 220px;
  padding: 12px 28px;
  font-size: 15px;
}

.alert {
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 14px;
}

.alert-error {
  background: var(--danger-light);
  color: var(--danger);
  border: 1px solid #fca5a5;
}

.recent-card h2 {
  font-size: 18px;
  margin-bottom: 16px;
}

.mono {
  font-family: monospace;
  font-size: 12px;
}

.btn-sm {
  padding: 4px 12px;
  font-size: 12px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
