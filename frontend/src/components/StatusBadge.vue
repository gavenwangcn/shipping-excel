<template>
  <span class="badge" :class="'badge-' + status">{{ label }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ status: String })

const label = computed(() => {
  const map = {
    pending: '等待中',
    processing: '处理中',
    completed: '已完成',
    failed: '失败',
  }
  return map[props.status] || props.status
})
</script>
