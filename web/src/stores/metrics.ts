import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Metrics, HealthStatus } from '@/types'
import { metricsApi, createWebSocketConnection } from '@/api/metrics'

export const useMetricsStore = defineStore('metrics', () => {
  // State
  const metrics = ref<Metrics | null>(null)
  const health = ref<HealthStatus | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const wsConnected = ref(false)
  const lastUpdate = ref<string>('')

  // WebSocket 实例
  let ws: WebSocket | null = null

  // Getters
  const cpuUsage = computed(() => metrics.value?.cpu.usage_percent ?? 0)
  const memoryUsage = computed(() => metrics.value?.memory.used_percent ?? 0)
  const diskList = computed(() => metrics.value?.disk ?? [])
  const netList = computed(() => metrics.value?.net ?? [])
  const hostName = computed(() => metrics.value?.host ?? 'unknown')

  // Actions
  async function fetchMetrics() {
    loading.value = true
    error.value = null
    try {
      const response = await metricsApi.getMetrics()
      metrics.value = response.data
      lastUpdate.value = new Date().toISOString()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch metrics'
      console.error('Failed to fetch metrics:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchHealth() {
    try {
      const response = await metricsApi.getHealth()
      health.value = response.data
    } catch (e) {
      console.error('Failed to fetch health:', e)
    }
  }

  function connectWebSocket() {
    if (ws) {
      ws.close()
    }

    ws = createWebSocketConnection(
      (data) => {
        metrics.value = data
        lastUpdate.value = new Date().toISOString()
        wsConnected.value = true
      },
      () => {
        wsConnected.value = false
      },
      () => {
        wsConnected.value = false
        // 自动重连
        setTimeout(connectWebSocket, 3000)
      }
    )
  }

  function disconnectWebSocket() {
    if (ws) {
      ws.close()
      ws = null
      wsConnected.value = false
    }
  }

  // 初始化
  async function init() {
    await fetchMetrics()
    await fetchHealth()
    connectWebSocket()
  }

  return {
    // State
    metrics,
    health,
    loading,
    error,
    wsConnected,
    lastUpdate,
    // Getters
    cpuUsage,
    memoryUsage,
    diskList,
    netList,
    hostName,
    // Actions
    fetchMetrics,
    fetchHealth,
    connectWebSocket,
    disconnectWebSocket,
    init
  }
})
