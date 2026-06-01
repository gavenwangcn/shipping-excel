<template>
  <div class="container">
    <div class="page-header">
      <router-link to="/" class="back-link">← 返回上传</router-link>
      <h1>任务详情</h1>
      <p class="job-id">任务 ID: {{ id }}</p>
    </div>

    <div v-if="loading" class="card loading-card">
      <div class="spinner-lg"></div>
      <p>加载任务信息...</p>
    </div>

    <template v-else-if="job">
      <!-- 进度卡片 -->
      <div class="card progress-card">
        <div class="progress-header">
          <StatusBadge :status="job.status" />
          <span class="progress-text">{{ job.message }}</span>
        </div>

        <div v-if="job.status === 'processing' || job.status === 'completed'" class="progress-section">
          <div class="progress-stats">
            <span>{{ job.progress }} / {{ job.total }} 个 HS CODE</span>
            <span>{{ progressPercent }}%</span>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
          </div>
          <p v-if="job.current_hs" class="current-hs">当前处理: {{ job.current_hs }}</p>
        </div>

        <div v-if="job.status === 'failed'" class="alert alert-error">
          {{ job.error_msg || '任务执行失败' }}
        </div>
      </div>

      <!-- 下载压缩包 -->
      <div v-if="job.status === 'completed' && job.zip_file_name" class="card download-card">
        <div class="download-inner">
          <div class="download-info">
            <span class="download-icon">📦</span>
            <div>
              <h2>下载报关 Excel 压缩包</h2>
              <p class="download-meta">
                {{ job.zip_file_name }} · 共 {{ job.file_count || files.length }} 个文件
                <span v-if="job.output_batch_name"> · 批次 {{ job.output_batch_name }}</span>
              </p>
            </div>
          </div>
          <a :href="zipDownloadUrl" class="btn btn-primary download-btn" download>
            下载 ZIP
          </a>
        </div>
      </div>

      <!-- 生成文件列表 -->
      <div v-if="files.length" class="card files-card">
        <div class="section-header">
          <h2>文件明细 ({{ files.length }})</h2>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>文件名</th>
                <th>HS CODE</th>
                <th>CI NO.</th>
                <th>数据行数</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="f in files" :key="f.id">
                <td>{{ f.file_name }}</td>
                <td><code>{{ f.hs_code }}</code></td>
                <td>{{ f.ci_no }}</td>
                <td>{{ f.row_count }}</td>
                <td>
                  <a :href="downloadUrl(f.id)" class="btn btn-primary btn-sm" download>
                    下载
                  </a>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 源数据预览 -->
      <div class="card data-card">
        <div class="section-header">
          <h2>源数据预览</h2>
          <span class="total-badge">共 {{ dataTotal }} 行</span>
        </div>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>行号</th>
                <th>HS CODE</th>
                <th>CI NO.</th>
                <th>PART No</th>
                <th>英文描述</th>
                <th>数量</th>
                <th>单价</th>
                <th>TYPE</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in dataRows" :key="row.id">
                <td>{{ row.row_num }}</td>
                <td><code>{{ row.hs_code }}</code></td>
                <td>{{ row.ci_no }}</td>
                <td>{{ row.part_no }}</td>
                <td class="desc-cell">{{ row.desc_en }}</td>
                <td>{{ row.qty }}</td>
                <td>{{ row.unit_price }}</td>
                <td>{{ row.type }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="dataTotal > pageSize" class="pagination">
          <button class="btn btn-outline btn-sm" :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
          <span>第 {{ page }} 页</span>
          <button class="btn btn-outline btn-sm" :disabled="page * pageSize >= dataTotal" @click="changePage(page + 1)">下一页</button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getJob, getJobFiles, getJobData, downloadFileUrl, downloadZipUrl } from '../api'
import StatusBadge from '../components/StatusBadge.vue'

const props = defineProps({ id: String })

const job = ref(null)
const files = ref([])
const dataRows = ref([])
const dataTotal = ref(0)
const loading = ref(true)
const page = ref(1)
const pageSize = 50
let pollTimer = null

const progressPercent = computed(() => {
  if (!job.value || !job.value.total) return 0
  return Math.round((job.value.progress / job.value.total) * 100)
})

const zipDownloadUrl = computed(() => downloadZipUrl(props.id))

function downloadUrl(fileId) {
  return downloadFileUrl(props.id, fileId)
}

async function fetchJob() {
  try {
    job.value = await getJob(props.id)
    if (job.value.status === 'completed' || job.value.status === 'failed') {
      stopPoll()
    }
  } catch {
    /* ignore */
  }
}

async function fetchFiles() {
  try {
    files.value = await getJobFiles(props.id)
  } catch {
    /* ignore */
  }
}

async function fetchData() {
  try {
    const res = await getJobData(props.id, page.value, pageSize)
    dataRows.value = res.data || []
    dataTotal.value = res.total || 0
  } catch {
    /* ignore */
  }
}

function changePage(p) {
  page.value = p
  fetchData()
}

async function refreshAll() {
  await fetchJob()
  await fetchFiles()
  await fetchData()
}

function startPoll() {
  pollTimer = setInterval(refreshAll, 1500)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  loading.value = true
  await refreshAll()
  loading.value = false
  if (job.value && (job.value.status === 'pending' || job.value.status === 'processing')) {
    startPoll()
  }
})

onUnmounted(stopPoll)
</script>

<style scoped>
.page-header {
  margin-bottom: 28px;
}

.back-link {
  font-size: 14px;
  color: var(--text-secondary);
  display: inline-block;
  margin-bottom: 12px;
}

.back-link:hover {
  color: var(--primary);
}

.page-header h1 {
  font-size: 28px;
  font-weight: 700;
}

.job-id {
  color: var(--text-secondary);
  font-size: 13px;
  font-family: monospace;
  margin-top: 4px;
}

.loading-card {
  text-align: center;
  padding: 60px;
  color: var(--text-secondary);
}

.spinner-lg {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}

.progress-card {
  margin-bottom: 24px;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.progress-text {
  font-size: 15px;
  color: var(--text-secondary);
}

.progress-section {
  margin-top: 8px;
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.current-hs {
  margin-top: 10px;
  font-size: 13px;
  color: var(--primary);
  font-weight: 500;
}

.alert-error {
  padding: 12px 16px;
  background: var(--danger-light);
  color: var(--danger);
  border-radius: 8px;
  font-size: 14px;
}

.files-card, .data-card {
  margin-bottom: 24px;
}

.download-card {
  margin-bottom: 24px;
  background: linear-gradient(135deg, #eff6ff 0%, #f0fdf4 100%);
  border-color: #bfdbfe;
}

.download-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
}

.download-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.download-icon {
  font-size: 36px;
}

.download-info h2 {
  font-size: 18px;
  margin-bottom: 4px;
}

.download-meta {
  font-size: 13px;
  color: var(--text-secondary);
}

.download-btn {
  min-width: 140px;
  padding: 12px 24px;
  font-size: 15px;
  white-space: nowrap;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-header h2 {
  font-size: 18px;
}

.total-badge {
  font-size: 13px;
  color: var(--text-secondary);
  background: var(--bg);
  padding: 4px 12px;
  border-radius: 999px;
}

code {
  font-family: monospace;
  font-size: 12px;
  background: var(--bg);
  padding: 2px 6px;
  border-radius: 4px;
}

.desc-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-sm {
  padding: 4px 14px;
  font-size: 12px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 16px;
  font-size: 14px;
  color: var(--text-secondary);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
