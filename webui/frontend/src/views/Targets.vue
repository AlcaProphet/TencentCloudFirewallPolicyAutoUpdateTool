<script setup lang="ts">
import { NDataTable, NButton } from 'naive-ui'
import { ref, onMounted } from 'vue'

const targets = ref<any[]>([])

onMounted(async () => {
  const res = await fetch('/api/targets')
  targets.value = await res.json() || []
})

const columns = [
  { title: '云产品', key: 'CloudType' },
  { title: '资源ID', key: 'ResourceID' },
  { title: '地域', key: 'Region' },
]
</script>

<template>
  <div>
    <h2>云资源管理</h2>
    <NDataTable :columns="columns" :data="targets" :bordered="true" />
  </div>
</template>
