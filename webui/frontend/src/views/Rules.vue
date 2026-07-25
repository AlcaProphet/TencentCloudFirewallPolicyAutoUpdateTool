<script setup lang="ts">
import { NDataTable } from 'naive-ui'
import { ref, onMounted } from 'vue'

const rules = ref<any[]>([])

onMounted(async () => {
  const res = await fetch('/api/rules')
  rules.value = await res.json() || []
})

const columns = [
  { title: '域名', key: 'Host' },
  { title: '协议', key: 'Protocol' },
  { title: '端口', key: 'Ports' },
  { title: '动作', key: 'Action' },
  { title: '备注', key: 'Comment' },
]
</script>

<template>
  <div>
    <h2>域名规则</h2>
    <NDataTable :columns="columns" :data="rules" :bordered="true" />
  </div>
</template>
