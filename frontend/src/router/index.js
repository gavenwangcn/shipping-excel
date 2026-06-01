import { createRouter, createWebHistory } from 'vue-router'
import UploadPage from '../views/UploadPage.vue'
import JobPage from '../views/JobPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'upload', component: UploadPage },
    { path: '/jobs/:id', name: 'job', component: JobPage, props: true },
  ],
})

export default router
