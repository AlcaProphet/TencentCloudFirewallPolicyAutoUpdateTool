<script setup lang="ts">
import { NCard, NStatistic, NGrid, NGi } from 'naive-ui'
import { ref, onMounted } from 'vue'

const health = ref('...')

onMounted(async () => {
  try {
    const res = await fetch('/api/health')
    const data = await res.json()
    health.value = data.status
  } catch {
    health.value = '离线'
  }
})
</script>

<template>
  <div>
    <h2>仪表盘</h2>
    <NGrid :cols="3" :x-gap="16">
      <NGi>
        <NCard>
          <NStatistic label="服务状态" :value="health" />
        </NCard>
      </NGi>
    </NGrid>
  </div>
</template>
