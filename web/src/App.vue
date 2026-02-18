<template>
  <div class="app-container">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metrics'

const metricsStore = useMetricsStore()

onMounted(() => {
  metricsStore.init()
})

onUnmounted(() => {
  metricsStore.disconnectWebSocket()
})
</script>

<style scoped>
.app-container {
  min-height: 100vh;
  padding: 0;
}
</style>
