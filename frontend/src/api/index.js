import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 120000,
})

export async function uploadFiles(sourceFile, templateFile) {
  const form = new FormData()
  form.append('source', sourceFile)
  form.append('template', templateFile)
  const { data } = await api.post('/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export async function getJob(id) {
  const { data } = await api.get(`/jobs/${id}`)
  return data
}

export async function getJobFiles(id) {
  const { data } = await api.get(`/jobs/${id}/files`)
  return data
}

export async function getJobData(id, page = 1, pageSize = 50) {
  const { data } = await api.get(`/jobs/${id}/data`, {
    params: { page, page_size: pageSize },
  })
  return data
}

export async function listJobs() {
  const { data } = await api.get('/jobs')
  return data
}

export function downloadFileUrl(jobId, fileId) {
  return `/api/jobs/${jobId}/files/${fileId}/download`
}

export function downloadZipUrl(jobId) {
  return `/api/jobs/${jobId}/download`
}

export default api
